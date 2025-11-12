package audit

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

var db *sql.DB

// Init initializes the audit logging system
func Init(database *sql.DB) error {
	db = database

	// Create audit logs table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS audit_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		username TEXT,
		ip_address TEXT,
		action TEXT NOT NULL,
		resource TEXT,
		result TEXT,
		details TEXT
	);

	CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_logs(timestamp);
	CREATE INDEX IF NOT EXISTS idx_audit_username ON audit_logs(username);
	CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create audit logs table: %w", err)
	}

	return nil
}

// LogEntry represents an audit log entry
type LogEntry struct {
	ID        int64
	Timestamp time.Time
	Username  string
	IPAddress string
	Action    string
	Resource  string
	Result    string
	Details   string
}

// Log writes an audit log entry
func Log(username, ipAddress, action, resource, result, details string) error {
	if db == nil {
		return fmt.Errorf("audit logging not initialized")
	}

	query := `
		INSERT INTO audit_logs (username, ip_address, action, resource, result, details)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := db.Exec(query, username, ipAddress, action, resource, result, details)
	if err != nil {
		log.Printf("Failed to write audit log: %v", err)
		return err
	}

	return nil
}

// LogSuccess logs a successful action
func LogSuccess(username, ipAddress, action, resource string) error {
	return Log(username, ipAddress, action, resource, "success", "")
}

// LogFailure logs a failed action
func LogFailure(username, ipAddress, action, resource, reason string) error {
	return Log(username, ipAddress, action, resource, "failure", reason)
}

// GetRecent retrieves recent audit log entries
func GetRecent(limit int) ([]LogEntry, error) {
	if db == nil {
		return nil, fmt.Errorf("audit logging not initialized")
	}

	query := `
		SELECT id, timestamp, username, ip_address, action, resource, result, details
		FROM audit_logs
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LogEntry
	for rows.Next() {
		var entry LogEntry
		err := rows.Scan(
			&entry.ID,
			&entry.Timestamp,
			&entry.Username,
			&entry.IPAddress,
			&entry.Action,
			&entry.Resource,
			&entry.Result,
			&entry.Details,
		)
		if err != nil {
			log.Printf("Error scanning audit log: %v", err)
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// Cleanup removes audit logs older than the specified number of days
func Cleanup(daysToKeep int) error {
	if db == nil {
		return fmt.Errorf("audit logging not initialized")
	}

	query := `DELETE FROM audit_logs WHERE timestamp < datetime('now', '-' || ? || ' days')`

	result, err := db.Exec(query, daysToKeep)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("Cleaned up %d old audit log entries", rowsAffected)
	}

	return nil
}
