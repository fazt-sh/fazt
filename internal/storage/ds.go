package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SQLDocStore implements DocStore using SQLite.
type SQLDocStore struct {
	db *sql.DB
}

// NewSQLDocStore creates a new SQLite-backed document store.
func NewSQLDocStore(db *sql.DB) *SQLDocStore {
	return &SQLDocStore{db: db}
}

// Insert adds a new document to a collection.
func (s *SQLDocStore) Insert(ctx context.Context, appID, collection string, doc map[string]interface{}) (string, error) {
	// Generate ID if not provided
	id, ok := doc["id"].(string)
	if !ok || id == "" {
		id = uuid.New().String()
	}

	// Remove id from doc data (stored separately)
	docCopy := make(map[string]interface{})
	for k, v := range doc {
		if k != "id" {
			docCopy[k] = v
		}
	}

	dataJSON, err := json.Marshal(docCopy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal document: %w", err)
	}

	query := `
		INSERT INTO app_docs (app_id, collection, id, data, created_at, updated_at)
		VALUES (?, ?, ?, ?, strftime('%s', 'now'), strftime('%s', 'now'))
	`
	_, err = s.db.ExecContext(ctx, query, appID, collection, id, string(dataJSON))
	if err != nil {
		return "", fmt.Errorf("failed to insert document: %w", err)
	}

	return id, nil
}

// Find retrieves documents matching a query.
func (s *SQLDocStore) Find(ctx context.Context, appID, collection string, query map[string]interface{}) ([]Document, error) {
	qb := NewQueryBuilder()
	whereClause, args, err := qb.Build(query)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	// Prepend app_id and collection to args
	fullArgs := make([]interface{}, 0, len(args)+2)
	fullArgs = append(fullArgs, appID, collection)
	fullArgs = append(fullArgs, args...)

	sqlQuery := fmt.Sprintf(`
		SELECT id, data, created_at, updated_at FROM app_docs
		WHERE app_id = ? AND collection = ? AND %s
		ORDER BY created_at DESC
	`, whereClause)

	rows, err := s.db.QueryContext(ctx, sqlQuery, fullArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var id, dataJSON string
		var createdAt, updatedAt int64
		if err := rows.Scan(&id, &dataJSON, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
			return nil, fmt.Errorf("failed to unmarshal document: %w", err)
		}

		docs = append(docs, Document{
			ID:        id,
			Data:      data,
			CreatedAt: time.Unix(createdAt, 0),
			UpdatedAt: time.Unix(updatedAt, 0),
		})
	}

	return docs, nil
}

// FindOne retrieves a single document by ID.
func (s *SQLDocStore) FindOne(ctx context.Context, appID, collection, id string) (*Document, error) {
	query := `
		SELECT id, data, created_at, updated_at FROM app_docs
		WHERE app_id = ? AND collection = ? AND id = ?
	`
	var docID, dataJSON string
	var createdAt, updatedAt int64
	err := s.db.QueryRowContext(ctx, query, appID, collection, id).Scan(&docID, &dataJSON, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query document: %w", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(dataJSON), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal document: %w", err)
	}

	return &Document{
		ID:        docID,
		Data:      data,
		CreatedAt: time.Unix(createdAt, 0),
		UpdatedAt: time.Unix(updatedAt, 0),
	}, nil
}

// Update modifies documents matching a query.
func (s *SQLDocStore) Update(ctx context.Context, appID, collection string, query, changes map[string]interface{}) (int64, error) {
	// Build the query to find matching documents
	qb := NewQueryBuilder()
	whereClause, whereArgs, err := qb.Build(query)
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	// Build the update expression
	ub := NewUpdateBuilder()
	updateExpr, updateArgs, err := ub.Build("data", changes)
	if err != nil {
		return 0, fmt.Errorf("failed to build update: %w", err)
	}

	// Combine args: update args first, then where args
	allArgs := make([]interface{}, 0, len(updateArgs)+len(whereArgs)+2)
	allArgs = append(allArgs, updateArgs...)
	allArgs = append(allArgs, appID, collection)
	allArgs = append(allArgs, whereArgs...)

	sqlQuery := fmt.Sprintf(`
		UPDATE app_docs
		SET data = %s, updated_at = strftime('%%s', 'now')
		WHERE app_id = ? AND collection = ? AND %s
	`, updateExpr, whereClause)

	result, err := s.db.ExecContext(ctx, sqlQuery, allArgs...)
	if err != nil {
		return 0, fmt.Errorf("failed to update documents: %w", err)
	}

	return result.RowsAffected()
}

// Delete removes documents matching a query.
func (s *SQLDocStore) Delete(ctx context.Context, appID, collection string, query map[string]interface{}) (int64, error) {
	qb := NewQueryBuilder()
	whereClause, args, err := qb.Build(query)
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	// Prepend app_id and collection to args
	fullArgs := make([]interface{}, 0, len(args)+2)
	fullArgs = append(fullArgs, appID, collection)
	fullArgs = append(fullArgs, args...)

	sqlQuery := fmt.Sprintf(`
		DELETE FROM app_docs
		WHERE app_id = ? AND collection = ? AND %s
	`, whereClause)

	result, err := s.db.ExecContext(ctx, sqlQuery, fullArgs...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete documents: %w", err)
	}

	return result.RowsAffected()
}

// Count returns the number of documents matching a query.
func (s *SQLDocStore) Count(ctx context.Context, appID, collection string, query map[string]interface{}) (int64, error) {
	qb := NewQueryBuilder()
	whereClause, args, err := qb.Build(query)
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	// Prepend app_id and collection to args
	fullArgs := make([]interface{}, 0, len(args)+2)
	fullArgs = append(fullArgs, appID, collection)
	fullArgs = append(fullArgs, args...)

	sqlQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM app_docs
		WHERE app_id = ? AND collection = ? AND %s
	`, whereClause)

	var count int64
	err = s.db.QueryRowContext(ctx, sqlQuery, fullArgs...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count documents: %w", err)
	}

	return count, nil
}

// Collections returns all collection names for an app.
func (s *SQLDocStore) Collections(ctx context.Context, appID string) ([]string, error) {
	query := `
		SELECT DISTINCT collection FROM app_docs
		WHERE app_id = ?
		ORDER BY collection
	`
	rows, err := s.db.QueryContext(ctx, query, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to query collections: %w", err)
	}
	defer rows.Close()

	var collections []string
	for rows.Next() {
		var collection string
		if err := rows.Scan(&collection); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		collections = append(collections, collection)
	}

	return collections, nil
}
