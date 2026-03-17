package config

import (
	"testing"
)

func parseTestConfig(yamlData string) (*Config, error) {
	providers, err := parseProviders([]byte(yamlData))
	if err != nil {
		return nil, err
	}

	cfg := NewConfig()
	cfg.Providers = providers
	return cfg, nil
}

func TestConfigProvidersOrderByYAML(t *testing.T) {
	yamlData := `
gd:
  "path1":
    - to: "remote:path1"
local:
  "path2":
    - to: "remote:path2"
`

	providers, err := parseProviders([]byte(yamlData))
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	cfg := NewConfig()
	cfg.Providers = providers

	providersFromCfg := cfg.GetProviders()

	if len(providersFromCfg) != 2 {
		t.Fatalf("Expected 2 providers, got %d", len(providersFromCfg))
	}

	if providersFromCfg[0].Name != "gd" {
		t.Errorf("Expected first provider to be 'gd', got '%s'", providersFromCfg[0].Name)
	}

	if providersFromCfg[1].Name != "local" {
		t.Errorf("Expected second provider to be 'local', got '%s'", providersFromCfg[1].Name)
	}
}

func TestConfigPathsOrderByYAML(t *testing.T) {
	yamlData := `
provider1:
  "pathA":
    - to: "remote:destA"
  "pathB":
    - to: "remote:destB"
  "pathC":
    - to: "remote:destC"
`

	providers, err := parseProviders([]byte(yamlData))
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	cfg := NewConfig()
	cfg.Providers = providers

	providersFromCfg := cfg.GetProviders()
	if len(providersFromCfg) != 1 {
		t.Fatalf("Expected 1 provider, got %d", len(providersFromCfg))
	}

	paths := providersFromCfg[0].Data

	expectedPaths := []string{"pathA", "pathB", "pathC"}
	if len(paths) != len(expectedPaths) {
		t.Fatalf("Expected %d paths, got %d", len(expectedPaths), len(paths))
	}

	for i, pathData := range paths {
		if i >= len(expectedPaths) {
			t.Fatalf("Path iteration length mismatch")
		}
		if pathData.Name != expectedPaths[i] {
			t.Errorf("Expected path %d to be '%s', got '%s'", i+1, expectedPaths[i], pathData.Name)
		}
	}
}

func TestConfigGetAllDestinationsOrderByYAML(t *testing.T) {
	yamlData := `
local:
  "/path/to/local":
    - to: "gd:syncerman/scenario1/"
gd:
  "syncerman/scenario1/":
    - to: "yd:syncerman/scenario1/"
yd:
  "syncerman/scenario1/":
    - to: "/path/to/local2"
`

	providers, err := parseProviders([]byte(yamlData))
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	cfg := NewConfig()
	cfg.Providers = providers

	targets := cfg.GetAllDestinations()

	if len(targets) != 3 {
		t.Fatalf("Expected 3 targets, got %d", len(targets))
	}

	expectedTargets := []struct {
		provider string
		path     string
		dest     string
	}{
		{"local", "/path/to/local", "gd:syncerman/scenario1/"},
		{"gd", "syncerman/scenario1/", "yd:syncerman/scenario1/"},
		{"yd", "syncerman/scenario1/", "/path/to/local2"},
	}

	for i, expected := range expectedTargets {
		if i >= len(targets) {
			t.Fatalf("Target iteration length mismatch at index %d", i)
		}

		if targets[i].SourceProvider != expected.provider {
			t.Errorf("Target %d: expected provider '%s', got '%s'", i, expected.provider, targets[i].SourceProvider)
		}

		if targets[i].SourcePath != expected.path {
			t.Errorf("Target %d: expected path '%s', got '%s'", i, expected.path, targets[i].SourcePath)
		}

		if targets[i].Destination.To != expected.dest {
			t.Errorf("Target %d: expected destination '%s', got '%s'", i, expected.dest, targets[i].Destination.To)
		}
	}
}

func TestLinearTwoHopChain(t *testing.T) {
	yamlData := `
localA:
  "/source":
    - to: "remoteA:/backup"
remoteA:
  "/backup":
    - to: "remoteB:/archive"
`

	cfg, err := parseTestConfig(yamlData)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	targets := cfg.GetAllDestinations()

	if len(targets) != 2 {
		t.Fatalf("Expected 2 targets for 2-hop chain, got %d", len(targets))
	}

	if targets[0].SourceProvider != "localA" {
		t.Errorf("Expected first target to be from 'localA', got '%s'", targets[0].SourceProvider)
	}

	if targets[0].Destination.To != "remoteA:/backup" {
		t.Errorf("Expected first target destination to be 'remoteA:/backup', got '%s'", targets[0].Destination.To)
	}

	if targets[1].SourceProvider != "remoteA" {
		t.Errorf("Expected second target to be from 'remoteA', got '%s'", targets[1].SourceProvider)
	}

	if targets[1].Destination.To != "remoteB:/archive" {
		t.Errorf("Expected second target destination to be 'remoteB:/archive', got '%s'", targets[1].Destination.To)
	}
}

