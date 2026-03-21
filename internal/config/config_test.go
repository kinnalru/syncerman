package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/kinnalru/syncerman/internal/errors"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()
	if config == nil {
		t.Fatal("NewConfig() returned nil")
	}
	if config.Jobs == nil {
		t.Error("Jobs slice is nil")
	}
}

func TestLoadConfig(t *testing.T) {
	yamlContent := `
jobs:
  job1:
    tasks:
      - from: "local:/path"
        to:
          - path: "remote:backup"
`
	tmpfile := createTempConfigFile(yamlContent, t)
	defer func() { _ = os.Remove(tmpfile) }()

	config, err := LoadConfig(tmpfile)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if len(config.Jobs) != 1 {
		t.Errorf("expected 1 job, got %d", len(config.Jobs))
	}
	job := config.Jobs[0]
	if job.ID != "job1" {
		t.Errorf("expected job ID job1, got %s", job.ID)
	}
	if job.Name != "job1" {
		t.Errorf("expected default job Name job1, got %s", job.Name)
	}
	if !job.Enabled {
		t.Error("expected default job Enabled to be true")
	}
	if job.Priority != 10 {
		t.Errorf("expected default Priority 10, got %d", job.Priority)
	}
	if len(job.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(job.Tasks))
	}
	task := job.Tasks[0]
	if !task.Enabled {
		t.Error("expected default task Enabled to be true")
	}
	if len(task.To) != 1 {
		t.Errorf("expected 1 destination, got %d", len(task.To))
	}
}

func TestLoadConfigSorting(t *testing.T) {
	yamlData := `
jobs:
  job_b:
    priority: 20
    tasks:
      - from: "local:/b"
        to:
          - path: "remote:/b"
  job_a:
    priority: 5
    tasks:
      - from: "local:/a"
        to:
          - path: "remote:/a"
  job_c:
    priority: 5
    tasks:
      - from: "local:/c"
        to:
          - path: "remote:/c"
`
	config, err := LoadConfigFromData([]byte(yamlData))
	if err != nil {
		t.Fatalf("LoadConfigFromData() error = %v", err)
	}

	if len(config.Jobs) != 3 {
		t.Fatalf("expected 3 jobs, got %d", len(config.Jobs))
	}

	// Priority 5 first, then by ID (job_a < job_c), then Priority 20
	if config.Jobs[0].ID != "job_a" {
		t.Errorf("expected job_a at index 0, got %s", config.Jobs[0].ID)
	}
	if config.Jobs[1].ID != "job_c" {
		t.Errorf("expected job_c at index 1, got %s", config.Jobs[1].ID)
	}
	if config.Jobs[2].ID != "job_b" {
		t.Errorf("expected job_b at index 2, got %s", config.Jobs[2].ID)
	}
}

