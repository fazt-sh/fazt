package media

import (
	"encoding/binary"
	"testing"
)

// buildMP4 constructs a minimal MP4 file with the given codecs and dimensions.
func buildMP4(brand string, videoCodec string, audioCodec string, width, height uint16, timescale uint32, duration uint32) []byte {
	var data []byte

	// ftyp box
	ftyp := makeBox("ftyp", []byte(brand+"\x00\x00\x00\x00"))
	data = append(data, ftyp...)

	// Build moov contents
	var moovContent []byte

	// mvhd box (version 0)
	mvhdPayload := make([]byte, 100) // version(1)+flags(3)+created(4)+modified(4)+timescale(4)+duration(4)+...
	binary.BigEndian.PutUint32(mvhdPayload[8:12], timescale)
	binary.BigEndian.PutUint32(mvhdPayload[12:16], duration)
	mvhd := makeFullBox("mvhd", 0, mvhdPayload)
	moovContent = append(moovContent, mvhd...)

	// Video track
	if videoCodec != "" {
		videoTrack := buildTrack("vide", videoCodec, width, height)
		moovContent = append(moovContent, videoTrack...)
	}

	// Audio track
	if audioCodec != "" {
		audioTrack := buildTrack("soun", audioCodec, 0, 0)
		moovContent = append(moovContent, audioTrack...)
	}

	moov := makeBox("moov", moovContent)
	data = append(data, moov...)

	return data
}

// buildTrack constructs a trak → mdia → hdlr + minf → stbl → stsd chain.
func buildTrack(handlerType, codec string, width, height uint16) []byte {
	// stsd entry (sample description)
	var entryPayload []byte
	if handlerType == "vide" {
		// Visual sample entry: reserved(6) + data_ref_idx(2) + pre_defined(2) +
		// reserved(2) + pre_defined(12) + width(2) + height(2) + ...
		entryPayload = make([]byte, 70) // enough for dimensions
		binary.BigEndian.PutUint16(entryPayload[24:26], width)
		binary.BigEndian.PutUint16(entryPayload[26:28], height)
	} else {
		entryPayload = make([]byte, 20) // audio sample entry basics
	}
	entry := makeBox(codec, entryPayload)

	// stsd content after version+flags: entry_count(4) + entries
	stsdPayload := make([]byte, 4)
	binary.BigEndian.PutUint32(stsdPayload[0:4], 1) // 1 entry
	stsdPayload = append(stsdPayload, entry...)
	stsd := makeFullBox("stsd", 0, stsdPayload)

	stbl := makeBox("stbl", stsd)

	// hdlr content after version+flags: pre_defined(4) + handler_type(4) + reserved(12)
	hdlrPayload := make([]byte, 20)
	copy(hdlrPayload[4:8], handlerType)
	hdlr := makeFullBox("hdlr", 0, hdlrPayload)

	minf := makeBox("minf", stbl)

	mdiaContent := append(hdlr, minf...)
	mdia := makeBox("mdia", mdiaContent)

	trak := makeBox("trak", mdia)
	return trak
}

// makeBox creates an ISO BMFF box with type and payload.
func makeBox(boxType string, payload []byte) []byte {
	size := 8 + len(payload)
	buf := make([]byte, size)
	binary.BigEndian.PutUint32(buf[0:4], uint32(size))
	copy(buf[4:8], boxType)
	copy(buf[8:], payload)
	return buf
}

// makeFullBox creates a full box (has version+flags before content).
func makeFullBox(boxType string, version byte, content []byte) []byte {
	// Full box content = version(1) + flags(3) + rest
	payload := make([]byte, 4+len(content))
	payload[0] = version
	copy(payload[4:], content)
	return makeBox(boxType, payload)
}

func TestProbeH264(t *testing.T) {
	data := buildMP4("isom", "avc1", "mp4a", 1920, 1080, 1000, 30000)

	info, err := ProbeVideo(data)
	if err != nil {
		t.Fatalf("ProbeVideo error: %v", err)
	}

	if info.Container != "mp4" {
		t.Errorf("container = %q, want mp4", info.Container)
	}
	if info.VideoCodec != "h264" {
		t.Errorf("videoCodec = %q, want h264", info.VideoCodec)
	}
	if info.AudioCodec != "aac" {
		t.Errorf("audioCodec = %q, want aac", info.AudioCodec)
	}
	if info.Width != 1920 {
		t.Errorf("width = %d, want 1920", info.Width)
	}
	if info.Height != 1080 {
		t.Errorf("height = %d, want 1080", info.Height)
	}
	if info.Duration != 30.0 {
		t.Errorf("duration = %f, want 30.0", info.Duration)
	}
	if !info.Compatible {
		t.Error("compatible = false, want true")
	}
}

