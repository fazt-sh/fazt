package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/handlers/testutil"
)

func setupAPITest(t *testing.T) {
	t.Helper()
	silenceTestLogs(t)

	db := setupTestDB(t)
	database.SetDB(db)
	t.Cleanup(func() {
		db.Close()
		database.SetDB(nil)
	})
}

// --- StatsHandler ---

func TestStatsHandler_Empty(t *testing.T) {
	setupAPITest(t)

	req := httptest.NewRequest("GET", "/api/stats", nil)
	resp := httptest.NewRecorder()
	StatsHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	if data == nil {
		t.Fatal("Expected stats data")
	}
}

func TestStatsHandler_WithEvents(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	createTestEvent(t, db, "example.com", "pageview")
	createTestEvent(t, db, "example.com", "click")
	createTestEvent(t, db, "other.com", "pageview")

	req := httptest.NewRequest("GET", "/api/stats", nil)
	resp := httptest.NewRecorder()
	StatsHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	// total_events_all_time should be 3
	if total, ok := data["total_events_all_time"].(float64); !ok || total != 3 {
		t.Errorf("Expected 3 total events, got %v", data["total_events_all_time"])
	}
}

func TestStatsHandler_MethodNotAllowed(t *testing.T) {
	setupAPITest(t)

	req := httptest.NewRequest("POST", "/api/stats", nil)
	resp := httptest.NewRecorder()
	StatsHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

// --- EventsHandler ---

func TestEventsHandler_Empty(t *testing.T) {
	setupAPITest(t)

	req := httptest.NewRequest("GET", "/api/events", nil)
	resp := httptest.NewRecorder()
	EventsHandler(resp, req)

	arr := testutil.CheckSuccessArray(t, resp, http.StatusOK)
	if len(arr) != 0 {
		t.Errorf("Expected 0 events, got %d", len(arr))
	}
}

func TestEventsHandler_WithEvents(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	createTestEvent(t, db, "example.com", "pageview")
	createTestEvent(t, db, "other.com", "click")

	req := httptest.NewRequest("GET", "/api/events", nil)
	resp := httptest.NewRecorder()
	EventsHandler(resp, req)

	arr := testutil.CheckSuccessArray(t, resp, http.StatusOK)
	if len(arr) != 2 {
		t.Errorf("Expected 2 events, got %d", len(arr))
	}
}

func TestEventsHandler_FilterByDomain(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	createTestEvent(t, db, "example.com", "pageview")
	createTestEvent(t, db, "other.com", "click")

	req := httptest.NewRequest("GET", "/api/events?domain=example.com", nil)
	resp := httptest.NewRecorder()
	EventsHandler(resp, req)

	arr := testutil.CheckSuccessArray(t, resp, http.StatusOK)
	if len(arr) != 1 {
		t.Errorf("Expected 1 event for domain filter, got %d", len(arr))
	}
}

func TestEventsHandler_FilterBySourceType(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	createTestEvent(t, db, "example.com", "pageview")

	req := httptest.NewRequest("GET", "/api/events?source_type=web", nil)
	resp := httptest.NewRecorder()
	EventsHandler(resp, req)

	arr := testutil.CheckSuccessArray(t, resp, http.StatusOK)
	if len(arr) != 1 {
		t.Errorf("Expected 1 event for source_type filter, got %d", len(arr))
	}
}

func TestEventsHandler_Pagination(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	for i := 0; i < 5; i++ {
		createTestEvent(t, db, "example.com", "pageview")
	}

	req := httptest.NewRequest("GET", "/api/events?limit=2&offset=0", nil)
	resp := httptest.NewRecorder()
	EventsHandler(resp, req)

	arr := testutil.CheckSuccessArray(t, resp, http.StatusOK)
	if len(arr) != 2 {
		t.Errorf("Expected 2 events with limit=2, got %d", len(arr))
	}
}

func TestEventsHandler_MethodNotAllowed(t *testing.T) {
	setupAPITest(t)

	req := httptest.NewRequest("POST", "/api/events", nil)
	resp := httptest.NewRecorder()
	EventsHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

// --- DomainsHandler ---

func TestDomainsHandler_Empty(t *testing.T) {
	setupAPITest(t)

	req := httptest.NewRequest("GET", "/api/domains", nil)
	resp := httptest.NewRecorder()
	DomainsHandler(resp, req)

	arr := testutil.CheckSuccessArray(t, resp, http.StatusOK)
	if len(arr) != 0 {
		t.Errorf("Expected 0 domains, got %d", len(arr))
	}
}

func TestDomainsHandler_WithEvents(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	createTestEvent(t, db, "example.com", "pageview")
	createTestEvent(t, db, "example.com", "click")
	createTestEvent(t, db, "other.com", "pageview")

	req := httptest.NewRequest("GET", "/api/domains", nil)
	resp := httptest.NewRecorder()
	DomainsHandler(resp, req)

	arr := testutil.CheckSuccessArray(t, resp, http.StatusOK)
	if len(arr) != 2 {
		t.Errorf("Expected 2 domains, got %d", len(arr))
	}
}

// --- TagsHandler ---

func TestTagsHandler_Empty(t *testing.T) {
	setupAPITest(t)

	req := httptest.NewRequest("GET", "/api/tags", nil)
	resp := httptest.NewRecorder()
	TagsHandler(resp, req)

	arr := testutil.CheckSuccessArray(t, resp, http.StatusOK)
	if len(arr) != 0 {
		t.Errorf("Expected 0 tags, got %d", len(arr))
	}
}

func TestTagsHandler_WithTags(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	// Insert event with tags
	_, err := db.Exec(`INSERT INTO events (domain, event_type, source_type, path, tags) VALUES (?, ?, ?, ?, ?)`,
		"example.com", "pageview", "web", "/", "blog,tech")
	if err != nil {
		t.Fatalf("Failed to insert event with tags: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/tags", nil)
	resp := httptest.NewRecorder()
	TagsHandler(resp, req)

	arr := testutil.CheckSuccessArray(t, resp, http.StatusOK)
	if len(arr) < 1 {
		t.Errorf("Expected at least 1 tag, got %d", len(arr))
	}
}

// --- RedirectsHandler ---

func TestRedirectsHandler_ListEmpty(t *testing.T) {
	setupAPITest(t)

	req := httptest.NewRequest("GET", "/api/redirects", nil)
	resp := httptest.NewRecorder()
	RedirectsHandler(resp, req)

	arr := testutil.CheckSuccessArray(t, resp, http.StatusOK)
	if len(arr) != 0 {
		t.Errorf("Expected 0 redirects, got %d", len(arr))
	}
}

func TestRedirectsHandler_Create(t *testing.T) {
	setupAPITest(t)

	body := map[string]interface{}{
		"slug":        "test-slug",
		"destination": "https://example.com",
		"tags":        []string{"marketing"},
	}

	req := testutil.JSONRequest("POST", "/api/redirects", body)
	resp := httptest.NewRecorder()
	RedirectsHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusCreated)
	testutil.AssertFieldEquals(t, data, "slug", "test-slug")
	testutil.AssertFieldEquals(t, data, "destination", "https://example.com")
}

func TestRedirectsHandler_CreateDuplicateSlug(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	createTestRedirect(t, db, "dup-slug", "https://example.com")

	body := map[string]interface{}{
		"slug":        "dup-slug",
		"destination": "https://other.com",
	}

	req := testutil.JSONRequest("POST", "/api/redirects", body)
	resp := httptest.NewRecorder()
	RedirectsHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for duplicate slug, got %d", resp.Code)
	}
}

func TestRedirectsHandler_CreateMissingFields(t *testing.T) {
	setupAPITest(t)

	body := map[string]interface{}{
		"slug": "no-dest",
	}

	req := testutil.JSONRequest("POST", "/api/redirects", body)
	resp := httptest.NewRecorder()
	RedirectsHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing destination, got %d", resp.Code)
	}
}

