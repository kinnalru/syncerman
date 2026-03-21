package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateOldConfig(t *testing.T) {
	oldYaml := `
local:
  "./cloud/mirror/folder":
    - to: gdrive:folders/folder1
      args: []
      resync: false
    - to: ydisk:folders/folder1
      args: []
      resync: false
gdrive:
  "folders/folder1":
    - to: ydisk:folders/folder1
      args: []
      resync: false
`
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	err := os.WriteFile(configPath, []byte(oldYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	err = MigrateOldConfig(configPath)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Check if backup exists
	if _, err := os.Stat(configPath + ".bak"); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}

	// Verify new content
	newData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read new config: %v", err)
	}

	content := string(newData)
	if !strings.Contains(content, "jobs:") {
		t.Error("Migrated config does not contain 'jobs:' key")
	}
	if !strings.Contains(content, "gdrive_migration:") && !strings.Contains(content, "local_migration:") {
		t.Error("Migrated config does not contain expected job keys")
	}
	if !strings.Contains(content, "from: local:./cloud/mirror/folder") {
		t.Error("Migrated config does not contain correct source path format")
	}
}

func TestMigrateAlreadyNewConfig(t *testing.T) {
	newYaml := `jobs:
  default:
    tasks:
      - from: "local:/a"
        to:
          - path: "remote:/b"
`
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	err := os.WriteFile(configPath, []byte(newYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	err = MigrateOldConfig(configPath)
	if err == nil {
		t.Fatal("Expected error when migrating already updated config, got nil")
	}
	if !strings.Contains(err.Error(), "already contains 'jobs' key") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}
