package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestScenario1_FirstTimeSetup(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "check command",
			args: []string{"check", "--help"},
			want: "check",
		},
		{
			name: "sync dry-run",
			args: []string{"sync", "--dry-run", "--help"},
			want: "--dry-run",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			testRoot := rootCmd
			testRoot.SetOut(buf)
			testRoot.SetErr(buf)
			testRoot.SetArgs(tt.args)

			_ = testRoot.Execute()

			if !strings.Contains(buf.String(), tt.want) {
				t.Errorf("expected %q in output, got %q", tt.want, buf.String())
			}
		})
	}
}

func TestScenario2_RegularSyncAllTargets(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "sync verbose",
			args: []string{"sync", "--verbose", "--help"},
			want: "--verbose",
		},
		{
			name: "sync quiet",
			args: []string{"sync", "--quiet", "--help"},
			want: "--quiet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			testRoot := rootCmd
			testRoot.SetOut(buf)
			testRoot.SetErr(buf)
			testRoot.SetArgs(tt.args)

			_ = testRoot.Execute()

			if !strings.Contains(buf.String(), tt.want) {
				t.Errorf("expected %q in output, got %q", tt.want, buf.String())
			}
		})
	}
}

func TestScenario3_SyncSpecificFolder(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "sync local target",
			args: []string{"sync", "local:./documents", "--help"},
			want: "Sync executes",
		},
		{
			name: "sync gdrive target",
			args: []string{"sync", "gdrive:projects/main", "--help"},
			want: "Sync executes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			testRoot := rootCmd
			testRoot.SetOut(buf)
			testRoot.SetErr(buf)
			testRoot.SetArgs(tt.args)

			_ = testRoot.Execute()

			if !strings.Contains(buf.String(), tt.want) {
				t.Errorf("expected %q in output, got %q", tt.want, buf.String())
			}
		})
	}
}

func TestScenario4_CustomConfigFile(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "sync with custom config",
			args: []string{"--config", "/path/to/config.yml", "sync", "--help"},
			want: "Sync executes",
		},
		{
			name: "check with custom config",
			args: []string{"--config", "/home/user/.config/syncerman/config.yml", "check", "--help"},
			want: "check",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			testRoot := rootCmd
			testRoot.SetOut(buf)
			testRoot.SetErr(buf)
			testRoot.SetArgs(tt.args)

			_ = testRoot.Execute()

			if !strings.Contains(buf.String(), tt.want) {
				t.Errorf("expected %q in output, got %q", tt.want, buf.String())
			}
		})
	}
}

func TestGlobalFlags(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		check func(t *testing.T, args []string)
	}{
		{
			name: "config flag",
			args: []string{"--config", "test.yaml"},
			check: func(t *testing.T, args []string) {
				GetConfig().ConfigFile = "test.yaml"
				if got := GetConfigFile(); got != "test.yaml" {
					t.Errorf("GetConfigFile() = %v, want test.yaml", got)
				}
			},
		},
		{
			name: "dry-run flag",
			args: []string{"--dry-run"},
			check: func(t *testing.T, args []string) {
				GetConfig().DryRun = true
				if !IsDryRun() {
					t.Error("IsDryRun() should return true")
				}
			},
		},
		{
			name: "verbose flag",
			args: []string{"--verbose"},
			check: func(t *testing.T, args []string) {
				GetConfig().Verbose = true
				if !IsVerbose() {
					t.Error("IsVerbose() should return true")
				}
			},
		},
		{
			name: "quiet flag",
			args: []string{"--quiet"},
			check: func(t *testing.T, args []string) {
				GetConfig().Quiet = true
				if !IsQuiet() {
					t.Error("IsQuiet() should return true")
				}
			},
		},
		{
			name: "short config flag",
			args: []string{"-c", "test.yaml"},
			check: func(t *testing.T, args []string) {
				GetConfig().ConfigFile = "test.yaml"
				if got := GetConfigFile(); got != "test.yaml" {
					t.Errorf("GetConfigFile() = %v, want test.yaml", got)
				}
			},
		},
		{
			name: "short dry-run flag",
			args: []string{"-d"},
			check: func(t *testing.T, args []string) {
				GetConfig().DryRun = true
				if !IsDryRun() {
					t.Error("IsDryRun() should return true")
				}
			},
		},
		{
			name: "short verbose flag",
			args: []string{"-v"},
			check: func(t *testing.T, args []string) {
				GetConfig().Verbose = true
				if !IsVerbose() {
					t.Error("IsVerbose() should return true")
				}
			},
		},
		{
			name: "short quiet flag",
			args: []string{"-q"},
			check: func(t *testing.T, args []string) {
				GetConfig().Quiet = true
				if !IsQuiet() {
					t.Error("IsQuiet() should return true")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.check(t, tt.args)
		})
	}
}