func TestRedirectsHandler_ListWithData(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	createTestRedirect(t, db, "slug1", "https://a.com")
	createTestRedirect(t, db, "slug2", "https://b.com")

	req := httptest.NewRequest("GET", "/api/redirects", nil)
	resp := httptest.NewRecorder()
	RedirectsHandler(resp, req)

	arr := testutil.CheckSuccessArray(t, resp, http.StatusOK)
	if len(arr) != 2 {
		t.Errorf("Expected 2 redirects, got %d", len(arr))
	}
}

func TestRedirectsHandler_InvalidJSON(t *testing.T) {
	setupAPITest(t)

	req := httptest.NewRequest("POST", "/api/redirects", nil)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	RedirectsHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for empty body, got %d", resp.Code)
	}
}

func TestRedirectsHandler_MethodNotAllowed(t *testing.T) {
	setupAPITest(t)

	req := httptest.NewRequest("DELETE", "/api/redirects", nil)
	resp := httptest.NewRecorder()
	RedirectsHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

// --- WebhooksHandler ---

func TestWebhooksHandler_ListEmpty(t *testing.T) {
	setupAPITest(t)

	req := httptest.NewRequest("GET", "/api/webhooks", nil)
	resp := httptest.NewRecorder()
	WebhooksHandler(resp, req)

	arr := testutil.CheckSuccessArray(t, resp, http.StatusOK)
	if len(arr) != 0 {
		t.Errorf("Expected 0 webhooks, got %d", len(arr))
	}
}

func TestWebhooksHandler_Create(t *testing.T) {
	setupAPITest(t)

	body := map[string]interface{}{
		"name":     "test-webhook",
		"endpoint": "github-push",
		"secret":   "mysecret",
	}

	req := testutil.JSONRequest("POST", "/api/webhooks", body)
	resp := httptest.NewRecorder()
	WebhooksHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusCreated)
	testutil.AssertFieldEquals(t, data, "name", "test-webhook")
	testutil.AssertFieldEquals(t, data, "endpoint", "github-push")
	testutil.AssertFieldEquals(t, data, "has_secret", true)
	testutil.AssertFieldEquals(t, data, "is_active", true)
}

