package scanner

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLScanner implements Scanner for MySQL databases.
type MySQLScanner struct {
	db *sql.DB
}

// NewMySQLScanner connects to a MySQL database.
// DSN format: user:password@tcp(localhost:3306)/dbname
func NewMySQLScanner(ctx context.Context, dsn string) (*MySQLScanner, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &MySQLScanner{db: db}, nil
}

// ScanSchema queries information_schema.columns for MySQL.
func (m *MySQLScanner) ScanSchema(ctx context.Context) ([]Column, error) {
	query := `
		SELECT 
			TABLE_NAME, 
			COLUMN_NAME, 
			COLUMN_TYPE, 
			COLUMN_COMMENT 
		FROM information_schema.columns 
		WHERE TABLE_SCHEMA = DATABASE();
	`

	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying schema: %w", err)
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		if err := rows.Scan(&col.Table, &col.Name, &col.DataType, &col.Comment); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		columns = append(columns, col)
	}

	return columns, rows.Err()
}

// SampleData pulls 100 rows for heuristic scanning.
func (m *MySQLScanner) SampleData(ctx context.Context, table string) ([]map[string]interface{}, error) {
	// Simple for MVP
	query := fmt.Sprintf("SELECT * FROM %s LIMIT 100", table)
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}

	for rows.Next() {
		// Create a slice of interface{}'s to represent each column,
		// and a second slice to contain pointers to each item in the columns slice.
		columnsData := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columnsData {
			columnPointers[i] = &columnsData[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			val := columnsData[i]

			// MySQL driver often returns []byte for strings
			b, ok := val.([]byte)
			if ok {
				rowMap[colName] = string(b)
			} else {
				rowMap[colName] = val
			}
		}
		results = append(results, rowMap)
	}

	return results, rows.Err()
}

// Close closes the database connection.
func (m *MySQLScanner) Close() {
	if m.db != nil {
		m.db.Close()
	}
}
