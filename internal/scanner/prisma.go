package scanner

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// PrismaScanner implements Scanner by parsing schema.prisma.
type PrismaScanner struct {
	path string
}

// NewPrismaScanner creates a Prisma scanner.
func NewPrismaScanner(path string) *PrismaScanner {
	if path == "" {
		path = "schema.prisma" // default
	}
	return &PrismaScanner{path: path}
}

// ScanSchema parses the schema.prisma file via Regex.
func (p *PrismaScanner) ScanSchema(ctx context.Context) ([]Column, error) {
	data, err := os.ReadFile(p.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", p.path, err)
	}
	content := string(data)

	var columns []Column
	
	// Extremely simplistic regex for Prisma models
	// Looks for `model Name { ... }` blocks
	modelRegex := regexp.MustCompile(`(?s)model\s+([A-Za-z0-9_]+)\s*\{([^}]+)\}`)
	// Looks for field lines: `fieldName Type @attrs /// [hornfels: pii=X]`
	fieldRegex := regexp.MustCompile(`(?m)^\s*([A-Za-z0-9_]+)\s+([A-Za-z0-9_?]+)(.*)$`)
	
	models := modelRegex.FindAllStringSubmatch(content, -1)
	for _, m := range models {
		table := m[1]
		body := m[2]
		
		lines := strings.Split(body, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "//") && !strings.HasPrefix(line, "///") {
				continue // skip empties and regular comments, but allow Prisma doc comments (///)
			}
			
			// Find field
			fm := fieldRegex.FindStringSubmatch(line)
			if len(fm) >= 3 {
				colName := fm[1]
				colType := fm[2]
				rest := fm[3]
				
				// A proper AST parser would be better, but regex serves MVP well here.
				columns = append(columns, Column{
					Table:    table,
					Name:     colName,
					DataType: colType,
					Comment:  rest, // The tags will be in 'rest'
				})
			}
		}
	}

	return columns, nil
}

// SampleData is a no-op for Prisma since we aren't connected to a live database.
func (p *PrismaScanner) SampleData(ctx context.Context, table string) ([]map[string]interface{}, error) {
	return nil, nil // No live data available
}

func (p *PrismaScanner) Close() {}