func TestWebhooksHandler_CreateWithoutSecret(t *testing.T) {
	setupAPITest(t)

	body := map[string]interface{}{
		"name":     "no-secret-hook",
		"endpoint": "slack-events",
	}

	req := testutil.JSONRequest("POST", "/api/webhooks", body)
	resp := httptest.NewRecorder()
	WebhooksHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusCreated)
	testutil.AssertFieldEquals(t, data, "has_secret", false)
}

func TestWebhooksHandler_CreateDuplicateEndpoint(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	createTestWebhook(t, db, "existing", "dup-endpoint")

	body := map[string]interface{}{
		"name":     "another",
		"endpoint": "dup-endpoint",
	}

	req := testutil.JSONRequest("POST", "/api/webhooks", body)
	resp := httptest.NewRecorder()
	WebhooksHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for duplicate endpoint, got %d", resp.Code)
	}
}

func TestWebhooksHandler_CreateMissingFields(t *testing.T) {
	setupAPITest(t)

	body := map[string]interface{}{
		"name": "no-endpoint",
	}

	req := testutil.JSONRequest("POST", "/api/webhooks", body)
	resp := httptest.NewRecorder()
	WebhooksHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing endpoint, got %d", resp.Code)
	}
}

func TestWebhooksHandler_ListWithData(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	createTestWebhook(t, db, "hook1", "ep1")
	createTestWebhook(t, db, "hook2", "ep2")

	req := httptest.NewRequest("GET", "/api/webhooks", nil)
	resp := httptest.NewRecorder()
	WebhooksHandler(resp, req)

	arr := testutil.CheckSuccessArray(t, resp, http.StatusOK)
	if len(arr) != 2 {
		t.Errorf("Expected 2 webhooks, got %d", len(arr))
	}
}

// --- parseInt ---

func TestParseInt(t *testing.T) {
	tests := []struct {
		input      string
		defaultVal int
		expected   int
	}{
		{"", 50, 50},
		{"10", 50, 10},
		{"abc", 50, 50},
		{"0", 50, 0},
		{"-1", 50, -1},
	}

	for _, tt := range tests {
		result := parseInt(tt.input, tt.defaultVal)
		if result != tt.expected {
			t.Errorf("parseInt(%q, %d) = %d, want %d", tt.input, tt.defaultVal, result, tt.expected)
		}
	}
}

// --- DeleteRedirectHandler ---

func TestDeleteRedirectHandler_Success(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	id := createTestRedirect(t, db, "del-slug", "https://example.com")

	req := httptest.NewRequest("DELETE", "/api/redirects/1", nil)
	req.SetPathValue("id", fmt.Sprintf("%d", id))
	resp := httptest.NewRecorder()
	DeleteRedirectHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "message", "Redirect deleted")
}

func TestDeleteRedirectHandler_NotFound(t *testing.T) {
	setupAPITest(t)

	req := httptest.NewRequest("DELETE", "/api/redirects/999", nil)
	req.SetPathValue("id", "999")
	resp := httptest.NewRecorder()
	DeleteRedirectHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusNotFound, "REDIRECT_NOT_FOUND")
}

