package config_test

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/dimmerz92/eavesdrop/internal/config"
)

func TestConfig_Validate(t *testing.T) {
	t.Run("test default config", func(t *testing.T) {
		cfg := config.DefaultConfig("")

		if err := cfg.Validate(); err != nil {
			t.Errorf("expected nil, got err: %v", err)
		}
	})

	t.Run("test custom config", func(t *testing.T) {
		cfg := config.Config{
			Root:      ".",
			Build:     ".",
			Run:       ".",
			Proxy:     true,
			AppPort:   8000,
			ProxyPort: 8001,
		}

		if err := cfg.Validate(); err != nil {
			t.Errorf("expected nil, got err: %v", err)
		}
	})

	t.Run("test bad root", func(t *testing.T) {
		cfg := config.Config{
			Build:     ".",
			Run:       ".",
			Proxy:     true,
			AppPort:   8000,
			ProxyPort: 8001,
		}

		if err := cfg.Validate(); err == nil {
			t.Error("expected err, got nil")
		}
	})

	t.Run("test bad build", func(t *testing.T) {
		cfg := config.Config{
			Root:      ".",
			Run:       ".",
			Proxy:     true,
			AppPort:   8000,
			ProxyPort: 8001,
		}

		if err := cfg.Validate(); err == nil {
			t.Error("expected err, got nil")
		}
	})

	t.Run("test bad run", func(t *testing.T) {
		cfg := config.Config{
			Root:      ".",
			Build:     ".",
			Proxy:     true,
			AppPort:   8000,
			ProxyPort: 8001,
		}

		if err := cfg.Validate(); err == nil {
			t.Error("expected err, got nil")
		}
	})

	t.Run("test bad app port", func(t *testing.T) {
		cfg := config.Config{
			Root:      ".",
			Build:     ".",
			Run:       ".",
			Proxy:     true,
			ProxyPort: 8001,
		}

		if err := cfg.Validate(); err == nil {
			t.Error("expected err, got nil")
		}
	})

	t.Run("test bad proxy port", func(t *testing.T) {
		cfg := config.Config{
			Root:    ".",
			Build:   ".",
			Run:     ".",
			Proxy:   true,
			AppPort: 8000,
		}

		if err := cfg.Validate(); err == nil {
			t.Error("expected err, got nil")
		}
	})
}

