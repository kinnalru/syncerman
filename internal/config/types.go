package config

import (
	"sort"

	"gopkg.in/yaml.v3"
)

// Destination represents a single target path with sync options.
type Destination struct {
	Path   string   `yaml:"path"`
	Args   []string `yaml:"args"`
	Resync bool     `yaml:"resync"`
}

// Task represents a synchronization task from a single source to multiple destinations.
type Task struct {
	From    string        `yaml:"from"`
	To      []Destination `yaml:"to"`
	Enabled bool          `yaml:"enabled"`
}

// UnmarshalYAML implements custom unmarshaling to provide default values for Task.
func (t *Task) UnmarshalYAML(node *yaml.Node) error {
	type rawTask Task
	raw := rawTask{
		Enabled: true,
	}
	if err := node.Decode(&raw); err != nil {
		return err
	}
	*t = Task(raw)
	return nil
}

// Job represents a group of tasks that should be executed together.
type Job struct {
	ID       string `yaml:"-"`
	Name     string `yaml:"name"`
	Enabled  bool   `yaml:"enabled"`
	Priority int    `yaml:"priority"`
	Tasks    []Task `yaml:"tasks"`
}

// UnmarshalYAML implements custom unmarshaling to provide default values for Job.
func (j *Job) UnmarshalYAML(node *yaml.Node) error {
	type rawJob Job
	raw := rawJob{
		Enabled:  true,
		Priority: 10,
	}
	if err := node.Decode(&raw); err != nil {
		return err
	}
	*j = Job(raw)
	return nil
}

// Config represents the main configuration structure containing all jobs.
type Config struct {
	Jobs []Job `yaml:"-"`
}

// NewConfig creates and returns a new empty Config instance.
func NewConfig() *Config {
	return &Config{
		Jobs: make([]Job, 0),
	}
}

// UnmarshalYAML implements custom unmarshaling to map the "jobs" key to a sorted slice.
func (c *Config) UnmarshalYAML(node *yaml.Node) error {
	var raw struct {
		Jobs map[string]yaml.Node `yaml:"jobs"`
	}
	if err := node.Decode(&raw); err != nil {
		return err
	}

	for id, jobNode := range raw.Jobs {
		var job Job
		if err := jobNode.Decode(&job); err != nil {
			return err
		}
		job.ID = id
		if job.Name == "" {
			job.Name = id
		}
		c.Jobs = append(c.Jobs, job)
	}

	// Sort jobs by priority ascending.
	// We use string ID for stable sorting if priorities are equal.
	sort.SliceStable(c.Jobs, func(i, j int) bool {
		if c.Jobs[i].Priority == c.Jobs[j].Priority {
			return c.Jobs[i].ID < c.Jobs[j].ID
		}
		return c.Jobs[i].Priority < c.Jobs[j].Priority
	})

	return nil
}

// GetJobs returns all jobs sorted by priority.
func (c *Config) GetJobs() []Job {
	return c.Jobs
}
