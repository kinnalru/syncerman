package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateCommand(t *testing.T) {
	oldYaml := `
local:
  "./cloud":
    - to: gd:cloud
`
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	err := os.WriteFile(configPath, []byte(oldYaml), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	commandConfig = NewCommandConfig()
	testRoot := rootCmd
	testRoot.AddCommand(migrateCmd)

	buf := new(bytes.Buffer)
	testRoot.SetOut(buf)
	testRoot.SetErr(buf)
	testRoot.SetArgs([]string{"migrate", configPath})

	GetLogger().SetQuiet(true)

	err = testRoot.Execute()
	if err != nil {
		t.Fatalf("migrate command failed: %v", err)
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
}

func TestMigrateCommandNoArgs(t *testing.T) {
	commandConfig = NewCommandConfig()
	commandConfig.ConfigFile = "/non/existent/path/for/test.yml"

	testRoot := rootCmd

	buf := new(bytes.Buffer)
	testRoot.SetOut(buf)
	testRoot.SetErr(buf)
	testRoot.SetArgs([]string{"migrate"})

	GetLogger().SetQuiet(true)

	err := testRoot.Execute()
	if err == nil {
		t.Fatal("Expected error when no args and config missing, got nil")
	}
}
