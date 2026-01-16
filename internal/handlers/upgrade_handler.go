package handlers

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fazt-sh/fazt/internal/api"
	"github.com/fazt-sh/fazt/internal/config"
)

// UpgradeResponse is the response from the upgrade endpoint
type UpgradeResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	CurrentVersion string `json:"current_version"`
	NewVersion     string `json:"new_version,omitempty"`
	Action         string `json:"action,omitempty"` // "upgraded", "already_latest", "check_only"
}

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// UpgradeHandler handles POST /api/upgrade
// Query params:
//   - check=true: Only check for updates, don't upgrade
func UpgradeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.ErrorResponse(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed", "")
		return
	}

	checkOnly := r.URL.Query().Get("check") == "true"
	currentVersion := config.Version

	// Get latest release from GitHub
	latestRelease, err := getLatestRelease()
	if err != nil {
		api.ErrorResponse(w, http.StatusBadGateway, "GITHUB_ERROR", "Failed to check latest release: "+err.Error(), "")
		return
	}

	latestVersion := strings.TrimPrefix(latestRelease.TagName, "v")

	// Check if already on latest
	if latestVersion == currentVersion {
		api.Success(w, http.StatusOK, UpgradeResponse{
			Success:        true,
			Message:        "Already running the latest version",
			CurrentVersion: currentVersion,
			NewVersion:     latestVersion,
			Action:         "already_latest",
		})
		return
	}

	// If check only, return without upgrading
	if checkOnly {
		api.Success(w, http.StatusOK, UpgradeResponse{
			Success:        true,
			Message:        fmt.Sprintf("Update available: %s -> %s", currentVersion, latestVersion),
			CurrentVersion: currentVersion,
			NewVersion:     latestVersion,
			Action:         "check_only",
		})
		return
	}

	// Find the right asset for this platform
	assetName := fmt.Sprintf("fazt-%s-%s-%s.tar.gz", latestRelease.TagName, runtime.GOOS, runtime.GOARCH)
	var downloadURL string
	for _, asset := range latestRelease.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		api.ErrorResponse(w, http.StatusNotFound, "ASSET_NOT_FOUND",
			fmt.Sprintf("No release asset found for %s/%s", runtime.GOOS, runtime.GOARCH), "")
		return
	}

	// Download and extract the new binary
	newBinaryPath, err := downloadAndExtract(downloadURL)
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "DOWNLOAD_ERROR", "Failed to download update: "+err.Error(), "")
		return
	}
	defer os.Remove(newBinaryPath)

	// Get current binary path
	currentBinary, err := os.Executable()
	if err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "PATH_ERROR", "Failed to get current binary path: "+err.Error(), "")
		return
	}
	currentBinary, _ = filepath.EvalSymlinks(currentBinary)

	// Stage new binary in same directory as target (required for atomic rename)
	// This avoids "text file busy" error when replacing a running binary
	targetDir := filepath.Dir(currentBinary)
	stagingPath := filepath.Join(targetDir, ".fazt.new")

	if err := copyFile(newBinaryPath, stagingPath); err != nil {
		api.ErrorResponse(w, http.StatusInternalServerError, "STAGING_ERROR", "Failed to stage new binary: "+err.Error(), "")
		return
	}

	// Set permissions before rename
	if err := os.Chmod(stagingPath, 0755); err != nil {
		os.Remove(stagingPath)
		api.ErrorResponse(w, http.StatusInternalServerError, "CHMOD_ERROR", "Failed to set permissions: "+err.Error(), "")
		return
	}

	// Atomic rename - this works on running binaries because it replaces
	// the directory entry, not the file. The old binary keeps running from
	// its inode until the process exits.
	if err := os.Rename(stagingPath, currentBinary); err != nil {
		os.Remove(stagingPath)
		api.ErrorResponse(w, http.StatusInternalServerError, "REPLACE_ERROR", "Failed to replace binary: "+err.Error(), "")
		return
	}

	// Set capabilities for port binding (Linux only)
	if runtime.GOOS == "linux" {
		exec.Command("setcap", "CAP_NET_BIND_SERVICE=+eip", currentBinary).Run()
	}

	// Send success response before restarting
	api.Success(w, http.StatusOK, UpgradeResponse{
		Success:        true,
		Message:        fmt.Sprintf("Upgraded from %s to %s. Service will restart.", currentVersion, latestVersion),
		CurrentVersion: currentVersion,
		NewVersion:     latestVersion,
		Action:         "upgraded",
	})

	// Flush the response
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Restart the service after a short delay to allow response to be sent.
	// Uses an external shell process with sleep so the delay happens outside Go,
	// ensuring the restart happens even if the Go process is interrupted.
	// The sudoers.d/fazt file grants NOPASSWD access for systemctl restart.
	exec.Command("sh", "-c", "sleep 1 && sudo systemctl restart fazt").Start()
}

func getLatestRelease() (*GitHubRelease, error) {
	resp, err := http.Get("https://api.github.com/repos/fazt-sh/fazt/releases/latest")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

func downloadAndExtract(url string) (string, error) {
	// Download
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned %d", resp.StatusCode)
	}

	// Create temp file for archive
	tmpArchive, err := os.CreateTemp("", "fazt-*.tar.gz")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpArchive.Name())

	if _, err := io.Copy(tmpArchive, resp.Body); err != nil {
		tmpArchive.Close()
		return "", err
	}
	tmpArchive.Close()

	// Extract
	return extractTarGz(tmpArchive.Name())
}

func extractTarGz(archivePath string) (string, error) {
	f, err := os.Open(archivePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	// Create temp file for binary
	tmpBinary, err := os.CreateTemp("", "fazt-binary-*")
	if err != nil {
		return "", err
	}
	tmpBinary.Close()

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Look for the fazt binary
		if header.Typeflag == tar.TypeReg && (header.Name == "fazt" || strings.HasSuffix(header.Name, "/fazt")) {
			out, err := os.Create(tmpBinary.Name())
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return "", err
			}
			out.Close()
			return tmpBinary.Name(), nil
		}
	}

	return "", fmt.Errorf("fazt binary not found in archive")
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	return err
}
