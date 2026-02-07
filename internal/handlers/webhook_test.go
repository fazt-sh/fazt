package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/handlers/testutil"
)

func setupWebhookTest(t *testing.T) {
	t.Helper()
	silenceTestLogs(t)

	db := setupTestDB(t)
	database.SetDB(db)
	t.Cleanup(func() {
		db.Close()
		database.SetDB(nil)
	})
}

// createWebhookWithSecret creates a webhook with a specific secret
func createWebhookWithSecret(t *testing.T, name, endpoint, secret string) int64 {
	t.Helper()
	db := database.GetDB()

	result, err := db.Exec(`
		INSERT INTO webhooks (name, endpoint, secret, is_active)
		VALUES (?, ?, ?, 1)
	`, name, endpoint, secret)
	if err != nil {
		t.Fatalf("Failed to create webhook: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}

// createInactiveWebhook creates a disabled webhook
func createInactiveWebhook(t *testing.T, name, endpoint string) int64 {
	t.Helper()
	db := database.GetDB()

	result, err := db.Exec(`
		INSERT INTO webhooks (name, endpoint, secret, is_active)
		VALUES (?, ?, '', 0)
	`, name, endpoint)
	if err != nil {
		t.Fatalf("Failed to create inactive webhook: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}

func computeHMAC(body, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(body))
	return hex.EncodeToString(mac.Sum(nil))
}

// --- WebhookHandler ---

func TestWebhookHandler_Success_NoSecret(t *testing.T) {
	setupWebhookTest(t)
	createWebhookWithSecret(t, "test-hook", "github-push", "")

	body := `{"event":"push","repository":"test/repo"}`
	req := httptest.NewRequest("POST", "/webhook/github-push", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	WebhookHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "message", "Webhook received")
	testutil.AssertFieldEquals(t, data, "webhook", "test-hook")
}

func TestWebhookHandler_Success_WithSignature(t *testing.T) {
	setupWebhookTest(t)
	secret := "my-webhook-secret"
	createWebhookWithSecret(t, "signed-hook", "signed-ep", secret)

	body := `{"event":"push"}`
	signature := computeHMAC(body, secret)

	req := httptest.NewRequest("POST", "/webhook/signed-ep", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", signature)
	resp := httptest.NewRecorder()
	WebhookHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "message", "Webhook received")
}

func TestWebhookHandler_InvalidSignature(t *testing.T) {
	setupWebhookTest(t)
	createWebhookWithSecret(t, "sig-hook", "sig-ep", "real-secret")

	body := `{"event":"push"}`
	req := httptest.NewRequest("POST", "/webhook/sig-ep", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Signature", "invalid-signature")
	resp := httptest.NewRecorder()
	WebhookHandler(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid signature, got %d", resp.Code)
	}
}

func TestWebhookHandler_MissingSignature(t *testing.T) {
	setupWebhookTest(t)
	createWebhookWithSecret(t, "missing-sig-hook", "missing-sig-ep", "a-secret")

	body := `{"event":"push"}`
	req := httptest.NewRequest("POST", "/webhook/missing-sig-ep", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// No X-Webhook-Signature header
	resp := httptest.NewRecorder()
	WebhookHandler(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for missing signature, got %d", resp.Code)
	}
}

func TestWebhookHandler_NotFound(t *testing.T) {
	setupWebhookTest(t)

	req := httptest.NewRequest("POST", "/webhook/nonexistent", strings.NewReader(`{}`))
	resp := httptest.NewRecorder()
	WebhookHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusNotFound, "WEBHOOK_NOT_FOUND")
}

func TestWebhookHandler_Inactive(t *testing.T) {
	setupWebhookTest(t)
	createInactiveWebhook(t, "disabled-hook", "disabled-ep")

	req := httptest.NewRequest("POST", "/webhook/disabled-ep", strings.NewReader(`{}`))
	resp := httptest.NewRecorder()
	WebhookHandler(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for inactive webhook, got %d", resp.Code)
	}
}

func TestWebhookHandler_EmptyEndpoint(t *testing.T) {
	setupWebhookTest(t)

	req := httptest.NewRequest("POST", "/webhook/", strings.NewReader(`{}`))
	resp := httptest.NewRecorder()
	WebhookHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for empty endpoint, got %d", resp.Code)
	}
}

func TestWebhookHandler_MethodNotAllowed(t *testing.T) {
	setupWebhookTest(t)

	req := httptest.NewRequest("GET", "/webhook/test", nil)
	resp := httptest.NewRecorder()
	WebhookHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestWebhookHandler_NonJSONBody(t *testing.T) {
	setupWebhookTest(t)
	createWebhookWithSecret(t, "raw-hook", "raw-ep", "")

	req := httptest.NewRequest("POST", "/webhook/raw-ep", strings.NewReader("plain text body"))
	resp := httptest.NewRecorder()
	WebhookHandler(resp, req)

	// Should still succeed — non-JSON body stored as raw
	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "message", "Webhook received")
}

func TestWebhookHandler_EventTypeExtraction(t *testing.T) {
	setupWebhookTest(t)
	createWebhookWithSecret(t, "event-hook", "event-ep", "")

	// Test with "event" field
	body := `{"event":"deployment"}`
	req := httptest.NewRequest("POST", "/webhook/event-ep", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	WebhookHandler(resp, req)

	testutil.CheckSuccess(t, resp, http.StatusOK)
}

// --- verifySignature ---

func TestVerifySignature_Valid(t *testing.T) {
	secret := "test-secret"
	body := []byte("hello world")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	if !verifySignature(body, secret, signature) {
		t.Error("Expected valid signature to verify")
	}
}

func TestVerifySignature_Invalid(t *testing.T) {
	if verifySignature([]byte("hello"), "secret", "invalid") {
		t.Error("Expected invalid signature to fail")
	}
}

func TestVerifySignature_EmptyBody(t *testing.T) {
	secret := "test-secret"
	body := []byte("")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	signature := hex.EncodeToString(mac.Sum(nil))

	if !verifySignature(body, secret, signature) {
		t.Error("Expected empty body signature to verify")
	}
}
