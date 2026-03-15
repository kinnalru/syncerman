package cmd

import (
	"os"

	"gitlab.com/kinnalru/syncerman/internal/config"
)

func loadAndValidateConfig() (*config.Config, error) {
	log := GetLogger()

	configPath := discoverConfigPath()
	if configPath == "" {
		log.Error("No configuration file found (use --config to specify)")
		return nil, &ExitCodeError{Code: exitCodeFileNotFound, Err: os.ErrNotExist}
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Error("Failed to load configuration: %v", err)
		return nil, &ExitCodeError{Code: exitCodeConfigError, Err: err}
	}

	if err := cfg.Validate(); err != nil {
		log.Error("Configuration validation failed: %v", err)
		return nil, &ExitCodeError{Code: exitCodeConfigError, Err: err}
	}

	return cfg, nil
}
