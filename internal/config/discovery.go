package config

import (
	"fmt"
	"os"
	"path/filepath"

	"syncerman/internal/errors"
)

const defaultConfigName = ".syncerman.yml"

var defaultConfigFiles = []string{
	defaultConfigName,
}

// DiscoverConfigPath discovers and validates the configuration file path.
//
// If a custom path is provided, it validates that the file exists at that location.
// Otherwise, it searches for the default configuration file .syncerman.yml in the
// current directory only.
//
// Parameters:
//   - customPath: optional custom path to a configuration file. If empty,
//     the function searches for .syncerman.yml in the current directory.
//
// Returns:
//   - string: the resolved configuration file path
//   - error: error if configuration file is not found
//
// Default search: current directory for .syncerman.yml
func DiscoverConfigPath(customPath string) (string, error) {
	if customPath != "" {
		if err := validateConfigPath(customPath); err != nil {
			return "", err
		}
		return customPath, nil
	}

	return findDefaultConfig()
}

// findDefaultConfig searches for the default configuration file .syncerman.yml in the current directory.
//
// The search is limited to the current working directory only - it does not traverse
// parent directories.
//
// Returns:
//   - string: the found configuration file path
//   - error: error if no configuration file is found in the current directory
//
// Default configuration file searched: .syncerman.yml
func findDefaultConfig() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.NewConfigError("failed to get current working directory", err)
	}

	configPath := searchInDirectory(cwd)
	if configPath != "" {
		return configPath, nil
	}

	return "", errors.NewConfigError(fmt.Sprintf("no configuration file found in current directory (searching for: %s)", defaultConfigName), nil)
}

// searchInDirectory searches for .syncerman.yml in the specified directory.
//
// Parameters:
//   - dir: directory path to search in
//
// Returns:
//   - string: the found config file path, or empty string if not found
//
// The function checks for .syncerman.yml only.
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
		return errors.NewConfigError(fmt.Sprintf("configuration file not found: %s", path), err)
	}
	return nil
}
