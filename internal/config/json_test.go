package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dimmerz92/eavesdrop/internal/config"
)

func TestJsonConfig(t *testing.T) {
	tmp := t.TempDir()

	t.Run("create a json config in tmp", func(t *testing.T) {
		if err := config.GenerateJsonConfig(tmp); err != nil {
			t.Fatalf("failed to generate json config: %v", err)
		}

		if file, err := os.Stat(filepath.Join(tmp, config.JSON_CONFIG)); err != nil {
			t.Fatalf("json config not found in tmp: %v", err)
		} else if file.Name() != config.JSON_CONFIG {
			t.Fatalf("name error, wanted %s, got %s", config.JSON_CONFIG, file.Name())
		}
	})

	t.Run("read a valid json config", func(t *testing.T) {
		path := filepath.Join(tmp, config.JSON_CONFIG)
		expected := config.DefaultConfig(path)
		cfg, err := config.ReadJsonConfig(path)
		if err != nil {
			t.Fatalf("failed to read json config: %v", err)
		}

		if !reflect.DeepEqual(expected, cfg) {
			t.Fatalf("read config not equal to default")
		}
	})

	t.Run("read an invalid json config", func(t *testing.T) {
		invalid := filepath.Join(tmp, "invalid.json")
		file, err := os.Create(invalid)
		if err != nil {
			t.Fatalf("failed to create json file: %v", err)
		} else {
			if _, err = file.Write([]byte("{")); err != nil {
				t.Fatalf("failed to write to json file: %v", err)
			}
		}
		defer file.Close()

		_, err = config.ReadJsonConfig(filepath.Join(invalid))
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})

	t.Run("read non-existant json config", func(t *testing.T) {
		notExists := filepath.Join(tmp, "not-exists.json")
		_, err := config.ReadJsonConfig(notExists)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
}
