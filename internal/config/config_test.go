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
	if config.Providers == nil {
		t.Error("Providers map is nil")
	}
}

func TestConfigAddProvider(t *testing.T) {
	config := createTestConfig()
	addTestProvider(config, "gdrive", "./test")

	if len(config.Providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(config.Providers))
	}

	providers := config.GetProviders()
	if providers[0].Name != "gdrive" {
		t.Error("gdrive provider not found")
	}
}

func TestConfigGetProviders(t *testing.T) {
	config := createTestConfig()
	addTestProvider(config, "gdrive", "./test")

	providers := config.GetProviders()
	if len(providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(providers))
	}
}

func TestConfigGetPaths(t *testing.T) {
	config := createTestConfig()
	paths := PathMap{"./test": []Destination{{To: "ydisk:test"}}}
	config.AddProvider("gdrive", paths)

	foundPaths, ok := config.GetPaths("gdrive")
	if !ok {
		t.Error("paths not found for provider gdrive")
	}

	if len(foundPaths) != 1 {
		t.Errorf("expected 1 path, got %d", len(foundPaths))
	}

	_, ok = config.GetPaths("nonexistent")
	if ok {
		t.Error("expected false for nonexistent provider")
	}
}

func TestConfigGetDestinations(t *testing.T) {
	config := createTestConfig()
	paths := PathMap{
		"./test": []Destination{
			{To: "ydisk:test"},
			{To: "dropbox:test"},
		},
	}
	config.AddProvider("gdrive", paths)

	destinations, ok := config.GetDestinations("gdrive", "./test")
	if !ok {
		t.Error("destinations not found")
	}

	if len(destinations) != 2 {
		t.Errorf("expected 2 destinations, got %d", len(destinations))
	}

	_, ok = config.GetDestinations("gdrive", "nonexistent")
	if ok {
		t.Error("expected false for nonexistent path")
	}
}

func TestLoadConfig(t *testing.T) {
	yamlContent := `
gdrive:
  "./test":
    - to: ydisk:test
      args: []
      resync: false
  `

	tmpfile := createTempConfigFile(yamlContent, t)
	defer func() { _ = os.Remove(tmpfile) }()

	config, err := LoadConfig(tmpfile)
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if config == nil {
		t.Fatal("config is nil")
	}

	if len(config.Providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(config.Providers))
	}
}

func TestLoadConfigFileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/config.yml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}

	if !errors.IsConfigError(err) {
		t.Error("expected ConfigError")
	}
}

func TestLoadConfigFromData(t *testing.T) {
	yamlData := `
gdrive:
  "./test":
    - to: ydisk:test
`

	config, err := LoadConfigFromData([]byte(yamlData))
	if err != nil {
		t.Fatalf("LoadConfigFromData() error = %v", err)
	}

	if len(config.Providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(config.Providers))
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	invalidYAML := `
gdrive:
  "./test":
    - invalid yaml syntax
`

	_, err := LoadConfigFromData([]byte(invalidYAML))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}

	if !errors.IsConfigError(err) {
		t.Error("expected ConfigError")
	}
}

func TestLoadConfigFileInvalidYAML(t *testing.T) {
	invalidYAML := `
gdrive:
  "./test":
    - invalid yaml syntax
`

	tmpfile := createTempConfigFile(invalidYAML, t)
	defer func() { _ = os.Remove(tmpfile) }()

	_, err := LoadConfig(tmpfile)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}

	if !errors.IsConfigError(err) {
		t.Error("expected ConfigError")
	}
}