func TestDeleteRedirectHandler_MissingID(t *testing.T) {
	setupAPITest(t)

	req := httptest.NewRequest("DELETE", "/api/redirects/", nil)
	resp := httptest.NewRecorder()
	DeleteRedirectHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestDeleteRedirectHandler_InvalidID(t *testing.T) {
	setupAPITest(t)

	req := httptest.NewRequest("DELETE", "/api/redirects/abc", nil)
	req.SetPathValue("id", "abc")
	resp := httptest.NewRecorder()
	DeleteRedirectHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

// --- Webhook CRUD (webhooks.go) ---

func TestDeleteWebhookHandler_Success(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	id := createTestWebhook(t, db, "del-hook", "del-endpoint")

	req := httptest.NewRequest("DELETE", "/api/webhooks/1", nil)
	req.SetPathValue("id", fmt.Sprintf("%d", id))
	resp := httptest.NewRecorder()
	DeleteWebhookHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "message", "Webhook deleted")
}

func TestDeleteWebhookHandler_NotFound(t *testing.T) {
	setupAPITest(t)

	req := httptest.NewRequest("DELETE", "/api/webhooks/999", nil)
	req.SetPathValue("id", "999")
	resp := httptest.NewRecorder()
	DeleteWebhookHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusNotFound, "WEBHOOK_NOT_FOUND")
}

func TestDeleteWebhookHandler_InvalidID(t *testing.T) {
	setupAPITest(t)

	req := httptest.NewRequest("DELETE", "/api/webhooks/abc", nil)
	req.SetPathValue("id", "abc")
	resp := httptest.NewRecorder()
	DeleteWebhookHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

func TestUpdateWebhookHandler_Success(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	id := createTestWebhook(t, db, "upd-hook", "upd-endpoint")

	newName := "updated-hook"
	body := map[string]interface{}{
		"name": newName,
	}

	req := testutil.JSONRequest("PUT", "/api/webhooks/1", body)
	req.SetPathValue("id", fmt.Sprintf("%d", id))
	resp := httptest.NewRecorder()
	UpdateWebhookHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "name", newName)
}

func TestUpdateWebhookHandler_ToggleActive(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	id := createTestWebhook(t, db, "toggle-hook", "toggle-endpoint")

	body := map[string]interface{}{
		"is_active": false,
	}

	req := testutil.JSONRequest("PUT", "/api/webhooks/1", body)
	req.SetPathValue("id", fmt.Sprintf("%d", id))
	resp := httptest.NewRecorder()
	UpdateWebhookHandler(resp, req)

	data := testutil.CheckSuccess(t, resp, http.StatusOK)
	testutil.AssertFieldEquals(t, data, "is_active", false)
}

func TestUpdateWebhookHandler_EndpointConflict(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	createTestWebhook(t, db, "hook1", "endpoint-a")
	id2 := createTestWebhook(t, db, "hook2", "endpoint-b")

	body := map[string]interface{}{
		"endpoint": "endpoint-a",
	}

	req := testutil.JSONRequest("PUT", "/api/webhooks/2", body)
	req.SetPathValue("id", fmt.Sprintf("%d", id2))
	resp := httptest.NewRecorder()
	UpdateWebhookHandler(resp, req)

	if resp.Code != http.StatusConflict {
		t.Errorf("Expected 409 conflict, got %d. Body: %s", resp.Code, resp.Body.String())
	}
}

func TestUpdateWebhookHandler_NotFound(t *testing.T) {
	setupAPITest(t)

	body := map[string]interface{}{
		"name": "ghost",
	}

	req := testutil.JSONRequest("PUT", "/api/webhooks/999", body)
	req.SetPathValue("id", "999")
	resp := httptest.NewRecorder()
	UpdateWebhookHandler(resp, req)

	testutil.CheckError(t, resp, http.StatusNotFound, "WEBHOOK_NOT_FOUND")
}

func TestUpdateWebhookHandler_InvalidJSON(t *testing.T) {
	setupAPITest(t)
	db := database.GetDB()
	id := createTestWebhook(t, db, "bad-json-hook", "bad-json-ep")

	req := httptest.NewRequest("PUT", "/api/webhooks/1", nil)
	req.SetPathValue("id", fmt.Sprintf("%d", id))
	resp := httptest.NewRecorder()
	UpdateWebhookHandler(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", resp.Code)
	}
}

