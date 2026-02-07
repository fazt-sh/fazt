// Package media — video probe support.
// Parses MP4/MOV box structure to detect codec, dimensions, and duration.
// Pure Go, zero dependencies, no ffmpeg or CGo required.
package media

import (
	"encoding/binary"
	"errors"
	"math"
)

// VideoInfo holds the result of probing a video file.
type VideoInfo struct {
	Container  string  `json:"container"`
	VideoCodec string  `json:"videoCodec"`
	AudioCodec string  `json:"audioCodec"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Duration   float64 `json:"duration"`
	Compatible bool    `json:"compatible"`
}

// ErrNotVideo is returned when the data is not a recognized video format.
var ErrNotVideo = errors.New("not a recognized video format")

// ProbeVideo parses video data and returns codec/dimension/duration info.
// Supports MP4, MOV, and detects WebM containers.
func ProbeVideo(data []byte) (*VideoInfo, error) {
	if len(data) < 8 {
		return nil, ErrNotVideo
	}

	// WebM/Matroska: EBML header magic bytes
	if data[0] == 0x1A && data[1] == 0x45 && data[2] == 0xDF && data[3] == 0xA3 {
		return probeWebM(data)
	}

	// MP4/MOV: look for ftyp or moov box at start
	boxType := string(data[4:8])
	if boxType == "ftyp" || boxType == "moov" || boxType == "free" || boxType == "mdat" || boxType == "wide" {
		return probeMP4(data)
	}

	return nil, ErrNotVideo
}

// probeMP4 walks the ISO BMFF box structure to extract video metadata.
func probeMP4(data []byte) (*VideoInfo, error) {
	info := &VideoInfo{Container: "mp4"}

	walkBoxes(data, func(boxType string, boxData []byte, fullData []byte) {
		switch boxType {
		case "ftyp":
			info.Container = parseFtyp(boxData)
		case "moov":
			parseMoov(fullData, info)
		}
	})

	if info.VideoCodec == "" && info.AudioCodec == "" {
		return nil, ErrNotVideo
	}

	info.Compatible = info.VideoCodec == "h264" &&
		(info.AudioCodec == "aac" || info.AudioCodec == "")

	return info, nil
}

// probeWebM detects WebM container and attempts basic codec detection from
// the EBML DocType element. Full EBML parsing is not implemented — we detect
// the container and set compatible=false since WebM codecs (VP8/VP9/AV1)
// aren't universally compatible.
func probeWebM(data []byte) (*VideoInfo, error) {
	info := &VideoInfo{
		Container:  "webm",
		Compatible: false,
	}

	// Scan for DocType string "webm" or "matroska" in the EBML header.
	// Also look for codec IDs as strings in the data.
	s := string(data[:min(4096, len(data))])

	if containsBytes(s, "matroska") {
		info.Container = "matroska"
	}

	// Codec IDs in Matroska are stored as strings like "V_VP9", "V_VP8", "V_AV1"
	if containsBytes(s, "V_VP9") {
		info.VideoCodec = "vp9"
	} else if containsBytes(s, "V_VP8") {
		info.VideoCodec = "vp8"
	} else if containsBytes(s, "V_AV1") {
		info.VideoCodec = "av1"
	}

	if containsBytes(s, "A_OPUS") {
		info.AudioCodec = "opus"
	} else if containsBytes(s, "A_VORBIS") {
		info.AudioCodec = "vorbis"
	}

	if info.VideoCodec == "" {
		// We know it's a WebM/Matroska container but can't identify the codec
		// from the header scan. Still return the container info.
		info.VideoCodec = "unknown"
	}

	return info, nil
}

// walkBoxes iterates over ISO BMFF boxes at the current level.
// callback receives: boxType, content after header, and full box data (header+content).
func walkBoxes(data []byte, callback func(boxType string, content []byte, fullBox []byte)) {
	offset := 0
	for offset+8 <= len(data) {
		size := int(binary.BigEndian.Uint32(data[offset : offset+4]))
		boxType := string(data[offset+4 : offset+8])

		headerSize := 8

		if size == 1 {
			// 64-bit extended size
			if offset+16 > len(data) {
				break
			}
			size = int(binary.BigEndian.Uint64(data[offset+8 : offset+16]))
			headerSize = 16
		} else if size == 0 {
			// Box extends to end of data
			size = len(data) - offset
		}

		if size < headerSize || offset+size > len(data) {
			break
		}

		fullBox := data[offset : offset+size]
		content := data[offset+headerSize : offset+size]
		callback(boxType, content, fullBox)

		offset += size
	}
}

// parseFtyp extracts the container type from ftyp box content.
func parseFtyp(data []byte) string {
	if len(data) < 4 {
		return "mp4"
	}
	brand := string(data[0:4])
	switch brand {
	case "qt  ":
		return "mov"
	case "M4V ", "M4VH", "M4VP":
		return "m4v"
	case "M4A ":
		return "m4a"
	default:
		return "mp4"
	}
}

// parseMoov walks into moov → trak → mdia to find codecs, dimensions, duration.
func parseMoov(moovFull []byte, info *VideoInfo) {
	if len(moovFull) < 8 {
		return
	}
	// moovFull includes the box header, content starts after it
	content := moovFull[8:]

	walkBoxes(content, func(boxType string, boxData []byte, fullBox []byte) {
		switch boxType {
		case "mvhd":
			parseMvhd(boxData, info)
		case "trak":
			parseTrak(fullBox, info)
		}
	})
}

// parseMvhd extracts duration and timescale from the movie header box.
func parseMvhd(data []byte, info *VideoInfo) {
	if len(data) < 4 {
		return
	}

	version := data[0]
	// Skip version (1) + flags (3)
	off := 4

	var timescale uint32
	var duration uint64

	if version == 0 {
		// v0: created (4) + modified (4) + timescale (4) + duration (4)
		if off+16 > len(data) {
			return
		}
		timescale = binary.BigEndian.Uint32(data[off+8 : off+12])
		duration = uint64(binary.BigEndian.Uint32(data[off+12 : off+16]))
	} else {
		// v1: created (8) + modified (8) + timescale (4) + duration (8)
		if off+28 > len(data) {
			return
		}
		timescale = binary.BigEndian.Uint32(data[off+16 : off+20])
		duration = binary.BigEndian.Uint64(data[off+20 : off+28])
	}

	if timescale > 0 && duration > 0 {
		sec := float64(duration) / float64(timescale)
		info.Duration = math.Round(sec*10) / 10 // round to 1 decimal
	}
}

// parseTrak walks into trak → mdia → minf → stbl → stsd.
func parseTrak(trakFull []byte, info *VideoInfo) {
	if len(trakFull) < 8 {
		return
	}
	content := trakFull[8:]

	walkBoxes(content, func(boxType string, _ []byte, fullBox []byte) {
		if boxType == "mdia" {
			parseMdia(fullBox, info)
		}
	})
}

// parseMdia walks into mdia → hdlr (to determine track type) + minf.
func parseMdia(mdiaFull []byte, info *VideoInfo) {
	if len(mdiaFull) < 8 {
		return
	}
	content := mdiaFull[8:]

	var trackType string // "vide" or "soun"

	walkBoxes(content, func(boxType string, boxData []byte, fullBox []byte) {
		switch boxType {
		case "hdlr":
			trackType = parseHdlr(boxData)
		case "minf":
			if trackType != "" {
				parseMinf(fullBox, trackType, info)
			}
		}
	})
}

// parseHdlr extracts the handler type (vide, soun, etc).
func parseHdlr(data []byte) string {
	// version (1) + flags (3) + pre_defined (4) + handler_type (4)
	if len(data) < 12 {
		return ""
	}
	return string(data[8:12])
}

// parseMinf walks into minf → stbl.
func parseMinf(minfFull []byte, trackType string, info *VideoInfo) {
	if len(minfFull) < 8 {
		return
	}
	content := minfFull[8:]

	walkBoxes(content, func(boxType string, _ []byte, fullBox []byte) {
		if boxType == "stbl" {
			parseStbl(fullBox, trackType, info)
		}
	})
}

// parseStbl walks into stbl → stsd.
func parseStbl(stblFull []byte, trackType string, info *VideoInfo) {
	if len(stblFull) < 8 {
		return
	}
	content := stblFull[8:]

	walkBoxes(content, func(boxType string, boxData []byte, _ []byte) {
		if boxType == "stsd" {
			parseStsd(boxData, trackType, info)
		}
	})
}

// parseStsd extracts codec info from sample description entries.
func parseStsd(data []byte, trackType string, info *VideoInfo) {
	// version (1) + flags (3) + entry_count (4)
	if len(data) < 8 {
		return
	}

	entryCount := binary.BigEndian.Uint32(data[4:8])
	if entryCount == 0 {
		return
	}

	// Parse first sample entry
	off := 8
	if off+8 > len(data) {
		return
	}

	entrySize := int(binary.BigEndian.Uint32(data[off : off+4]))
	entryType := string(data[off+4 : off+8])

	if entrySize < 8 || off+entrySize > len(data) {
		return
	}

	switch trackType {
	case "vide":
		info.VideoCodec = videoCodecName(entryType)
		// Visual sample entry: after 8-byte header:
		// reserved (6) + data_ref_idx (2) + pre_defined (2) + reserved (2) +
		// pre_defined (12) + width (2) + height (2)
		entry := data[off:]
		if len(entry) >= 32 {
			info.Width = int(binary.BigEndian.Uint16(entry[32:34]))
			info.Height = int(binary.BigEndian.Uint16(entry[34:36]))
		}
	case "soun":
		info.AudioCodec = audioCodecName(entryType)
	}
}

// videoCodecName maps fourcc to human-readable codec name.
func videoCodecName(fourcc string) string {
	switch fourcc {
	case "avc1", "avc3":
		return "h264"
	case "hvc1", "hev1":
		return "hevc"
	case "vp08":
		return "vp8"
	case "vp09":
		return "vp9"
	case "av01":
		return "av1"
	default:
		return fourcc
	}
}

// audioCodecName maps fourcc to human-readable codec name.
func audioCodecName(fourcc string) string {
	switch fourcc {
	case "mp4a":
		return "aac"
	case "Opus":
		return "opus"
	case "ac-3":
		return "ac3"
	case "ec-3":
		return "eac3"
	case "fLaC":
		return "flac"
	default:
		return fourcc
	}
}

// containsBytes is a simple string search (avoids importing strings package).
func containsBytes(s, substr string) bool {
	return len(substr) <= len(s) && searchString(s, substr) >= 0
}

func searchString(s, sub string) int {
	n := len(sub)
	for i := 0; i <= len(s)-n; i++ {
		if s[i:i+n] == sub {
			return i
		}
	}
	return -1
}
