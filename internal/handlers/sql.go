package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fazt-sh/fazt/internal/database"
)

// SQLRequest represents a SQL query request
type SQLRequest struct {
	Query string `json:"query"`
	Write bool   `json:"write"`
	Limit int    `json:"limit"`
}

// SQLResponse represents a SQL query response for SELECT queries
type SQLResponse struct {
	Columns []string        `json:"columns"`
	Rows    [][]interface{} `json:"rows"`
	Count   int             `json:"count"`
	TimeMS  int64           `json:"time_ms"`
}

// SQLWriteResponse represents a response for write operations
type SQLWriteResponse struct {
	Affected int64 `json:"affected"`
	TimeMS   int64 `json:"time_ms"`
}

// HandleSQL executes SQL queries against the database
func HandleSQL(w http.ResponseWriter, r *http.Request) {
	// Require API key auth (bypasses AdminMiddleware for remote peer access)
	if !requireAPIKeyAuth(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req SQLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Query == "" {
		http.Error(w, "Query required", http.StatusBadRequest)
		return
	}

	// Default limit
	if req.Limit == 0 {
		req.Limit = 100
	}

	// Check if query is a mutation
	isMutation := isWriteQuery(req.Query)
	if isMutation && !req.Write {
		http.Error(w, "Write operations require write: true", http.StatusBadRequest)
		return
	}

	// Get database
	db := database.GetDB()
	if db == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	// Execute query
	start := time.Now()

	if isMutation {
		result, err := db.Exec(req.Query)
		if err != nil {
			http.Error(w, fmt.Sprintf("Query error: %v", err), http.StatusBadRequest)
			return
		}

		affected, _ := result.RowsAffected()
		elapsed := time.Since(start)

		response := SQLWriteResponse{
			Affected: affected,
			TimeMS:   elapsed.Milliseconds(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// SELECT query
	rows, err := db.Query(req.Query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query error: %v", err), http.StatusBadRequest)
		return
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, fmt.Sprintf("Column error: %v", err), http.StatusInternalServerError)
		return
	}

	// Read all rows
	var results [][]interface{}
	for rows.Next() {
		if len(results) >= req.Limit {
			break
		}

		// Create slice of interface{} to hold each column value
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}

		// Convert byte arrays to strings for JSON serialization
		row := make([]interface{}, len(columns))
		for i, val := range values {
			if b, ok := val.([]byte); ok {
				row[i] = string(b)
			} else {
				row[i] = val
			}
		}
		results = append(results, row)
	}

	elapsed := time.Since(start)

	// Build response
	response := SQLResponse{
		Columns: columns,
		Rows:    results,
		Count:   len(results),
		TimeMS:  elapsed.Milliseconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// isWriteQuery checks if a query is a write operation
func isWriteQuery(query string) bool {
	q := strings.TrimSpace(strings.ToUpper(query))
	return strings.HasPrefix(q, "INSERT") ||
		strings.HasPrefix(q, "UPDATE") ||
		strings.HasPrefix(q, "DELETE") ||
		strings.HasPrefix(q, "DROP") ||
		strings.HasPrefix(q, "CREATE") ||
		strings.HasPrefix(q, "ALTER")
}
