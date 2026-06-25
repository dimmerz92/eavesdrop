package ev

import (
	"regexp"
	"strings"

	"github.com/dimmerz92/eavesdrop/internal/components"
)

type Excluder struct {
	root  string
	ops   Op
	dirs  components.Set[string]
	files components.Set[string]
	regex []*regexp.Regexp
}

// NewExcluder returns a new Excluder rooted at at the given root.
func NewExcluder(root string) *Excluder {
	if strings.TrimSpace(root) == "" {
		root = "."
	}

	return &Excluder{
		root:  root,
		dirs:  make(components.Set[string]),
		files: make(components.Set[string]),
	}
}

// ShouldIgnore returns true when operation or path should be skipped.
func (e *Excluder) ShouldIgnore(event Event) bool {
	if event.Has(e.ops) {
		return true
	}

	for dir := range e.dirs {
		if event.Path() == dir || components.IsRelative(dir, event.Path()) {
			return true
		}
	}

	if event.info == nil || !event.info.IsDir() {
		if _, ok := e.files[event.Path()]; ok {
			return true
		}
	}

	for _, pattern := range e.regex {
		if pattern.MatchString(event.path) {
			return true
		}
	}

	return false
}

// WithOps adds file operations to the excluder that should be excluded.
func (e *Excluder) WithOps(ops ...Op) *Excluder {
	for _, op := range ops {
		e.ops |= op
	}
	return e
}

// WithDirs adds directory paths relative to the excluder root to exclude.
// Only the exact specified path and its contents are excluded.
func (e *Excluder) WithDirs(dirs ...string) *Excluder {
	for _, dir := range dirs {
		e.dirs[dir] = struct{}{}
	}
	return e
}

// WithFiles adds file paths relative to the excluder root to exclude.
// Only the exact path is excluded.
func (e *Excluder) WithFiles(files ...string) *Excluder {
	for _, file := range files {
		e.files[file] = struct{}{}
	}
	return e
}

// WithRegex adds regular expression patterns matched against the full path.
// Use this for name-based matching at any depth in the tree.
func (e *Excluder) WithRegex(regex ...string) *Excluder {
	for _, pattern := range regex {
		e.regex = append(e.regex, regexp.MustCompile(pattern))
	}
	return e
}
