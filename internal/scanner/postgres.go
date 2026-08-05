package scanner

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresScanner implements Scanner for PostgreSQL databases.
type PostgresScanner struct {
	pool *pgxpool.Pool
}

// NewPostgresScanner connects to a PostgreSQL database.
func NewPostgresScanner(ctx context.Context, dbURL string) (*PostgresScanner, error) {
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &PostgresScanner{pool: pool}, nil
}

// ScanSchema queries pg_class, pg_attribute, and pg_description.
func (p *PostgresScanner) ScanSchema(ctx context.Context) ([]Column, error) {
	query := `
		SELECT 
			c.relname AS table_name,
			a.attname AS column_name,
			COALESCE(d.description, '') AS pii_tag,
			t.typname AS data_type
		FROM pg_class c
		JOIN pg_attribute a ON c.oid = a.attrelid
		JOIN pg_type t ON a.atttypid = t.oid
		LEFT JOIN pg_description d ON c.oid = d.objoid AND a.attnum = d.objsubid
		WHERE c.relkind = 'r' AND a.attnum > 0 AND c.relnamespace = 'public'::regnamespace;
	`

	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying schema: %w", err)
	}
	defer rows.Close()

	var columns []Column
	for rows.Next() {
		var col Column
		if err := rows.Scan(&col.Table, &col.Name, &col.Comment, &col.DataType); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}

// SampleData pulls 100 rows for heuristic scanning.
func (p *PostgresScanner) SampleData(ctx context.Context, table string) ([]map[string]interface{}, error) {
	// Note: In a production app, the table name MUST be aggressively sanitized or passed securely
	// to prevent SQL injection. For Phase 1, we rely on the schema scan's safe list of tables.
	query := fmt.Sprintf("SELECT * FROM %s LIMIT 100", table) // Simplistic for MVP
	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	fields := rows.FieldDescriptions()

	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		
		rowMap := make(map[string]interface{})
		for i, field := range fields {
			rowMap[field.Name] = values[i]
		}
		results = append(results, rowMap)
	}
	
	return results, rows.Err()
}

// Close closes the connection pool.
func (p *PostgresScanner) Close() {
	if p.pool != nil {
		p.pool.Close()
	}
}
