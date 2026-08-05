package policy

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Baseline represents the .hornfels-baseline.yaml state file.
// It tracks columns that existed before Hornfels was installed, ignoring them.
type Baseline struct {
	IgnoredColumns map[string][]string `yaml:"ignored_columns"` // Table -> List of Column Names
}

// LoadBaseline reads the baseline state file.
func LoadBaseline(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty baseline if file does not exist
			return &Baseline{
				IgnoredColumns: make(map[string][]string),
			}, nil
		}
		return nil, err
	}

	var b Baseline
	if err := yaml.Unmarshal(data, &b); err != nil {
		return nil, err
	}

	if b.IgnoredColumns == nil {
		b.IgnoredColumns = make(map[string][]string)
	}

	return &b, nil
}

// IsIgnored checks if a specific column in a table is ignored by the baseline.
func (b *Baseline) IsIgnored(table, column string) bool {
	cols, exists := b.IgnoredColumns[table]
	if !exists {
		return false
	}
	for _, c := range cols {
		if c == column {
			return true
		}
	}
	return false
}
