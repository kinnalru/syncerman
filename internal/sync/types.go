package sync

import (
	"context"
	"strings"

	"gitlab.com/kinnalru/syncerman/internal/config"
	"gitlab.com/kinnalru/syncerman/internal/logger"
	"gitlab.com/kinnalru/syncerman/internal/rclone"
)

func joinErrorMessages(errors []error, separator string) string {
	if len(errors) == 0 {
		return ""
	}
	messages := make([]string, len(errors))
	for i, err := range errors {
		messages[i] = err.Error()
	}
	return strings.Join(messages, separator)
}

// SyncTarget represents a single synchronization operation between source and destinations.
type SyncTarget struct {
	JobID       string             // The ID of the job this target belongs to
	JobName     string             // The name of the job this target belongs to
	Provider    string             // Source provider name
	SourcePath  string             // Path within source provider
	Destination config.Destination // Single destination configuration
	Resync      bool               // Whether to use --resync flag
}

// SyncOptions contains options for sync operations.
type SyncOptions struct {
	DryRun     bool   // Perform trial run without changes
	ConfigPath string // Path to configuration file
	Verbose    bool   // Enable verbose output
	Quiet      bool   // Suppress non-error output
}

// SyncResult represents the result of a sync operation.
type SyncResult struct {
	Target     SyncTarget // The sync target that was processed
	Success    bool       // Whether the sync completed successfully
	Error      error      // Error if the operation failed
	FirstRun   bool       // Whether the sync was retried due to first-run error
	RetryCount int        // Number of retries performed for this target
}

// SyncEngine defines the interface for synchronization operations.
type SyncEngine interface {
	// Run executes a single sync operation.
	Run(ctx context.Context, target SyncTarget, options SyncOptions) (*SyncResult, error)

	// RunAll executes all sync operations from configuration.
	RunAll(ctx context.Context, config *config.Config, options SyncOptions) ([]*SyncResult, error)

	// Validate checks if configuration and targets are valid for sync.
	Validate(ctx context.Context, config *config.Config) error
}

// Engine is the default implementation of SyncEngine.
type Engine struct {
	config *config.Config
	rclone rclone.Executor
	logger logger.Logger
	dryRun bool
}

func (e *Engine) SetDryRun(dryRun bool) {
	e.dryRun = dryRun
}

func NewEngine(cfg *config.Config, exec rclone.Executor, log logger.Logger) *Engine {
	return &Engine{
		config: cfg,
		rclone: exec,
		logger: log,
	}
}

// SyncEngineFromConfig creates a sync engine from configuration file path.
func SyncEngineFromConfig(configPath string, exec rclone.Executor, log logger.Logger) (*Engine, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	cfgErr := cfg.Validate()
	if cfgErr != nil {
		return nil, cfgErr
	}

	return NewEngine(cfg, exec, log), nil
}
