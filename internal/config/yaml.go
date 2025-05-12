package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const YAML_CONFIG = "eavesdrop.yaml"

// GenerateYamlConfig saves a copy of the default Config struct as a yaml file in the given output directory.
func GenerateYamlConfig(outPath string) error {
	path := filepath.Join(outPath, YAML_CONFIG)
	config, err := yaml.Marshal(DefaultConfig(path))
	if err != nil {
		return fmt.Errorf("failed to marshal config to yaml: %w", err)
	}

	// save the default config to the given output direcory
	if err = os.WriteFile(path, config, 0644); err != nil {
		return fmt.Errorf("failed to write config to yaml file: %w", err)
	}

	return nil
}

// ReadYamlConfig reads the data in the given file, unmarshals it to a *Config struct, and returns it.
func ReadYamlConfig(inPath string) (*Config, error) {
	// read in the given file path
	file, err := os.ReadFile(inPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read yaml config: %w", err)
	} else if len(file) == 0 {
		return nil, fmt.Errorf("yaml config is empty")
	}

	config := &Config{cfg: inPath}
	if err = yaml.Unmarshal(file, config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml to config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	return config, nil
}