func TestGetConfig(t *testing.T) {
	tmp := t.TempDir()

	t.Run("no path given", func(t *testing.T) {
		expected := config.DefaultConfig("")
		cfg, err := config.GetConfig("")
		if err != nil {
			t.Fatalf("got err when expected nil: %v", err)
		}

		if !reflect.DeepEqual(expected, cfg) {
			t.Errorf("expected %+v got %+v", expected, cfg)
		}
	})

	t.Run("get json config", func(t *testing.T) {
		path := filepath.Join(tmp, config.JSON_CONFIG)
		expected := config.DefaultConfig(path)

		if err := config.GenerateJsonConfig(tmp); err != nil {
			t.Fatalf("failed to generate json config: %v", err)
		}

		cfg, err := config.GetConfig(path)
		if err != nil {
			t.Fatalf("got err when expected nil: %v", err)
		}

		if !reflect.DeepEqual(expected, cfg) {
			t.Errorf("expected %+v got %+v", expected, cfg)
		}
	})

	t.Run("get toml config", func(t *testing.T) {
		path := filepath.Join(tmp, config.TOML_CONFIG)
		expected := config.DefaultConfig(path)

		if err := config.GenerateTomlConfig(tmp); err != nil {
			t.Fatalf("failed to generate toml config: %v", err)
		}

		cfg, err := config.GetConfig(path)
		if err != nil {
			t.Fatalf("got err when expected nil: %v", err)
		}

		if !reflect.DeepEqual(expected, cfg) {
			t.Errorf("expected %+v got %+v", expected, cfg)
		}
	})

	t.Run("get yaml config", func(t *testing.T) {
		path := filepath.Join(tmp, config.YAML_CONFIG)
		expected := config.DefaultConfig(path)

		if err := config.GenerateYamlConfig(tmp); err != nil {
			t.Fatalf("failed to generate yaml config: %v", err)
		}

		cfg, err := config.GetConfig(path)
		if err != nil {
			t.Fatalf("got err when expected nil: %v", err)
		}

		if !reflect.DeepEqual(expected, cfg) {
			t.Errorf("expected %+v got %+v", expected, cfg)
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		path := filepath.Join(tmp, "not_here.json")
		_, err := config.GetConfig(path)
		if err == nil {
			t.Fatalf("expected err, got nil")
		}
	})
}

func TestGenerateConfig(t *testing.T) {
	tmp := t.TempDir()

	t.Run("blank ext", func(t *testing.T) {
		path := filepath.Join(tmp, config.JSON_CONFIG)
		expected := config.DefaultConfig(path)

		if err := config.GenerateConfig(tmp, ""); err != nil {
			t.Fatalf("failed to generate config: %v", err)
		}

		cfg, err := config.GetConfig(path)
		if err != nil {
			t.Fatalf("config not generated to given path: %v", err)
		}

		if !reflect.DeepEqual(expected, cfg) {
			t.Errorf("expected %+v got %+v", expected, cfg)
		}
	})

	t.Run(".json ext", func(t *testing.T) {
		path := filepath.Join(tmp, config.JSON_CONFIG)
		expected := config.DefaultConfig(path)

		if err := config.GenerateConfig(tmp, ".json"); err != nil {
			t.Fatalf("failed to generate config: %v", err)
		}

		cfg, err := config.GetConfig(path)
		if err != nil {
			t.Fatalf("config not generated to given path: %v", err)
		}

		if !reflect.DeepEqual(expected, cfg) {
			t.Errorf("expected %+v got %+v", expected, cfg)
		}
	})

	t.Run("json ext", func(t *testing.T) {
		path := filepath.Join(tmp, config.JSON_CONFIG)
		expected := config.DefaultConfig(path)

		if err := config.GenerateConfig(tmp, "json"); err != nil {
			t.Fatalf("failed to generate config: %v", err)
		}

		cfg, err := config.GetConfig(path)
		if err != nil {
			t.Fatalf("config not generated to given path: %v", err)
		}

		if !reflect.DeepEqual(expected, cfg) {
			t.Errorf("expected %+v got %+v", expected, cfg)
		}
	})

	t.Run(".toml ext", func(t *testing.T) {
		path := filepath.Join(tmp, config.TOML_CONFIG)
		expected := config.DefaultConfig(path)

		if err := config.GenerateConfig(tmp, ".toml"); err != nil {
			t.Fatalf("failed to generate config: %v", err)
		}

		cfg, err := config.GetConfig(path)
		if err != nil {
			t.Fatalf("config not generated to given path: %v", err)
		}

		if !reflect.DeepEqual(expected, cfg) {
			t.Errorf("expected %+v got %+v", expected, cfg)
		}
	})

	t.Run("toml ext", func(t *testing.T) {
		path := filepath.Join(tmp, config.TOML_CONFIG)
		expected := config.DefaultConfig(path)

		if err := config.GenerateConfig(tmp, "toml"); err != nil {
			t.Fatalf("failed to generate config: %v", err)
		}

		cfg, err := config.GetConfig(path)
		if err != nil {
			t.Fatalf("config not generated to given path: %v", err)
		}

		if !reflect.DeepEqual(expected, cfg) {
			t.Errorf("expected %+v got %+v", expected, cfg)
		}
	})

	t.Run(".yaml ext", func(t *testing.T) {
		path := filepath.Join(tmp, config.YAML_CONFIG)
		expected := config.DefaultConfig(path)

		if err := config.GenerateConfig(tmp, ".yaml"); err != nil {
			t.Fatalf("failed to generate config: %v", err)
		}

		cfg, err := config.GetConfig(path)
		if err != nil {
			t.Fatalf("config not generated to given path: %v", err)
		}

		if !reflect.DeepEqual(expected, cfg) {
			t.Errorf("expected %+v got %+v", expected, cfg)
		}
	})

	t.Run("yaml ext", func(t *testing.T) {
		path := filepath.Join(tmp, config.YAML_CONFIG)
		expected := config.DefaultConfig(path)

		if err := config.GenerateConfig(tmp, "yaml"); err != nil {
			t.Fatalf("failed to generate config: %v", err)
		}

		cfg, err := config.GetConfig(path)
		if err != nil {
			t.Fatalf("config not generated to given path: %v", err)
		}

		if !reflect.DeepEqual(expected, cfg) {
			t.Errorf("expected %+v got %+v", expected, cfg)
		}
	})

	t.Run("invalid ext", func(t *testing.T) {
		if err := config.GenerateConfig(tmp, ".txt"); err == nil {
			t.Fatalf("expected err got nil")
		}
	})
}
