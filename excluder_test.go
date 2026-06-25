package ev_test

import (
	"io/fs"
	"path/filepath"
	"testing"
	"time"

	"github.com/dimmerz92/eavesdrop"
)

type mockDirInfo struct{}

func (m mockDirInfo) Name() string       { return "" }
func (m mockDirInfo) Size() int64        { return 0 }
func (m mockDirInfo) Mode() fs.FileMode  { return 0 }
func (m mockDirInfo) ModTime() time.Time { return time.Time{} }
func (m mockDirInfo) IsDir() bool        { return true }
func (m mockDirInfo) Sys() any           { return nil }

func TestExcluder_ShouldIgnore(t *testing.T) {
	root := t.TempDir()
	excludedDir := filepath.Join(root, "vendor")
	excludedFile := filepath.Join(root, "go.sum")
	otherDir := filepath.Join(root, "cmd")
	otherFile := filepath.Join(root, "main.go")

	tests := []struct {
		name     string
		excluder *ev.Excluder
		event    ev.Event
		expected bool
	}{
		{
			name:     "no options never ignores",
			excluder: ev.NewExcluder(root),
			event:    ev.NewEvent(ev.WRITE, otherFile, mockFileInfo{}),
			expected: false,
		},
		{
			name:     "matching op excluded",
			excluder: ev.NewExcluder(root).WithOps(ev.WRITE),
			event:    ev.NewEvent(ev.WRITE, otherFile, mockFileInfo{}),
			expected: true,
		},
		{
			name:     "non-matching op not excluded",
			excluder: ev.NewExcluder(root).WithOps(ev.CHMOD),
			event:    ev.NewEvent(ev.WRITE, otherFile, mockFileInfo{}),
			expected: false,
		},
		{
			name:     "one of multiple excluded ops matches",
			excluder: ev.NewExcluder(root).WithOps(ev.CHMOD, ev.WRITE),
			event:    ev.NewEvent(ev.WRITE, otherFile, mockFileInfo{}),
			expected: true,
		},
		{
			name:     "excluded dir matched",
			excluder: ev.NewExcluder(root).WithDirs(excludedDir),
			event:    ev.NewEvent(ev.CREATE, excludedDir, mockDirInfo{}),
			expected: true,
		},
		{
			name:     "non-excluded dir not matched",
			excluder: ev.NewExcluder(root).WithDirs(excludedDir),
			event:    ev.NewEvent(ev.CREATE, otherDir, mockDirInfo{}),
			expected: false,
		},
		{
			name:     "excluded file matched",
			excluder: ev.NewExcluder(root).WithFiles(excludedFile),
			event:    ev.NewEvent(ev.WRITE, excludedFile, mockFileInfo{}),
			expected: true,
		},
		{
			name:     "non-excluded file not matched",
			excluder: ev.NewExcluder(root).WithFiles(excludedFile),
			event:    ev.NewEvent(ev.WRITE, otherFile, mockFileInfo{}),
			expected: false,
		},
		{
			name:     "regex matches path",
			excluder: ev.NewExcluder(root).WithRegex(`\.env$`),
			event:    ev.NewEvent(ev.CREATE, filepath.Join(root, ".env"), mockFileInfo{}),
			expected: true,
		},
		{
			name:     "regex does not match path",
			excluder: ev.NewExcluder(root).WithRegex(`\.env$`),
			event:    ev.NewEvent(ev.WRITE, otherFile, mockFileInfo{}),
			expected: false,
		},
		{
			name:     "relative excluded dir matched exactly",
			excluder: ev.NewExcluder(".").WithDirs("vendor"),
			event:    ev.NewEvent(ev.CREATE, "vendor", mockDirInfo{}),
			expected: true,
		},
		{
			name:     "relative excluded dir child matched",
			excluder: ev.NewExcluder(".").WithDirs("vendor"),
			event:    ev.NewEvent(ev.CREATE, filepath.Join("vendor", "pkg", "foo.go"), mockFileInfo{}),
			expected: true,
		},
		{
			name:     "relative excluded dir sibling not matched",
			excluder: ev.NewExcluder(".").WithDirs("vendor"),
			event:    ev.NewEvent(ev.CREATE, "cmd", mockDirInfo{}),
			expected: false,
		},
		{
			name:     "relative excluded file matched",
			excluder: ev.NewExcluder(".").WithFiles("go.sum"),
			event:    ev.NewEvent(ev.WRITE, "go.sum", mockFileInfo{}),
			expected: true,
		},
		{
			name:     "relative excluded file not matched for other file",
			excluder: ev.NewExcluder(".").WithFiles("go.sum"),
			event:    ev.NewEvent(ev.WRITE, "main.go", mockFileInfo{}),
			expected: false,
		},
		{
			name:     "relative excluded file not matched in subdirectory",
			excluder: ev.NewExcluder(".").WithFiles("go.sum"),
			event:    ev.NewEvent(ev.WRITE, filepath.Join("vendor", "go.sum"), mockFileInfo{}),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.excluder.ShouldIgnore(test.event); got != test.expected {
				t.Errorf("ShouldIgnore() = %v, expected %v", got, test.expected)
			}
		})
	}
}