func TestConfigValidateErrors(t *testing.T) {
	tests := []struct {
		name        string
		setupConfig func() *Config
		expectError bool
	}{
		{
			name: "empty config",
			setupConfig: func() *Config {
				return NewConfig()
			},
			expectError: true,
		},
		{
			name: "empty provider name",
			setupConfig: func() *Config {
				config := NewConfig()
				config.AddProvider("", PathMap{"./test": []Destination{{To: "ydisk:test"}}})
				return config
			},
			expectError: true,
		},
		{
			name: "empty paths",
			setupConfig: func() *Config {
				config := NewConfig()
				config.AddProvider("gdrive", PathMap{})
				return config
			},
			expectError: true,
		},
		{
			name: "empty path",
			setupConfig: func() *Config {
				config := NewConfig()
				config.AddProvider("gdrive", PathMap{
					"": []Destination{{To: "ydisk:test"}},
				})
				return config
			},
			expectError: true,
		},
		{
			name: "empty destinations",
			setupConfig: func() *Config {
				config := NewConfig()
				config.AddProvider("gdrive", PathMap{
					"./test": []Destination{},
				})
				return config
			},
			expectError: true,
		},
		{
			name: "empty destination to",
			setupConfig: func() *Config {
				config := NewConfig()
				config.AddProvider("gdrive", PathMap{
					"./test": []Destination{{To: ""}},
				})
				return config
			},
			expectError: true,
		},
		{
			name: "invalid destination format",
			setupConfig: func() *Config {
				config := NewConfig()
				config.AddProvider("gdrive", PathMap{
					"./test": []Destination{{To: "invalid"}},
				})
				return config
			},
			expectError: true,
		},
		{
			name: "empty argument",
			setupConfig: func() *Config {
				config := NewConfig()
				config.AddProvider("gdrive", PathMap{
					"./test": []Destination{
						{To: "ydisk:test", Args: []string{""}},
					},
				})
				return config
			},
			expectError: true,
		},
		{
			name: "valid config",
			setupConfig: func() *Config {
				config := NewConfig()
				config.AddProvider("gdrive", PathMap{
					"./test": []Destination{
						{To: "ydisk:test", Args: []string{"--arg1"}},
					},
				})
				return config
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setupConfig()
			err := config.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if !errors.IsValidationError(err) {
					t.Error("expected ValidationError")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestDiscoverConfigPathCustom(t *testing.T) {
	yamlContent := `gdrive: {}`
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
		yamlContent := `gdrive: {}`
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
		yamlContent := `gdrive: {}`
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

		yamlContent := `gdrive: {}`
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
	tmpfile := createTempConfigFile(`gdrive: {}`, t)
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

func TestValidateConfigPathPermissionDenied(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-perm-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "config.yml")
	if err := os.WriteFile(testFile, []byte("gdrive: {}"), 0600); err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(testFile, 0000); err != nil {
		t.Skipf("cannot set file permissions: %v", err)
	}
	defer os.Chmod(testFile, 0600)

	err = validateConfigPath(testFile)
	if err == nil {
		t.Skip("permission restrictions not enforced in this environment (possibly running as root)")
	}

	if !errors.IsConfigError(err) {
		t.Error("expected ConfigError")
	}
}

func TestSearchInDirectory(t *testing.T) {
	runInTempDir(func() {
		yamlContent := `gdrive: {}`

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

func TestSyncTarget(t *testing.T) {
	dest := Destination{
		To:     "ydisk:test",
		Args:   []string{"--arg1"},
		Resync: true,
	}

	target := SyncTarget{
		SourceProvider: "gdrive",
		SourcePath:     "./test",
		Destination:    dest,
	}

	if target.SourceProvider != "gdrive" {
		t.Errorf("expected gdrive, got %s", target.SourceProvider)
	}

	if target.SourcePath != "./test" {
		t.Errorf("expected ./test, got %s", target.SourcePath)
	}

	if target.Destination.To != "ydisk:test" {
		t.Errorf("expected ydisk:test, got %s", target.Destination.To)
	}

	if !target.Destination.Resync {
		t.Error("expected resync to be true")
	}
}

func TestConfigAddProviderWithNilProviders(t *testing.T) {
	config := &Config{}
	paths := PathMap{"./test": []Destination{{To: "ydisk:test"}}}
	config.AddProvider("gdrive", paths)

	if len(config.Providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(config.Providers))
	}
}

func TestConfigGetPathsWithNilProviders(t *testing.T) {
	config := &Config{}
	_, ok := config.GetPaths("gdrive")
	if ok {
		t.Error("expected false for nil providers")
	}
}

func TestConfigGetDestinationsWithNonexistentProvider(t *testing.T) {
	config := createTestConfig()
	_, ok := config.GetDestinations("nonexistent", "./test")
	if ok {
		t.Error("expected false for nonexistent provider")
	}
}

func TestConfigGetAllDestinations(t *testing.T) {
	config := NewConfig()
	paths := PathMap{
		"./test": []Destination{
			{To: "ydisk:test"},
			{To: "dropbox:test"},
		},
		"./data": []Destination{
			{To: "gdrive:backup"},
		},
	}
	config.AddProvider("gdrive", paths)

	paths2 := PathMap{
		"./files": []Destination{
			{To: "local:/backup"},
		},
	}
	config.AddProvider("dropbox", paths2)

	targets := config.GetAllDestinations()

	expectedCount := 4
	if len(targets) != expectedCount {
		t.Errorf("expected %d targets, got %d", expectedCount, len(targets))
	}

	foundProviders := make(map[string]bool)
	for _, target := range targets {
		foundProviders[target.SourceProvider] = true
	}

	if !foundProviders["gdrive"] || !foundProviders["dropbox"] {
		t.Error("expected to find both gdrive and dropbox providers")
	}
}

func TestConfigGetAllDestinationsEmpty(t *testing.T) {
	config := NewConfig()
	targets := config.GetAllDestinations()

	if len(targets) != 0 {
		t.Errorf("expected 0 targets, got %d", len(targets))
	}
}

func TestConfigGetAllDestinationsNilProviders(t *testing.T) {
	config := &Config{}
	targets := config.GetAllDestinations()

	if len(targets) != 0 {
		t.Errorf("expected 0 targets, got %d", len(targets))
	}
}

func TestConfigCountTotalTargets(t *testing.T) {
	config := NewConfig()
	paths := PathMap{
		"./test": []Destination{
			{To: "ydisk:test"},
			{To: "dropbox:test"},
		},
		"./data": []Destination{
			{To: "gdrive:backup"},
		},
	}
	config.AddProvider("gdrive", paths)

	paths2 := PathMap{
		"./files": []Destination{
			{To: "local:/backup"},
		},
	}
	config.AddProvider("dropbox", paths2)

	total := config.countTotalTargets()
	expected := 4
	if total != expected {
		t.Errorf("expected %d targets, got %d", expected, total)
	}
}

func TestConfigCountTotalTargetsEmpty(t *testing.T) {
	config := NewConfig()
	total := config.countTotalTargets()

	if total != 0 {
		t.Errorf("expected 0 targets, got %d", total)
	}
}

func TestConfigCountTotalTargetsNilProviders(t *testing.T) {
	config := &Config{}
	total := config.countTotalTargets()

	if total != 0 {
		t.Errorf("expected 0 targets, got %d", total)
	}
}

func TestParseProviderMapValid(t *testing.T) {
	yamlData := `
gdrive:
  "./test":
    - to: ydisk:test
dropbox:
  "./files":
    - to: local:/backup
`

	providers, err := parseProviders([]byte(yamlData))
	if err != nil {
		t.Fatalf("parseProviders() error = %v", err)
	}

	if len(providers) != 2 {
		t.Errorf("expected 2 providers, got %d", len(providers))
	}

	if providers[0].Name != "gdrive" {
		t.Error("gdrive provider not found")
	}

	if providers[1].Name != "dropbox" {
		t.Error("dropbox provider not found")
	}
}

func TestParseProviderMapInvalidYAML(t *testing.T) {
	invalidYAML := `
gdrive:
  "./test":
    - invalid: yaml: syntax: error
`

	_, err := parseProviders([]byte(invalidYAML))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestIsValidDestinationFormat(t *testing.T) {
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
			got := isValidDestinationFormat(tt.dest)
			if got != tt.want {
				t.Errorf("isValidDestinationFormat(%q) = %v, want %v", tt.dest, got, tt.want)
			}
		})
	}
}

func createTestConfig() *Config {
	return NewConfig()
}

func addTestProvider(cfg *Config, name string, path string) {
	paths := PathMap{path: []Destination{}}
	cfg.AddProvider(name, paths)
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
