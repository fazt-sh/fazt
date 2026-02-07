package media

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/fazt-sh/fazt/internal/debug"
	"github.com/fazt-sh/fazt/internal/system"
)

// VariantPrefix is the blob path prefix for transcoded H.264 variants.
const VariantPrefix = "_v/h264/"

// VariantPath returns the blob path for an H.264 transcoded variant.
func VariantPath(originalPath string) string {
	return VariantPrefix + originalPath
}

var (
	transcodeSem  chan struct{}
	transcodeOnce sync.Once
)

func getTranscodeSem() chan struct{} {
	transcodeOnce.Do(func() {
		n := system.GetLimits().Video.Concurrency
		if n <= 0 {
			n = 1
		}
		transcodeSem = make(chan struct{}, n)
	})
	return transcodeSem
}

// TranscodeResult describes the outcome of a transcode request.
type TranscodeResult struct {
	Status string `json:"status"` // "queued", "compatible", "no_ffmpeg", "too_large", "not_video"
}

// StoreFunc stores a blob. Used as callback from background transcoding.
type StoreFunc func(ctx context.Context, path string, data []byte, mime string) error

// QueueTranscode checks if video data needs transcoding and queues it if so.
// Returns immediately — transcoding happens in the background.
// storeResult is called from the background goroutine to persist the variant.
func QueueTranscode(appID, blobPath string, data []byte, mime string, storeResult StoreFunc) TranscodeResult {
	limits := system.GetLimits().Video

	if !limits.FFmpegAvailable {
		return TranscodeResult{Status: "no_ffmpeg"}
	}

	if !IsVideoContentType(mime) {
		return TranscodeResult{Status: "not_video"}
	}

	if len(data) > limits.MaxInputMB*1024*1024 {
		return TranscodeResult{Status: "too_large"}
	}

	info, err := ProbeVideo(data)
	if err != nil {
		return TranscodeResult{Status: "not_video"}
	}

	if info.Compatible {
		return TranscodeResult{Status: "compatible"}
	}

	// Queue background transcode
	go func() {
		sem := getTranscodeSem()
		sem <- struct{}{} // blocking acquire
		defer func() { <-sem }()

		bgCtx := context.Background()
		result, err := TranscodeToH264(bgCtx, data, limits.OutputMaxHeight)
		if err != nil {
			debug.Log("media", "transcode failed for %s/%s: %v", appID, blobPath, err)
			return
		}

		variantPath := VariantPath(blobPath)
		if err := storeResult(bgCtx, variantPath, result, "video/mp4"); err != nil {
			debug.Log("media", "failed to store variant for %s/%s: %v", appID, blobPath, err)
			return
		}

		debug.Log("media", "transcoded %s/%s → H.264 (%d → %d bytes)", appID, blobPath, len(data), len(result))
	}()

	return TranscodeResult{Status: "queued"}
}

// TranscodeToH264 runs ffmpeg to convert video data to H.264+AAC MP4.
// Blocks until complete. Runs at nice +19 with -threads 1.
func TranscodeToH264(ctx context.Context, input []byte, maxHeight int) ([]byte, error) {
	tmpDir, err := os.MkdirTemp("", "fazt-transcode-*")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	inputPath := filepath.Join(tmpDir, "input")
	outputPath := filepath.Join(tmpDir, "output.mp4")

	if err := os.WriteFile(inputPath, input, 0600); err != nil {
		return nil, fmt.Errorf("write input: %w", err)
	}

	args := []string{
		"-n", "19", "ffmpeg",
		"-i", inputPath,
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", "23",
		"-threads", "1",
		"-c:a", "aac",
		"-movflags", "+faststart",
	}

	if maxHeight > 0 {
		args = append(args, "-vf", fmt.Sprintf("scale=-2:'min(%d,ih)'", maxHeight))
	}

	args = append(args, "-y", outputPath)

	cmd := exec.CommandContext(ctx, "nice", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg: %v: %s", err, string(output[:min(500, len(output))]))
	}

	return os.ReadFile(outputPath)
}

// IsVideoContentType returns true if the content type is a video type.
func IsVideoContentType(contentType string) bool {
	return len(contentType) > 6 && contentType[:6] == "video/"
}
