package activity

import (
	"database/sql"
	"encoding/json"
	"strings"
	"time"
)

// LogEntry represents a stored activity log entry
type LogEntry struct {
	ID           int64                  `json:"id"`
	Timestamp    time.Time              `json:"timestamp"`
	ActorType    string                 `json:"actor_type"`
	ActorID      string                 `json:"actor_id,omitempty"`
	ActorIP      string                 `json:"actor_ip,omitempty"`
	ActorUA      string                 `json:"actor_ua,omitempty"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id,omitempty"`
	Action       string                 `json:"action"`
	Result       string                 `json:"result"`
	Weight       int                    `json:"weight"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// QueryParams defines filters for querying activity logs
// All filters can be used with any command (list, cleanup, export)
type QueryParams struct {
	// Weight filters
	MinWeight *int
	MaxWeight *int

	// Resource filters
	ResourceType string
	ResourceID   string
	AppID        string // Matches resources belonging to an app (app:X, X:key, X.domain/path)

	// Actor filters
	ActorType string // user/system/api_key/anonymous
	ActorID   string
	UserID    string // Alias for ActorType=user + ActorID

	// Action filter
	Action string
	Result string // success/failure

	// Time filters
	Since *time.Time
	Until *time.Time

	// Pagination
	Limit  int
	Offset int
}

// DefaultLimit is the default number of entries to return
const DefaultLimit = 20

// MaxLimit is the maximum number of entries that can be returned
const MaxLimit = 1000

// Stats contains activity log statistics
type Stats struct {
	TotalCount    int           `json:"total_count"`
	CountByWeight map[int]int   `json:"count_by_weight"`
	OldestEntry   *time.Time    `json:"oldest_entry,omitempty"`
	NewestEntry   *time.Time    `json:"newest_entry,omitempty"`
	SizeEstimate  int64         `json:"size_estimate_bytes"`
}

// Query retrieves activity logs with filtering and pagination
func Query(db *sql.DB, params QueryParams) ([]LogEntry, int, error) {
	// Set defaults
	if params.Limit <= 0 {
		params.Limit = DefaultLimit
	}
	if params.Limit > MaxLimit {
		params.Limit = MaxLimit
	}

	// Normalize UserID to ActorType+ActorID
	if params.UserID != "" {
		params.ActorType = ActorUser
		params.ActorID = params.UserID
	}

	// Build query
	query := `SELECT id, timestamp, actor_type, actor_id, actor_ip, actor_ua, resource_type, resource_id, action, result, weight, details FROM activity_log WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM activity_log WHERE 1=1`
	args := []interface{}{}

	// Weight filters
	if params.MinWeight != nil {
		query += " AND weight >= ?"
		countQuery += " AND weight >= ?"
		args = append(args, *params.MinWeight)
	}
	if params.MaxWeight != nil {
		query += " AND weight <= ?"
		countQuery += " AND weight <= ?"
		args = append(args, *params.MaxWeight)
	}

	// Resource filters
	if params.ResourceType != "" {
		query += " AND resource_type = ?"
		countQuery += " AND resource_type = ?"
		args = append(args, params.ResourceType)
	}
	if params.ResourceID != "" {
		query += " AND resource_id = ?"
		countQuery += " AND resource_id = ?"
		args = append(args, params.ResourceID)
	}

	// App filter: matches resources belonging to an app
	// - resource_type='app' AND resource_id=appID (deploys)
	// - resource_id starts with 'appID:' (KV, docs: appID:key)
	// - resource_id starts with 'appID.' (pages: appID.domain.com/path)
	if params.AppID != "" {
		appFilter := " AND ((resource_type = 'app' AND resource_id = ?) OR resource_id LIKE ? OR resource_id LIKE ?)"
		query += appFilter
		countQuery += appFilter
		args = append(args, params.AppID, params.AppID+":%", params.AppID+".%")
	}

	// Actor filters
	if params.ActorType != "" {
		query += " AND actor_type = ?"
		countQuery += " AND actor_type = ?"
		args = append(args, params.ActorType)
	}
	if params.ActorID != "" {
		query += " AND actor_id = ?"
		countQuery += " AND actor_id = ?"
		args = append(args, params.ActorID)
	}

	// Action/result filters
	if params.Action != "" {
		query += " AND action = ?"
		countQuery += " AND action = ?"
		args = append(args, params.Action)
	}
	if params.Result != "" {
		query += " AND result = ?"
		countQuery += " AND result = ?"
		args = append(args, params.Result)
	}

	// Time filters
	if params.Since != nil {
		query += " AND timestamp >= ?"
		countQuery += " AND timestamp >= ?"
		args = append(args, params.Since.Unix())
	}
	if params.Until != nil {
		query += " AND timestamp <= ?"
		countQuery += " AND timestamp <= ?"
		args = append(args, params.Until.Unix())
	}

	// Get total count
	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	if err := db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Add ordering and pagination
	query += " ORDER BY timestamp DESC LIMIT ? OFFSET ?"
	args = append(args, params.Limit, params.Offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []LogEntry
	for rows.Next() {
		var e LogEntry
		var ts int64
		var actorID, actorIP, actorUA, resourceID sql.NullString
		var details sql.NullString

		if err := rows.Scan(&e.ID, &ts, &e.ActorType, &actorID, &actorIP, &actorUA, &e.ResourceType, &resourceID, &e.Action, &e.Result, &e.Weight, &details); err != nil {
			continue
		}

		e.Timestamp = time.Unix(ts, 0)
		if actorID.Valid {
			e.ActorID = actorID.String
		}
		if actorIP.Valid {
			e.ActorIP = actorIP.String
		}
		if actorUA.Valid {
			e.ActorUA = actorUA.String
		}
		if resourceID.Valid {
			e.ResourceID = resourceID.String
		}
		if details.Valid && details.String != "" {
			json.Unmarshal([]byte(details.String), &e.Details)
		}

		entries = append(entries, e)
	}

	return entries, total, nil
}

// Cleanup removes activity logs based on filters
// Supports all QueryParams filters for targeted cleanup
func Cleanup(db *sql.DB, params QueryParams, dryRun bool) (int64, error) {
	// Normalize UserID
	if params.UserID != "" {
		params.ActorType = ActorUser
		params.ActorID = params.UserID
	}

	// Build WHERE clause
	where := "WHERE 1=1"
	args := []interface{}{}

	if params.MinWeight != nil {
		where += " AND weight >= ?"
		args = append(args, *params.MinWeight)
	}
	if params.MaxWeight != nil {
		where += " AND weight <= ?"
		args = append(args, *params.MaxWeight)
	}
	if params.ResourceType != "" {
		where += " AND resource_type = ?"
		args = append(args, params.ResourceType)
	}
	if params.ResourceID != "" {
		where += " AND resource_id = ?"
		args = append(args, params.ResourceID)
	}
	if params.AppID != "" {
		where += " AND ((resource_type = 'app' AND resource_id = ?) OR resource_id LIKE ? OR resource_id LIKE ?)"
		args = append(args, params.AppID, params.AppID+":%", params.AppID+".%")
	}
	if params.ActorType != "" {
		where += " AND actor_type = ?"
		args = append(args, params.ActorType)
	}
	if params.ActorID != "" {
		where += " AND actor_id = ?"
		args = append(args, params.ActorID)
	}
	if params.Action != "" {
		where += " AND action = ?"
		args = append(args, params.Action)
	}
	if params.Result != "" {
		where += " AND result = ?"
		args = append(args, params.Result)
	}
	if params.Since != nil {
		where += " AND timestamp >= ?"
		args = append(args, params.Since.Unix())
	}
	if params.Until != nil {
		where += " AND timestamp <= ?"
		args = append(args, params.Until.Unix())
	}

	if dryRun {
		var count int64
		err := db.QueryRow("SELECT COUNT(*) FROM activity_log "+where, args...).Scan(&count)
		return count, err
	}

	result, err := db.Exec("DELETE FROM activity_log "+where, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// CleanupByWeightAndAge is a convenience function for simple weight+age cleanup
func CleanupByWeightAndAge(db *sql.DB, maxWeight int, olderThan time.Duration, dryRun bool) (int64, error) {
	until := time.Now().Add(-olderThan)
	return Cleanup(db, QueryParams{
		MaxWeight: &maxWeight,
		Until:     &until,
	}, dryRun)
}

// GetStats retrieves activity log statistics (unfiltered)
func GetStats(db *sql.DB) (*Stats, error) {
	return GetStatsFiltered(db, QueryParams{})
}

// GetStatsFiltered retrieves activity log statistics with optional filters
func GetStatsFiltered(db *sql.DB, params QueryParams) (*Stats, error) {
	stats := &Stats{
		CountByWeight: make(map[int]int),
	}

	// Normalize UserID
	if params.UserID != "" {
		params.ActorType = ActorUser
		params.ActorID = params.UserID
	}

	// Build WHERE clause
	where := "WHERE 1=1"
	args := []interface{}{}

	if params.MinWeight != nil {
		where += " AND weight >= ?"
		args = append(args, *params.MinWeight)
	}
	if params.MaxWeight != nil {
		where += " AND weight <= ?"
		args = append(args, *params.MaxWeight)
	}
	if params.ResourceType != "" {
		where += " AND resource_type = ?"
		args = append(args, params.ResourceType)
	}
	if params.ResourceID != "" {
		where += " AND resource_id = ?"
		args = append(args, params.ResourceID)
	}
	if params.AppID != "" {
		where += " AND ((resource_type = 'app' AND resource_id = ?) OR resource_id LIKE ? OR resource_id LIKE ?)"
		args = append(args, params.AppID, params.AppID+":%", params.AppID+".%")
	}
	if params.ActorType != "" {
		where += " AND actor_type = ?"
		args = append(args, params.ActorType)
	}
	if params.ActorID != "" {
		where += " AND actor_id = ?"
		args = append(args, params.ActorID)
	}
	if params.Action != "" {
		where += " AND action = ?"
		args = append(args, params.Action)
	}
	if params.Result != "" {
		where += " AND result = ?"
		args = append(args, params.Result)
	}
	if params.Since != nil {
		where += " AND timestamp >= ?"
		args = append(args, params.Since.Unix())
	}
	if params.Until != nil {
		where += " AND timestamp <= ?"
		args = append(args, params.Until.Unix())
	}

	// Total count
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	if err := db.QueryRow("SELECT COUNT(*) FROM activity_log "+where, countArgs...).Scan(&stats.TotalCount); err != nil {
		return nil, err
	}

	// Count by weight
	weightArgs := make([]interface{}, len(args))
	copy(weightArgs, args)
	rows, err := db.Query("SELECT weight, COUNT(*) FROM activity_log "+where+" GROUP BY weight ORDER BY weight", weightArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var weight, count int
		if err := rows.Scan(&weight, &count); err != nil {
			continue
		}
		stats.CountByWeight[weight] = count
	}

	// Oldest and newest within filter
	timeArgs := make([]interface{}, len(args))
	copy(timeArgs, args)
	var oldest, newest sql.NullInt64
	db.QueryRow("SELECT MIN(timestamp), MAX(timestamp) FROM activity_log "+where, timeArgs...).Scan(&oldest, &newest)
	if oldest.Valid {
		t := time.Unix(oldest.Int64, 0)
		stats.OldestEntry = &t
	}
	if newest.Valid {
		t := time.Unix(newest.Int64, 0)
		stats.NewestEntry = &t
	}

	// Size estimate (rough - based on matching row count)
	stats.SizeEstimate = int64(stats.TotalCount * 200)

	return stats, nil
}

// GetRecent retrieves the most recent activity log entries
func GetRecent(db *sql.DB, limit int) ([]LogEntry, error) {
	entries, _, err := Query(db, QueryParams{Limit: limit})
	return entries, err
}

// GetByResource retrieves activity for a specific resource
func GetByResource(db *sql.DB, resourceType, resourceID string, limit int) ([]LogEntry, error) {
	entries, _, err := Query(db, QueryParams{
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Limit:        limit,
	})
	return entries, err
}

// GetByApp retrieves activity for a specific app
func GetByApp(db *sql.DB, appID string, limit int) ([]LogEntry, error) {
	entries, _, err := Query(db, QueryParams{
		AppID: appID,
		Limit: limit,
	})
	return entries, err
}

// GetByUser retrieves activity for a specific user
func GetByUser(db *sql.DB, userID string, limit int) ([]LogEntry, error) {
	entries, _, err := Query(db, QueryParams{
		UserID: userID,
		Limit:  limit,
	})
	return entries, err
}

// DescribeFilters returns a human-readable description of active filters
func DescribeFilters(params QueryParams) string {
	var parts []string

	if params.MinWeight != nil {
		parts = append(parts, "weight >= "+string(rune('0'+*params.MinWeight)))
	}
	if params.MaxWeight != nil {
		parts = append(parts, "weight <= "+string(rune('0'+*params.MaxWeight)))
	}
	if params.AppID != "" {
		parts = append(parts, "app="+params.AppID)
	}
	if params.UserID != "" {
		parts = append(parts, "user="+params.UserID)
	} else if params.ActorID != "" {
		parts = append(parts, "actor="+params.ActorID)
	}
	if params.ActorType != "" && params.UserID == "" {
		parts = append(parts, "actor_type="+params.ActorType)
	}
	if params.ResourceType != "" {
		parts = append(parts, "type="+params.ResourceType)
	}
	if params.ResourceID != "" {
		parts = append(parts, "resource="+params.ResourceID)
	}
	if params.Action != "" {
		parts = append(parts, "action="+params.Action)
	}
	if params.Result != "" {
		parts = append(parts, "result="+params.Result)
	}
	if params.Since != nil {
		parts = append(parts, "since="+params.Since.Format("2006-01-02"))
	}
	if params.Until != nil {
		parts = append(parts, "until="+params.Until.Format("2006-01-02"))
	}

	if len(parts) == 0 {
		return "all entries"
	}
	return strings.Join(parts, ", ")
}
