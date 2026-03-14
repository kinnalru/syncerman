package config

import (
	"os"
	"path/filepath"

	"syncerman/internal/errors"
)

const (
	DefaultConfigName   = "configuration.yml"
	AlternateConfigName = "config.yml"
	HiddenConfigName    = ".syncerman.yml"
)

var defaultConfigFiles = []string{
	DefaultConfigName,
	AlternateConfigName,
	HiddenConfigName,
}

func DiscoverConfigPath(customPath string) (string, error) {
	if customPath != "" {
		if err := validateConfigPath(customPath); err != nil {
			return "", err
		}
		return customPath, nil
	}

	return findDefaultConfig()
}

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

func searchInDirectory(dir string) string {
	for _, configFile := range defaultConfigFiles {
		path := filepath.Join(dir, configFile)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func validateConfigPath(path string) error {
	if _, err := os.Stat(path); err != nil {
		return errors.NewConfigError("configuration file not found: "+path, err)
	}
	return nil
}
