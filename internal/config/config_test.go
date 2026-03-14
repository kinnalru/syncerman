package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"syncerman/internal/errors"
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
	config := NewConfig()
	paths := PathMap{"./test": []Destination{}}
	config.AddProvider("gdrive", paths)

	if len(config.Providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(config.Providers))
	}

	if _, ok := config.GetProviders()["gdrive"]; !ok {
		t.Error("gdrive provider not found")
	}
}

func TestConfigGetProviders(t *testing.T) {
	config := NewConfig()
	paths := PathMap{"./test": []Destination{}}
	config.AddProvider("gdrive", paths)

	providers := config.GetProviders()
	if len(providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(providers))
	}
}

func TestConfigGetPaths(t *testing.T) {
	config := NewConfig()
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
	config := NewConfig()
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

func TestConfigGetAllDestinations(t *testing.T) {
	config := NewConfig()
	paths := PathMap{
		"./test": []Destination{
			{To: "ydisk:test"},
		},
		"./other": []Destination{
			{To: "dropbox:other"},
		},
	}
	config.AddProvider("gdrive", paths)
	config.AddProvider("local", PathMap{"./local": []Destination{{To: "gdrive:local"}}})

	targets := config.GetAllDestinations()
	expected := 3
	if len(targets) != expected {
		t.Errorf("expected %d targets, got %d", expected, len(targets))
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

	tmpfile, err := os.CreateTemp("", "config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(yamlContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	config, err := LoadConfig(tmpfile.Name())
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

func TestConfigValidateEmptyConfig(t *testing.T) {
	config := NewConfig()
	err := config.Validate()
	if err == nil {
		t.Error("expected error for empty config")
	}

	if !errors.IsValidationError(err) {
		t.Error("expected ValidationError")
	}
}

func TestConfigValidateEmptyProviderName(t *testing.T) {
	config := NewConfig()
	config.Providers = ProviderMap{
		"": PathMap{"./test": []Destination{{To: "ydisk:test"}}},
	}

	err := config.Validate()
	if err == nil {
		t.Error("expected error for empty provider name")
	}

	if !errors.IsValidationError(err) {
		t.Error("expected ValidationError")
	}
}

func TestConfigValidateEmptyPaths(t *testing.T) {
	config := NewConfig()
	config.Providers = ProviderMap{
		"gdrive": PathMap{},
	}

	err := config.Validate()
	if err == nil {
		t.Error("expected error for empty paths")
	}

	if !errors.IsValidationError(err) {
		t.Error("expected ValidationError")
	}
}

func TestConfigValidateEmptyPath(t *testing.T) {
	config := NewConfig()
	config.Providers = ProviderMap{
		"gdrive": PathMap{
			"": []Destination{{To: "ydisk:test"}},
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("expected error for empty path")
	}

	if !errors.IsValidationError(err) {
		t.Error("expected ValidationError")
	}
}

func TestConfigValidateEmptyDestinations(t *testing.T) {
	config := NewConfig()
	config.Providers = ProviderMap{
		"gdrive": PathMap{
			"./test": []Destination{},
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("expected error for empty destinations")
	}

	if !errors.IsValidationError(err) {
		t.Error("expected ValidationError")
	}
}

func TestConfigValidateEmptyDestinationTo(t *testing.T) {
	config := NewConfig()
	config.Providers = ProviderMap{
		"gdrive": PathMap{
			"./test": []Destination{{To: ""}},
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("expected error for empty destination to")
	}

	if !errors.IsValidationError(err) {
		t.Error("expected ValidationError")
	}
}

func TestConfigValidateInvalidDestinationFormat(t *testing.T) {
	config := NewConfig()
	config.Providers = ProviderMap{
		"gdrive": PathMap{
			"./test": []Destination{{To: "invalid"}},
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("expected error for invalid destination format")
	}

	if !errors.IsValidationError(err) {
		t.Error("expected ValidationError")
	}
}

func TestConfigValidateEmptyArgument(t *testing.T) {
	config := NewConfig()
	config.Providers = ProviderMap{
		"gdrive": PathMap{
			"./test": []Destination{
				{To: "ydisk:test", Args: []string{""}},
			},
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("expected error for empty argument")
	}

	if !errors.IsValidationError(err) {
		t.Error("expected ValidationError")
	}
}

func TestConfigValidateValidConfig(t *testing.T) {
	config := NewConfig()
	config.Providers = ProviderMap{
		"gdrive": PathMap{
			"./test": []Destination{
				{To: "ydisk:test", Args: []string{"--arg1"}},
			},
		},
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("unexpected error for valid config: %v", err)
	}
}

func TestDiscoverConfigPathCustom(t *testing.T) {
	yamlContent := `gdrive: {}`
	tmpfile, err := os.CreateTemp("", "custom-config-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(yamlContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	path, err := DiscoverConfigPath(tmpfile.Name())
	if err != nil {
		t.Fatalf("DiscoverConfigPath() error = %v", err)
	}

	if path != tmpfile.Name() {
		t.Errorf("expected %s, got %s", tmpfile.Name(), path)
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
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	tmpDir, err := os.MkdirTemp("", "test-config-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	yamlContent := `gdrive: {}`
	if err := os.WriteFile("configuration.yml", []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	path, err := DiscoverConfigPath("")
	if err != nil {
		t.Fatalf("DiscoverConfigPath() error = %v", err)
	}

	if !strings.Contains(path, "configuration.yml") {
		t.Errorf("expected configuration.yml in path, got %s", path)
	}
}

func TestDiscoverConfigPathPriority(t *testing.T) {
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	tmpDir, err := os.MkdirTemp("", "test-config-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	yamlContent := `gdrive: {}`
	if err := os.WriteFile(".syncerman.yml", []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("config.yml", []byte(yamlContent), 0644); err != nil {
		t.Fatal(err)
	}

	path, err := DiscoverConfigPath("")
	if err != nil {
		t.Fatalf("DiscoverConfigPath() error = %v", err)
	}

	expectedFile := "config.yml"
	if filepath.Base(path) != expectedFile {
		t.Errorf("expected %s, got %s", expectedFile, filepath.Base(path))
	}
}

func TestDiscoverConfigPathNotFound(t *testing.T) {
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalWd)

	tmpDir, err := os.MkdirTemp("", "test-config-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	_, err = DiscoverConfigPath("")
	if err == nil {
		t.Error("expected error when no config found")
	}

	if !errors.IsConfigError(err) {
		t.Error("expected ConfigError")
	}
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
