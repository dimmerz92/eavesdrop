package config_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/dimmerz92/eavesdrop/internal/config"
)

func TestJSON(t *testing.T) {
	tmp := t.TempDir()

	t.Run("test generate", func(t *testing.T) {
		err := config.GenerateJsonConfig(tmp)
		if err != nil {
			t.Fatalf("failed to generate json config: %v", err)
		}

		path := filepath.Join(tmp, config.JSON_CONFIG)
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

		path := filepath.Join(tmp, config.JSON_CONFIG)

		err = os.WriteFile(path, data, 0644)
		if err != nil {
			t.Fatalf("failed to save json: %v", err)
		}

		config, err := config.ReadJsonConfig(path)
		if err != nil {
			t.Fatalf("failed to read json config: %v", err)
		}

		if !reflect.DeepEqual(generated, config) {
			t.Fatalf("expected\n%#v\n\ngot\n%#v", generated, config)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, err := config.ReadJsonConfig("nonexistent.json")
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

		_, err = config.ReadJsonConfig(path)
		if err == nil || !strings.Contains(err.Error(), "failed to unmarshal json to config") {
			t.Errorf("expected empty file error, got: %v", err)
		}
	})

}
