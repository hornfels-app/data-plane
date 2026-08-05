package policy

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the .hornfels.yaml configuration file.
type Config struct {
	Enforce    bool     `yaml:"enforce"`     // If true, exit 1 on unclassified columns
	Exclude    []string `yaml:"exclude"`     // Regex patterns for tables to completely ignore
	StrictMode bool     `yaml:"strict_mode"` // If true, requires explicit [hornfels: pii=X] on every column
}

// LoadConfig reads the .hornfels.yaml file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file does not exist
			return &Config{
				Enforce:    true,
				StrictMode: true,
			}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
