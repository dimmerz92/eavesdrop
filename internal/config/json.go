package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const JSON_CONFIG = "eavesdrop.json"

func GenerateJsonConfig(outPath string) error {
	path := filepath.Join(outPath, JSON_CONFIG)

	config, err := json.MarshalIndent(DefaultConfig(), "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal config to json: %w", err)
	}

	err = os.WriteFile(path, config, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config to json file: %w", err)
	}

	return nil
}

func ReadJsonConfig(inPath string) (Config, error) {
	file, err := os.ReadFile(inPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read json config: %w", err)
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal json to config: %w", err)
	}

	return config, nil
}
