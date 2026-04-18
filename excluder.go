package eavesdrop

import (
	"io/fs"
	"path/filepath"
	"regexp"
)

type Excluder interface {
	ShouldIgnore(file fs.FileInfo) bool
}

type excluder struct {
	root  string
	dirs  Set[string]
	files Set[string]
	regex []*regexp.Regexp
}

type ExcluderOption func(*excluder)

func WithDirs(dirs ...string) ExcluderOption {
	return func(e *excluder) { e.dirs = ToSet(dirs...) }
}

func WithFiles(files ...string) ExcluderOption {
	return func(e *excluder) { e.files = ToSet(files...) }
}

func WithRegex(regex ...string) ExcluderOption {
	return func(e *excluder) {
		for _, pattern := range regex {
			e.regex = append(e.regex, regexp.MustCompile(pattern))
		}
	}
}

// NewExcluder returns a new instance of the default Excluder implementation.
func NewExcluder(root string, opts ...ExcluderOption) *excluder {
	excluder := &excluder{}

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

// ShouldIgnore checks the given file info agains the exclude rules and returns true if it should be ignored.
func (e *excluder) ShouldIgnore(file fs.FileInfo) bool {
	switch file.IsDir() {
	case true:
		rel, _ := filepath.Rel(e.root, file.Name())
		for dir := range e.dirs {
			if rel == dir || IsRelative(dir, rel) {
				return true
			}
		}

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
