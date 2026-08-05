package scanner

import (
	"context"
	"regexp"
	"strings"
)

// Column represents a single column from the database schema.
type Column struct {
	Table    string
	Name     string
	DataType string
	Comment  string
}

// Scanner defines the interface for database-specific schema extractors.
type Scanner interface {
	// ScanSchema returns a list of all columns in the target database.
	ScanSchema(ctx context.Context) ([]Column, error)
	// SampleData retrieves up to 100 sample rows for a specific table to pass to the heuristics engine.
	SampleData(ctx context.Context, table string) ([]map[string]interface{}, error)
	// Close releases database resources.
	Close()
}

var piiTagRegex = regexp.MustCompile(`(?i)\[hornfels:\s*pii=(true|false)\]`)

// HasHornfelsTag checks if the comment contains the [hornfels: pii=X] tag.
// Returns (hasTag, isPII).
func HasHornfelsTag(comment string) (bool, bool) {
	matches := piiTagRegex.FindStringSubmatch(comment)
	if len(matches) < 2 {
		return false, false
	}
	val := strings.ToLower(matches[1])
	return true, val == "true"
}
