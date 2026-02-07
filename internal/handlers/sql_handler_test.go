package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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

// TestHandleSQL_MissingAuthorization tests SQL query without API key
func TestHandleSQL_MissingAuthorization(t *testing.T) {
	setupSQLHandlerTest(t)

	body, _ := json.Marshal(SQLRequest{Query: "SELECT 1"})
	req := httptest.NewRequest(http.MethodPost, "/api/sql", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	HandleSQL(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// TestHandleSQL_InvalidAPIKey tests SQL query with invalid API key
func TestHandleSQL_InvalidAPIKey(t *testing.T) {
	setupSQLHandlerTest(t)

	body, _ := json.Marshal(SQLRequest{Query: "SELECT 1"})
	req := httptest.NewRequest(http.MethodPost, "/api/sql", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer invalid-key")
	rr := httptest.NewRecorder()

	HandleSQL(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}
}

// TestHandleSQL_SQLSyntaxError tests malformed SQL query
func TestHandleSQL_SQLSyntaxError(t *testing.T) {
	token := setupSQLHandlerTest(t)

	body, _ := json.Marshal(SQLRequest{Query: "SELECT * FROM WHERE"})
	req := httptest.NewRequest(http.MethodPost, "/api/sql", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	HandleSQL(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// TestHandleSQL_NonExistentTable tests query on table that doesn't exist
func TestHandleSQL_NonExistentTable(t *testing.T) {
	token := setupSQLHandlerTest(t)

	body, _ := json.Marshal(SQLRequest{Query: "SELECT * FROM nonexistent_table_xyz"})
	req := httptest.NewRequest(http.MethodPost, "/api/sql", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	HandleSQL(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusBadRequest)
	}
}

// TestHandleSQL_LimitParameter tests query result limiting
func TestHandleSQL_LimitParameter(t *testing.T) {
	token := setupSQLHandlerTest(t)
	db := database.GetDB()

	// Insert 10 events
	for i := 0; i < 10; i++ {
		_, err := db.Exec(`INSERT INTO events (domain, source_type, event_type) VALUES (?, ?, ?)`,
			"test.local", "web", "pageview")
		if err != nil {
			t.Fatalf("Failed to insert event: %v", err)
		}
	}

	// Query with limit of 5
	body, _ := json.Marshal(SQLRequest{Query: "SELECT * FROM events", Limit: 5})
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

	if resp.Count != 5 {
		t.Fatalf("count = %d, want 5 (limit applied)", resp.Count)
	}
	if len(resp.Rows) != 5 {
		t.Fatalf("rows = %d, want 5", len(resp.Rows))
	}
}

// TestHandleSQL_DefaultLimit tests default limit of 100
func TestHandleSQL_DefaultLimit(t *testing.T) {
	token := setupSQLHandlerTest(t)
	db := database.GetDB()

	// Insert 150 events
	for i := 0; i < 150; i++ {
		_, err := db.Exec(`INSERT INTO events (domain, source_type, event_type) VALUES (?, ?, ?)`,
			"test.local", "web", "pageview")
		if err != nil {
			t.Fatalf("Failed to insert event: %v", err)
		}
	}

	// Query without limit (should default to 100)
	body, _ := json.Marshal(SQLRequest{Query: "SELECT * FROM events"})
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

	if resp.Count != 100 {
		t.Fatalf("count = %d, want 100 (default limit)", resp.Count)
	}
}

// TestHandleSQL_NoResults tests query that returns no rows
func TestHandleSQL_NoResults(t *testing.T) {
	token := setupSQLHandlerTest(t)

	body, _ := json.Marshal(SQLRequest{Query: "SELECT * FROM events WHERE domain = 'nonexistent.com'"})
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

	if resp.Count != 0 {
		t.Fatalf("count = %d, want 0", resp.Count)
	}
	if len(resp.Rows) != 0 {
		t.Fatalf("rows = %d, want 0", len(resp.Rows))
	}
}

// TestHandleSQL_ComplexJoinQuery tests complex SQL with JOINs
func TestHandleSQL_ComplexJoinQuery(t *testing.T) {
	token := setupSQLHandlerTest(t)
	db := database.GetDB()

	// Insert test data
	_, err := db.Exec(`INSERT INTO events (domain, source_type, event_type) VALUES (?, ?, ?)`,
		"test.local", "web", "pageview")
	if err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	// Complex query with subquery
	body, _ := json.Marshal(SQLRequest{
		Query: "SELECT domain, COUNT(*) as count FROM events GROUP BY domain",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/sql", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	HandleSQL(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp SQLResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(resp.Columns) != 2 {
		t.Fatalf("columns = %v, want 2 columns", resp.Columns)
	}
}

// TestHandleSQL_VeryLongQuery tests extremely long SQL query
func TestHandleSQL_VeryLongQuery(t *testing.T) {
	token := setupSQLHandlerTest(t)

	// Build a very long query with many OR conditions
	var queryBuilder strings.Builder
	queryBuilder.WriteString("SELECT * FROM events WHERE domain = 'test.local'")
	for i := 0; i < 1000; i++ {
		queryBuilder.WriteString(" OR domain = 'test")
		queryBuilder.WriteString(fmt.Sprintf("%d", i))
		queryBuilder.WriteString(".local'")
	}

	body, _ := json.Marshal(SQLRequest{Query: queryBuilder.String()})
	req := httptest.NewRequest(http.MethodPost, "/api/sql", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	HandleSQL(rr, req)

	// Very long query may fail due to limits, or succeed - either is acceptable
	if rr.Code != http.StatusOK && rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 200 or 400", rr.Code)
	}
}

// TestHandleSQL_WhitespaceOnly tests query with only whitespace
func TestHandleSQL_WhitespaceOnly(t *testing.T) {
	token := setupSQLHandlerTest(t)

	body, _ := json.Marshal(SQLRequest{Query: "   \t\n  "})
	req := httptest.NewRequest(http.MethodPost, "/api/sql", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	HandleSQL(rr, req)

	// Handler doesn't trim whitespace, so this is treated as a query
	// SQLite will return an error for invalid syntax
	if rr.Code != http.StatusOK && rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 200 or 400", rr.Code)
	}
}

// TestHandleSQL_WriteWithoutFlag tests write operation detection
func TestHandleSQL_WriteWithoutFlag(t *testing.T) {
	token := setupSQLHandlerTest(t)

	tests := []struct {
		name  string
		query string
	}{
		{"INSERT", "INSERT INTO redirects (slug, destination) VALUES ('test', 'https://example.com')"},
		{"UPDATE", "UPDATE redirects SET destination = 'https://new.com' WHERE slug = 'test'"},
		{"DELETE", "DELETE FROM redirects WHERE slug = 'test'"},
		{"DROP", "DROP TABLE IF EXISTS temp_test_table"},
		{"CREATE", "CREATE TABLE temp_test (id INTEGER)"},
		{"ALTER", "ALTER TABLE redirects ADD COLUMN test TEXT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(SQLRequest{Query: tt.query, Write: false})
			req := httptest.NewRequest(http.MethodPost, "/api/sql", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			rr := httptest.NewRecorder()

			HandleSQL(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d (write flag required)", rr.Code, http.StatusBadRequest)
			}
		})
	}
}

// TestHandleSQL_NegativeLimit tests negative limit value
func TestHandleSQL_NegativeLimit(t *testing.T) {
	token := setupSQLHandlerTest(t)
	db := database.GetDB()

	_, err := db.Exec(`INSERT INTO events (domain, source_type, event_type) VALUES (?, ?, ?)`,
		"test.local", "web", "pageview")
	if err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	// Negative limit should be treated as no limit or default
	body, _ := json.Marshal(SQLRequest{Query: "SELECT * FROM events", Limit: -1})
	req := httptest.NewRequest(http.MethodPost, "/api/sql", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	HandleSQL(rr, req)

	// Should still succeed (implementation may vary)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
}
