package sync

import (
	"context"
	"syncerman/internal/config"
	"syncerman/internal/rclone"
)

// SyncTarget represents a single synchronization operation between source and destinations.
type SyncTarget struct {
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
	logger Logger
	dryRun bool
}

// Logger interface for sync engine logging.
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// SetDryRun sets the dry-run mode for the engine.
// When dry-run is enabled, sync operations will only show what would be changed.
func (e *Engine) SetDryRun(dryRun bool) {
	e.dryRun = dryRun
}

// NewEngine creates a new sync engine with the given configuration and rclone executor.
func NewEngine(cfg *config.Config, exec rclone.Executor, log Logger) *Engine {
	if log == nil {
		log = &defaultLogger{}
	}
	return &Engine{
		config: cfg,
		rclone: exec,
		logger: log,
	}
}

// SyncEngineFromConfig creates a sync engine from configuration file path.
func SyncEngineFromConfig(configPath string, exec rclone.Executor, log Logger) (*Engine, error) {
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

// defaultLogger provides a no-op logger implementation.
type defaultLogger struct{}

func (l *defaultLogger) Debug(msg string, args ...interface{}) {}
func (l *defaultLogger) Info(msg string, args ...interface{})  {}
func (l *defaultLogger) Warn(msg string, args ...interface{})  {}
func (l *defaultLogger) Error(msg string, args ...interface{}) {}
