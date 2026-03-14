package rclone

import (
	"context"
	"strings"
)

// ListRemotes executes the rclone listremotes command and returns a list of remote names.
// The returned remote names have the trailing colon removed.
// Returns an empty slice if no remotes are configured (not an error).
// Returns an error if the rclone command fails.
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
		line = strings.TrimSpace(line)
		if line != "" {
			remoteName := strings.TrimSuffix(line, ":")
			remotes = append(remotes, remoteName)
		}
	}

	return remotes, nil
}

// RemoteExists checks if a specific remote name exists in the configured list of remotes.
// Returns true if the remote exists, false otherwise.
// Returns an error if the rclone command fails.
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
