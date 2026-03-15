package config

import (
	"fmt"
	"os"

	"gitlab.com/kinnalru/syncerman/internal/errors"

	"gopkg.in/yaml.v3"
)

// LoadConfig loads and parses configuration from a YAML file.
//
// It reads the configuration file at the specified path, parses it as YAML,
// and returns a Config object populated with the provider definitions.
//
// Parameters:
//
//	path - path to the configuration file (YAML format)
//
// Returns:
//
//	*Config - configuration object containing provider definitions
//	error - error if configuration loading or parsing fails
//
// Error cases:
//   - File not found or cannot be read (os.ReadFile error)
//   - Invalid YAML syntax in the configuration file
//   - Configuration structure errors (invalid provider format)
//
// Implementation details:
//   - Uses gopkg.in/yaml.v3 for YAML parsing with order preservation
//   - Reads file into memory before parsing
//   - Creates a new Config instance and populates Providers field
//   - Preserves exact YAML document order in the OrderedProviders slice
//
// Example usage:
//
//	config, err := LoadConfig("config.yaml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Loaded %d providers\n", len(config.Providers))
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.NewConfigError(fmt.Sprintf("failed to read configuration file: %s", path), err)
	}

	providers, err := parseProviders(data)
	if err != nil {
		return nil, errors.NewConfigError(fmt.Sprintf("failed to parse configuration file: %s", path), err)
	}

	config := NewConfig()
	config.Providers = providers

	return config, nil
}

// LoadConfigFromData loads and parses configuration from byte data.
//
// It parses the provided byte data as YAML and returns a Config object
// populated with the provider definitions. This is useful for testing
// scenarios, inline configurations, or when configuration data comes from
// memory rather than a file.
//
// Parameters:
//
//	data - configuration data as bytes (YAML format)
//
// Returns:
//
//	*Config - configuration object containing provider definitions
//	error - error if configuration parsing fails
//
// Error cases:
//   - Invalid YAML syntax in the data
//   - Configuration structure errors (invalid provider format)
//
// Implementation details:
//   - Uses gopkg.in/yaml.v3 for YAML parsing with order preservation
//   - Creates a new Config instance and populates Providers field
//   - Does not perform file I/O operations
//   - Preserves exact YAML document order in the OrderedProviders slice
//
// Example usage:
//
//	yamlData := []byte("providers:\n  local: /path/to/dir")
//	config, err := LoadConfigFromData(yamlData)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Loaded %d providers\n", len(config.Providers))
func LoadConfigFromData(data []byte) (*Config, error) {
	providers, err := parseProviders(data)
	if err != nil {
		return nil, errors.NewConfigError("failed to parse configuration data", err)
	}

	config := NewConfig()
	config.Providers = providers

	return config, nil
}

func parseProviders(data []byte) (OrderedProviders, error) {
	var providers OrderedProviders
	if err := yaml.Unmarshal(data, &providers); err != nil {
		return nil, err
	}
	return providers, nil
}
