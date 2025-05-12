package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const TOML_CONFIG = "eavesdrop.toml"

// GenerateTomlConfig saves a copy of the default Config struct as a toml file in the given output directory.
func GenerateTomlConfig(outPath string) error {
	path := filepath.Join(outPath, TOML_CONFIG)
	config, err := toml.Marshal(DefaultConfig(path))
	if err != nil {
		return fmt.Errorf("failed to marshal config to toml: %w", err)
	}

	if err = os.WriteFile(path, config, 0644); err != nil {
		return fmt.Errorf("failed to write config to toml file: %w", err)
	}

	return nil
}

// ReadTomlConfig reads the data in the given file, unmarshals it to a *Config struct, and returns it.
func ReadTomlConfig(inPath string) (*Config, error) {
	file, err := os.ReadFile(inPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read toml config: %w", err)
	} else if len(file) == 0 {
		return nil, fmt.Errorf("toml config is empty")
	}

	config := &Config{cfg: inPath}
	if err = toml.Unmarshal(file, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal toml to config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return config, nil
}
