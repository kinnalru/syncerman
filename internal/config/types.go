package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// PathWithKey preserves path name and its destination configurations.
// This structure is used internally to maintain YAML ordering while
// providing backward-compatible accessor methods.
type PathWithKey struct {
	Name   string        `yaml:"-"`
	Values []Destination `yaml:",inline"`
}

// PathMap maps source paths to their corresponding destination configurations.
//
// PathMap is a map type where each key is a source path string (relative to the provider's root)
// and the value is a slice of Destination objects representing all targets for that source path.
//
// Keys:
//   - source path: A string representing the path within the provider to sync from.
//     This can be a file path, directory path, or empty string for the root.
//     Examples: "", "documents", "projects/syncerman", "photos/vacation".
//
// Values:
//   - slice of Destination: A collection of Destination objects that define where and how
//     to synchronize the source path. Multiple destinations can be configured for a single
//     source path, allowing for redundant backups or multi-site distribution.
//
// Example:
//
//	PathMap{
//	    "documents": []Destination{
//	        {To: "gdrive:backup/docs", Args: []string{"--fast-list"}},
//	        {To: "local:/backup/documents", Resync: true},
//	    },
//	}
type PathMap map[string][]Destination

// OrderedPaths preserves path order from YAML configuration.
//
// This type implements yaml.Unmarshaler to ensure that when YAML is loaded,
// the paths are stored in the exact order they appear in the configuration file.
// This is critical for maintaining configuration order when multiple destinations
// are configured within the same provider.
//
// Example:
//
//	yamlData:
//	  provider1:
//	    "path1":  # Should execute 1st
//	      - to: "dest1"
//	    "path2":  # Should execute 2nd
//	      - to: "dest2"
//
//	After unmarshaling, OrderedPaths preserves [path1, path2] order
type OrderedPaths []PathWithKey

// UnmarshalYAML implements yaml.Unmarshaler to preserve path order from YAML.
//
// It parses the YAML mapping node and extracts path names in document order,
// preserving the exact order from the configuration file. This ensures that
// paths are iterated in the same order they appear in the configuration.
//
// If the YAML node is not a mapping, an error is returned.
// Path configurations are decoded as slices of Destination objects.
func (op *OrderedPaths) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("paths must be a mapping node")
	}

	*op = make(OrderedPaths, 0, len(node.Content)/2)

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		var destinations []Destination
		if err := valueNode.Decode(&destinations); err != nil {
			return fmt.Errorf("failed to decode path %s: %w", keyNode.Value, err)
		}

		*op = append(*op, PathWithKey{
			Name:   keyNode.Value,
			Values: destinations,
		})
	}

	return nil
}

// toPathMap converts OrderedPaths to PathMap for backward compatibility.
func (op OrderedPaths) toPathMap() PathMap {
	result := make(PathMap, len(op))
	for _, path := range op {
		result[path.Name] = path.Values
	}
	return result
}

// toOrderedPaths converts PathMap to OrderedPaths.
func toOrderedPaths(pm PathMap) OrderedPaths {
	result := make(OrderedPaths, 0, len(pm))
	for name, values := range pm {
		result = append(result, PathWithKey{
			Name:   name,
			Values: values,
		})
	}
	return result
}

// ProviderWithKey preserves provider name and its path configuration.
// This structure is used internally to maintain YAML ordering while
// providing backward-compatible accessor methods.
type ProviderWithKey struct {
	Name string       `yaml:"-"`
	Data OrderedPaths `yaml:",inline"`
}

// OrderedProviders preserves provider order from YAML configuration.
//
// This type implements yaml.Unmarshaler to ensure that when YAML is loaded,
// the providers are stored in the exact order they appear in the configuration file.
// This is critical for linear synchronization chains where execution order matters.
//
// Example:
//
//	yamlData:
//	  local:  # Should execute 1st
//	    '/path': [...]
//	  gd:     # Should execute 2nd
//	    '/path': [...]
//
//	After unmarshaling, OrderedProviders preserves [local, gd] order
type OrderedProviders []ProviderWithKey

// UnmarshalYAML implements yaml.Unmarshaler to preserve provider order from YAML.
//
// It parses the YAML mapping node and extracts provider names in document order,
// preserving the exact order from the configuration file. This ensures that
// linear synchronization chains execute in the correct sequence.
//
// If the YAML node is not a mapping, an error is returned.
// Provider configurations are decoded using the OrderedPaths type to preserve order.
func (op *OrderedProviders) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("providers must be a mapping node")
	}

	*op = make(OrderedProviders, 0, len(node.Content)/2)

	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valueNode := node.Content[i+1]

		var paths OrderedPaths
		if err := valueNode.Decode(&paths); err != nil {
			return fmt.Errorf("failed to decode provider %s: %w", keyNode.Value, err)
		}

		*op = append(*op, ProviderWithKey{
			Name: keyNode.Value,
			Data: paths,
		})
	}

	return nil
}

