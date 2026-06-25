package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const YAML_CONFIG = "eavesdrop.yaml"

func GenerateYamlConfig(outPath string) error {
	path := filepath.Join(outPath, YAML_CONFIG)

	config, err := yaml.Marshal(DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to marshal config to yaml: %w", err)
	}

	err = os.WriteFile(path, config, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config to yaml file: %w", err)
	}

	return nil
}

func ReadYamlConfig(inPath string) (Config, error) {
	file, err := os.ReadFile(inPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read yaml config: %w", err)
	} else if len(file) == 0 {
		return Config{}, fmt.Errorf("yaml config is empty")
	}

	config := Config{}
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal yaml to config: %w", err)
	}

	return config, nil
}
