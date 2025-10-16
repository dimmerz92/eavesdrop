package eavesdrop

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

const (
	JSON_CONFIG = ".eavesdrop.json"
	TOML_CONFIG = ".eavesdrop.toml"
	YAML_CONFIG = ".eavesdrop.yaml"
)

type Config struct {
	RootDir    string          `json:"root_dir" toml:"root_dir" yaml:"root_dir"`
	Tmp        bool            `json:"tmp" toml:"tmp" yaml:"tmp"`
	CleanupTmp bool            `json:"cleanup_tmp" toml:"cleanup_tmp" yaml:"cleanup_tmp"`
	Exclude    ExcluderConfig  `json:"exclude" toml:"exclude" yaml:"exclude"`
	Watchers   []WatcherConfig `json:"watchers" toml:"watchers" yaml:"watchers"`
	Proxy      ProxyConfig     `json:"proxy" toml:"proxy" yaml:"proxy"`
}

// DefaultConfig returns a config initialised to helpful defaults.
func DefaultConfig() Config {
	return Config{
		RootDir: ".",
		Exclude: ExcluderConfig{
			Dirs:  []string{"data", "dist", "node_modules", "tmp"},
			Files: []string{},
			Regex: []string{
				`^\.?(\/?|\\?)(?:\w+(\/|\\))*(\.\w+)$`, // dotfiles on windows or unix at any hierarchy
				`^.+\.sqlite$`, `^.+\.wal$`, `^.+\.shm$`,
			},
		},
		Watchers: []WatcherConfig{{
			Name:      "watcher",
			FileTypes: []string{},
			FileNames: []string{},
			Exclude: ExcluderConfig{
				Dirs:  []string{},
				Files: []string{},
				Regex: []string{},
			},
			Tasks:             []string{},
			Service:           "",
			RunOnStart:        true,
			MaxTaskTime:       2000,
			MaxServiceTimeout: 5000,
			DebounceDelay:     300,
			TriggerRefresh:    false,
		}},
		Proxy: ProxyConfig{
			Enabled:   false,
			AppPort:   8000,
			ProxyPort: 8001,
		},
	}
}

// Validate checks to make sure the Config fields are valid.
func (c *Config) Validate() error {
	if c.RootDir == "" {
		return fmt.Errorf("root directory required. use '.' for the current working directory")
	}

	watchers := make(map[string]struct{})
	for _, watcher := range c.Watchers {
		err := watcher.Validate()
		if err != nil {
			return err
		}

		if _, ok := watchers[watcher.Name]; ok {
			return fmt.Errorf("two watchers with the same name: %s", watcher.Name)
		}

		watchers[watcher.Name] = struct{}{}
	}

	return c.Proxy.Validate()
}

// GetConfig returns the config for the given path or the default config if the path is an empty string.
func GetConfig(path string) (Config, error) {
	var err error
	var config Config
	switch ext := filepath.Ext(path); ext {
	case ".json":
		config, err = ReadJsonConfig(path)
	case ".yaml":
		config, err = ReadYamlConfig(path)
	case ".toml":
		config, err = ReadTomlConfig(path)
	default:
		err = fmt.Errorf("please use .json, .yaml, or .toml, not %s", ext)
	}

	return config, err
}

// GenerateConfig creates a default config and saves it at the given path and ext. If no ext defined, defaults to json.
func GenerateConfig(path, ext string) error {
	var err error
	switch ext {
	case ".json", "json", "":
		err = GenerateJsonConfig(path)
	case ".yaml", "yaml", ".yml", "yml":
		err = GenerateYamlConfig(path)
	case ".toml", "toml":
		err = GenerateTomlConfig(path)
	default:
		err = fmt.Errorf("invalid extension: %s", ext)
	}

	return err
}

// GenerateJsonConfig saves a copy of the default Config struct as a json file in the given output directory.
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

// ReadJsonConfig reads the data in the given file, unmarshals it to a *Config struct, and returns it.
func ReadJsonConfig(inPath string) (Config, error) {
	file, err := os.ReadFile(inPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read json config: %w", err)
	}

	config := Config{}
	err = json.Unmarshal(file, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal json to config: %w", err)
	}

	err = config.Validate()
	if err != nil {
		return Config{}, fmt.Errorf("validation error: %w", err)
	}

	return config, nil
}

// GenerateTomlConfig saves a copy of the default Config struct as a toml file in the given output directory.
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

// ReadTomlConfig reads the data in the given file, unmarshals it to a *Config struct, and returns it.
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

	err = config.Validate()
	if err != nil {
		return Config{}, fmt.Errorf("validation error: %w", err)
	}

	return config, nil
}

// GenerateYamlConfig saves a copy of the default Config struct as a yaml file in the given output directory.
func GenerateYamlConfig(outPath string) error {
	path := filepath.Join(outPath, YAML_CONFIG)

	config, err := yaml.Marshal(DefaultConfig())
	if err != nil {
		return fmt.Errorf("failed to marshal config to yaml: %w", err)
	}

	// save the default config to the given output direcory
	err = os.WriteFile(path, config, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config to yaml file: %w", err)
	}

	return nil
}

// ReadYamlConfig reads the data in the given file, unmarshals it to a *Config struct, and returns it.
func ReadYamlConfig(inPath string) (Config, error) {
	// read in the given file path
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

	err = config.Validate()
	if err != nil {
		return Config{}, fmt.Errorf("validation error: %w", err)
	}

	return config, nil
}
