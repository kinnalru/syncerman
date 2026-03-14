package rclone

import (
	"strings"
)

// BisyncOptions contains optional parameters for rclone bisync command.
type BisyncOptions struct {
	DryRun bool
	Resync bool
	Args   []string
}

// BisyncArgs builds arguments for rclone bisync command.
type BisyncArgs struct {
	src     string
	dst     string
	options *BisyncOptions
}

// NewBisyncArgs creates a new BisyncArgs with source, destination, and options.
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
func (b *BisyncArgs) WithResync() *BisyncArgs {
	b.options.Resync = true
	return b
}

// WithDryRun adds --dry-run flag to the bisync command.
func (b *BisyncArgs) WithDryRun() *BisyncArgs {
	b.options.DryRun = true
	return b
}

// WithArgs adds additional arguments to the bisync command.
func (b *BisyncArgs) WithArgs(args ...string) *BisyncArgs {
	b.options.Args = append(b.options.Args, args...)
	return b
}

// Build returns the complete list of arguments for rclone bisync command.
// The arguments include all standard flags from OVERALL.md plus any user-specified args.
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

func (b *BisyncArgs) buildStandardFlags() []string {
	return []string{
		"--create-empty-src-dirs",
		"--compare=size,modtime",
		"--no-slow-hash",
		"-MvP",
		"--drive-skip-gdocs",
		"--fix-case",
		"--ignore-listing-checksum",
		"--fast-list",
		"--transfers=10",
		"--resilient",
	}
}

// String returns the bisync command as a string (for logging/debugging).
// Format: "rclone bisync [--flags...] source_dest destination"
func (b *BisyncArgs) String() string {
	args := b.Build()
	return "rclone " + strings.Join(args, " ")
}

func cmdString(args []string) string {
	result := ""
	for _, arg := range args {
		if result == "" {
			result = arg
		} else {
			result += " " + arg
		}
	}
	return result
}
