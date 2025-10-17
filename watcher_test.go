package eavesdrop_test

import (
	"testing"

	"github.com/dimmerz92/eavesdrop"
)

func TestWatcherConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  eavesdrop.WatcherConfig
		wantErr bool
	}{
		{
			name:    "missing name",
			config:  eavesdrop.WatcherConfig{},
			wantErr: true,
		},
		{
			name:    "missing file types and file names",
			config:  eavesdrop.WatcherConfig{Name: "watcher1"},
			wantErr: true,
		},
		{
			name:    "missing tasks and service",
			config:  eavesdrop.WatcherConfig{Name: "watcher2", FileTypes: []string{".go"}},
			wantErr: true,
		},
		{
			name: "negative max task time",
			config: eavesdrop.WatcherConfig{
				Name:        "watcher3",
				FileTypes:   []string{".go"},
				Tasks:       []string{"echo hello"},
				MaxTaskTime: -100,
			},
			wantErr: true,
		},
		{
			name: "negative service timeout",
			config: eavesdrop.WatcherConfig{
				Name:              "watcher4",
				FileTypes:         []string{".go"},
				Tasks:             []string{"echo hello"},
				MaxServiceTimeout: -100,
			},
			wantErr: true,
		},
		{
			name: "negative debounce delay",
			config: eavesdrop.WatcherConfig{
				Name:          "watcher5",
				FileTypes:     []string{".go"},
				Tasks:         []string{"echo hello"},
				DebounceDelay: -100,
			},
			wantErr: true,
		},
		{
			name: "valid config with task",
			config: eavesdrop.WatcherConfig{
				Name:      "watcher6",
				FileTypes: []string{".go"},
				Tasks:     []string{"echo hello"},
			},
			wantErr: false,
		},
		{
			name: "valid config with service",
			config: eavesdrop.WatcherConfig{
				Name:      "watcher7",
				FileTypes: []string{".go"},
				Service:   "echo hello",
			},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.Validate()
			if test.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			} else if !test.wantErr && err != nil {
				t.Fatalf("expected nil, got %v", err)
			}
		})
	}
}
