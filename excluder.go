package ev

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dimmerz92/eavesdrop/internal/components"
)

type Excluder interface {
	// ShouldIgnore returns true when path or operation should be skipped.
	ShouldIgnore(e Event) bool
}

type excluder struct {
	root  string
	ops   Op
	dirs  components.Set[string]
	files components.Set[string]
	regex []*regexp.Regexp
}

// ExcluderOption configures an Excluder.
type ExcluderOption func(*excluder)

// WithOp adds file operations to the excluder that should be excluded.
func WithOp(ops ...Op) ExcluderOption {
	return func(e *excluder) {
		for _, op := range ops {
			e.ops |= op
		}
	}
}

// WithDirs adds directory paths relative to the excluder root to exclude.
// Only the exact specified path and its contents are excluded.
//
// e.g., "cmd" excludes ./cmd but not ./src/cmd.
// Use WithRegex for name-based matching at any depth.
func WithDirs(dirs ...string) ExcluderOption {
	return func(e *excluder) {
		if e.dirs == nil {
			e.dirs = components.Set[string]{}
		}
		for _, dir := range dirs {
			abs, err := filepath.Abs(dir)
			if err != nil {
				panic(err)
			}
			e.dirs[abs] = struct{}{}
		}
	}
}

// WithFiles adds file paths relative to the excluder root to exclude.
// Only the exact path is excluded.
//
// e.g., "go.sum" excludes ./go.sum but not ./vendor/go.sum.
// Use WithRegex for name-based matching at any depth.
func WithFiles(files ...string) ExcluderOption {
	return func(e *excluder) {
		if e.files == nil {
			e.files = components.Set[string]{}
		}
		for _, file := range files {
			abs, err := filepath.Abs(file)
			if err != nil {
				panic(err)
			}
			e.files[abs] = struct{}{}
		}
	}
}

// WithRegex adds regular expression patterns matched against the full path.
// Use this for name-based matching at any depth in the tree.
//
// e.g., to exclude every file named ".env" regardless of location.
func WithRegex(regex ...string) ExcluderOption {
	return func(e *excluder) {
		for _, pattern := range regex {
			e.regex = append(e.regex, regexp.MustCompile(pattern))
		}
	}
}

// NewExcluder returns a new Excluder rooted at root.
// Paths passed to ShouldIgnore are compared relative to root when checking
// dirs and files.
func NewExcluder(root string, opts ...ExcluderOption) (Excluder, error) {
	if strings.TrimSpace(root) == "" {
		root = "."
	}

	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("excluder: unable to determine absolute root")
	}

	e := &excluder{root: root}

	for _, opt := range opts {
		opt(e)
	}

	if e.dirs == nil {
		e.dirs = components.Set[string]{}
	}

	if e.files == nil {
		e.files = components.Set[string]{}
	}

	return e, nil
}

// ShouldIgnore returns true if the event should be ignored based on the
// excluder configuration.
func (e *excluder) ShouldIgnore(event Event) bool {
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
