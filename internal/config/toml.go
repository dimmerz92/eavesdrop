package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const TOML_CONFIG = "eavesdrop.toml"

func GenerateTomlConfig(outPath string) error {
	path := filepath.Join(outPath, TOML_CONFIG)

	config, err := toml.Marshal(DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to marshal config to toml: %w", err)
	}

	err = os.WriteFile(path, config, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config to toml file: %w", err)
	}

	return nil
}

func ReadTomlConfig(inPath string) (Config, error) {
	file, err := os.ReadFile(inPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read toml config: %w", err)
	} else if len(file) == 0 {
		return Config{}, fmt.Errorf("toml config is empty")
	}

	config := Config{}
	err = toml.Unmarshal(file, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal toml to config: %w", err)
	}

	return config, nil
}
