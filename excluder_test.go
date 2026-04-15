package eavesdrop_test

import (
	"io/fs"
	"testing"

	"github.com/dimmerz92/eavesdrop"
)

type mockFileInfo struct {
	fs.FileInfo
	isDir bool
	name  string
}

func (mfi *mockFileInfo) IsDir() bool  { return mfi.isDir }
func (mfi *mockFileInfo) Name() string { return mfi.name }

func TestExcluder(t *testing.T) {
	tests := []struct {
		name   string
		dirs   []string
		files  []string
		regex  []string
		values map[mockFileInfo]bool
	}{
		{name: "no constraints", values: make(map[mockFileInfo]bool)},
		{
			name: "dirs",
			dirs: []string{"tmp", "node_modules", ".git"},
			values: map[mockFileInfo]bool{
				{name: "tmp", isDir: true}:          true,
				{name: "node_modules", isDir: true}: true,
				{name: ".git", isDir: true}:         true,
				{name: "cmd", isDir: true}:          false,
			},
		},
		{
			name:  "files",
			files: []string{".env", ".DS_Store"},
			values: map[mockFileInfo]bool{
				{name: ".env"}:      true,
				{name: ".DS_Store"}: true,
				{name: "main.go"}:   false,
			},
		},
		{
			name:  "regex",
			regex: []string{"^\\.?(\\/?|\\\\?)(?:\\w+(\\/|\\\\))*(\\.\\w+)$", "^.+\\.sqlite$"},
			values: map[mockFileInfo]bool{
				{name: ".env"}:      true,
				{name: "db.sqlite"}: true,
				{name: "main.go"}:   false,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			excluder := eavesdrop.NewExcluder(
				eavesdrop.WithDirs(test.dirs...),
				eavesdrop.WithFiles(test.files...),
				eavesdrop.WithRegex(test.regex...),
			)

			for file, expected := range test.values {
				if got := excluder.ShouldIgnore(&file); got != expected {
					t.Errorf("expected %t for %s, got %t", expected, file.Name(), got)
				}
			}
		})
	}
}
