package config

import (
	"os"
	"path/filepath"

	"syncerman/internal/errors"
)

// DefaultConfigName is the primary configuration file name searched when no custom path is provided.
const DefaultConfigName = "configuration.yml"

// AlternateConfigName is a common alternative configuration file name used as a fallback when DefaultConfigName is not found.
const AlternateConfigName = "config.yml"

// HiddenConfigName is the hidden configuration file name, typically used for user-specific overrides.
const HiddenConfigName = ".syncerman.yml"

var defaultConfigFiles = []string{
	DefaultConfigName,
	AlternateConfigName,
	HiddenConfigName,
}

// DiscoverConfigPath discovers and validates the configuration file path.
//
// If a custom path is provided, it validates that the file exists at that location.
// Otherwise, it searches for default configuration files in the current directory
// and parent directories.
//
// Parameters:
//   - customPath: optional custom path to a configuration file. If empty,
//     the function searches for default configuration files.
//
// Returns:
//   - string: the resolved configuration file path
//   - error: error if configuration file is not found
//
// Default search order: current directory and parent directories for .syncerman.yml,
// config.yml, configuration.yml
func DiscoverConfigPath(customPath string) (string, error) {
	if customPath != "" {
		if err := validateConfigPath(customPath); err != nil {
			return "", err
		}
		return customPath, nil
	}

	return findDefaultConfig()
}

// findDefaultConfig searches for default configuration files in current and parent directories.
//
// The search starts in the current working directory and travels upward through
// the directory tree until reaching the root directory. For each directory visited,
// it checks for the presence of any default configuration file.
//
// Returns:
//   - string: the found configuration file path
//   - error: error if no configuration file is found in the search path
//
// Default configuration files searched (in order):
//   - configuration.yml
//   - config.yml
//   - .syncerman.yml
func findDefaultConfig() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.NewConfigError("failed to get current directory", err)
	}

	configPath := searchInDirectory(cwd)
	if configPath != "" {
		return configPath, nil
	}

	parent := filepath.Dir(cwd)
	for parent != cwd && parent != "/" {
		configPath = searchInDirectory(parent)
		if configPath != "" {
			return configPath, nil
		}
		cwd = parent
		parent = filepath.Dir(cwd)
	}

	return "", errors.NewConfigError("no configuration file found (searched for: configuration.yml, config.yml, .syncerman.yml)", nil)
}

// searchInDirectory searches for any default config file in a specific directory.
//
// It checks each default configuration file name in the predefined order
// and returns the path to the first file found.
//
// Parameters:
//   - dir: directory path to search in
//
// Returns:
//   - string: the found config file path, or empty string if not found
//
// The function checks files in this order:
//   - configuration.yml
//   - config.yml
//   - .syncerman.yml
func searchInDirectory(dir string) string {
	for _, configFile := range defaultConfigFiles {
		path := filepath.Join(dir, configFile)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// validateConfigPath validates that a configuration file exists at the specified path.
//
// Parameters:
//   - path: path to the configuration file to validate
//
// Returns:
//   - error: error if the file doesn't exist at the specified path, nil if valid
func validateConfigPath(path string) error {
	if _, err := os.Stat(path); err != nil {
		return errors.NewConfigError("configuration file not found: "+path, err)
	}
	return nil
}
