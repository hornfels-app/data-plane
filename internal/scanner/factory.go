package scanner

import (
	"context"
	"strings"
)

// NewScannerFromURL detects the database type from the URL and returns the appropriate scanner.
func NewScannerFromURL(ctx context.Context, dbURL string) (Scanner, error) {
	if strings.HasPrefix(dbURL, "postgres://") || strings.HasPrefix(dbURL, "postgresql://") {
		return NewPostgresScanner(ctx, dbURL)
	}

	// If it doesn't look like postgres, assume it's a MySQL DSN (user:pass@tcp(host)/db)
	return NewMySQLScanner(ctx, dbURL)
}
