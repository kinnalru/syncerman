package rclone

import (
	"strings"
	"testing"
)

func TestNewBisyncArgs(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		dst     string
		options *BisyncOptions
	}{
		{
			name:    "basic args",
			src:     "gdrive:path",
			dst:     "s3:path",
			options: &BisyncOptions{},
		},
		{
			name:    "nil options",
			src:     "gdrive:path",
			dst:     "s3:path",
			options: nil,
		},
		{
			name: "with flags",
			src:  "gdrive:path",
			dst:  "s3:path",
			options: &BisyncOptions{
				Resync: true,
				DryRun: true,
				Args:   []string{"--extra-flag"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := NewBisyncArgs(tt.src, tt.dst, tt.options)

			if args.src != tt.src {
				t.Errorf("NewBisyncArgs() src = %v, want %v", args.src, tt.src)
			}

			if args.dst != tt.dst {
				t.Errorf("NewBisyncArgs() dst = %v, want %v", args.dst, tt.dst)
			}

			if args.options == nil {
				t.Errorf("NewBisyncArgs() options is nil")
			}
		})
	}
}

func TestBisyncArgs_WithResync(t *testing.T) {
	args := NewBisyncArgs("gdrive:path", "s3:path", nil)
	args.WithResync()

	if !args.options.Resync {
		t.Errorf("WithResync() Resync = false, want true")
	}

	built := args.Build()
	hasResync := false
	for _, arg := range built {
		if arg == "--resync" {
			hasResync = true
			break
		}
	}

	if !hasResync {
		t.Errorf("WithResync() build() does not contain --resync flag")
	}
}

func TestBisyncArgs_WithDryRun(t *testing.T) {
	args := NewBisyncArgs("gdrive:path", "s3:path", nil)
	args.WithDryRun()

	if !args.options.DryRun {
		t.Errorf("WithDryRun() DryRun = false, want true")
	}

	built := args.Build()
	hasDryRun := false
	for _, arg := range built {
		if arg == "--dry-run" {
			hasDryRun = true
			break
		}
	}

	if !hasDryRun {
		t.Errorf("WithDryRun() build() does not contain --dry-run flag")
	}
}

func TestBisyncArgs_WithArgs(t *testing.T) {
	args := NewBisyncArgs("gdrive:path", "s3:path", nil)
	args.WithArgs("--extra-flag", "--another-flag")

	expectedArgs := []string{"--extra-flag", "--another-flag"}
	if len(args.options.Args) != len(expectedArgs) {
		t.Errorf("WithArgs() args = %v, want %v", args.options.Args, expectedArgs)
	}

	for i, arg := range args.options.Args {
		if arg != expectedArgs[i] {
			t.Errorf("WithArgs() args[%d] = %v, want %v", i, arg, expectedArgs[i])
		}
	}
}

func TestBisyncArgs_Build(t *testing.T) {
	tests := []struct {
		name          string
		src           string
		dst           string
		options       *BisyncOptions
		requiredFlags []string
	}{
		{
			name:    "basic build",
			src:     "gdrive:docs",
			dst:     "s3:backup/docs",
			options: &BisyncOptions{},
			requiredFlags: []string{
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
			},
		},
		{
			name: "with resync",
			src:  "local:/path",
			dst:  "ydisk:/path",
			options: &BisyncOptions{
				Resync: true,
			},
			requiredFlags: []string{
				"--create-empty-src-dirs",
				"--resync",
			},
		},
		{
			name: "with dry run",
			src:  "gdrive:data",
			dst:  "s3:data",
			options: &BisyncOptions{
				DryRun: true,
			},
			requiredFlags: []string{
				"--create-empty-src-dirs",
				"--dry-run",
			},
		},
		{
			name: "with args",
			src:  "gdrive:photos",
			dst:  "ydisk:photos",
			options: &BisyncOptions{
				Args: []string{"--max-age", "30d"},
			},
			requiredFlags: []string{
				"--create-empty-src-dirs",
				"--max-age",
				"30d",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := NewBisyncArgs(tt.src, tt.dst, tt.options)
			built := args.Build()

			if built[0] != "bisync" {
				t.Errorf("Build()[0] = %v, want 'bisync'", built[0])
			}

			hasSrc := false
			hasDst := false
			for _, arg := range built {
				if arg == tt.src {
					hasSrc = true
				}
				if arg == tt.dst {
					hasDst = true
				}
			}

			if !hasSrc {
				t.Errorf("Build() does not contain source: %s", tt.src)
			}

			if !hasDst {
				t.Errorf("Build() does not contain destination: %s", tt.dst)
			}

			for _, requiredFlag := range tt.requiredFlags {
				hasFlag := false
				for _, arg := range built {
					if arg == requiredFlag {
						hasFlag = true
						break
					}
				}
				if !hasFlag {
					t.Errorf("Build() does not contain required flag: %s", requiredFlag)
				}
			}
		})
	}
}

func TestBuildStandardFlags(t *testing.T) {
	args := NewBisyncArgs("src", "dst", nil)
	built := args.Build()

	expectedFlags := []string{
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

	for _, expectedFlag := range expectedFlags {
		hasFlag := false
		for _, arg := range built {
			if arg == expectedFlag {
				hasFlag = true
				break
			}
		}
		if !hasFlag {
			t.Errorf("buildStandardFlags() does not contain flag: %s", expectedFlag)
		}
	}
}

func TestBisyncArgs_String(t *testing.T) {
	args := NewBisyncArgs("gdrive:docs", "s3:backup", nil)
	cmdStr := args.String()

	if !strings.HasPrefix(cmdStr, "rclone bisync") {
		t.Errorf("String() = %v, want to start with 'rclone bisync'", cmdStr)
	}

	if !strings.Contains(cmdStr, "gdrive:docs") {
		t.Errorf("String() = %v, want to contain 'gdrive:docs'", cmdStr)
	}

	if !strings.Contains(cmdStr, "s3:backup") {
		t.Errorf("String() = %v, want to contain 's3:backup'", cmdStr)
	}
}

func TestBisyncArgs_Chaining(t *testing.T) {
	args := NewBisyncArgs("src", "dst", nil).
		WithResync().
		WithDryRun().
		WithArgs("--extra")

	built := args.Build()

	if !containsFlag(built, "--resync") {
		t.Errorf("Chained flags do not include --resync")
	}

	if !containsFlag(built, "--dry-run") {
		t.Errorf("Chained flags do not include --dry-run")
	}

	if !containsFlag(built, "--extra") {
		t.Errorf("Chained flags do not include --extra")
	}
}

func containsFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

func TestBisyncArgs_ModifyAfterBuild(t *testing.T) {
	args := NewBisyncArgs("src", "dst", nil)
	built1 := args.Build()

	args.WithResync()
	built2 := args.Build()

	if containsFlag(built1, "--resync") {
		t.Errorf("Build() before WithResync() contains --resync")
	}

	if !containsFlag(built2, "--resync") {
		t.Errorf("Build() after WithResync() does not contain --resync")
	}
}
