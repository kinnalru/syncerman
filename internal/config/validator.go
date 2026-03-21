package config

import (
	"fmt"
	"strings"

	"gitlab.com/kinnalru/syncerman/internal/errors"
)

// Validate performs comprehensive validation of the configuration structure.
func (c *Config) Validate() error {
	if len(c.Jobs) == 0 {
		return errors.NewValidationError("configuration is empty: at least one job must be defined. "+
			"Example:\njobs:\n  backup:\n    tasks:\n      - from: 'local:/path'\n        to:\n          - path: 'remote:backup'", nil)
	}

	for _, job := range c.Jobs {
		if job.ID == "" {
			return errors.NewValidationError("job ID cannot be empty", nil)
		}

		if len(job.Tasks) == 0 {
			return errors.NewValidationError(fmt.Sprintf("job %q has no tasks defined. "+
				"Each job must specify at least one task. "+
				"Example:\n  %q:\n    tasks:\n      - from: \"local:/path\"\n        to:\n          - path: \"remote:backup\"", job.ID, job.ID), nil)
		}

		for tIdx, task := range job.Tasks {
			if task.From == "" {
				return errors.NewValidationError(fmt.Sprintf("task 'from' path cannot be empty for job %q (task #%d). "+
					"Each task must specify a source location.", job.ID, tIdx), nil)
			}

			if !isValidFormat(task.From) {
				return errors.NewValidationError(fmt.Sprintf("invalid task 'from' format %q for job %q (task #%d). "+
					"Must be one of:\n  - Remote provider: 'provider:path'\n  - Local path: '/absolute/path' or './relative/path'",
					task.From, job.ID, tIdx), nil)
			}

			if len(task.To) == 0 {
				return errors.NewValidationError(fmt.Sprintf("no destinations defined for job %q task #%d. "+
					"Each task must have at least one destination in the 'to' list.",
					job.ID, tIdx), nil)
			}

			for dIdx, dest := range task.To {
				if err := validateDestination(job.ID, tIdx, dest, dIdx); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func validateDestination(jobID string, taskIdx int, dest Destination, destIdx int) error {
	if dest.Path == "" {
		return errors.NewValidationError(fmt.Sprintf("destination 'path' field cannot be empty for job %q task #%d (destination #%d). "+
			"Each destination must specify where to sync to.",
			jobID, taskIdx, destIdx), nil)
	}

	if !isValidFormat(dest.Path) {
		return errors.NewValidationError(fmt.Sprintf("invalid destination 'path' format %q for job %q task #%d (destination #%d). "+
			"Must be one of:\n  - Remote provider: 'provider:path'\n  - Local path: '/absolute/path' or './relative/path'",
			dest.Path, jobID, taskIdx, destIdx), nil)
	}

	for i, arg := range dest.Args {
		if arg == "" {
			return errors.NewValidationError(fmt.Sprintf("destination args element #%d cannot be empty for job %q task #%d (destination #%d).",
				i, jobID, taskIdx, destIdx), nil)
		}
	}

	return nil
}

func isValidFormat(path string) bool {
	if strings.Contains(path, ":") {
		return true
	}
	if strings.HasPrefix(path, ".") || strings.HasPrefix(path, "/") {
		return true
	}
	return false
}
