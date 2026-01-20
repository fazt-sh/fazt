package storage

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Query operators
const (
	OpEq       = "$eq"
	OpNe       = "$ne"
	OpGt       = "$gt"
	OpLt       = "$lt"
	OpGte      = "$gte"
	OpLte      = "$lte"
	OpIn       = "$in"
	OpNin      = "$nin"
	OpContains = "$contains"
	OpSet      = "$set"
	OpUnset    = "$unset"
	OpInc      = "$inc"
)

// QueryBuilder converts MongoDB-style queries to SQL WHERE clauses.
type QueryBuilder struct {
	conditions []string
	args       []interface{}
}

// NewQueryBuilder creates a new query builder.
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		conditions: make([]string, 0),
		args:       make([]interface{}, 0),
	}
}

// Build converts a query map to a SQL WHERE clause and arguments.
func (qb *QueryBuilder) Build(query map[string]interface{}) (string, []interface{}, error) {
	for field, value := range query {
		if err := qb.parseField(field, value); err != nil {
			return "", nil, err
		}
	}

	if len(qb.conditions) == 0 {
		return "1=1", nil, nil
	}

	return strings.Join(qb.conditions, " AND "), qb.args, nil
}

func (qb *QueryBuilder) parseField(field string, value interface{}) error {
	// Handle reserved fields that are stored in columns, not JSON
	if field == "id" {
		qb.conditions = append(qb.conditions, "id = ?")
		qb.args = append(qb.args, value)
		return nil
	}

	// Check if value is an operator object
	if opMap, ok := value.(map[string]interface{}); ok {
		for op, opVal := range opMap {
			if strings.HasPrefix(op, "$") {
				cond, args, err := qb.buildOperator(field, op, opVal)
				if err != nil {
					return err
				}
				qb.conditions = append(qb.conditions, cond)
				qb.args = append(qb.args, args...)
			}
		}
		return nil
	}

	// Simple equality
	qb.conditions = append(qb.conditions, fmt.Sprintf("json_extract(data, '$.%s') = ?", escapeJSONPath(field)))
	qb.args = append(qb.args, value)
	return nil
}

func (qb *QueryBuilder) buildOperator(field, op string, value interface{}) (string, []interface{}, error) {
	// Handle reserved fields that are stored in columns, not JSON
	var fieldExpr string
	if field == "id" {
		fieldExpr = "id"
	} else {
		fieldExpr = fmt.Sprintf("json_extract(data, '$.%s')", escapeJSONPath(field))
	}
	jsonPath := fieldExpr

	switch op {
	case OpEq:
		return fmt.Sprintf("%s = ?", jsonPath), []interface{}{value}, nil

	case OpNe:
		return fmt.Sprintf("(%s IS NULL OR %s != ?)", jsonPath, jsonPath), []interface{}{value}, nil

	case OpGt:
		return fmt.Sprintf("%s > ?", jsonPath), []interface{}{value}, nil

	case OpLt:
		return fmt.Sprintf("%s < ?", jsonPath), []interface{}{value}, nil

	case OpGte:
		return fmt.Sprintf("%s >= ?", jsonPath), []interface{}{value}, nil

	case OpLte:
		return fmt.Sprintf("%s <= ?", jsonPath), []interface{}{value}, nil

	case OpIn:
		arr, ok := value.([]interface{})
		if !ok {
			return "", nil, fmt.Errorf("$in requires an array")
		}
		if len(arr) == 0 {
			return "0=1", nil, nil // Empty $in matches nothing
		}
		placeholders := make([]string, len(arr))
		for i := range arr {
			placeholders[i] = "?"
		}
		return fmt.Sprintf("%s IN (%s)", jsonPath, strings.Join(placeholders, ",")), arr, nil

	case OpNin:
		arr, ok := value.([]interface{})
		if !ok {
			return "", nil, fmt.Errorf("$nin requires an array")
		}
		if len(arr) == 0 {
			return "1=1", nil, nil // Empty $nin matches everything
		}
		placeholders := make([]string, len(arr))
		for i := range arr {
			placeholders[i] = "?"
		}
		return fmt.Sprintf("(%s IS NULL OR %s NOT IN (%s))", jsonPath, jsonPath, strings.Join(placeholders, ",")), arr, nil

	case OpContains:
		// For array contains, use json_each to check if value exists in array
		return fmt.Sprintf("EXISTS (SELECT 1 FROM json_each(%s) WHERE value = ?)", jsonPath), []interface{}{value}, nil

	default:
		return "", nil, fmt.Errorf("unknown operator: %s", op)
	}
}

// UpdateBuilder converts MongoDB-style update operations to SQL.
type UpdateBuilder struct {
	sets []string
	args []interface{}
}

// NewUpdateBuilder creates a new update builder.
func NewUpdateBuilder() *UpdateBuilder {
	return &UpdateBuilder{
		sets: make([]string, 0),
		args: make([]interface{}, 0),
	}
}

// Build converts update operations to SQL SET clause parts.
// Returns the modified JSON expression and arguments.
func (ub *UpdateBuilder) Build(currentData string, changes map[string]interface{}) (string, []interface{}, error) {
	result := currentData

	for op, value := range changes {
		switch op {
		case OpSet:
			fields, ok := value.(map[string]interface{})
			if !ok {
				return "", nil, fmt.Errorf("$set requires an object")
			}
			for field, val := range fields {
				result = fmt.Sprintf("json_set(%s, '$.%s', json(?))", result, escapeJSONPath(field))
				jsonVal, err := marshalJSONValue(val)
				if err != nil {
					return "", nil, err
				}
				ub.args = append(ub.args, jsonVal)
			}

		case OpUnset:
			fields, ok := value.(map[string]interface{})
			if !ok {
				return "", nil, fmt.Errorf("$unset requires an object")
			}
			for field := range fields {
				result = fmt.Sprintf("json_remove(%s, '$.%s')", result, escapeJSONPath(field))
			}

		case OpInc:
			fields, ok := value.(map[string]interface{})
			if !ok {
				return "", nil, fmt.Errorf("$inc requires an object")
			}
			for field, incVal := range fields {
				// json_set with COALESCE to handle missing fields
				jsonPath := fmt.Sprintf("$.%s", escapeJSONPath(field))
				result = fmt.Sprintf("json_set(%s, '%s', COALESCE(json_extract(%s, '%s'), 0) + ?)",
					result, jsonPath, result, jsonPath)
				ub.args = append(ub.args, incVal)
			}

		default:
			// If not an operator, treat as direct $set
			if !strings.HasPrefix(op, "$") {
				result = fmt.Sprintf("json_set(%s, '$.%s', json(?))", result, escapeJSONPath(op))
				jsonVal, err := marshalJSONValue(value)
				if err != nil {
					return "", nil, err
				}
				ub.args = append(ub.args, jsonVal)
			}
		}
	}

	return result, ub.args, nil
}

// escapeJSONPath escapes a field name for use in JSON path expressions.
func escapeJSONPath(field string) string {
	// Handle dots in field names by quoting
	if strings.Contains(field, ".") {
		parts := strings.Split(field, ".")
		for i, part := range parts {
			parts[i] = part
		}
		return strings.Join(parts, ".")
	}
	return field
}

// marshalJSONValue converts a value to a JSON string for json_set.
func marshalJSONValue(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