func TestProbeHEVC(t *testing.T) {
	data := buildMP4("isom", "hvc1", "mp4a", 3840, 2160, 600, 7200)

	info, err := ProbeVideo(data)
	if err != nil {
		t.Fatalf("ProbeVideo error: %v", err)
	}

	if info.VideoCodec != "hevc" {
		t.Errorf("videoCodec = %q, want hevc", info.VideoCodec)
	}
	if info.Width != 3840 {
		t.Errorf("width = %d, want 3840", info.Width)
	}
	if info.Height != 2160 {
		t.Errorf("height = %d, want 2160", info.Height)
	}
	if info.Duration != 12.0 {
		t.Errorf("duration = %f, want 12.0", info.Duration)
	}
	if info.Compatible {
		t.Error("compatible = true, want false (HEVC)")
	}
}

func TestProbeMOV(t *testing.T) {
	data := buildMP4("qt  ", "avc1", "mp4a", 1280, 720, 30000, 900000)

	info, err := ProbeVideo(data)
	if err != nil {
		t.Fatalf("ProbeVideo error: %v", err)
	}

	if info.Container != "mov" {
		t.Errorf("container = %q, want mov", info.Container)
	}
	if !info.Compatible {
		t.Error("compatible = false, want true (H.264+AAC)")
	}
}

func TestProbeVideoOnly(t *testing.T) {
	// H.264 with no audio — should still be compatible
	data := buildMP4("isom", "avc1", "", 640, 480, 1000, 5000)

	info, err := ProbeVideo(data)
	if err != nil {
		t.Fatalf("ProbeVideo error: %v", err)
	}

	if info.VideoCodec != "h264" {
		t.Errorf("videoCodec = %q, want h264", info.VideoCodec)
	}
	if info.AudioCodec != "" {
		t.Errorf("audioCodec = %q, want empty", info.AudioCodec)
	}
	if !info.Compatible {
		t.Error("compatible = false, want true (H.264, no audio)")
	}
}

func TestProbeVP9(t *testing.T) {
	data := buildMP4("isom", "vp09", "", 1920, 1080, 1000, 10000)

	info, err := ProbeVideo(data)
	if err != nil {
		t.Fatalf("ProbeVideo error: %v", err)
	}

	if info.VideoCodec != "vp9" {
		t.Errorf("videoCodec = %q, want vp9", info.VideoCodec)
	}
	if info.Compatible {
		t.Error("compatible = true, want false")
	}
}

func TestProbeAV1(t *testing.T) {
	data := buildMP4("isom", "av01", "Opus", 1920, 1080, 1000, 60000)

	info, err := ProbeVideo(data)
	if err != nil {
		t.Fatalf("ProbeVideo error: %v", err)
	}

	if info.VideoCodec != "av1" {
		t.Errorf("videoCodec = %q, want av1", info.VideoCodec)
	}
	if info.AudioCodec != "opus" {
		t.Errorf("audioCodec = %q, want opus", info.AudioCodec)
	}
	if info.Compatible {
		t.Error("compatible = true, want false")
	}
}

func TestProbeWebM(t *testing.T) {
	// Minimal WebM-like data: EBML header + codec ID string
	data := []byte{0x1A, 0x45, 0xDF, 0xA3} // EBML magic
	data = append(data, make([]byte, 100)...)
	copy(data[20:], "webm")
	copy(data[50:], "V_VP9")
	copy(data[70:], "A_OPUS")

	info, err := ProbeVideo(data)
	if err != nil {
		t.Fatalf("ProbeVideo error: %v", err)
	}

	if info.Container != "webm" {
		t.Errorf("container = %q, want webm", info.Container)
	}
	if info.VideoCodec != "vp9" {
		t.Errorf("videoCodec = %q, want vp9", info.VideoCodec)
	}
	if info.AudioCodec != "opus" {
		t.Errorf("audioCodec = %q, want opus", info.AudioCodec)
	}
	if info.Compatible {
		t.Error("compatible = true, want false")
	}
}

func TestProbeNotVideo(t *testing.T) {
	// Random data
	data := []byte("this is not a video file at all")

	_, err := ProbeVideo(data)
	if err != ErrNotVideo {
		t.Errorf("error = %v, want ErrNotVideo", err)
	}
}

func TestProbeTooShort(t *testing.T) {
	_, err := ProbeVideo([]byte{0x00, 0x01})
	if err != ErrNotVideo {
		t.Errorf("error = %v, want ErrNotVideo", err)
	}
}

func TestProbeDuration(t *testing.T) {
	tests := []struct {
		name      string
		timescale uint32
		duration  uint32
		want      float64
	}{
		{"30fps_10sec", 30000, 300000, 10.0},
		{"24fps_60sec", 24000, 1440000, 60.0},
		{"1000_5.5sec", 1000, 5500, 5.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := buildMP4("isom", "avc1", "", 640, 480, tt.timescale, tt.duration)
			info, err := ProbeVideo(data)
			if err != nil {
				t.Fatalf("ProbeVideo error: %v", err)
			}
			if info.Duration != tt.want {
				t.Errorf("duration = %f, want %f", info.Duration, tt.want)
			}
		})
	}
}
