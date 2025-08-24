package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load loads the configuration from the given path.
// If the file doesn't exist, it creates a default config file and returns it.
func Load(path string) (*Config, error) {
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %v", err)
		}
		path = filepath.Join(homeDir, ".gotrack", "config.yaml")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := Default()
			if err := cfg.Save(path); err != nil {
				return nil, fmt.Errorf("failed to create default config: %v", err)
			}
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &cfg, nil
}

// Save saves the configuration to the given path.
// If no path is provided, it uses the default config path.
func (c *Config) Save(path string) error {
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(homeDir, ".gotrack", "config.yaml")

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
