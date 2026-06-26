package config_test

import (
	"testing"

	"github.com/dimmerz92/eavesdrop/v2/internal/config"
)

func generateConfig() config.Config {
	config := config.DefaultConfig()
	config.Watchers[0].Filetypes = []string{".go"}
	config.Watchers[0].Shell.Tasks = []string{"echo hello"}

	return config
}

func TestDefaultConfig(t *testing.T) {
	config := config.DefaultConfig()

	if config.RootDir != "." {
		t.Fatalf("expected RootDir to be '.', got '%s'", config.RootDir)
	}

	if len(config.Watchers) == 0 {
		t.Fatalf("expected at least one watcher in default config")
	}
}
