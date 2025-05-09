package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dimmerz92/eavesdrop/internal/config"
)

func TestTomlConfig(t *testing.T) {
	tmp := t.TempDir()

	t.Run("create a toml config in tmp", func(t *testing.T) {
		if err := config.GenerateTomlConfig(tmp); err != nil {
			t.Fatalf("failed to generate toml config: %v", err)
		}

		if file, err := os.Stat(filepath.Join(tmp, config.TOML_CONFIG)); err != nil {
			t.Fatalf("toml config not found in tmp: %v", err)
		} else if file.Name() != config.TOML_CONFIG {
			t.Fatalf("name error, wanted %s, got %s", config.TOML_CONFIG, file.Name())
		}
	})

	t.Run("read a valid toml config", func(t *testing.T) {
		path := filepath.Join(tmp, config.TOML_CONFIG)
		expected := config.DefaultConfig(path)
		cfg, err := config.ReadTomlConfig(path)
		if err != nil {
			t.Fatalf("failed to read toml config: %v", err)
		}

		if !reflect.DeepEqual(expected, cfg) {
			t.Log(expected)
			t.Log(cfg)
			t.Fatalf("read config not equal to default")
		}
	})

	t.Run("read an invalid toml config", func(t *testing.T) {
		invalid := filepath.Join(tmp, "invalid.toml")
		file, err := os.Create(invalid)
		if err != nil {
			t.Fatalf("failed to create toml file: %v", err)
		} else {
			if _, err = file.Write([]byte("")); err != nil {
				t.Fatalf("failed to write to toml file: %v", err)
			}
		}
		defer file.Close()

		_, err = config.ReadTomlConfig(filepath.Join(invalid))
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})

	t.Run("read non-existant toml config", func(t *testing.T) {
		notExists := filepath.Join(tmp, "not-exists.toml")
		_, err := config.ReadTomlConfig(notExists)
		if err == nil {
			t.Fatalf("expected error, got none")
		}
	})
}
