package config

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
// The Providers field holds a map of provider names to their associated path configurations.
// Each provider can have multiple source paths, and each path can have multiple destinations.
//
// The Config structure is validated to ensure all providers have valid path configurations
// before synchronization operations are performed.
//
// Use NewConfig() to create a new Config instance rather than creating one directly.
type Config struct {
	Providers ProviderMap `yaml:"-"`
}

// NewConfig creates and returns a new Config instance.
//
// It initializes an empty Providers map, allowing providers to be added later
// using the AddProvider method.
//
// Returns:
//   - *Config: A pointer to the newly created Config instance
func NewConfig() *Config {
	return &Config{
		Providers: make(ProviderMap),
	}
}

// AddProvider adds a provider with its paths to the configuration.
//
// If the Providers map is nil, it will be initialized before adding the provider.
// This method allows for dynamic configuration of providers and their associated paths.
//
// Parameters:
//   - name: The unique identifier for the provider
//   - paths: A PathMap containing the source paths and their destinations for this provider
func (c *Config) AddProvider(name string, paths PathMap) {
	if c.Providers == nil {
		c.Providers = make(ProviderMap)
	}
	c.Providers[name] = paths
}

// GetProviders returns all configured providers.
//
// This method provides access to the complete ProviderMap, which contains
// all providers and their associated path configurations.
//
// Returns:
//   - ProviderMap: A map of all provider names to their PathMap configurations
func (c *Config) GetProviders() ProviderMap {
	return c.Providers
}

// GetPaths retrieves paths for a specific provider.
//
// This method looks up a provider by name and returns its associated path map.
// If the provider does not exist or the Providers map is nil, it returns false
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
	paths, ok := c.Providers[provider]
	return paths, ok
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

// GetAllDestinations converts the configuration into a list of all sync targets.
//
// This method iterates through all configured providers, their source paths, and all destinations
// for each path, building a flat list of SyncTarget objects representing every configured sync operation.
//
// The implementation maintains the nested structure by flattening it:
//   - Level 1: Iterates through all providers in ProviderMap
//   - Level 2: For each provider, iterates through all source paths in PathMap
//   - Level 3: For each source path, iterates through all destinations in the slice
//
// Returns:
//   - []SyncTarget: A slice containing all configured sync targets. Each SyncTarget contains
//     the source provider name, source path, and destination configuration. The order follows
//     the iteration order of the underlying maps, which is not guaranteed to be stable.
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
	var targets []SyncTarget
	for provider, paths := range c.Providers {
		for path, destinations := range paths {
			for _, dest := range destinations {
				targets = append(targets, SyncTarget{
					SourceProvider: provider,
					SourcePath:     path,
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
