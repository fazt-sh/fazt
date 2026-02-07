package media

import (
	"bytes"
	"container/list"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/url"
	"testing"
)

func makeTestJPEG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 100, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	return buf.Bytes()
}

func makeTestPNG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 0, G: 200, B: 100, A: 255})
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// --- TransformOpts tests ---

func TestParseTransformQuery_WidthOnly(t *testing.T) {
	q := url.Values{"w": {"800"}}
	opts := ParseTransformQuery(q)
	if opts.Width != 800 {
		t.Errorf("Width = %d, want 800", opts.Width)
	}
	if opts.Height != 0 {
		t.Errorf("Height = %d, want 0", opts.Height)
	}
	if !opts.HasTransform() {
		t.Error("HasTransform() = false, want true")
	}
}

func TestParseTransformQuery_HeightOnly(t *testing.T) {
	q := url.Values{"h": {"600"}}
	opts := ParseTransformQuery(q)
	if opts.Height != 600 {
		t.Errorf("Height = %d, want 600", opts.Height)
	}
	if opts.Width != 0 {
		t.Errorf("Width = %d, want 0", opts.Width)
	}
}

func TestParseTransformQuery_Full(t *testing.T) {
	q := url.Values{"w": {"200"}, "h": {"200"}, "fit": {"cover"}, "q": {"70"}}
	opts := ParseTransformQuery(q)
	if opts.Width != 200 || opts.Height != 200 {
		t.Errorf("size = %dx%d, want 200x200", opts.Width, opts.Height)
	}
	if opts.Fit != "cover" {
		t.Errorf("Fit = %q, want cover", opts.Fit)
	}
	if opts.Quality != 70 {
		t.Errorf("Quality = %d, want 70", opts.Quality)
	}
}

func TestParseTransformQuery_Empty(t *testing.T) {
	opts := ParseTransformQuery(url.Values{})
	if opts.HasTransform() {
		t.Error("HasTransform() = true for empty query")
	}
}

func TestParseTransformQuery_InvalidValues(t *testing.T) {
	q := url.Values{"w": {"abc"}, "h": {"-5"}, "fit": {"invalid"}, "q": {"200"}}
	opts := ParseTransformQuery(q)
	if opts.Width != 0 {
		t.Errorf("Width = %d, want 0 for invalid", opts.Width)
	}
	if opts.Height != 0 {
		t.Errorf("Height = %d, want 0 for negative", opts.Height)
	}
	if opts.Fit != "" {
		t.Errorf("Fit = %q, want empty for invalid", opts.Fit)
	}
	if opts.Quality != 0 {
		t.Errorf("Quality = %d, want 0 for out-of-range", opts.Quality)
	}
}

func TestParseTransformQuery_MaxDimension(t *testing.T) {
	q := url.Values{"w": {"5000"}}
	opts := ParseTransformQuery(q)
	if opts.Width != 0 {
		t.Errorf("Width = %d, want 0 for >4096", opts.Width)
	}
}

// --- CacheKey tests ---

func TestCacheKey_Deterministic(t *testing.T) {
	opts := TransformOpts{Width: 800, Height: 0, Fit: "contain", Quality: 85}
	key1 := opts.CacheKey()
	key2 := opts.CacheKey()
	if key1 != key2 {
		t.Errorf("CacheKey not deterministic: %q != %q", key1, key2)
	}
}

func TestCacheKey_DefaultValues(t *testing.T) {
	opts := TransformOpts{Width: 200, Height: 200}
	key := opts.CacheKey()
	// Should use defaults for fit=contain, q=85
	if key != "200x200_contain_q85" {
		t.Errorf("CacheKey = %q, want 200x200_contain_q85", key)
	}
}

func TestCacheKey_WithFitAndQuality(t *testing.T) {
	opts := TransformOpts{Width: 200, Height: 200, Fit: "cover", Quality: 70}
	key := opts.CacheKey()
	if key != "200x200_cover_q70" {
		t.Errorf("CacheKey = %q, want 200x200_cover_q70", key)
	}
}

// --- IsImageContentType tests ---

func TestIsImageContentType(t *testing.T) {
	tests := []struct {
		ct   string
		want bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"image/gif", true},
		{"image/webp", true},
		{"application/json", false},
		{"text/html", false},
		{"application/octet-stream", false},
		{"", false},
	}
	for _, tt := range tests {
		got := IsImageContentType(tt.ct)
		if got != tt.want {
			t.Errorf("IsImageContentType(%q) = %v, want %v", tt.ct, got, tt.want)
		}
	}
}

