package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const JSON_CONFIG = ".eavesdrop.json"

// GenerateJsonConfig saves a copy of the default Config struct as a json file in the given output directory.
func GenerateJsonConfig(outPath string) error {
	path := filepath.Join(outPath, JSON_CONFIG)
	config, err := json.MarshalIndent(DefaultConfig(path), "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal config to json: %w", err)
	}

	if err = os.WriteFile(path, config, 0644); err != nil {
		return fmt.Errorf("failed to write config to json file: %w", err)
	}

	return nil
}

// ReadJsonConfig reads the data in the given file, unmarshals it to a *Config struct, and returns it.
func ReadJsonConfig(inPath string) (*Config, error) {
	file, err := os.ReadFile(inPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read json config: %w", err)
	}

	config := &Config{cfg: inPath}
	if err = json.Unmarshal(file, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json to config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	path, err := filepath.Abs(config.Root)
	if err != nil {
		return nil, fmt.Errorf("could not determine absolute path for root: %v", err)
	}
	config.Root = path

	return config, nil
}
