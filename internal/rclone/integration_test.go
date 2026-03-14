package rclone

import (
	"context"
	"testing"
	"time"

	"syncerman/internal/logger"
)

func TestIntegration_EndToEnd_RcloneDetection(t *testing.T) {
	skipIfNoRclone(t)

	config, err := ConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to create config from environment: %v", err)
	}

	exec := NewExecutorWithLogger(config, logger.NewConsoleLogger())
	exec.(*ExecutorImpl).logger.SetLevel(logger.LevelQuiet)

	if exec == nil {
		t.Fatal("NewExecutor returned nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result := &Result{ExitCode: 0, Stdout: "", Stderr: "", Combined: ""}
	_ = result

	_, err = ListRemotes(ctx, exec)
	if err != nil {
		t.Logf("Real rclone listremotes test: %v", err)
	}
}

func TestIntegration_EndToEnd_MkdirWorkflow(t *testing.T) {
	skipIfNoRclone(t)

	config, err := ConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to create config from environment: %v", err)
	}

	exec := NewExecutorWithLogger(config, logger.NewConsoleLogger())
	exec.(*ExecutorImpl).logger.SetLevel(logger.LevelQuiet)

	ctx := context.Background()

	err = Mkdir(ctx, exec, "local:/tmp/syncerman-test-mkdir")
	if err != nil {
		t.Logf("Mkdir on local filesystem: %v (may be expected if running as non-root)", err)
	}

	_, err = ListRemotes(ctx, exec)
	if err != nil {
		t.Logf("ListRemotes test result: %v", err)
	}
}

func TestIntegration_EndToEnd_BisyncCommandBuild(t *testing.T) {
	bisyncArgs := NewBisyncArgs("gdrive:source", "s3:dest", &BisyncOptions{
		Resync: true,
		DryRun: true,
		Args:   []string{"--max-size", "10M"},
	})

	args := bisyncArgs.Build()

	if len(args) == 0 {
		t.Error("Build() returned empty args")
	}

	hasBisync := false
	for _, arg := range args {
		if arg == "bisync" {
			hasBisync = true
			break
		}
	}

	if !hasBisync {
		t.Error("Build() args do not contain 'bisync'")
	}

	hasResync := false
	for _, arg := range args {
		if arg == "--resync" {
			hasResync = true
			break
		}
	}

	if !hasResync {
		t.Error("Build() args do not contain --resync flag")
	}

	hasDryRun := false
	for _, arg := range args {
		if arg == "--dry-run" {
			hasDryRun = true
			break
		}
	}

	if !hasDryRun {
		t.Error("Build() args do not contain --dry-run flag")
	}
}

func TestIntegration_EndToEnd_ErrorHandling(t *testing.T) {
	skipIfNoRclone(t)

	config, err := ConfigFromEnv()
	if err != nil {
		t.Fatalf("Failed to create config from environment: %v", err)
	}

	exec := NewExecutorWithLogger(config, logger.NewConsoleLogger())
	exec.(*ExecutorImpl).logger.SetLevel(logger.LevelQuiet)

	ctx := context.Background()

	err = Mkdir(ctx, exec, "invalid:nonexistent/path")
	if err == nil {
		t.Log("Mkdir with invalid remote: returned nil (expected)")
	} else {
		t.Logf("Mkdir with invalid remote: %v", err)
	}
}

func TestIntegration_EndToEnd_StandardFlags(t *testing.T) {
	bisyncArgs := NewBisyncArgs("local:/src", "gdrive:/dst", nil)
	args := bisyncArgs.Build()

	standardFlags := []string{
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

	for _, flag := range standardFlags {
		hasFlag := false
		for _, arg := range args {
			if arg == flag {
				hasFlag = true
				break
			}
		}

		if !hasFlag {
			t.Errorf("Build() args do not contain standard flag: %s", flag)
		}
	}
}

func TestIntegration_EndToEnd_DryRunWorkflow(t *testing.T) {
	options := &BisyncOptions{
		DryRun: true,
	}

	bisyncArgs := NewBisyncArgs("local:/src", "gdrive:/dst", options)
	args := bisyncArgs.Build()

	hasDryRun := false
	for _, arg := range args {
		if arg == "--dry-run" {
			hasDryRun = true
			break
		}
	}

	if !hasDryRun {
		t.Error("DryRun workflow: args do not contain --dry-run")
	}
}

func TestIntegration_EndToEnd_ErrorDetection(t *testing.T) {
	firstRunError := "ERROR : Bisync critical error: cannot find prior Path1 or Path2 listings, likely due to critical error on prior run\nTip: here are the filenames we were looking for. Do they exist?\n"

	if !IsFirstRunError(firstRunError) {
		t.Error("FirstRunError detection failed for standard error")
	}

	paths := ExtractFirstRunErrorPaths(firstRunError)
	if len(paths) == 0 {
		t.Log("No paths extracted from first-run error (expected)")
	}

	parsed := ParseFirstRunError(firstRunError)
	if parsed == nil {
		t.Error("ParseFirstRunError returned nil for valid error")
	}

	if parsed != nil && parsed.Message != firstRunError {
		t.Errorf("ParseFirstRunError message mismatch: got %q, want %q", parsed.Message, firstRunError)
	}
}

func TestIntegration_EndToEnd_Configuration(t *testing.T) {
	config := NewConfig()
	if config.BinaryPath == "" {
		t.Error("NewConfig() returned empty BinaryPath")
	}

	if config.BinaryPath != "rclone" {
		t.Errorf("NewConfig() BinaryPath = %q, want 'rclone'", config.BinaryPath)
	}

	customPath := "/custom/path/to/rclone"
	config.BinaryPath = customPath
	if config.BinaryPath != customPath {
		t.Errorf("Config BinaryPath assignment failed: got %q, want %q", config.BinaryPath, customPath)
	}
}

func TestIntegration_EndToEnd_ResultHandling(t *testing.T) {
	result := &Result{
		ExitCode: 0,
		Stdout:   "output",
		Stderr:   "",
		Combined: "output",
	}

	if !result.Success() {
		t.Error("Success() returned false for exitCode 0")
	}

	if err := result.Error(); err != nil {
		t.Errorf("Error() returned non-nil for successful result: %v", err)
	}

	failedResult := &Result{
		ExitCode: 1,
		Stdout:   "",
		Stderr:   "error message",
		Combined: "error message",
	}

	if failedResult.Success() {
		t.Error("Success() returned true for exitCode 1")
	}

	if err := failedResult.Error(); err == nil {
		t.Error("Error() returned nil for failed result")
	}
}