// Destination represents a sync destination configuration.
//
// Destination specifies where a source file or directory should be synchronized to,
// along with optional arguments and behavior flags for the sync operation.
//
// Fields:
//   - To: The destination path in format "provider:path" (e.g., "gdrive:backup") or a local filesystem path.
//     This field is required and specifies the target location for synchronization.
//   - Args: Optional additional rclone command-line arguments to pass during the sync operation.
//     Common examples include "--fast-list", "--max-age 30d", or "--exclude *.tmp".
//     This field is optional and defaults to an empty slice.
//   - Resync: Optional flag to force a complete synchronization from the source.
//     When true, the sync operation will be performed without any previous state checks.
//     This is useful for initial syncs or when corruption is suspected.
//     This field is optional and defaults to false.
//
// Example:
//
//	dest := Destination{
//	    To:     "gdrive:backup/documents",
//	    Args:   []string{"--ignore-existing"},
//	    Resync: true,
//	}
type Destination struct {
	To     string   `yaml:"to"`
	Args   []string `yaml:"args"`
	Resync bool     `yaml:"resync"`
}

// ProviderMap maps provider names to their path configurations.
//
// ProviderMap is a map type where each key is a provider name (matching rclone remote names)
// and the value is a PathMap containing all source paths and their destinations for that provider.
//
// Keys:
//   - provider name: A string representing the rclone remote name or local filesystem identifier.
//     This must match a configured remote in the rclone configuration file.
//     Examples: "gdrive", "dropbox", "onedrive", "local", or "s3:mybucket".
//
// Values:
//   - PathMap: A map of source paths to their destination configurations.
//     Each PathMap can contain multiple source paths, and each path can have multiple
//     destinations configured.
//
// Example:
//
//	ProviderMap{
//	    "gdrive": PathMap{
//	        "documents": []Destination{
//	            {To: "local:/backup/docs", Resync: true},
//	        },
//	    },
//	    "dropbox": PathMap{
//	        "files": []Destination{
//	            {To: "gdrive:backup/dropbox"},
//	        },
//	    },
//	}
type ProviderMap map[string]PathMap

// Config represents the main configuration structure for synchronization targets.
//
// The Providers field holds an ordered list of provider names and their path configurations.
// Each provider can have multiple source paths, and each path can have multiple destinations.
//
// The ordered list preserves the exact order from the YAML configuration file, which is
// critical for linear synchronization chains where execution order matters.
//
// The Config structure is validated to ensure all providers have valid path configurations
// before synchronization operations are performed.
//
// Use NewConfig() to create a new Config instance rather than creating one directly.
type Config struct {
	// Providers holds providers in YAML configuration order.
	// The yaml:"-" tag indicates that this field is not unmarshaled directly from YAML.
	// Instead, the LoadConfig and LoadConfigFromData functions manually unmarshal the
	// OrderedProviders to preserve configuration order and provide better error messages.
	Providers OrderedProviders `yaml:"-"`
}

// NewConfig creates and returns a new Config instance.
//
// It initializes an empty Providers slice, allowing providers to be added later
// using the AddProvider method.
//
// Returns:
//   - *Config: A pointer to the newly created Config instance
func NewConfig() *Config {
	return &Config{
		Providers: make(OrderedProviders, 0),
	}
}

// AddProvider adds a provider with its paths to the configuration.
//
// If the Providers slice is nil, it will be initialized before adding the provider.
// This method allows for dynamic configuration of providers and their associated paths.
// The provider is appended to the end of the ordered list.
//
// Parameters:
//   - name: The unique identifier for the provider
//   - paths: A PathMap containing the source paths and their destinations for this provider
func (c *Config) AddProvider(name string, paths PathMap) {
	if c.Providers == nil {
		c.Providers = make(OrderedProviders, 0)
	}
	c.Providers = append(c.Providers, ProviderWithKey{
		Name: name,
		Data: toOrderedPaths(paths),
	})
}

// GetProviders returns all configured providers in YAML order.
//
// This method provides access to the complete list of providers, which contains
// all providers and their associated path configurations in the exact order
// they appear in the YAML configuration file. This is critical for linear
// synchronization chains where execution order must match configuration order.
//
// Returns:
//   - OrderedProviders: A slice of all providers with their PathMap configurations
func (c *Config) GetProviders() OrderedProviders {
	return c.Providers
}

// GetProvidersMap returns a legacy map of all configured providers.
//
// DEPRECATED: Use GetProviders() instead for order-preserving access.
// This method is provided for backward compatibility only.
//
// Returns:
//   - ProviderMap: A map of all provider names to their PathMap configurations
func (c *Config) GetProvidersMap() ProviderMap {
	if c.Providers == nil {
		return make(ProviderMap)
	}
	result := make(ProviderMap, len(c.Providers))
	for _, provider := range c.Providers {
		result[provider.Name] = provider.Data.toPathMap()
	}
	return result
}

