// Package image provides server-side image processing for fazt serverless.
// Pure Go implementation using golang.org/x/image — no CGO required.
package image

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	_ "golang.org/x/image/webp" // WebP decode support
)

// Format represents an output image format.
type Format string

const (
	FormatJPEG Format = "jpeg"
	FormatPNG  Format = "png"
)

// Fit determines how an image is resized to fit target dimensions.
type Fit string

const (
	FitContain Fit = "contain" // Fit within bounds, preserve aspect ratio (default)
	FitCover   Fit = "cover"   // Fill bounds, crop excess
	FitFill    Fit = "fill"    // Stretch to exact dimensions
)

// Result holds the output of an image operation.
type Result struct {
	Data   []byte
	Width  int
	Height int
	Format Format
}

// Decode reads image data from raw bytes.
// Supports JPEG, PNG, GIF, and WebP input.
func Decode(data []byte) (image.Image, Format, error) {
	r := bytes.NewReader(data)
	img, formatStr, err := image.Decode(r)
	if err != nil {
		return nil, "", fmt.Errorf("decode: %w", err)
	}

	// Map format string to our Format type
	var format Format
	switch formatStr {
	case "jpeg":
		format = FormatJPEG
	case "png":
		format = FormatPNG
	default:
		// GIF, WebP, etc. — default to JPEG for output
		format = FormatJPEG
	}

	return img, format, nil
}

// Encode writes an image to bytes in the specified format.
func Encode(w io.Writer, img image.Image, format Format, quality int) error {
	switch format {
	case FormatPNG:
		return png.Encode(w, img)
	case FormatJPEG:
		if quality <= 0 || quality > 100 {
			quality = 85
		}
		return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
	default:
		return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
	}
}

// EncodeToBytes is a convenience wrapper around Encode.
func EncodeToBytes(img image.Image, format Format, quality int) ([]byte, error) {
	var buf bytes.Buffer
	if err := Encode(&buf, img, format, quality); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DetectFormat guesses the output format from input MIME type.
// Falls back to JPEG if unknown.
func DetectFormat(mime string) Format {
	switch mime {
	case "image/png":
		return FormatPNG
	case "image/gif":
		return FormatJPEG // GIF → JPEG for processed output
	case "image/webp":
		return FormatJPEG // WebP → JPEG (no pure-Go WebP encoder)
	default:
		return FormatJPEG
	}
}

// Register GIF decoder (imported for side effect).
func init() {
	image.RegisterFormat("gif", "GIF8", gif.Decode, gif.DecodeConfig)
}
