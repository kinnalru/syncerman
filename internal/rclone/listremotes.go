package rclone

import (
	"context"
	"strings"
)

// parseRemoteLine parses a single line of rclone listremotes output.
// It trims whitespace and removes the trailing colon to extract the clean remote name.
//
// Parameters:
//
//	line - A single line from rclone listremotes output
//
// Returns:
//
//	string - The clean remote name without trailing colon
//	bool   - true if a valid remote was extracted, false for empty lines
func parseRemoteLine(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", false
	}
	remoteName := strings.TrimSuffix(line, ":")
	return remoteName, true
}

// ListRemotes executes the rclone listremotes command and returns a list of remote names.
//
// Purpose: Executes "rclone listremotes" to retrieve all configured remote storages from
// the rclone configuration file. This is useful for validating remote names or providing
// remote selection options to users.
//
// Parameters:
//
//	ctx      - Context for controlling cancellation and timeouts during command execution
//	executor - Executor interface implementation for running rclone commands
//
// Returns:
//
//	[]string - Slice of remote names with trailing colons removed (e.g., "dropbox" instead of "dropbox:")
//	error    - Error if the rclone command fails during execution
//
// Implementation details:
//   - Executes "rclone listremotes" subcommand via the provided executor
//   - Parses command output line by line, ignoring empty lines
//   - Removes trailing colons from remote names (rclone output format includes colons)
//   - Returns empty slice if no remotes are configured (not treated as an error)
//
// Error cases:
//   - Returns error if the executor.Run call fails (context cancellation, binary issues)
//   - Treats non-zero exit codes as "no remotes" and returns empty slice
func ListRemotes(ctx context.Context, executor Executor) ([]string, error) {
	result, err := executor.Run(ctx, "listremotes")
	if err != nil {
		return nil, err
	}

	if result.ExitCode != 0 {
		return []string{}, nil
	}

	lines := strings.Split(result.Stdout, "\n")
	remotes := make([]string, 0, len(lines))

	for _, line := range lines {
		if remoteName, ok := parseRemoteLine(line); ok {
			remotes = append(remotes, remoteName)
		}
	}

	return remotes, nil
}

// RemoteExists checks if a specific remote name exists in the configured list of remotes.
//
// Purpose:
//
//	Validates remote names from user configuration files against the actual
//	rclone configuration. This prevents execution failures due to typos or
//	misconfigured remote names in the application configuration.
//
// Parameters:
//
//	ctx        - Context for controlling cancellation and timeouts during command execution
//	executor   - Executor interface implementation for running rclone commands
//	remoteName - The name of the remote to check (without trailing colon)
//
// Returns:
//
//	bool  - true if the remote exists in rclone configuration, false otherwise
//	error - Error if the ListRemotes call fails
//
// Implementation:
//   - Calls ListRemotes to retrieve all configured remotes
//   - Performs case-sensitive comparison against the provided remote name
//   - Does not add or remove colons; expects plain remote name input
//
// Use case:
//
//	Typically used during application initialization or configuration validation
//	to ensure all remote names referenced in jobs/sync targets are actually
//	configured in rclone before attempting execution.
func RemoteExists(ctx context.Context, executor Executor, remoteName string) (bool, error) {
	remotes, err := ListRemotes(ctx, executor)
	if err != nil {
		return false, err
	}

	for _, remote := range remotes {
		if remote == remoteName {
			return true, nil
		}
	}

	return false, nil
}
