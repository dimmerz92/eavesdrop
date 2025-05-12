package config_test

import (
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