// --- SnapToStep tests ---

func TestSnapToStep_ExactMultiple(t *testing.T) {
	opts := TransformOpts{Width: 800}
	snapped := SnapToStep(opts, 50)
	if snapped.Width != 800 {
		t.Errorf("Width = %d, want 800 (exact multiple)", snapped.Width)
	}
}

func TestSnapToStep_SnapsToNearest(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{120, 150},  // rounds up
		{130, 150},  // rounds up
		{101, 150},  // just over → next step
		{1, 50},     // very small → minimum step
		{49, 50},    // just below
		{51, 100},   // just above
		{1919, 1950},
		{1950, 1950}, // exact multiple unchanged
	}
	for _, tt := range tests {
		snapped := SnapToStep(TransformOpts{Width: tt.input}, 50)
		if snapped.Width != tt.want {
			t.Errorf("SnapToStep(%d) = %d, want %d", tt.input, snapped.Width, tt.want)
		}
	}
}

func TestSnapToStep_ScalesHeight(t *testing.T) {
	opts := TransformOpts{Width: 320, Height: 240}
	snapped := SnapToStep(opts, 50)
	// 320 → rounds up to 350. Height should scale: 240 * 350/320 = 263
	if snapped.Width != 350 {
		t.Errorf("Width = %d, want 350", snapped.Width)
	}
	if snapped.Height != 263 {
		t.Errorf("Height = %d, want 263", snapped.Height)
	}
}

func TestSnapToStep_NoWidth(t *testing.T) {
	opts := TransformOpts{Height: 600}
	snapped := SnapToStep(opts, 50)
	if snapped.Height != 600 {
		t.Errorf("Height = %d, want 600 (unchanged)", snapped.Height)
	}
}

func TestSnapToStep_ZeroStep(t *testing.T) {
	opts := TransformOpts{Width: 500}
	snapped := SnapToStep(opts, 0)
	if snapped.Width != 500 {
		t.Errorf("Width = %d, want 500 (no step)", snapped.Width)
	}
}

// --- ProcessImage tests ---

func TestProcessImage_ResizeJPEG(t *testing.T) {
	data := makeTestJPEG(800, 600)
	opts := TransformOpts{Width: 400}
	processed, mime, err := ProcessImage(data, opts)
	if err != nil {
		t.Fatalf("ProcessImage: %v", err)
	}
	if mime != "image/jpeg" {
		t.Errorf("mime = %q, want image/jpeg", mime)
	}
	if len(processed) == 0 {
		t.Error("processed data is empty")
	}
	// Verify it's smaller than original
	if len(processed) >= len(data) {
		t.Errorf("processed (%d) should be smaller than original (%d)", len(processed), len(data))
	}
}

func TestProcessImage_ResizePNG(t *testing.T) {
	data := makeTestPNG(400, 300)
	opts := TransformOpts{Width: 200}
	processed, mime, err := ProcessImage(data, opts)
	if err != nil {
		t.Fatalf("ProcessImage: %v", err)
	}
	if mime != "image/png" {
		t.Errorf("mime = %q, want image/png", mime)
	}
	if len(processed) == 0 {
		t.Error("processed data is empty")
	}
}

