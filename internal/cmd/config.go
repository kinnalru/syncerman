package cmd

import (
	"gitlab.com/kinnalru/syncerman/internal/config"
	"gitlab.com/kinnalru/syncerman/internal/rclone"
	"gitlab.com/kinnalru/syncerman/internal/sync"
)

func wrapError(code int, err error, prefix string) *ExitCodeError {
	return &ExitCodeError{Code: code, Err: err}
}

func loadAndValidateConfig() (*config.Config, error) {
	log := GetLogger()

	configPath, err := config.DiscoverConfigPath(GetConfigFile())
	if err != nil {
		log.Error("No configuration file found (use --config to specify): %v", err)
		return nil, wrapError(exitCodeFileNotFound, err, "")
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Error("Failed to load configuration: %v", err)
		return nil, wrapError(exitCodeConfigError, err, "")
	}

	if err := cfg.Validate(); err != nil {
		log.Error("Configuration validation failed: %v", err)
		return nil, wrapError(exitCodeConfigError, err, "")
	}

	return cfg, nil
}

func createEngine(cfg *config.Config) *sync.Engine {
	executor := rclone.NewExecutor(rclone.NewConfig())
	return sync.NewEngine(cfg, executor, GetLogger())
}
