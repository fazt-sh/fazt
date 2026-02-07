package image

import (
	"fmt"
	"image"
	"image/draw"
	"math"

	xdraw "golang.org/x/image/draw"
)

// ResizeOpts configures a resize operation.
type ResizeOpts struct {
	Width   int    // Target width (0 = auto from height)
	Height  int    // Target height (0 = auto from width)
	Fit     Fit    // Resize strategy
	Format  Format // Output format (empty = same as input)
	Quality int    // JPEG quality 1-100 (0 = default 85)
}

// Resize scales image data to the given dimensions.
// At least one of Width or Height must be > 0.
func Resize(data []byte, opts ResizeOpts) (*Result, error) {
	if opts.Width <= 0 && opts.Height <= 0 {
		return nil, fmt.Errorf("resize: width or height required")
	}

	src, srcFormat, err := Decode(data)
	if err != nil {
		return nil, err
	}

	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	if opts.Fit == "" {
		opts.Fit = FitContain
	}

	outFormat := opts.Format
	if outFormat == "" {
		outFormat = srcFormat
	}

	targetW, targetH := opts.Width, opts.Height

	// Calculate missing dimension from aspect ratio
	if targetW <= 0 {
		targetW = int(math.Round(float64(srcW) * float64(targetH) / float64(srcH)))
	}
	if targetH <= 0 {
		targetH = int(math.Round(float64(srcH) * float64(targetW) / float64(srcW)))
	}

	var dst *image.RGBA

	// Skip resize if source is already smaller than target
	if targetW >= srcW && targetH >= srcH && opts.Fit != FitFill {
		encoded, encErr := EncodeToBytes(src, outFormat, opts.Quality)
		if encErr != nil {
			return nil, fmt.Errorf("resize encode: %w", encErr)
		}
		return &Result{Data: encoded, Width: srcW, Height: srcH, Format: outFormat}, nil
	}

	switch opts.Fit {
	case FitFill:
		dst = image.NewRGBA(image.Rect(0, 0, targetW, targetH))
		xdraw.BiLinear.Scale(dst, dst.Bounds(), src, bounds, xdraw.Over, nil)

	case FitCover:
		dst = resizeCover(src, targetW, targetH)

	default: // FitContain
		dst = resizeContain(src, targetW, targetH)
	}

	encoded, err := EncodeToBytes(dst, outFormat, opts.Quality)
	if err != nil {
		return nil, fmt.Errorf("resize encode: %w", err)
	}

	return &Result{
		Data:   encoded,
		Width:  dst.Bounds().Dx(),
		Height: dst.Bounds().Dy(),
		Format: outFormat,
	}, nil
}

// Thumbnail generates a square thumbnail with center-crop (cover fit).
func Thumbnail(data []byte, size int) (*Result, error) {
	if size <= 0 {
		return nil, fmt.Errorf("thumbnail: size must be > 0")
	}
	return Resize(data, ResizeOpts{
		Width:   size,
		Height:  size,
		Fit:     FitCover,
		Quality: 70,
	})
}

// resizeContain scales the image to fit within targetW x targetH,
// preserving aspect ratio. The output may be smaller than the target.
func resizeContain(src image.Image, targetW, targetH int) *image.RGBA {
	bounds := src.Bounds()
	srcW := float64(bounds.Dx())
	srcH := float64(bounds.Dy())

	scale := math.Min(float64(targetW)/srcW, float64(targetH)/srcH)
	outW := int(math.Round(srcW * scale))
	outH := int(math.Round(srcH * scale))

	dst := image.NewRGBA(image.Rect(0, 0, outW, outH))
	xdraw.BiLinear.Scale(dst, dst.Bounds(), src, bounds, xdraw.Over, nil)
	return dst
}

// resizeCover scales and center-crops to exactly fill targetW x targetH.
func resizeCover(src image.Image, targetW, targetH int) *image.RGBA {
	bounds := src.Bounds()
	srcW := float64(bounds.Dx())
	srcH := float64(bounds.Dy())

	// Scale so shortest side fills the target
	scale := math.Max(float64(targetW)/srcW, float64(targetH)/srcH)
	scaledW := int(math.Round(srcW * scale))
	scaledH := int(math.Round(srcH * scale))

	// Scale up
	scaled := image.NewRGBA(image.Rect(0, 0, scaledW, scaledH))
	xdraw.BiLinear.Scale(scaled, scaled.Bounds(), src, bounds, xdraw.Over, nil)

	// Center-crop to target
	cropX := (scaledW - targetW) / 2
	cropY := (scaledH - targetH) / 2
	cropRect := image.Rect(cropX, cropY, cropX+targetW, cropY+targetH)

	dst := image.NewRGBA(image.Rect(0, 0, targetW, targetH))
	draw.Draw(dst, dst.Bounds(), scaled, cropRect.Min, draw.Src)
	return dst
}
