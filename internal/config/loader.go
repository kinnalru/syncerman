package config

import (
	"os"

	"gopkg.in/yaml.v3"
	"syncerman/internal/errors"
)

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.NewConfigError("failed to read configuration file", err)
	}

	var providerMap ProviderMap
	if err := yaml.Unmarshal(data, &providerMap); err != nil {
		return nil, errors.NewConfigError("failed to parse configuration file", err)
	}

	config := NewConfig()
	config.Providers = providerMap

	return config, nil
}

func LoadConfigFromData(data []byte) (*Config, error) {
	var providerMap ProviderMap
	if err := yaml.Unmarshal(data, &providerMap); err != nil {
		return nil, errors.NewConfigError("failed to parse configuration data", err)
	}

	config := NewConfig()
	config.Providers = providerMap

	return config, nil
}