// GetPaths retrieves paths for a specific provider.
//
// This method looks up a provider by name and returns its associated path map.
// If the provider does not exist or the Providers slice is nil, it returns false
// as the second return value.
//
// Parameters:
//   - provider: The name of the provider to look up
//
// Returns:
//   - PathMap: The map of paths for the provider (or nil if not found)
//   - bool: True if the provider exists, false otherwise
func (c *Config) GetPaths(provider string) (PathMap, bool) {
	if c.Providers == nil {
		return nil, false
	}
	for _, p := range c.Providers {
		if p.Name == provider {
			return p.Data.toPathMap(), true
		}
	}
	return nil, false
}

// GetDestinations retrieves destinations for a specific provider and path.
//
// This method navigates through the provider hierarchy to find all destinations
// configured for a given provider's source path. If either the provider or path
// doesn't exist, it returns false as the second return value.
//
// Parameters:
//   - provider: The name of the provider
//   - path: The source path to retrieve destinations for
//
// Returns:
//   - []Destination: A slice of Destination objects (or nil if not found)
//   - bool: True if the provider and path exist, false otherwise
func (c *Config) GetDestinations(provider string, path string) ([]Destination, bool) {
	paths, ok := c.GetPaths(provider)
	if !ok {
		return nil, false
	}
	destinations, ok := paths[path]
	return destinations, ok
}

// countTotalTargets returns the total number of sync targets across all providers.
//
// This helper method iterates through the Providers slice counting all destinations
// for each provider and path combination. It's used by GetAllDestinations to
// pre-allocate the result slice for better performance.
//
// Returns:
//   - int: The total number of sync targets configured
func (c *Config) countTotalTargets() int {
	if c.Providers == nil {
		return 0
	}

	total := 0
	for _, provider := range c.Providers {
		for _, pathData := range provider.Data {
			total += len(pathData.Values)
		}
	}
	return total
}

// GetAllDestinations converts the configuration into a list of all sync targets.
//
// This method iterates through all configured providers, their source paths, and all destinations
// for each path, building a flat list of SyncTarget objects representing every configured sync operation.
//
// The implementation maintains the nested structure by flattening it:
//   - Level 1: Iterates through all providers in OrderedProviders (preserves YAML order)
//   - Level 2: For each provider, iterates through all source paths in OrderedPaths
//   - Level 3: For each source path, iterates through all destinations in the slice
//
// Returns:
//   - []SyncTarget: A slice containing all configured sync targets. Each SyncTarget contains
//     the source provider name, source path, and destination configuration. The order follows
//     the YAML configuration order thanks to OrderedProviders.
//
// Example usage:
//
//	config := NewConfig()
//	config.AddProvider("gdrive", PathMap{
//	    "documents": []Destination{{To: "local:/backup"}},
//	})
//	targets := config.GetAllDestinations()
//	// targets[0] will be a SyncTarget with SourceProvider: "gdrive",
//	// SourcePath: "documents", and Destination: {To: "local:/backup"}
func (c *Config) GetAllDestinations() []SyncTarget {
	total := c.countTotalTargets()
	targets := make([]SyncTarget, 0, total)

	for _, provider := range c.Providers {
		for _, pathData := range provider.Data {
			for _, dest := range pathData.Values {
				targets = append(targets, SyncTarget{
					SourceProvider: provider.Name,
					SourcePath:     pathData.Name,
					Destination:    dest,
				})
			}
		}
	}
	return targets
}

// SyncTarget represents a complete sync target configuration.
//
// SyncTarget encapsulates all information needed to perform a synchronization operation between
// a source provider and destination. It is typically created by flattening the nested configuration
// structure of ProviderMap -> PathMap -> Destination.
//
// Fields:
//   - SourceProvider: The name of the source provider (rclone remote name).
//     This identifies where the content is being synchronized from.
//     Examples: "gdrive", "dropbox", "onedrive", or "local".
//   - SourcePath: The path within the source provider to sync from.
//     This is the relative path from the provider's root, and can be empty for the root.
//     Examples: "", "documents", "projects/syncerman".
//   - Destination: The destination configuration specifying where to sync to.
//     This includes the destination path, optional arguments, and resync flag.
//
// Example:
//
//	target := SyncTarget{
//	    SourceProvider: "gdrive",
//	    SourcePath:     "documents",
//	    Destination: Destination{
//	        To:     "local:/backup/docs",
//	        Args:   []string{"--fast-list", "--ignore-existing"},
//	        Resync: true,
//	    },
//	}
type SyncTarget struct {
	SourceProvider string
	SourcePath     string
	Destination    Destination
}
