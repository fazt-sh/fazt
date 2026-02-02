package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fazt-sh/fazt/internal/database"
	"github.com/fazt-sh/fazt/internal/output"
)

// handleSQLCommand executes SQL queries against the local database
func handleSQLCommand(args []string) {
	fs := flag.NewFlagSet("sql", flag.ExitOnError)
	dbPath := fs.String("db", "", "Database path (default: configured database)")
	write := fs.Bool("write", false, "Allow write operations (INSERT, UPDATE, DELETE)")
	limit := fs.Int("limit", 100, "Maximum rows to return")

	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: SQL query required")
		fmt.Fprintln(os.Stderr, "Usage: fazt sql \"SELECT * FROM apps\" [--write] [--limit N]")
		os.Exit(1)
	}

	query := fs.Arg(0)

	// Check if query is a mutation
	isMutation := isWriteQuery(query)
	if isMutation && !*write {
		fmt.Fprintln(os.Stderr, "Error: Write operations require --write flag")
		fmt.Fprintln(os.Stderr, "Query appears to be: INSERT, UPDATE, DELETE, or DROP")
		os.Exit(1)
	}

	// Open database
	var dbPathResolved string
	if *dbPath != "" {
		dbPathResolved = *dbPath
	} else {
		dbPathResolved = database.ResolvePath("")
	}

	if err := database.Init(dbPathResolved); err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	db := database.GetDB()
	defer database.Close()

	// Execute query
	start := time.Now()

	if isMutation {
		result, err := db.Exec(query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error executing query: %v\n", err)
			os.Exit(1)
		}

		affected, _ := result.RowsAffected()
		elapsed := time.Since(start)

		fmt.Printf("%d row(s) affected (%dms)\n", affected, elapsed.Milliseconds())
		return
	}

	// SELECT query
	rows, err := db.Query(query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing query: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting columns: %v\n", err)
		os.Exit(1)
	}

	// Read all rows
	var results [][]string
	for rows.Next() {
		if len(results) >= *limit {
			break
		}

		// Create slice of interface{} to hold each column value
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			fmt.Fprintf(os.Stderr, "Error scanning row: %v\n", err)
			os.Exit(1)
		}

		// Convert values to strings
		row := make([]string, len(columns))
		for i, val := range values {
			row[i] = valueToString(val)
		}
		results = append(results, row)
	}

	elapsed := time.Since(start)

	// Format output
	renderer := getRenderer()

	// Prepare data for JSON
	data := map[string]interface{}{
		"columns": columns,
		"rows":    results,
		"count":   len(results),
		"time_ms": elapsed.Milliseconds(),
	}

	// Build markdown table
	table := &output.Table{
		Headers: columns,
		Rows:    results,
	}

	md := output.NewMarkdown().
		H1("Query Results").
		Table(table).
		Para(fmt.Sprintf("%d rows (%dms)", len(results), elapsed.Milliseconds())).
		String()

	renderer.Print(md, data)
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

// valueToString converts a database value to string
func valueToString(val interface{}) string {
	if val == nil {
		return "NULL"
	}

	switch v := val.(type) {
	case []byte:
		return string(v)
	case int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%f", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}
