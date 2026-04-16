package eavesdrop_test

import (
	"sync"
	"testing"

	"github.com/dimmerz92/eavesdrop"
)

func TestWatcherConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		watcherName string
		config      []eavesdrop.WatcherOption
		panic       bool
	}{
		{
			name:        "missing name",
			watcherName: "   ",
			panic:       true,
		},
		{
			name:        "missing file types and file names",
			watcherName: "test1",
			panic:       true,
		},
		{
			name:        "missing tasks and service",
			watcherName: "test2",
			config:      []eavesdrop.WatcherOption{eavesdrop.WithWatchedFiletypes(".go")},
			panic:       true,
		},
		{
			name:        "with tasks",
			watcherName: "test3",
			config: []eavesdrop.WatcherOption{
				eavesdrop.WithWatchedFiletypes(".go"),
				eavesdrop.WithTasks("echo 'hello'"),
			},
		},
		{
			name:        "with service",
			watcherName: "test4",
			config: []eavesdrop.WatcherOption{
				eavesdrop.WithWatchedFiletypes(".go"),
				eavesdrop.WithService("sleep 1000; echo hello"),
			},
		},
	}

	for _, test := range tests {
		var mu sync.Mutex
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if test.panic && r == nil {
					t.Fatalf("expected panic")
				}
				if !test.panic && r != nil {
					t.Fatalf("unexpected panic")
				}
			}()

			_ = eavesdrop.NewWatcher(t.Context(), test.watcherName, &mu, test.config...)
		})
	}
}
