package sync

import (
	"context"
	"fmt"
	"strings"
	"syncerman/internal/config"
	"syncerman/internal/rclone"
)

// ValidationErrors represents collection of validation errors.
type ValidationErrors []error

// Error returns formatted error message from all validation errors.
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}

	messages := make([]string, len(ve))
	for i, err := range ve {
		messages[i] = err.Error()
	}

	return fmt.Sprintf("validation errors: %s", strings.Join(messages, "; "))
}

// ValidateTargets checks that all providers and paths in config are valid.
// It verifies that providers exist in rclone and paths are configured.
func (e *Engine) ValidateTargets(ctx context.Context, config *config.Config) error {
	var errs ValidationErrors

	providers := config.GetProviders()
	if len(providers) == 0 {
		return fmt.Errorf("no providers configured")
	}

	for provider := range providers {
		if provider == "" {
			errs = append(errs, fmt.Errorf("provider name cannot be empty"))
			continue
		}

		if provider == "local" {
			continue
		}

		exists, err := rclone.RemoteExists(ctx, e.rclone, provider)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to verify provider %s: %w", provider, err))
			continue
		}

		if !exists {
			errs = append(errs, fmt.Errorf("provider %s not found in rclone configuration", provider))
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

// ExpandTargets expands configuration YAML into a list of sync targets.
// Each source:path + destination combination becomes a SyncTarget.
func (e *Engine) ExpandTargets(config *config.Config) ([]*SyncTarget, error) {
	var targets []*SyncTarget
	var errs ValidationErrors

	providers := config.GetProviders()

	for provider, pathMap := range providers {
		for sourcePath, destinations := range pathMap {
			if sourcePath == "" {
				errs = append(errs, fmt.Errorf("source path cannot be empty for provider %s", provider))
				continue
			}

			if len(destinations) == 0 {
				errs = append(errs, fmt.Errorf("no destinations configured for %s:%s", provider, sourcePath))
				continue
			}

			for _, dest := range destinations {
				if dest.To == "" {
					errs = append(errs, fmt.Errorf("destination 'to' cannot be empty for %s:%s", provider, sourcePath))
					continue
				}

				target := &SyncTarget{
					Provider:    provider,
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

// ProviderExists checks if a provider name exists in rclone configuration.
// Local provider always returns true.
func (e *Engine) ProviderExists(ctx context.Context, provider string) (bool, error) {
	if provider == "local" {
		return true, nil
	}

	return rclone.RemoteExists(ctx, e.rclone, provider)
}

// FormatRemote formats provider and path into remote path format.
// For local provider, returns just the path. For remotes, returns "provider:path".
func FormatRemote(provider, path string) string {
	if provider == "local" {
		return path
	}
	return fmt.Sprintf("%s:%s", provider, path)
}

// ParseRemote parses a remote path string into provider and path components.
// Returns empty provider if format is invalid.
func ParseRemote(remote string) (provider, path string, err error) {
	if !strings.Contains(remote, ":") {
		return "local", remote, nil
	}

	parts := strings.SplitN(remote, ":", 2)
	if len(parts) != 2 {
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
