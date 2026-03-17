package rclone

// Package rclone provides integration with the rclone CLI for bidirectional synchronization.
//
// Architecture
//
// This package interfaces with the rclone command-line tool through direct execution,
// providing a clean abstraction over rclone's command complexity while maintaining full
// access to its capabilities. The package follows a simple execution model where
// rclone commands are built, executed, and their output captured for parsing.
//
// Command Execution Model
//
// The Executor interface defines the contract for running rclone commands:
//   - Run(ctx context.Context, args ...string) (*Result, error)
//
// ExecutorImpl provides the standard implementation using Go's exec.Command.
// It manages the rclone binary location, executes commands with context support for
// cancellation, and captures stdout/stderr streams. Commands are logged at debug level
// for troubleshooting.
//
// Output Parsing Strategy
//
// Results are captured in the Result struct:
//   - ExitCode: Process exit status (0 = success)
//   - Stdout: Standard output buffer
//   - Stderr: Standard error buffer
//   - Combined: Concatenated stdout and stderr
//
// Result provides convenience methods:
//   - Success(): Returns true if ExitCode == 0
//   - Error(): Returns error object if command failed, nil otherwise
//
// Error Detection and Handling
//
// Command failures are detected through:
//   - Non-zero exit codes from rclone
//   - Context cancellation during execution
//   - Invalid command start errors
//
// Errors are wrapped with syncerman errors package providing context about
// the failure type (rclone error, context cancellation, command start error).
//
// First-Run Error Detection
//
// The package implements automatic first-run error recognition for bisync operations.
// When rclone bisync runs for the first time, it lacks state files and produces
// a specific error pattern:
//
//   "cannot find prior Path1 or Path2 listings" and "here are the filenames"
//
// The package uses two regexp patterns to detect this condition:
//   - FirstRunErrorPatternListings: Matches the listings error
//   - FirstRunErrorPatternFilenames: Matches the filenames hint
//
// See guides/OVERALL.md:311-337 for full first-run handling details.
//
// Key Components
//
// Config
//   Configures the rclone binary location. The BinaryPath field defaults to "rclone",
//   which searches for rclone in the system PATH. Can be set to an absolute path
//   for specific installations.
//
// Executor
//   Interface for executing rclone commands. ExecutorImpl is the default implementation
//   with support for custom logging via NewExecutorWithLogger. All package functions
//   accept an Executor parameter for testability (can inject mock executor).
//
// Result
//   Represents the output of rclone command execution with ExitCode, Stdout, Stderr,
//   and Combined output. Provides Success() and Error() methods for status checking.
//
// BisyncArgs
//   Builder for constructing rclone bisync command arguments. Implements the standard
//   bisync template from guides/OVERALL.md:256-264:
//
//     rclone bisync <SRC> <DST> \
//       --create-empty-src-dirs \
//       --compare=size,modtime \
//       --no-slow-hash \
//       -Mv \
//       --drive-skip-gdocs \
//       --fix-case \
//       --ignore-listing-checksum \
//       --fast-list \
//       --transfers=10 \
//       --resilient
//
//   Supports fluent configuration with WithResync(), WithDryRun(), and WithArgs().
//   The String() method returns the full command for debugging.
//
// Remote Management
//   - ListRemotes: Executes "rclone listremotes" to get configured remotes
//   - RemoteExists: Checks if a specific remote name is configured
//   - Returns remote names with trailing colon removed (e.g., "gdrive" not "gdrive:")
//
// Directory Creation
//   - Mkdir: Executes "rclone mkdir <remote:path>" to create directories
//   - CreatePath: Creates directory structures with parent directory handling
//   - Treats "directory already exists" as success
//   - Returns formatted errors for creation failures
//
// Provider Handling
//
// Local Provider:
//   - Path format: ./path/to/folder or local:./path/to/folder
//   - No provider prefix needed for filesystem paths
//   - Paths starting with ./ are relative to current directory
//
// Remote Providers:
//   - Format: <provider>:<path>
//   - Provider name must match rclone remote configuration
//   - Examples: gdrive:documents, ydisk:backup, s3:bucket/folder
//
// See guides/OVERALL.md:250-337 for complete rclone integration details.
