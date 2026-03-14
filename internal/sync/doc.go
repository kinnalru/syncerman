// Package sync provides synchronization operation logic for Syncerman.
//
// It handles execution of sync operations, error detection, and recovery
// mechanisms for rclone bisync commands.
//
// Usage Example:
//
//	engine := sync.NewEngine(config, executor, logger)
//	ctx := context.Background()
//
//	target := sync.SyncTarget{
//	    Provider:   "gdrive",
//	    SourcePath: "docs",
//	    Destination: config.Destination{To: "s3:backup/docs"},
//	}
//
//	result, err := engine.Run(ctx, target, sync.SyncOptions{Verbose: true})
//	if err != nil {
//	    log.Fatalf("Sync failed: %v", err)
//	}
//
// Running Multiple Syncs:
//
//	cfg := config.LoadConfig("syncerman.yaml")
//	results, err := engine.RunAll(ctx, cfg, sync.SyncOptions{})
//	if len(results) > 0 {
//	    report := engine.CollectResults(results)
//	    fmt.Println(report.Format(true))
//	}
//
// Error Handling:
//
// The sync engine automatically handles first-run errors by retrying with --resync flag.
// If a sync fails due to missing state files, engine will attempt a retry
// before returning an error. Users can track retries via SyncResult.RetryCount.
//
// Dry-Run Mode:
//
// Enable dry-run mode to preview sync changes without applying them:
//
//	engine.SetDryRun(true)
//	// or
//	result, err := engine.Run(ctx, target, sync.SyncOptions{DryRun: true})
//
// Directory Creation:
//
// Use Prepare() to create all destination directories before sync:
//
//	err := engine.Prepare(ctx, cfg, sync.SyncOptions{})
//
// This can be called separately during initialization to ensure destinations
// exist before starting sync operations.
package sync
