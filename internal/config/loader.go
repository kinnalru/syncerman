package config

import (
	"fmt"
	"os"

	"gitlab.com/kinnalru/syncerman/internal/errors"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads and parses configuration from a YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.NewConfigError(fmt.Sprintf("failed to read configuration file: %s.\n\nSolutions:\n  - Check the file path spelling\n  - Ensure the file exists\n  - Verify file permissions (read access required)\n  - Check that the file is not a directory",
			path), err)
	}

	config, err := parseConfig(data)
	if err != nil {
		return nil, errors.NewConfigError(fmt.Sprintf("failed to parse configuration file: %s.\n\nThis typically indicates a YAML syntax error or invalid configuration structure. "+
			"Check the following:\n  - YAML indentation (use spaces, not tabs)\n  - Matching brackets, quotes, and colons\n  - Valid YAML structure (see documentation for examples)",
			path), err)
	}

	return config, nil
}

// LoadConfigFromData loads and parses configuration from byte data.
func LoadConfigFromData(data []byte) (*Config, error) {
	config, err := parseConfig(data)
	if err != nil {
		return nil, errors.NewConfigError("failed to parse configuration data.\n\nThis typically indicates a YAML syntax error or invalid configuration structure. "+
			"Check the following:\n  - YAML indentation (use spaces, not tabs)\n  - Matching brackets, quotes, and colons\n  - Valid YAML structure (see documentation for examples)", err)
	}

	return config, nil
}

func parseConfig(data []byte) (*Config, error) {
	config := NewConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML configuration: %w. Check YAML syntax and structure", err)
	}
	return config, nil
}
