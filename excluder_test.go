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

func TestNewExcluder(t *testing.T) {
	tests := []struct {
		name        string
		root        string
		expectedErr bool
	}{
		{"empty root defaults to cwd", "", false},
		{"whitespace root defaults to cwd", "   ", false},
		{"valid relative root", ".", false},
		{"valid absolute root", "/tmp", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			excl, err := ev.NewExcluder(test.root)
			if (err != nil) != test.expectedErr {
				t.Errorf("NewExcluder(%q) error = %v, expectedErr %v", test.root, err, test.expectedErr)
			}

			if !test.expectedErr && excl == nil {
				t.Error("NewExcluder() returned nil Excluder without error")
			}
		})
	}
}

func TestExcluder_ShouldIgnore(t *testing.T) {
	root := t.TempDir()
	excludedDir := filepath.Join(root, "vendor")
	excludedFile := filepath.Join(root, "go.sum")
	otherDir := filepath.Join(root, "cmd")
	otherFile := filepath.Join(root, "main.go")

	tests := []struct {
		name     string
		opts     []ev.ExcluderOption
		event    ev.Event
		expected bool
	}{
		{
			name:     "no options never ignores",
			opts:     nil,
			event:    ev.NewEvent(ev.WRITE, otherFile, mockFileInfo{}),
			expected: false,
		},
		{
			name:     "matching op excluded",
			opts:     []ev.ExcluderOption{ev.WithOp(ev.WRITE)},
			event:    ev.NewEvent(ev.WRITE, otherFile, mockFileInfo{}),
			expected: true,
		},
		{
			name:     "non-matching op not excluded",
			opts:     []ev.ExcluderOption{ev.WithOp(ev.CHMOD)},
			event:    ev.NewEvent(ev.WRITE, otherFile, mockFileInfo{}),
			expected: false,
		},
		{
			name:     "one of multiple excluded ops matches",
			opts:     []ev.ExcluderOption{ev.WithOp(ev.CHMOD, ev.WRITE)},
			event:    ev.NewEvent(ev.WRITE, otherFile, mockFileInfo{}),
			expected: true,
		},
		{
			name:     "excluded dir matched",
			opts:     []ev.ExcluderOption{ev.WithDirs(excludedDir)},
			event:    ev.NewEvent(ev.CREATE, excludedDir, mockDirInfo{}),
			expected: true,
		},
		{
			name:     "non-excluded dir not matched",
			opts:     []ev.ExcluderOption{ev.WithDirs(excludedDir)},
			event:    ev.NewEvent(ev.CREATE, otherDir, mockDirInfo{}),
			expected: false,
		},
		{
			name:     "excluded file matched",
			opts:     []ev.ExcluderOption{ev.WithFiles(excludedFile)},
			event:    ev.NewEvent(ev.WRITE, excludedFile, mockFileInfo{}),
			expected: true,
		},
		{
			name:     "non-excluded file not matched",
			opts:     []ev.ExcluderOption{ev.WithFiles(excludedFile)},
			event:    ev.NewEvent(ev.WRITE, otherFile, mockFileInfo{}),
			expected: false,
		},
		{
			name:     "regex matches path",
			opts:     []ev.ExcluderOption{ev.WithRegex(`\.env$`)},
			event:    ev.NewEvent(ev.CREATE, filepath.Join(root, ".env"), mockFileInfo{}),
			expected: true,
		},
		{
			name:     "regex does not match path",
			opts:     []ev.ExcluderOption{ev.WithRegex(`\.env$`)},
			event:    ev.NewEvent(ev.WRITE, otherFile, mockFileInfo{}),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			excl, err := ev.NewExcluder(root, test.opts...)
			if err != nil {
				t.Fatalf("NewExcluder() unexpected error: %v", err)
			}

			if got := excl.ShouldIgnore(test.event); got != test.expected {
				t.Errorf("ShouldIgnore() = %v, expected %v", got, test.expected)
			}
		})
	}
}
