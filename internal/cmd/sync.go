package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"syncerman/internal/config"
	"syncerman/internal/rclone"
	"syncerman/internal/sync"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize targets from configuration or single target",
	Long: `Sync executes bidirectional synchronization using rclone bisync.
	
When called without arguments, syncs all targets from configuration file.
When called with a target argument (provider:path), syncs only that specific target.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runSync(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) {
	log := GetLogger()

	configPath := getConfigPath()
	if configPath == "" {
		fmt.Fprintln(os.Stderr, "Error: No configuration file found (use --config to specify)")
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Error("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	opts := sync.SyncOptions{
		DryRun:  IsDryRun(),
		Verbose: IsVerbose(),
		Quiet:   IsQuiet(),
	}

	executor := rclone.NewExecutor(rclone.NewConfig())
	engine := sync.NewEngine(cfg, executor, log)
	ctx := context.Background()

	if len(args) == 0 {
		syncAllTargets(ctx, engine, cfg, opts)
	} else {
		syncSingleTarget(ctx, engine, args[0], opts)
	}
}

func syncAllTargets(ctx context.Context, engine *sync.Engine, cfg *config.Config, opts sync.SyncOptions) {
	log := GetLogger()
	if err := engine.Prepare(ctx, cfg, opts); err != nil && !opts.Quiet {
		log.Error("Failed to prepare directories: %v", err)
	}

	results, err := engine.RunAll(ctx, cfg, opts)
	if err != nil {
		report := engine.CollectResults(results)
		if !opts.Quiet {
			fmt.Fprintln(os.Stderr, report.FormatError())
		}
		os.Exit(1)
	}

	report := engine.CollectResults(results)
	if opts.Verbose || !opts.Quiet {
		fmt.Println(report.Format(opts.Verbose))
	}

	if report.ExitCode != 0 {
		os.Exit(report.ExitCode)
	}
}

func syncSingleTarget(ctx context.Context, engine *sync.Engine, targetArg string, opts sync.SyncOptions) {
	log := GetLogger()
	provider, path, err := sync.ParseRemote(targetArg)
	if err != nil {
		log.Error("Invalid target format: %v (expected: provider:path)", err)
		os.Exit(1)
	}

	configPath := getConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Error("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	targets, err := engine.ExpandTargets(cfg)
	if err != nil {
		log.Error("Failed to expand targets: %v", err)
		os.Exit(1)
	}

	var found *sync.SyncTarget
	for _, target := range targets {
		if target.Provider == provider && target.SourcePath == path {
			found = target
			break
		}
	}

	if found == nil {
		log.Error("Target %s:%s not found in configuration", provider, path)
		os.Exit(1)
	}

	if err := engine.Prepare(ctx, cfg, opts); err != nil && !opts.Quiet {
		log.Error("Failed to prepare directories: %v", err)
	}

	result, err := engine.Run(ctx, *found, opts)
	if err != nil {
		log.Error("Sync failed: %v", err)
		os.Exit(1)
	}

	report := engine.CollectResults([]*sync.SyncResult{result})
	if opts.Verbose || !opts.Quiet {
		fmt.Println(report.Format(opts.Verbose))
	}

	if report.ExitCode != 0 {
		os.Exit(report.ExitCode)
	}
}

func getConfigPath() string {
	if cfgFile != "" {
		return cfgFile
	}

	defaultPaths := []string{
		"./syncerman.yaml",
		"./syncerman.yml",
	}

	for _, path := range defaultPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}