func TestLinearThreeHopChain(t *testing.T) {
	yamlData := `
local:
  "/path/to/local":
    - to: "gd:syncerman/scenario1/"
gd:
  "syncerman/scenario1/":
    - to: "yd:syncerman/scenario1/"
yd:
  "syncerman/scenario1/":
    - to: "/path/to/local2"
`

	cfg, err := parseTestConfig(yamlData)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	targets := cfg.GetAllDestinations()

	if len(targets) != 3 {
		t.Fatalf("Expected 3 targets for 3-hop chain, got %d", len(targets))
	}

	expectedSequence := []struct {
		from string
		to   string
	}{
		{"local", "gd:syncerman/scenario1/"},
		{"gd", "yd:syncerman/scenario1/"},
		{"yd", "/path/to/local2"},
	}

	for i, expected := range expectedSequence {
		if targets[i].SourceProvider != expected.from {
			t.Errorf("Hop %d: expected source '%s', got '%s'", i+1, expected.from, targets[i].SourceProvider)
		}
		if targets[i].Destination.To != expected.to {
			t.Errorf("Hop %d: expected destination '%s', got '%s'", i+1, expected.to, targets[i].Destination.To)
		}
	}
}

func TestMultipleDestinationsSamePath(t *testing.T) {
	yamlData := `
local:
  "/documents":
    - to: "remoteA:/backup/docs"
    - to: "remoteB:/backup/docs"
    - to: "remoteC:/backup/docs"
`

	cfg, err := parseTestConfig(yamlData)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	targets := cfg.GetAllDestinations()

	if len(targets) != 3 {
		t.Fatalf("Expected 3 destinations, got %d", len(targets))
	}

	expectedDests := []string{"remoteA:/backup/docs", "remoteB:/backup/docs", "remoteC:/backup/docs"}
	for i, expectedDest := range expectedDests {
		if targets[i].Destination.To != expectedDest {
			t.Errorf("Destination %d: expected '%s', got '%s'", i+1, expectedDest, targets[i].Destination.To)
		}
	}
}

func TestGetPathsReturnsCorrectPath(t *testing.T) {
	yamlData := `
provider1:
  "path1":
    - to: "remote:dest1"
  "path2":
    - to: "remote:dest2"
provider2:
  "path3":
    - to: "remote:dest3"
`

	cfg, err := parseTestConfig(yamlData)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	paths, ok := cfg.GetPaths("provider1")
	if !ok {
		t.Fatalf("Expected provider1 to exist")
	}

	if len(paths) != 2 {
		t.Fatalf("Expected 2 paths for provider1, got %d", len(paths))
	}

	if _, ok := paths["path1"]; !ok {
		t.Errorf("Expected path 'path1' to exist in provider1")
	}

	if _, ok := paths["path2"]; !ok {
		t.Errorf("Expected path 'path2' to exist in provider1")
	}
}

func TestGetDestinationsReturnsCorrectDestinations(t *testing.T) {
	yamlData := `
provider1:
  "path1":
    - to: "remote:dest1"
    - to: "remote:dest2"
`

	cfg, err := parseTestConfig(yamlData)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	dests, ok := cfg.GetDestinations("provider1", "path1")
	if !ok {
		t.Fatalf("Expected provider1:path1 to exist")
	}

	if len(dests) != 2 {
		t.Fatalf("Expected 2 destinations, got %d", len(dests))
	}

	if dests[0].To != "remote:dest1" {
		t.Errorf("Expected first destination to be 'remote:dest1', got '%s'", dests[0].To)
	}

	if dests[1].To != "remote:dest2" {
		t.Errorf("Expected second destination to be 'remote:dest2', got '%s'", dests[1].To)
	}
}

