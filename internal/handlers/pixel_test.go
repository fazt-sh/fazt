package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- PixelHandler ---

func TestPixelHandler_Basic(t *testing.T) {
	silenceTestLogs(t)

	req := httptest.NewRequest("GET", "/pixel?domain=example.com", nil)
	resp := httptest.NewRecorder()
	PixelHandler(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.Code)
	}
	ct := resp.Header().Get("Content-Type")
	if ct != "image/gif" {
		t.Errorf("Expected Content-Type image/gif, got %s", ct)
	}
	// Should be a tiny GIF
	if resp.Body.Len() == 0 {
		t.Error("Expected non-empty body (GIF data)")
	}
}

func TestPixelHandler_NoCacheHeaders(t *testing.T) {
	silenceTestLogs(t)

	req := httptest.NewRequest("GET", "/pixel?domain=test.com", nil)
	resp := httptest.NewRecorder()
	PixelHandler(resp, req)

	cc := resp.Header().Get("Cache-Control")
	if cc != "no-cache, no-store, must-revalidate" {
		t.Errorf("Expected no-cache header, got %s", cc)
	}
	pragma := resp.Header().Get("Pragma")
	if pragma != "no-cache" {
		t.Errorf("Expected Pragma no-cache, got %s", pragma)
	}
}

func TestPixelHandler_WithTags(t *testing.T) {
	silenceTestLogs(t)

	req := httptest.NewRequest("GET", "/pixel?domain=test.com&tags=email,newsletter", nil)
	resp := httptest.NewRecorder()
	PixelHandler(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.Code)
	}
}

func TestPixelHandler_WithSource(t *testing.T) {
	silenceTestLogs(t)

	req := httptest.NewRequest("GET", "/pixel?domain=test.com&source=email-open", nil)
	resp := httptest.NewRecorder()
	PixelHandler(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.Code)
	}
}

func TestPixelHandler_NoDomain(t *testing.T) {
	silenceTestLogs(t)

	req := httptest.NewRequest("GET", "/pixel", nil)
	resp := httptest.NewRecorder()
	PixelHandler(resp, req)

	// Should still succeed with domain="unknown"
	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.Code)
	}
}

func TestPixelHandler_DomainFromReferer(t *testing.T) {
	silenceTestLogs(t)

	req := httptest.NewRequest("GET", "/pixel", nil)
	req.Header.Set("Referer", "https://mysite.com/page")
	resp := httptest.NewRecorder()
	PixelHandler(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", resp.Code)
	}
}

// --- extractDomainFromReferer ---

func TestExtractDomainFromReferer_HTTPS(t *testing.T) {
	result := extractDomainFromReferer("https://example.com/path/page")
	if result != "example.com" {
		t.Errorf("Expected example.com, got %s", result)
	}
}

func TestExtractDomainFromReferer_HTTP(t *testing.T) {
	result := extractDomainFromReferer("http://test.org/")
	if result != "test.org" {
		t.Errorf("Expected test.org, got %s", result)
	}
}

func TestExtractDomainFromReferer_NoProtocol(t *testing.T) {
	result := extractDomainFromReferer("plain.com/page")
	if result != "plain.com" {
		t.Errorf("Expected plain.com, got %s", result)
	}
}

func TestExtractDomainFromReferer_Empty(t *testing.T) {
	result := extractDomainFromReferer("")
	if result != "" {
		t.Errorf("Expected empty, got %s", result)
	}
}
