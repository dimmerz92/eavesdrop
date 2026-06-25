package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/dimmerz92/eavesdrop/internal/config"
)

func TestTOML(t *testing.T) {
	tmp := t.TempDir()

	t.Run("test generate", func(t *testing.T) {
		err := config.GenerateTomlConfig(tmp)
		if err != nil {
			t.Fatalf("failed to generate toml config: %v", err)
		}

		path := filepath.Join(tmp, config.TOML_CONFIG)
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

		path := filepath.Join(tmp, config.TOML_CONFIG)

		err = os.WriteFile(path, data, 0644)
		if err != nil {
			t.Fatalf("failed to save toml: %v", err)
		}

		config, err := config.ReadTomlConfig(path)
		if err != nil {
			t.Fatalf("failed to read toml config: %v", err)
		}

		if !reflect.DeepEqual(generated, config) {
			t.Fatalf("expected\n%#v\n\ngot\n%#v", generated, config)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := config.ReadTomlConfig("nonexistent.toml")
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

		_, err = config.ReadTomlConfig(path)
		if err == nil || !strings.Contains(err.Error(), "toml config is empty") {
			t.Errorf("expected empty file error, got: %v", err)
		}
	})
}
