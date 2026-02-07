// Package media provides on-demand image processing for the fazt platform.
// Images are transformed at serve time based on URL query parameters and
// cached in app_blobs for subsequent requests.
package media

import (
	"context"
	"strings"
	"sync"

	imgservice "github.com/fazt-sh/fazt/internal/services/image"
	"github.com/fazt-sh/fazt/internal/system"
)

// resizeSem is the global concurrency limiter for image resize operations.
// Initialized once from system.Limits.Media.Concurrency.
var (
	resizeSem  chan struct{}
	resizeOnce sync.Once
)

func getResizeSem() chan struct{} {
	resizeOnce.Do(func() {
		n := system.GetLimits().Media.Concurrency
		if n <= 0 {
			n = 2
		}
		resizeSem = make(chan struct{}, n)
	})
	return resizeSem
}

// IsImageContentType returns true if the content type is a processable image.
func IsImageContentType(contentType string) bool {
	switch {
	case strings.HasPrefix(contentType, "image/jpeg"),
		strings.HasPrefix(contentType, "image/png"),
		strings.HasPrefix(contentType, "image/gif"),
		strings.HasPrefix(contentType, "image/webp"):
		return true
	default:
		return false
	}
}

// ProcessImage resizes image data according to opts and returns the processed bytes + mime type.
func ProcessImage(data []byte, opts TransformOpts) ([]byte, string, error) {
	fit := imgservice.Fit(opts.Fit)
	if fit == "" {
		fit = imgservice.FitContain
	}

	result, err := imgservice.Resize(data, imgservice.ResizeOpts{
		Width:   opts.Width,
		Height:  opts.Height,
		Fit:     fit,
		Quality: opts.Quality,
	})
	if err != nil {
		return nil, "", err
	}

	mime := "image/jpeg"
	if result.Format == imgservice.FormatPNG {
		mime = "image/png"
	}

	return result.Data, mime, nil
}

// ProcessAndCache checks the cache for a variant, or processes the image and caches the result.
//
// All widths are snapped to the nearest step (rounded up) before processing,
// so every request is cacheable. This prevents cache-bypass attacks where
// an attacker requests arbitrary widths to force repeated resizes.
//
// Concurrency is limited by system.Limits.Media.Concurrency. If the resize queue is full,
// returns the original data unprocessed (graceful degradation).
func ProcessAndCache(ctx context.Context, cache *MediaCache, appID, originalPath string, data []byte, opts TransformOpts) ([]byte, string, error) {
	limits := system.GetLimits().Media

	// Reject images larger than the configured max
	if int64(len(data)) > limits.MaxSourceBytes {
		return data, "", nil
	}

	// Snap width to step — every request resolves to a cacheable size.
	// ?w=320 → 350, ?w=800 → 800 (already aligned). Always rounds up
	// so the frontend never gets a smaller image than requested.
	opts = SnapToStep(opts, limits.WidthStep)

	// Check cache (in-memory first, then DB)
	cached, mime, err := cache.Get(ctx, appID, originalPath, opts)
	if err == nil && cached != nil {
		return cached, mime, nil
	}

	// Acquire resize slot (non-blocking — if full, serve original)
	sem := getResizeSem()
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	default:
		// All resize slots busy — graceful degradation: serve original
		return data, "", nil
	}

	// Process at snapped size
	processed, mime, err := ProcessImage(data, opts)
	if err != nil {
		return nil, "", err
	}

	_ = cache.Put(ctx, appID, originalPath, opts, processed, mime)

	return processed, mime, nil
}