func TestOrderedProviders_TableDrivenTests(t *testing.T) {
	tests := []struct {
		name          string
		yamlData      string
		expectedOrder []struct {
			provider string
			path     string
			dest     string
		}
	}{
		{
			name: "two providers in alphabetical order",
			yamlData: `
alpha:
  "path1":
    - to: "dest1"
beta:
  "path2":
    - to: "dest2"
`,
			expectedOrder: []struct {
				provider string
				path     string
				dest     string
			}{
				{"alpha", "path1", "dest1"},
				{"beta", "path2", "dest2"},
			},
		},
		{
			name: "two providers in reverse alphabetical order",
			yamlData: `
zeta:
  "path1":
    - to: "dest1"
alpha:
  "path2":
    - to: "dest2"
`,
			expectedOrder: []struct {
				provider string
				path     string
				dest     string
			}{
				{"zeta", "path1", "dest1"},
				{"alpha", "path2", "dest2"},
			},
		},
		{
			name: "single provider with multiple paths",
			yamlData: `
provider1:
  "path1":
    - to: "dest1"
  "path2":
    - to: "dest2"
  "path3":
    - to: "dest3"
`,
			expectedOrder: []struct {
				provider string
				path     string
				dest     string
			}{
				{"provider1", "path1", "dest1"},
				{"provider1", "path2", "dest2"},
				{"provider1", "path3", "dest3"},
			},
		},
		{
			name: "multiple providers with single path each",
			yamlData: `
providerA:
  "path":
    - to: "destA"
providerB:
  "path":
    - to: "destB"
providerC:
  "path":
    - to: "destC"
`,
			expectedOrder: []struct {
				provider string
				path     string
				dest     string
			}{
				{"providerA", "path", "destA"},
				{"providerB", "path", "destB"},
				{"providerC", "path", "destC"},
			},
		},
		{
			name: "five providers complex chain",
			yamlData: `
p1:
  "path":
    - to: "p2:path"
p2:
  "path":
    - to: "p3:path"
p3:
  "path":
    - to: "p4:path"
p4:
  "path":
    - to: "p5:path"
p5:
  "path":
    - to: "final:path"
`,
			expectedOrder: []struct {
				provider string
				path     string
				dest     string
			}{
				{"p1", "path", "p2:path"},
				{"p2", "path", "p3:path"},
				{"p3", "path", "p4:path"},
				{"p4", "path", "p5:path"},
				{"p5", "path", "final:path"},
			},
		},
		{
			name: "provider with multiple destinations",
			yamlData: `
provider:
  "path":
    - to: "dest1"
    - to: "dest2"
    - to: "dest3"
`,
			expectedOrder: []struct {
				provider string
				path     string
				dest     string
			}{
				{"provider", "path", "dest1"},
				{"provider", "path", "dest2"},
				{"provider", "path", "dest3"},
			},
		},
		{
			name: "mixed local and remote providers",
			yamlData: `
local:
  "/path":
    - to: "remote:backup"
remote:
  "backup":
    - to: "s3:archive"
s3:
  "archive":
    - to: "/final/backup"
`,
			expectedOrder: []struct {
				provider string
				path     string
				dest     string
			}{
				{"local", "/path", "remote:backup"},
				{"remote", "backup", "s3:archive"},
				{"s3", "archive", "/final/backup"},
			},
		},
		{
			name: "single provider single destination",
			yamlData: `
provider:
  "path":
    - to: "dest"
`,
			expectedOrder: []struct {
				provider string
				path     string
				dest     string
			}{
				{"provider", "path", "dest"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseTestConfig(tt.yamlData)
			if err != nil {
				t.Fatalf("Failed to unmarshal YAML: %v", err)
			}

			targets := cfg.GetAllDestinations()

			if len(targets) != len(tt.expectedOrder) {
				t.Fatalf("Expected %d targets, got %d", len(tt.expectedOrder), len(targets))
			}

			for i, expected := range tt.expectedOrder {
				if i >= len(targets) {
					t.Fatalf("Target iteration length mismatch at index %d", i)
				}

				if targets[i].SourceProvider != expected.provider {
					t.Errorf("Target %d: expected provider '%s', got '%s'", i, expected.provider, targets[i].SourceProvider)
				}

				if targets[i].SourcePath != expected.path {
					t.Errorf("Target %d: expected path '%s', got '%s'", i, expected.path, targets[i].SourcePath)
				}

				if targets[i].Destination.To != expected.dest {
					t.Errorf("Target %d: expected destination '%s', got '%s'", i, expected.dest, targets[i].Destination.To)
				}
			}
		})
	}
}

func TestGetDestinations_NonExistentProvider(t *testing.T) {
	cfg := NewConfig()
	cfg.AddProvider("provider1", PathMap{
		"path1": []Destination{{To: "dest1"}},
	})

	dests, ok := cfg.GetDestinations("nonexistent", "path")
	if ok {
		t.Errorf("Expected false for nonexistent provider, got true")
	}

	if dests != nil {
		t.Errorf("Expected nil destinations for nonexistent provider, got %v", dests)
	}
}

func TestGetDestinations_NonExistentPath(t *testing.T) {
	cfg := NewConfig()
	cfg.AddProvider("provider1", PathMap{
		"path1": []Destination{{To: "dest1"}},
	})

	dests, ok := cfg.GetDestinations("provider1", "nonexistent")
	if ok {
		t.Errorf("Expected false for nonexistent path, got true")
	}

	if dests != nil {
		t.Errorf("Expected nil destinations for nonexistent path, got %v", dests)
	}
}

func TestGetPaths_NonExistentProvider(t *testing.T) {
	cfg := NewConfig()
	cfg.AddProvider("provider1", PathMap{
		"path1": []Destination{{To: "dest1"}},
	})

	paths, ok := cfg.GetPaths("nonexistent")
	if ok {
		t.Errorf("Expected false for nonexistent provider, got true")
	}

	if paths != nil {
		t.Errorf("Expected nil paths for nonexistent provider, got %v", paths)
	}
}
