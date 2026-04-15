package eavesdrop

import (
	"io/fs"
	"regexp"
)

type Excluder struct {
	dirs  Set[string]
	files Set[string]
	regex []*regexp.Regexp
}

type ExcluderOption func(*Excluder)

func WithDirs(dirs ...string) ExcluderOption {
	return func(e *Excluder) { e.dirs = ToSet(dirs...) }
}

func WithFiles(files ...string) ExcluderOption {
	return func(e *Excluder) { e.files = ToSet(files...) }
}

func WithRegex(regex ...string) ExcluderOption {
	return func(e *Excluder) {
		for _, pattern := range regex {
			e.regex = append(e.regex, regexp.MustCompile(pattern))
		}
	}
}

func NewExcluder(opts ...ExcluderOption) *Excluder {
	excluder := &Excluder{}

	for _, opt := range opts {
		opt(excluder)
	}

	if excluder.dirs == nil {
		excluder.dirs = Set[string]{}
	}

	if excluder.files == nil {
		excluder.files = Set[string]{}
	}

	return excluder
}

func (e *Excluder) ShouldIgnore(file fs.FileInfo) bool {
	switch file.IsDir() {
	case true:
		if _, ok := e.dirs[file.Name()]; ok {
			return true
		}

	default:
		if _, ok := e.files[file.Name()]; ok {
			return true
		}
	}

	for _, regex := range e.regex {
		if regex.MatchString(file.Name()) {
			return true
		}
	}

	return false
}