func TestProcessImage_CoverFit(t *testing.T) {
	data := makeTestJPEG(800, 600)
	opts := TransformOpts{Width: 200, Height: 200, Fit: "cover"}
	processed, _, err := ProcessImage(data, opts)
	if err != nil {
		t.Fatalf("ProcessImage cover: %v", err)
	}
	// Verify output by decoding
	img, _, err := image.Decode(bytes.NewReader(processed))
	if err != nil {
		t.Fatalf("decode processed: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 200 || b.Dy() != 200 {
		t.Errorf("size = %dx%d, want 200x200", b.Dx(), b.Dy())
	}
}

func TestProcessImage_QualityOverride(t *testing.T) {
	data := makeTestJPEG(800, 600)
	highQ, _, _ := ProcessImage(data, TransformOpts{Width: 400, Quality: 95})
	lowQ, _, _ := ProcessImage(data, TransformOpts{Width: 400, Quality: 20})
	if len(lowQ) >= len(highQ) {
		t.Errorf("low quality (%d) should be smaller than high quality (%d)", len(lowQ), len(highQ))
	}
}

// --- memCache tests ---

func TestMemCache_GetPut(t *testing.T) {
	mc := &memCache{
		items:   make(map[string]*list.Element),
		order:   list.New(),
		maxSize: 1024 * 1024, // 1MB
	}

	// Miss
	data, mime := mc.get("app1", "path1")
	if data != nil {
		t.Error("expected miss on empty cache")
	}

	// Put + hit
	mc.put("app1", "path1", []byte("hello"), "image/jpeg")
	data, mime = mc.get("app1", "path1")
	if string(data) != "hello" || mime != "image/jpeg" {
		t.Errorf("got %q/%q, want hello/image/jpeg", data, mime)
	}
}

func TestMemCache_Eviction(t *testing.T) {
	mc := &memCache{
		items:   make(map[string]*list.Element),
		order:   list.New(),
		maxSize: 1000,
	}

	// Fill with 240 bytes each (under 25% of 1000 = 250)
	mc.put("app1", "a", make([]byte, 240), "image/jpeg")
	mc.put("app1", "b", make([]byte, 240), "image/jpeg")
	mc.put("app1", "c", make([]byte, 240), "image/jpeg")
	mc.put("app1", "d", make([]byte, 240), "image/jpeg")
	// 960 bytes. Add one more — should evict "a"
	mc.put("app1", "e", make([]byte, 240), "image/jpeg")

	if data, _ := mc.get("app1", "a"); data != nil {
		t.Error("expected 'a' to be evicted")
	}
	if data, _ := mc.get("app1", "e"); data == nil {
		t.Error("expected 'e' to be present")
	}
}

func TestMemCache_LRUOrder(t *testing.T) {
	mc := &memCache{
		items:   make(map[string]*list.Element),
		order:   list.New(),
		maxSize: 900, // 4 × 200 = 800, 5th triggers eviction
	}

	mc.put("app1", "a", make([]byte, 200), "image/jpeg")
	mc.put("app1", "b", make([]byte, 200), "image/jpeg")
	mc.put("app1", "c", make([]byte, 200), "image/jpeg")
	mc.put("app1", "d", make([]byte, 200), "image/jpeg")

	// Access "a" to promote it
	mc.get("app1", "a")

	// Add "e" — should evict "b" (least recently used), not "a"
	mc.put("app1", "e", make([]byte, 200), "image/jpeg")

	if data, _ := mc.get("app1", "a"); data == nil {
		t.Error("expected 'a' to survive (recently accessed)")
	}
	if data, _ := mc.get("app1", "b"); data != nil {
		t.Error("expected 'b' to be evicted (LRU)")
	}
}

func TestMemCache_InvalidatePrefix(t *testing.T) {
	mc := &memCache{
		items:   make(map[string]*list.Element),
		order:   list.New(),
		maxSize: 1024 * 1024,
	}

	mc.put("app1", "_media/abc/200x0_contain_q85", []byte("thumb"), "image/jpeg")
	mc.put("app1", "_media/abc/800x0_contain_q85", []byte("medium"), "image/jpeg")
	mc.put("app1", "_media/def/200x0_contain_q85", []byte("other"), "image/jpeg")

	// Invalidate all variants for hash "abc"
	mc.invalidatePrefix("app1", "_media/abc/")

	if data, _ := mc.get("app1", "_media/abc/200x0_contain_q85"); data != nil {
		t.Error("expected abc/200 to be invalidated")
	}
	if data, _ := mc.get("app1", "_media/abc/800x0_contain_q85"); data != nil {
		t.Error("expected abc/800 to be invalidated")
	}
	// "def" should survive
	if data, _ := mc.get("app1", "_media/def/200x0_contain_q85"); data == nil {
		t.Error("expected def/200 to survive")
	}
}

func TestMemCache_SkipsOversizedEntries(t *testing.T) {
	mc := &memCache{
		items:   make(map[string]*list.Element),
		order:   list.New(),
		maxSize: 100,
	}

	// Entry larger than 25% of max (25 bytes) should be skipped
	mc.put("app1", "big", make([]byte, 30), "image/jpeg")
	if data, _ := mc.get("app1", "big"); data != nil {
		t.Error("expected oversized entry to be skipped")
	}
}
