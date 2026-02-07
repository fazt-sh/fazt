package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fazt-sh/fazt/internal/database"
)

func setupSQLHandlerTest(t *testing.T) string {
	t.Helper()

	silenceTestLogs(t)

	db := setupTestDB(t)
	database.SetDB(db)
	t.Cleanup(func() {
		db.Close()
		database.SetDB(nil)
	})

	token := "sql-token-123"
	insertTestAPIKey(t, db, token)
	return token
}

func TestHandleSQL_MethodNotAllowed(t *testing.T) {
	token := setupSQLHandlerTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sql", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	HandleSQL(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleSQL_InvalidJSON(t *testing.T) {
	token := setupSQLHandlerTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/sql", bytes.NewBufferString("{"))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	HandleSQL(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestHandleSQL_EmptyQuery(t *testing.T) {
	token := setupSQLHandlerTest(t)

	body, _ := json.Marshal(SQLRequest{Query: ""})
	req := httptest.NewRequest(http.MethodPost, "/api/sql", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	HandleSQL(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestHandleSQL_WriteRequiresFlag(t *testing.T) {
	token := setupSQLHandlerTest(t)

	body, _ := json.Marshal(SQLRequest{Query: "CREATE TABLE test_write (id INTEGER)"})
	req := httptest.NewRequest(http.MethodPost, "/api/sql", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	HandleSQL(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

func TestHandleSQL_SelectSuccess(t *testing.T) {
	token := setupSQLHandlerTest(t)

	db := database.GetDB()
	_, err := db.Exec(`INSERT INTO events (domain, source_type, event_type) VALUES (?, ?, ?)`, "test.local", "web", "pageview")
	if err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	body, _ := json.Marshal(SQLRequest{Query: "SELECT domain FROM events", Limit: 10})
	req := httptest.NewRequest(http.MethodPost, "/api/sql", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	HandleSQL(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var resp SQLResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(resp.Columns) != 1 || resp.Columns[0] != "domain" {
		t.Fatalf("columns = %v, want [domain]", resp.Columns)
	}
	if resp.Count != 1 {
		t.Fatalf("count = %d, want 1", resp.Count)
	}
	if len(resp.Rows) != 1 || len(resp.Rows[0]) != 1 {
		t.Fatalf("rows = %v, want 1 row", resp.Rows)
	}
}

func TestHandleSQL_WriteSuccess(t *testing.T) {
	token := setupSQLHandlerTest(t)

	body, _ := json.Marshal(SQLRequest{
		Query: "INSERT INTO redirects (slug, destination) VALUES ('a', 'https://example.com')",
		Write: true,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/sql", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	HandleSQL(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}

	var resp SQLWriteResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if resp.Affected != 1 {
		t.Fatalf("affected = %d, want 1", resp.Affected)
	}
}
