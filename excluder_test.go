package eavesdrop_test

import (
	"regexp"
	"testing"

	"github.com/dimmerz92/eavesdrop"
)

func TestExcluderConfig_ToExcluder(t *testing.T) {
	tests := []struct {
		name      string
		config    eavesdrop.ExcluderConfig
		wantErr   bool
		wantRegex []string
	}{
		{
			name: "Valid config with all fields",
			config: eavesdrop.ExcluderConfig{
				Dirs:  []string{"dir1", "dir2", ".git", "./.git"},
				Files: []string{"file1.txt"},
				Regex: []string{`.*\.log$`, `^tmp/`, `^\.?(\/?|\\?)(?:\w+(\/|\\))*(\.\w+)$`},
			},
			wantErr:   false,
			wantRegex: []string{`.*\.log$`, `^tmp/`, `^\.?(\/?|\\?)(?:\w+(\/|\\))*(\.\w+)$`},
		},
		{
			name: "Valid config with no regex",
			config: eavesdrop.ExcluderConfig{
				Dirs:  []string{"dir1"},
				Files: []string{"file1.txt"},
			},
			wantErr:   false,
			wantRegex: nil,
		},
		{
			name: "Invalid regex",
			config: eavesdrop.ExcluderConfig{
				Regex: []string{"[invalid"},
			},
			wantErr: true,
		},
		{
			name:    "Empty config",
			config:  eavesdrop.ExcluderConfig{},
			wantErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			excluder, err := test.config.ToExcluder("")

			if test.wantErr {
				if err == nil {
					t.Fatal("expected error but got nil")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, d := range test.config.Dirs {
				if _, ok := excluder.Dirs[d]; !ok {
					t.Errorf("expected dir %s to be in excluder.Dirs", d)
				}
			}
			for _, f := range test.config.Files {
				if _, ok := excluder.Files[f]; !ok {
					t.Errorf("expected file %s to be in excluder.Files", f)
				}
			}

			if len(test.wantRegex) != len(excluder.Regex) {
				t.Errorf("expected %d regexes, got %d", len(test.wantRegex), len(excluder.Regex))
			}
			for i, r := range test.wantRegex {
				if excluder.Regex[i].String() != r {
					t.Errorf("expected regex %q, got %q", r, excluder.Regex[i].String())
				}
			}
		})
	}
}

func TestExcluder_ShouldIgnore(t *testing.T) {
	excluder := &eavesdrop.Excluder{
		Dirs: map[string]struct{}{"ignore_dir": {}, ".git": {}},
		Files: map[string]struct{}{"ignore_file.txt": {},
			".env":   {},
			"./.env": {}, "dir/.env": {}, "./dir/.env": {},
			".\\.env": {}, "dir\\.env": {}, ".\\dir\\.env": {},
		},
		Regex: []*regexp.Regexp{
			regexp.MustCompile(`.*\.log$`),
			regexp.MustCompile(`^tmp/`),
			regexp.MustCompile(`^\.?(\/?|\\?)(?:\w+(\/|\\))*(\.\w+)$`),
		},
	}

	tests := []struct {
		name   string
		path   string
		isDir  bool
		expect bool
	}{
		{
			name:   "Ignore matching dir",
			path:   "ignore_dir",
			isDir:  true,
			expect: true,
		},
		{
			name:   "Ignore matching file",
			path:   "ignore_file.txt",
			isDir:  false,
			expect: true,
		},
		{
			name:   "Ignore matching regex .log",
			path:   "debug.log",
			isDir:  false,
			expect: true,
		},
		{
			name:   "Ignore matching regex tmp/",
			path:   "tmp/cache",
			isDir:  true,
			expect: true,
		},
		{
			name:   "Not ignored file",
			path:   "main.go",
			isDir:  false,
			expect: false,
		},
		{
			name:   "Not ignored dir",
			path:   "src",
			isDir:  true,
			expect: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := excluder.ShouldIgnore(test.path, test.isDir)
			if got != test.expect {
				t.Errorf("got %v; want %v", got, test.expect)
			}
		})
	}
}
