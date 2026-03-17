package rclone

import (
	"strings"
)

// BisyncOptions contains optional parameters for rclone bisync command.
// These options modify the behavior of the bisync command and are passed to NewBisyncArgs.
//
// Fields:
//   - DryRun: bool - trial run without making any changes to the destination
//   - Resync: bool - force initial sync mode, useful for first-time sync or recovery from desync
//   - Args: []string - additional rclone arguments for custom configuration
type BisyncOptions struct {
	DryRun bool
	Resync bool
	Args   []string
}

// BisyncArgs is a builder for rclone bisync command arguments.
// It follows a fluent builder pattern for easy configuration and chaining of options.
//
// Fields:
//   - src: source path for sync
//   - dst: destination path for sync
//   - options: optional bisync parameters that modify command behavior
type BisyncArgs struct {
	src     string
	dst     string
	options *BisyncOptions
}

// NewBisyncArgs creates a new BisyncArgs with source, destination, and options.
//
// Parameters:
//   - src: source path for synchronization in the format "provider:path"
//   - dst: destination path for synchronization in the format "provider:path"
//   - options: optional bisync parameters (nil creates empty options)
//
// Returns: *BisyncArgs configured builder instance
func NewBisyncArgs(src, dst string, options *BisyncOptions) *BisyncArgs {
	if options == nil {
		options = &BisyncOptions{}
	}
	return &BisyncArgs{
		src:     src,
		dst:     dst,
		options: options,
	}
}

// WithResync adds --resync flag to the bisync command.
//
// Use case: force initial sync mode when running bisync for the first time or recovering
// from a desynchronized state. This flag causes rclone to treat the two paths as not
// previously synchronized.
//
// Returns: *BisyncArgs for method chaining (fluent pattern)
func (b *BisyncArgs) WithResync() *BisyncArgs {
	b.options.Resync = true
	return b
}

// WithDryRun adds --dry-run flag to the bisync command.
//
// Use case: preview sync operations without making any changes to the destination.
// Useful for testing configuration and verifying what changes would be made before
// executing the actual sync.
//
// Returns: *BisyncArgs for method chaining (fluent pattern)
func (b *BisyncArgs) WithDryRun() *BisyncArgs {
	b.options.DryRun = true
	return b
}

// WithArgs adds additional arguments to the bisync command.
//
// Parameters: variable string args - custom rclone options or flags
//
// Use case: specify additional rclone configuration options not covered by the standard
// builder methods, such as --filter rules, --max-age, or any other rclone flag.
//
// Returns: *BisyncArgs for method chaining (fluent pattern)
func (b *BisyncArgs) WithArgs(args ...string) *BisyncArgs {
	b.options.Args = append(b.options.Args, args...)
	return b
}

// Build returns the complete list of arguments for rclone bisync command.
// This method constructs the full command argument list including standard flags,
// optional flags from options, and source/destination paths.
//
// Returns: []string of all command arguments ready for use with exec.Command
//
// Implementation details:
//   - Includes standard flags from guides/OVERALL.md:256-267
//   - Adds --resync if options.Resync is true
//   - Adds --dry-run if options.DryRun is true
//   - Adds source and destination paths
//   - Appends additional args if specified in options.Args
func (b *BisyncArgs) Build() []string {
	args := []string{
		"bisync",
	}

	args = append(args, b.buildStandardFlags()...)

	if b.options.Resync {
		args = append(args, "--resync")
	}

	if b.options.DryRun {
		args = append(args, "--dry-run")
	}

	args = append(args, b.src, b.dst)

	if len(b.options.Args) > 0 {
		args = append(args, b.options.Args...)
	}

	return args
}

// buildStandardFlags returns standard rclone bisync flags defined in OVERALL.md:256-267.
// These flags provide optimal configuration for efficient and reliable bidirectional sync.
//
// Returns: []string of standard flag strings
//
// Reference: guides/OVERALL.md:269-283 (Rclone options explanation)
func (b *BisyncArgs) buildStandardFlags() []string {
	return []string{
		"--create-empty-src-dirs",   // Sync creation and deletion of empty directories
		"--compare=size,modtime",    // Compare files by size and modification time instead of checksum
		"--no-slow-hash",            // Skip slow checksum calculations during listing
		"-Mv",                       // Preserve metadata, verbose output
		"--drive-skip-gdocs",        // Skip Google Docs files (Google Drive specific)
		"--fix-case",                // Force rename of case-insensitive destinations
		"--ignore-listing-checksum", // Don't use checksums for listings
		"--fast-list",               // Use faster directory listing that reduces API calls
		"--transfers=10",            // Run 10 parallel transfers for better performance
		"--resilient",               // Allow recovering from errors without full resync
	}
}

// String returns the bisync command as a string for logging and debugging purposes.
// This method provides a human-readable representation of the complete command.
//
// Returns: formatted command string
//
// Format: "rclone bisync [--flags...] source destination"
func (b *BisyncArgs) String() string {
	args := b.Build()
	return "rclone " + strings.Join(args, " ")
}
