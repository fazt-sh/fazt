package image

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"bytes"
	"testing"
)

// makeTestJPEG creates a solid-color JPEG for testing.
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

// makeTestPNG creates a solid-color PNG for testing.
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

func TestDecode_JPEG(t *testing.T) {
	data := makeTestJPEG(100, 80)
	img, format, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode JPEG: %v", err)
	}
	if format != FormatJPEG {
		t.Errorf("format = %q, want jpeg", format)
	}
	b := img.Bounds()
	if b.Dx() != 100 || b.Dy() != 80 {
		t.Errorf("size = %dx%d, want 100x80", b.Dx(), b.Dy())
	}
}

func TestDecode_PNG(t *testing.T) {
	data := makeTestPNG(200, 150)
	img, format, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode PNG: %v", err)
	}
	if format != FormatPNG {
		t.Errorf("format = %q, want png", format)
	}
	b := img.Bounds()
	if b.Dx() != 200 || b.Dy() != 150 {
		t.Errorf("size = %dx%d, want 200x150", b.Dx(), b.Dy())
	}
}

func TestDecode_InvalidData(t *testing.T) {
	_, _, err := Decode([]byte("not an image"))
	if err == nil {
		t.Fatal("expected error for invalid data")
	}
}

func TestResize_WidthOnly(t *testing.T) {
	data := makeTestJPEG(800, 600)
	result, err := Resize(data, ResizeOpts{Width: 400})
	if err != nil {
		t.Fatalf("Resize: %v", err)
	}
	if result.Width != 400 {
		t.Errorf("width = %d, want 400", result.Width)
	}
	if result.Height != 300 {
		t.Errorf("height = %d, want 300", result.Height)
	}
	if len(result.Data) == 0 {
		t.Error("result data is empty")
	}
}

func TestResize_HeightOnly(t *testing.T) {
	data := makeTestJPEG(800, 600)
	result, err := Resize(data, ResizeOpts{Height: 300})
	if err != nil {
		t.Fatalf("Resize: %v", err)
	}
	if result.Width != 400 {
		t.Errorf("width = %d, want 400", result.Width)
	}
	if result.Height != 300 {
		t.Errorf("height = %d, want 300", result.Height)
	}
}

func TestResize_Contain(t *testing.T) {
	data := makeTestJPEG(800, 400) // 2:1 landscape
	result, err := Resize(data, ResizeOpts{Width: 300, Height: 300, Fit: FitContain})
	if err != nil {
		t.Fatalf("Resize contain: %v", err)
	}
	// Should fit within 300x300, preserving 2:1 → 300x150
	if result.Width != 300 {
		t.Errorf("width = %d, want 300", result.Width)
	}
	if result.Height != 150 {
		t.Errorf("height = %d, want 150", result.Height)
	}
}

func TestResize_Cover(t *testing.T) {
	data := makeTestJPEG(800, 400) // 2:1 landscape
	result, err := Resize(data, ResizeOpts{Width: 200, Height: 200, Fit: FitCover})
	if err != nil {
		t.Fatalf("Resize cover: %v", err)
	}
	// Should fill 200x200 exactly (scale up shortest, crop)
	if result.Width != 200 {
		t.Errorf("width = %d, want 200", result.Width)
	}
	if result.Height != 200 {
		t.Errorf("height = %d, want 200", result.Height)
	}
}

func TestResize_Fill(t *testing.T) {
	data := makeTestJPEG(800, 400)
	result, err := Resize(data, ResizeOpts{Width: 300, Height: 300, Fit: FitFill})
	if err != nil {
		t.Fatalf("Resize fill: %v", err)
	}
	if result.Width != 300 || result.Height != 300 {
		t.Errorf("size = %dx%d, want 300x300", result.Width, result.Height)
	}
}

func TestResize_FormatConversion(t *testing.T) {
	data := makeTestPNG(100, 100)
	result, err := Resize(data, ResizeOpts{Width: 50, Format: FormatJPEG})
	if err != nil {
		t.Fatalf("Resize format: %v", err)
	}
	if result.Format != FormatJPEG {
		t.Errorf("format = %q, want jpeg", result.Format)
	}
	// Verify it's valid JPEG
	_, err = jpeg.Decode(bytes.NewReader(result.Data))
	if err != nil {
		t.Errorf("output is not valid JPEG: %v", err)
	}
}

func TestResize_NoDimensions(t *testing.T) {
	data := makeTestJPEG(100, 100)
	_, err := Resize(data, ResizeOpts{})
	if err == nil {
		t.Fatal("expected error when no dimensions provided")
	}
}

func TestThumbnail(t *testing.T) {
	data := makeTestJPEG(800, 600) // 4:3
	result, err := Thumbnail(data, 200)
	if err != nil {
		t.Fatalf("Thumbnail: %v", err)
	}
	if result.Width != 200 || result.Height != 200 {
		t.Errorf("size = %dx%d, want 200x200", result.Width, result.Height)
	}
	if len(result.Data) == 0 {
		t.Error("thumbnail data is empty")
	}
}

func TestThumbnail_Portrait(t *testing.T) {
	data := makeTestJPEG(400, 800) // portrait
	result, err := Thumbnail(data, 150)
	if err != nil {
		t.Fatalf("Thumbnail portrait: %v", err)
	}
	if result.Width != 150 || result.Height != 150 {
		t.Errorf("size = %dx%d, want 150x150", result.Width, result.Height)
	}
}

func TestThumbnail_InvalidSize(t *testing.T) {
	data := makeTestJPEG(100, 100)
	_, err := Thumbnail(data, 0)
	if err == nil {
		t.Fatal("expected error for size=0")
	}
}

func TestEncodeToBytes_PNG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	data, err := EncodeToBytes(img, FormatPNG, 0)
	if err != nil {
		t.Fatalf("EncodeToBytes PNG: %v", err)
	}
	// Verify it's valid PNG
	_, err = png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Errorf("output is not valid PNG: %v", err)
	}
}
