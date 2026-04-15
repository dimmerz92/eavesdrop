package internal_test

import (
	"testing"

	"github.com/dimmerz92/eavesdrop/internal"
)

func generateConfig() internal.Config {
	config := internal.DefaultConfig()
	config.Watchers[0].Filetypes = []string{".go"}
	config.Watchers[0].Shell.Tasks = []string{"echo hello"}

	return config
}

func TestDefaultConfig(t *testing.T) {
	config := internal.DefaultConfig()

	if config.RootDir != "." {
		t.Fatalf("expected RootDir to be '.', got '%s'", config.RootDir)
	}

	if len(config.Watchers) == 0 {
		t.Fatalf("expected at least one watcher in default config")
	}
}
