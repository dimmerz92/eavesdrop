package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/dimmerz92/eavesdrop/v2/internal/config"
	"gopkg.in/yaml.v3"
)

func TestYAML(t *testing.T) {
	tmp := t.TempDir()

	t.Run("test generate", func(t *testing.T) {
		err := config.GenerateYamlConfig(tmp)
		if err != nil {
			t.Fatalf("failed to generate yaml config: %v", err)
		}

		path := filepath.Join(tmp, config.YAML_CONFIG)
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

		path := filepath.Join(tmp, config.YAML_CONFIG)

		err = os.WriteFile(path, data, 0644)
		if err != nil {
			t.Fatalf("failed to save yaml: %v", err)
		}

		config, err := config.ReadYamlConfig(path)
		if err != nil {
			t.Fatalf("failed to read yaml config: %v", err)
		}

		if !reflect.DeepEqual(generated, config) {
			t.Fatalf("expected\n%#v\n\ngot\n%#v", generated, config)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := config.ReadYamlConfig("nonexistent.yaml")
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

		_, err = config.ReadYamlConfig(path)
		if err == nil || !strings.Contains(err.Error(), "yaml config is empty") {
			t.Errorf("expected empty file error, got: %v", err)
		}
	})
}