func TestConfigValidateErrors(t *testing.T) {
	tests := []struct {
		name        string
		yamlData    string
		expectError bool
	}{
		{
			name:        "empty config",
			yamlData:    `jobs: {}`,
			expectError: true,
		},
		{
			name: "empty tasks",
			yamlData: `
jobs:
  job1:
    tasks: []`,
			expectError: true,
		},
		{
			name: "empty from",
			yamlData: `
jobs:
  job1:
    tasks:
      - from: ""
        to:
          - path: "remote:backup"`,
			expectError: true,
		},
		{
			name: "empty destinations",
			yamlData: `
jobs:
  job1:
    tasks:
      - from: "local:/path"
        to: []`,
			expectError: true,
		},
		{
			name: "empty destination path",
			yamlData: `
jobs:
  job1:
    tasks:
      - from: "local:/path"
        to:
          - path: ""`,
			expectError: true,
		},
		{
			name: "invalid from format",
			yamlData: `
jobs:
  job1:
    tasks:
      - from: "invalid"
        to:
          - path: "remote:backup"`,
			expectError: true,
		},
		{
			name: "invalid destination format",
			yamlData: `
jobs:
  job1:
    tasks:
      - from: "local:/path"
        to:
          - path: "invalid"`,
			expectError: true,
		},
		{
			name: "empty argument",
			yamlData: `
jobs:
  job1:
    tasks:
      - from: "local:/path"
        to:
          - path: "remote:backup"
            args: [""]`,
			expectError: true,
		},
		{
			name: "valid config",
			yamlData: `
jobs:
  job1:
    tasks:
      - from: "local:/path"
        to:
          - path: "remote:backup"
            args: ["--arg1"]`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := LoadConfigFromData([]byte(tt.yamlData))
			if err != nil {
				// empty config might fail at parsing? No, it unmarshals as empty map.
				// let's see.
				if !tt.expectError {
					t.Fatalf("unexpected parsing error: %v", err)
				}
			}
			if config != nil {
				err = config.Validate()
				if tt.expectError && err == nil {
					t.Error("expected error but got none")
				} else if !tt.expectError && err != nil {
					t.Errorf("unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestDiscoverConfigPathCustom(t *testing.T) {
	yamlContent := `
jobs:
  job1:
    tasks:
      - from: "local:/path"
        to:
          - path: "remote:backup"
`
	tmpfile := createTempConfigFile(yamlContent, t)
	defer func() { _ = os.Remove(tmpfile) }()

	path, err := DiscoverConfigPath(tmpfile)
	if err != nil {
		t.Fatalf("DiscoverConfigPath() error = %v", err)
	}

	if path != tmpfile {
		t.Errorf("expected %s, got %s", tmpfile, path)
	}
}

func TestDiscoverConfigPathCustomNotFound(t *testing.T) {
	_, err := DiscoverConfigPath("/nonexistent/config.yml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}

	if !errors.IsConfigError(err) {
		t.Error("expected ConfigError")
	}
}

func TestDiscoverConfigPathDefault(t *testing.T) {
	runInTempDir(func() {
		yamlContent := `
jobs:
  job1:
    tasks:
      - from: "local:/path"
        to:
          - path: "remote:backup"
`
		if err := os.WriteFile(".syncerman.yml", []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		path, err := DiscoverConfigPath("")
		if err != nil {
			t.Fatalf("DiscoverConfigPath() error = %v", err)
		}

		if !strings.Contains(path, ".syncerman.yml") {
			t.Errorf("expected .syncerman.yml in path, got %s", path)
		}
	})
}

func TestDiscoverConfigPathDefaultOnly(t *testing.T) {
	runInTempDir(func() {
		yamlContent := `
jobs:
  job1:
    tasks:
      - from: "local:/path"
        to:
          - path: "remote:backup"
`
		if err := os.WriteFile(".syncerman.yml", []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		path, err := DiscoverConfigPath("")
		if err != nil {
			t.Fatalf("DiscoverConfigPath() error = %v", err)
		}

		expectedFile := ".syncerman.yml"
		if filepath.Base(path) != expectedFile {
			t.Errorf("expected %s, got %s", expectedFile, filepath.Base(path))
		}
	})
}

func TestDiscoverConfigPathNotFound(t *testing.T) {
	runInTempDir(func() {
		_, err := DiscoverConfigPath("")
		if err == nil {
			t.Error("expected error when no config found")
		}

		if !errors.IsConfigError(err) {
			t.Error("expected ConfigError")
		}
	})
}

func TestFindDefaultConfigNotInParentDirectory(t *testing.T) {
	runInTempDir(func() {
		parentDir, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		subDir := filepath.Join(parentDir, "subdir")
		if err := os.Mkdir(subDir, 0755); err != nil {
			t.Fatal(err)
		}

		yamlContent := `jobs: {}`
		if err := os.WriteFile(filepath.Join(parentDir, ".syncerman.yml"), []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		if err := os.Chdir(subDir); err != nil {
			t.Fatal(err)
		}

		_, err = findDefaultConfig()
		if err == nil {
			t.Fatal("expected error when config not found in current directory")
		}

		if !errors.IsConfigError(err) {
			t.Error("expected ConfigError")
		}
	})
}

func TestValidateConfigPathValid(t *testing.T) {
	tmpfile := createTempConfigFile(`jobs: {}`, t)
	defer func() { _ = os.Remove(tmpfile) }()

	err := validateConfigPath(tmpfile)
	if err != nil {
		t.Errorf("validateConfigPath() error = %v", err)
	}
}

func TestValidateConfigPathInvalid(t *testing.T) {
	err := validateConfigPath("/nonexistent/path.yml")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}

	if !errors.IsConfigError(err) {
		t.Error("expected ConfigError")
	}
}

func TestSearchInDirectory(t *testing.T) {
	runInTempDir(func() {
		yamlContent := `jobs: {}`

		if err := os.WriteFile(".syncerman.yml", []byte(yamlContent), 0644); err != nil {
			t.Fatal(err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		path := searchInDirectory(cwd)
		if path == "" {
			t.Error("expected to find config file")
		}

		if !strings.Contains(path, ".syncerman.yml") {
			t.Errorf("expected .syncerman.yml in path, got %s", path)
		}
	})
}

func TestSearchInDirectoryNotFound(t *testing.T) {
	runInTempDir(func() {
		cwd, err := os.Getwd()
		if err != nil {
			t.Fatal(err)
		}

		path := searchInDirectory(cwd)
		if path != "" {
			t.Errorf("expected empty string, got %s", path)
		}
	})
}

func TestIsValidFormat(t *testing.T) {
	tests := []struct {
		name string
		dest string
		want bool
	}{
		{
			name: "valid provider with path",
			dest: "gdrive:test",
			want: true,
		},
		{
			name: "valid provider with nested path",
			dest: "dropbox:backup/documents",
			want: true,
		},
		{
			name: "valid local relative path",
			dest: "./backup",
			want: true,
		},
		{
			name: "valid local absolute path",
			dest: "/backup/documents",
			want: true,
		},
		{
			name: "invalid no colon",
			dest: "gdrivetest",
			want: false,
		},
		{
			name: "invalid simple string",
			dest: "invalid",
			want: false,
		},
		{
			name: "valid provider with colon only",
			dest: "provider:",
			want: true,
		},
		{
			name: "valid parent directory path",
			dest: "../backup",
			want: true,
		},
		{
			name: "valid current directory",
			dest: ".",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidFormat(tt.dest)
			if got != tt.want {
				t.Errorf("isValidFormat(%q) = %v, want %v", tt.dest, got, tt.want)
			}
		})
	}
}

func createTempConfigFile(content string, t *testing.T) string {
	tmpfile, err := os.CreateTemp("", "config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	return tmpfile.Name()
}

func runInTempDir(fn func()) {
	originalWd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	defer func() { _ = os.Chdir(originalWd) }()

	tmpDir, err := os.MkdirTemp("", "test-config-*")
	if err != nil {
		panic(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := os.Chdir(tmpDir); err != nil {
		panic(err)
	}

	fn()
}
