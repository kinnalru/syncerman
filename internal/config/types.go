package config

type Destination struct {
	To     string   `yaml:"to"`
	Args   []string `yaml:"args"`
	Resync bool     `yaml:"resync"`
}

type PathMap map[string][]Destination

type ProviderMap map[string]PathMap

type Config struct {
	Providers ProviderMap `yaml:"-"`
}

func NewConfig() *Config {
	return &Config{
		Providers: make(ProviderMap),
	}
}

func (c *Config) AddProvider(name string, paths PathMap) {
	if c.Providers == nil {
		c.Providers = make(ProviderMap)
	}
	c.Providers[name] = paths
}

func (c *Config) GetProviders() ProviderMap {
	return c.Providers
}

func (c *Config) GetPaths(provider string) (PathMap, bool) {
	if c.Providers == nil {
		return nil, false
	}
	paths, ok := c.Providers[provider]
	return paths, ok
}

func (c *Config) GetDestinations(provider string, path string) ([]Destination, bool) {
	paths, ok := c.GetPaths(provider)
	if !ok {
		return nil, false
	}
	destinations, ok := paths[path]
	return destinations, ok
}

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

type SyncTarget struct {
	SourceProvider string
	SourcePath     string
	Destination    Destination
}
