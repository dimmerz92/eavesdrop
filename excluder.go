package eavesdrop

import (
	"fmt"
	"regexp"
)

type ExcluderConfig struct {
	Dirs  []string `json:"dirs" toml:"dirs" yaml:"dirs"`
	Files []string `json:"files" toml:"files" yaml:"files"`
	Regex []string `json:"regex" toml:"regex" yaml:"regex"`
}

// ToExcluder returns an initialised *Excluder or a regexp error on failure.
func (e *ExcluderConfig) ToExcluder(rootDir string) (*Excluder, error) {
	var regexes []*regexp.Regexp
	for _, pattern := range e.Regex {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("excluder error: %w", err)
		}
		regexes = append(regexes, regex)
	}

	return &Excluder{
		Dirs:  ToSet(e.Dirs),
		Files: ToSet(e.Files),
		Regex: regexes,
	}, nil
}

type Excluder struct {
	Dirs  map[string]struct{}
	Files map[string]struct{}
	Regex []*regexp.Regexp
}

// ShouldIgnore returns true if the path should be ignored, otherwise false.
func (e *Excluder) ShouldIgnore(path string, isDir bool) bool {
	if isDir {
		if _, ok := e.Dirs[path]; ok {
			return true
		}
	}

	if _, ok := e.Files[path]; ok {
		return true
	}

	for _, regex := range e.Regex {
		if regex.MatchString(path) {
			return true
		}
	}

	return false
}