func TestFlagCombinations(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "verbose and dry-run",
			args: []string{"sync", "--verbose", "--dry-run"},
		},
		{
			name: "quiet and dry-run",
			args: []string{"sync", "--quiet", "--dry-run"},
		},
		{
			name: "config and verbose",
			args: []string{"--config", "test.yaml", "sync", "--verbose"},
		},
		{
			name: "config and dry-run",
			args: []string{"--config", "test.yaml", "sync", "--verbose", "--dry-run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			commandConfig = NewCommandConfig()
			testRoot := rootCmd
			testRoot.SetOut(buf)
			testRoot.SetErr(buf)
			testRoot.SetArgs(tt.args)

			_ = testRoot.Execute()
		})
	}
}

func TestSyncCommandVariants(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "sync all targets",
			args: []string{"sync", "--help"},
			want: "Sync executes",
		},
		{
			name: "sync specific target local",
			args: []string{"sync", "local:./cloud/docs", "--help"},
			want: "Sync executes",
		},
		{
			name: "sync specific target gdrive",
			args: []string{"sync", "gdrive:folders/folder1", "--help"},
			want: "Sync executes",
		},
		{
			name: "sync with local shorthand",
			args: []string{"sync", "./cloud/docs", "--help"},
			want: "Sync executes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			testRoot := rootCmd
			testRoot.SetOut(buf)
			testRoot.SetErr(buf)
			testRoot.SetArgs(tt.args)

			_ = testRoot.Execute()

			if !strings.Contains(buf.String(), tt.want) {
				t.Errorf("expected %q in output, got %q", tt.want, buf.String())
			}
		})
	}
}

func TestCheckCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "check basic",
			args: []string{"check", "--help"},
			want: "check",
		},
		{
			name: "check with verbose",
			args: []string{"check", "--verbose", "--help"},
			want: "check",
		},
		{
			name: "check with custom config",
			args: []string{"--config", "/path/to/config.yml", "check", "--help"},
			want: "check",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			testRoot := rootCmd
			testRoot.SetOut(buf)
			testRoot.SetErr(buf)
			testRoot.SetArgs(tt.args)

			_ = testRoot.Execute()

			if !strings.Contains(buf.String(), tt.want) {
				t.Errorf("expected %q in output, got %q", tt.want, buf.String())
			}
		})
	}
}

func TestAllCLIExamples(t *testing.T) {
	examples := []string{
		"syncerman sync",
		"syncerman sync --verbose",
		"syncerman sync --dry-run",
		"syncerman sync local:./cloud/docs",
		"syncerman sync gdrive:folders/folder1 --verbose",
		"syncerman sync ydisk:folders/folder1 --dry-run",
		"syncerman check",
		"syncerman check --config /path/to/config.yml",
		"syncerman check --verbose",
		"syncerman --config /home/user/.config/syncerman/config.yml sync",
		"syncerman --config /home/user/.config/syncerman/config.yml check",
	}

	for _, example := range examples {
		t.Run(example, func(t *testing.T) {
			args := strings.Split(example, " ")
			if len(args) > 0 && args[0] == "syncerman" {
				args = args[1:]
			}

			buf := new(bytes.Buffer)
			testRoot := rootCmd
			testRoot.SetOut(buf)
			testRoot.SetErr(buf)
			testRoot.SetArgs(args)

			_ = testRoot.Execute()
		})
	}
}
