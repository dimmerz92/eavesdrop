package eavesdrop_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/dimmerz92/eavesdrop"
	"gopkg.in/yaml.v3"
)

func generateConfig() eavesdrop.Config {
	config := eavesdrop.DefaultConfig()
	config.Watchers[0].FileTypes = []string{".go"}
	config.Watchers[0].Tasks = []string{"echo hello"}

	return config
}

func TestDefaultConfig(t *testing.T) {
	config := eavesdrop.DefaultConfig()

	if config.RootDir != "." {
		t.Fatalf("expected RootDir to be '.', got '%s'", config.RootDir)
	}

	if len(config.Watchers) == 0 {
		t.Fatalf("expected at least one watcher in default config")
	}
}

func TestConfig_Validate(t *testing.T) {
	config := generateConfig()

	err := config.Validate()
	if err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}

	config.RootDir = ""
	err = config.Validate()
	if err == nil {
		t.Fatal("expected error for empty RootDir, got nil")
	}
}

func TestJSON(t *testing.T) {
	tmp := t.TempDir()

	t.Run("test generate", func(t *testing.T) {
		err := eavesdrop.GenerateJsonConfig(tmp)
		if err != nil {
			t.Fatalf("failed to generate json config: %v", err)
		}

		path := filepath.Join(tmp, eavesdrop.JSON_CONFIG)
		info, err := os.Stat(path)
		if err != nil || info.Size() == 0 {
			t.Fatal("expected json config to be written and non-empty")
		}
	})

	t.Run("test read", func(t *testing.T) {
		generated := generateConfig()

		data, err := json.MarshalIndent(generated, "", "\t")
		if err != nil {
			t.Fatalf("failed to marshal config to json: %v", err)
		}

		path := filepath.Join(tmp, eavesdrop.JSON_CONFIG)

		err = os.WriteFile(path, data, 0644)
		if err != nil {
			t.Fatalf("failed to save json: %v", err)
		}

		config, err := eavesdrop.ReadJsonConfig(path)
		if err != nil {
			t.Fatalf("failed to read json config: %v", err)
		}

		if !reflect.DeepEqual(generated, config) {
			t.Fatalf("expected\n%#v\n\ngot\n%#v", generated, config)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := eavesdrop.ReadJsonConfig("nonexistent.json")
		if err == nil || !strings.Contains(err.Error(), "failed to read json config") {
			t.Errorf("expected read error, got: %v", err)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "empty.json")

		err := os.WriteFile(path, []byte(""), 0644)
		if err != nil {
			t.Fatalf("failed to write empty json file: %v", err)
		}

		_, err = eavesdrop.ReadJsonConfig(path)
		if err == nil || !strings.Contains(err.Error(), "failed to unmarshal json to config") {
			t.Errorf("expected empty file error, got: %v", err)
		}
	})

}

func TestTOML(t *testing.T) {
	tmp := t.TempDir()

	t.Run("test generate", func(t *testing.T) {
		err := eavesdrop.GenerateTomlConfig(tmp)
		if err != nil {
			t.Fatalf("failed to generate toml config: %v", err)
		}

		path := filepath.Join(tmp, eavesdrop.TOML_CONFIG)
		info, err := os.Stat(path)
		if err != nil || info.Size() == 0 {
			t.Fatal("expected toml config to be written and non-empty")
		}
	})

	t.Run("test read", func(t *testing.T) {
		generated := generateConfig()

		data, err := toml.Marshal(generated)
		if err != nil {
			t.Fatalf("failed to marshal config to toml: %v", err)
		}

		path := filepath.Join(tmp, eavesdrop.TOML_CONFIG)

		err = os.WriteFile(path, data, 0644)
		if err != nil {
			t.Fatalf("failed to save toml: %v", err)
		}

		config, err := eavesdrop.ReadTomlConfig(path)
		if err != nil {
			t.Fatalf("failed to read toml config: %v", err)
		}

		if !reflect.DeepEqual(generated, config) {
			t.Fatalf("expected\n%#v\n\ngot\n%#v", generated, config)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := eavesdrop.ReadTomlConfig("nonexistent.toml")
		if err == nil || !strings.Contains(err.Error(), "failed to read toml config") {
			t.Errorf("expected read error, got: %v", err)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "empty.toml")

		err := os.WriteFile(path, []byte(""), 0644)
		if err != nil {
			t.Fatalf("failed to write empty toml file: %v", err)
		}

		_, err = eavesdrop.ReadTomlConfig(path)
		if err == nil || !strings.Contains(err.Error(), "toml config is empty") {
			t.Errorf("expected empty file error, got: %v", err)
		}
	})
}

func TestYAML(t *testing.T) {
	tmp := t.TempDir()

	t.Run("test generate", func(t *testing.T) {
		err := eavesdrop.GenerateYamlConfig(tmp)
		if err != nil {
			t.Fatalf("failed to generate yaml config: %v", err)
		}

		path := filepath.Join(tmp, eavesdrop.YAML_CONFIG)
		info, err := os.Stat(path)
		if err != nil || info.Size() == 0 {
			t.Fatal("expected yaml config to be written and non-empty")
		}
	})

	t.Run("test read", func(t *testing.T) {
		generated := generateConfig()

		data, err := yaml.Marshal(generated)
		if err != nil {
			t.Fatalf("failed to marshal config to yaml: %v", err)
		}

		path := filepath.Join(tmp, eavesdrop.YAML_CONFIG)

		err = os.WriteFile(path, data, 0644)
		if err != nil {
			t.Fatalf("failed to save yaml: %v", err)
		}

		config, err := eavesdrop.ReadYamlConfig(path)
		if err != nil {
			t.Fatalf("failed to read yaml config: %v", err)
		}

		if !reflect.DeepEqual(generated, config) {
			t.Fatalf("expected\n%#v\n\ngot\n%#v", generated, config)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := eavesdrop.ReadYamlConfig("nonexistent.yaml")
		if err == nil || !strings.Contains(err.Error(), "failed to read yaml config") {
			t.Errorf("expected read error, got: %v", err)
		}
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "empty.yaml")

		err := os.WriteFile(path, []byte(""), 0644)
		if err != nil {
			t.Fatalf("failed to write empty yaml file: %v", err)
		}

		_, err = eavesdrop.ReadYamlConfig(path)
		if err == nil || !strings.Contains(err.Error(), "yaml config is empty") {
			t.Errorf("expected empty file error, got: %v", err)
		}
	})
}
