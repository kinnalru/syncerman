package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmd(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "default help",
			args: []string{"--help"},
			want: "Syncerman is a CLI application",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			root := rootCmd
			root.SetOut(buf)
			root.SetErr(buf)
			root.SetArgs(tt.args)

			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(buf.String(), tt.want) {
				t.Errorf("expected %q in output, got %q", tt.want, buf.String())
			}
		})
	}
}

func TestPersistentFlags(t *testing.T) {
	root := rootCmd

	if root.PersistentFlags().Lookup("config") == nil {
		t.Error("config flag not found")
	}
	if root.PersistentFlags().Lookup("dry-run") == nil {
		t.Error("dry-run flag not found")
	}
	if root.PersistentFlags().Lookup("verbose") == nil {
		t.Error("verbose flag not found")
	}
	if root.PersistentFlags().Lookup("quiet") == nil {
		t.Error("quiet flag not found")
	}
}

func TestGetConfigFile(t *testing.T) {
	cfgFile = "test-config.yml"
	if got := GetConfigFile(); got != "test-config.yml" {
		t.Errorf("GetConfigFile() = %v, want test-config.yml", got)
	}
}

func TestGetLogger(t *testing.T) {
	log := GetLogger()
	if log == nil {
		t.Error("GetLogger() returned nil")
	}
}

func TestIsDryRun(t *testing.T) {
	tests := []struct {
		name  string
		value bool
	}{
		{"dry run enabled", true},
		{"dry run disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dryRun = tt.value
			if got := IsDryRun(); got != tt.value {
				t.Errorf("IsDryRun() = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestIsVerbose(t *testing.T) {
	tests := []struct {
		name  string
		value bool
	}{
		{"verbose enabled", true},
		{"verbose disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verbose = tt.value
			if got := IsVerbose(); got != tt.value {
				t.Errorf("IsVerbose() = %v, want %v", got, tt.value)
			}
		})
	}
}

func TestIsQuiet(t *testing.T) {
	tests := []struct {
		name  string
		value bool
	}{
		{"quiet enabled", true},
		{"quiet disabled", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quiet = tt.value
			if got := IsQuiet(); got != tt.value {
				t.Errorf("IsQuiet() = %v, want %v", got, tt.value)
			}
		})
	}
}
