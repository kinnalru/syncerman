package sync

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"gitlab.com/kinnalru/syncerman/internal/config"
	"gitlab.com/kinnalru/syncerman/internal/rclone"
)

const (
	localProvider    = "local"
	remoteDelimiter  = ":"
	remoteSplitCount = 2
)

// ValidationErrors represents collection of validation errors.
// Used to aggregate multiple validation failures into a single error return.
type ValidationErrors []error

// Error returns formatted error message from all validation errors.
// Concatenates all error messages into a single string separated by semicolons.
// Returns empty string if no validation errors exist.
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	return fmt.Sprintf("validation errors: %s", joinErrorMessages(ve, "; "))
}

// ValidateTargets checks that all providers and paths in config are valid.
// It verifies that providers exist in rclone configuration and paths are configured.
// Validates each provider by querying rclone, except for 'local' which is always valid.
// Validates jobs and tasks in the order they appear in the configuration.
func (e *Engine) ValidateTargets(ctx context.Context, config *config.Config) error {
	var errs ValidationErrors

	jobs := config.GetJobs()
	if len(jobs) == 0 {
		return fmt.Errorf("no jobs configured")
	}

	providerCache := make(map[string]bool)

	for _, job := range jobs {
		if !job.Enabled {
			continue
		}

		for _, task := range job.Tasks {
			if !task.Enabled {
				continue
			}

			if task.From == "" {
				errs = append(errs, fmt.Errorf("source 'from' cannot be empty in job %s", job.ID))
				continue
			}

			providerName, _, err := ParseRemote(task.From)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid source 'from' format %s in job %s: %w", task.From, job.ID, err))
				continue
			}

			if providerName == localProvider {
				continue
			}

			if _, checked := providerCache[providerName]; checked {
				continue
			}

			exists, err := rclone.RemoteExists(ctx, e.rclone, providerName)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to verify provider %s for job %s: %w", providerName, job.ID, err))
				providerCache[providerName] = true
				continue
			}

			if !exists {
				errs = append(errs, fmt.Errorf("provider %s not found in rclone configuration for job %s", providerName, job.ID))
			}
			providerCache[providerName] = true
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// ExpandTargets expands configuration YAML into a list of sync targets.
// Each task:from + destination combination becomes a SyncTarget.
// Validates source paths and destination targets, but does not validate rclone provider existence.
// Provider validation should be done separately using ValidateTargets().
// Returns error if any target is invalid, along with all validation errors found.
// Targets are returned in the exact order from YAML configuration respecting Priority.
// If filterJobIDs are provided, only targets belonging to those job IDs are returned.
func (e *Engine) ExpandTargets(config *config.Config, filterJobIDs ...string) ([]*SyncTarget, error) {
	var targets []*SyncTarget
	var errs ValidationErrors

	jobIDFilter := make(map[string]bool)
	for _, id := range filterJobIDs {
		jobIDFilter[id] = true
	}

	for _, job := range config.GetJobs() {
		if !job.Enabled {
			continue
		}

		if len(jobIDFilter) > 0 && !jobIDFilter[job.ID] {
			continue
		}

		for _, task := range job.Tasks {
			if !task.Enabled {
				continue
			}

			if task.From == "" {
				errs = append(errs, fmt.Errorf("source 'from' cannot be empty in job %s", job.ID))
				continue
			}

			if len(task.To) == 0 {
				errs = append(errs, fmt.Errorf("no destinations configured for %s in job %s", task.From, job.ID))
				continue
			}

			providerName, sourcePath, err := ParseRemote(task.From)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid source 'from' format %s in job %s: %w", task.From, job.ID, err))
				continue
			}

			// Create a sync target for each destination
			for _, dest := range task.To {
				if dest.Path == "" {
					errs = append(errs, fmt.Errorf("destination 'path' cannot be empty for %s in job %s", task.From, job.ID))
					continue
				}

				target := &SyncTarget{
					JobID:       job.ID,
					JobName:     job.Name,
					Provider:    providerName,
					SourcePath:  sourcePath,
					Destination: dest,
					Resync:      dest.Resync,
				}

				targets = append(targets, target)
			}
		}
	}

	if len(errs) > 0 {
		return nil, errs
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no valid sync targets found in configuration")
	}

	return targets, nil
}

// Validate calls ValidateTargets to check configuration validity.
func (e *Engine) Validate(ctx context.Context, config *config.Config) error {
	return e.ValidateTargets(ctx, config)
}

// RemoteProviderExists checks if a provider name exists in rclone configuration.
// Local provider always returns true.
func (e *Engine) RemoteProviderExists(ctx context.Context, provider string) (bool, error) {
	if provider == localProvider {
		return true, nil
	}

	return rclone.RemoteExists(ctx, e.rclone, provider)
}

// FormatRemote formats provider and path into remote path format.
// For local provider, returns just the path. For remotes, returns "provider:path".
func FormatRemote(provider, path string) string {
	if provider == localProvider {
		return path
	}
	return fmt.Sprintf("%s%s%s", provider, remoteDelimiter, path)
}

// ParseRemote parses a remote path string into provider and path components.
// Returns empty provider if format is invalid.
// Assumes 'local' provider if no colon is present.
// Validates that both provider and path are non-empty.
func ParseRemote(remote string) (provider, path string, err error) {
	// Local path (no colon found)
	if !strings.Contains(remote, remoteDelimiter) {
		return localProvider, remote, nil
	}

	// Remote path with format "provider:path"
	parts := strings.SplitN(remote, remoteDelimiter, remoteSplitCount)
	if len(parts) != remoteSplitCount {
		return "", "", fmt.Errorf("invalid remote format: %s", remote)
	}

	provider = parts[0]
	path = parts[1]

	if provider == "" {
		return "", "", fmt.Errorf("provider name cannot be empty in remote: %s", remote)
	}

	if path == "" {
		return "", "", fmt.Errorf("path cannot be empty in remote: %s", remote)
	}

	return provider, path, nil
}

// StripProviderHash removes provider hash suffix from rclone path format.
// Pattern matches: provider{ALPHANUMERIC_HASH}:rest_of_path
// Returns provider:rest_of_path if pattern matches, otherwise returns original path.
func StripProviderHash(path string) string {
	re := regexp.MustCompile(`^(\w+)\{[A-Za-z0-9]+\}:(.*)$`)
	matches := re.FindStringSubmatch(path)

	if len(matches) == 3 {
		return matches[1] + ":" + matches[2]
	}

	return path
}
