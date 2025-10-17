package eavesdrop

import (
	"fmt"
	"path/filepath"
	"regexp"
)

type ExcluderConfig struct {
	Dirs  []string `json:"dirs" toml:"dirs" yaml:"dirs"`
	Files []string `json:"files" toml:"files" yaml:"files"`
	Regex []string `json:"regex" toml:"regex" yaml:"regex"`
}

// ToExcluder returns an initialised *Excluder or a regexp error on failure.
func (e *ExcluderConfig) ToExcluder(rootdir string) (*Excluder, error) {
	// clean up filepaths
	for i, dir := range e.Dirs {
		e.Dirs[i] = filepath.Clean(dir)
	}

	for i, file := range e.Files {
		e.Files[i] = filepath.Clean(file)
	}

	var regexes []*regexp.Regexp
	for _, pattern := range e.Regex {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("excluder error: %w", err)
		}
		regexes = append(regexes, regex)
	}

	return &Excluder{
		RootDir: rootdir,
		Dirs:    ToSet(e.Dirs),
		Files:   ToSet(e.Files),
		Regex:   regexes,
	}, nil
}

type Excluder struct {
	RootDir string
	Dirs    map[string]struct{}
	Files   map[string]struct{}
	Regex   []*regexp.Regexp
}

// ShouldIgnore returns true if the path should be ignored, otherwise false.
// args:
// - path is the relative path to be checked.
// - isDir specifies whether the path is a directory.
func (e *Excluder) ShouldIgnore(path string, isDir bool) bool {
	relpath, err := filepath.Rel(e.RootDir, path)
	if path == "" || err != nil {
		return true
	}

	for dir := range e.Dirs {
		if relpath == dir || IsChild(dir, relpath) {
			return true
		}
	}

	if _, ok := e.Files[relpath]; ok {
		return true
	}

	for _, regex := range e.Regex {
		if regex.MatchString(relpath) {
			return true
		}
	}

	return false
}
