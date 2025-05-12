package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dimmerz92/eavesdrop/internal/config"
)

func TestYamlConfig(t *testing.T) {
	tmp := t.TempDir()

	t.Run("create a yaml config in tmp", func(t *testing.T) {
		if err := config.GenerateYamlConfig(tmp); err != nil {
			t.Fatalf("failed to generate yaml config: %v", err)
		}

		if file, err := os.Stat(filepath.Join(tmp, config.YAML_CONFIG)); err != nil {
			t.Fatalf("yaml config not found in tmp: %v", err)
		} else if file.Name() != config.YAML_CONFIG {
			t.Fatalf("name error, wanted %s, got %s", config.YAML_CONFIG, file.Name())
		}
	})

	t.Run("read a valid yaml config", func(t *testing.T) {
		path := filepath.Join(tmp, config.YAML_CONFIG)
		expected := config.DefaultConfig(path)
		cfg, err := config.ReadYamlConfig(path)
		if err != nil {
			t.Fatalf("failed to read yaml config: %v", err)
		}

		if !reflect.DeepEqual(expected, cfg) {
			t.Log(expected)
			t.Log(cfg)
			t.Fatalf("read config not equal to default")
		}
	})

	t.Run("read an invalid yaml config", func(t *testing.T) {
		invalid := filepath.Join(tmp, "invalid.yaml")
		file, err := os.Create(invalid)
		if err != nil {
			t.Fatalf("failed to create yaml file: %v", err)
		} else {
			if _, err = file.Write([]byte("")); err != nil {
				t.Fatalf("failed to write to yaml file: %v", err)
			}
		}
		defer file.Close()

		_, err = config.ReadYamlConfig(filepath.Join(invalid))
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})

	t.Run("read non-existant yaml config", func(t *testing.T) {
		notExists := filepath.Join(tmp, "not-exists.yaml")
		_, err := config.ReadYamlConfig(notExists)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
}
