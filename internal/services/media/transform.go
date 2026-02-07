package media

import (
	"context"
	"math"
	"net/url"
	"strconv"
)

type contextKey string

const queryKey contextKey = "media.query"

// WithQuery attaches URL query params to a context for media bindings.
func WithQuery(ctx context.Context, q url.Values) context.Context {
	return context.WithValue(ctx, queryKey, q)
}

// QueryFromContext extracts URL query params from context.
func QueryFromContext(ctx context.Context) url.Values {
	if q, ok := ctx.Value(queryKey).(url.Values); ok {
		return q
	}
	return nil
}

// TransformOpts holds parsed image transform parameters from query string.
type TransformOpts struct {
	Width   int
	Height  int
	Fit     string // "contain", "cover", "fill"
	Quality int    // 1-100, 0 = default (85)
}

// HasTransform returns true if any transform parameter is set.
func (t TransformOpts) HasTransform() bool {
	return t.Width > 0 || t.Height > 0
}

// CacheKey returns a deterministic string for caching this transform variant.
func (t TransformOpts) CacheKey() string {
	fit := t.Fit
	if fit == "" {
		fit = "contain"
	}
	q := t.Quality
	if q <= 0 {
		q = 85
	}
	return strconv.Itoa(t.Width) + "x" + strconv.Itoa(t.Height) + "_" + fit + "_q" + strconv.Itoa(q)
}

// ParseTransformQuery extracts transform options from URL query parameters.
//
//	?w=800        → width 800, auto height
//	?h=600        → height 600, auto width
//	?w=200&h=200&fit=cover  → thumbnail
//	?q=70         → quality override
func ParseTransformQuery(q url.Values) TransformOpts {
	opts := TransformOpts{}

	if w := q.Get("w"); w != "" {
		if n, err := strconv.Atoi(w); err == nil && n > 0 && n <= 4096 {
			opts.Width = n
		}
	}
	if h := q.Get("h"); h != "" {
		if n, err := strconv.Atoi(h); err == nil && n > 0 && n <= 4096 {
			opts.Height = n
		}
	}
	if fit := q.Get("fit"); fit != "" {
		switch fit {
		case "contain", "cover", "fill":
			opts.Fit = fit
		}
	}
	if quality := q.Get("q"); quality != "" {
		if n, err := strconv.Atoi(quality); err == nil && n >= 1 && n <= 100 {
			opts.Quality = n
		}
	}

	return opts
}

// SnapToStep rounds opts.Width to the nearest multiple of step.
// If the width is already a multiple (or no width is set), opts is returned unchanged.
// Height is scaled proportionally when width is snapped.
func SnapToStep(opts TransformOpts, step int) TransformOpts {
	if step <= 0 || opts.Width <= 0 {
		return opts
	}

	// Already a multiple
	if opts.Width%step == 0 {
		return opts
	}

	// Round up — always serve at least the requested size
	snapped := opts
	best := int(math.Ceil(float64(opts.Width)/float64(step))) * step
	if best > 4096 {
		best = 4096
	}

	// Scale height proportionally if both were set
	if opts.Height > 0 && opts.Width > 0 {
		snapped.Height = int(math.Round(float64(opts.Height) * float64(best) / float64(opts.Width)))
	}
	snapped.Width = best
	return snapped
}
