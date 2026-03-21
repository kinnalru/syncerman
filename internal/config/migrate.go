package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// LegacyConfig represents the old configuration format.
// It was defined as a map of providers, pointing to a map of source paths,
// which points to a slice of LegacyDestination.
type LegacyConfig map[string]map[string][]LegacyDestination

// LegacyDestination represents a destination in the old configuration format.
type LegacyDestination struct {
	To     string   `yaml:"to"`
	Args   []string `yaml:"args,omitempty"`
	Resync bool     `yaml:"resync,omitempty"`
}

// MigrateOldConfig reads an old format configuration file and writes it in the new job-centric format.
func MigrateOldConfig(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", filePath, err)
	}

	// We do a quick check to see if "jobs:" is present in the raw data, which indicates it's already new.
	if strings.Contains(string(data), "jobs:") {
		return fmt.Errorf("configuration file already contains 'jobs' key, skipping migration")
	}

	// Try to unmarshal into the old format.
	var oldConfig LegacyConfig
	err = yaml.Unmarshal(data, &oldConfig)
	if err != nil {
		return fmt.Errorf("failed to parse as old config format: %w", err)
	}

	if len(oldConfig) == 0 {
		return fmt.Errorf("old configuration is empty, nothing to migrate")
	}

	// To generate YAML with jobs as map, we need a wrapper struct since Config unmarshals differently
	type JobWrapper struct {
		Name     string `yaml:"name,omitempty"`
		Enabled  bool   `yaml:"enabled"`
		Priority int    `yaml:"priority,omitempty"`
		Tasks    []Task `yaml:"tasks"`
	}

	type ConfigWrapper struct {
		Jobs map[string]JobWrapper `yaml:"jobs"`
	}

	newConfig := ConfigWrapper{
		Jobs: make(map[string]JobWrapper),
	}

	for provider, sources := range oldConfig {
		jobName := fmt.Sprintf("%s_migration", provider)

		job := JobWrapper{
			Name:     fmt.Sprintf("Migrated from %s", provider),
			Enabled:  true,
			Priority: 10,
			Tasks:    []Task{},
		}

		for sourcePath, destinations := range sources {
			task := Task{
				From:    fmt.Sprintf("%s:%s", provider, sourcePath),
				Enabled: true,
				To:      make([]Destination, len(destinations)),
			}

			for i, dest := range destinations {
				task.To[i] = Destination{
					Path:   dest.To,
					Args:   dest.Args,
					Resync: dest.Resync,
				}
			}

			job.Tasks = append(job.Tasks, task)
		}

		newConfig.Jobs[jobName] = job
	}

	// Write back to the file
	newData, err := yaml.Marshal(&newConfig)
	if err != nil {
		return fmt.Errorf("failed to generate new configuration yaml: %w", err)
	}

	// Create a backup of the old file
	backupPath := filePath + ".bak"
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to create backup file %s: %w", backupPath, err)
	}

	if err := os.WriteFile(filePath, newData, 0644); err != nil {
		return fmt.Errorf("failed to write migrated config to %s: %w", filePath, err)
	}

	return nil
}
