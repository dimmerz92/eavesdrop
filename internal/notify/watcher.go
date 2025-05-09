package notify

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dimmerz92/eavesdrop/internal/config"
	"github.com/dimmerz92/eavesdrop/internal/utils"
	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	Config      *config.Config
	IgnoreDirs  map[string]struct{}
	IgnoreFiles map[string]struct{}
	IgnoreRegex []*regexp.Regexp
	WatchFiles  map[string]struct{}
	WatchExts   map[string]struct{}
	*fsnotify.Watcher
}

// NewWatcher returns a Watcher with the config and set up complete.
func NewWatcher(cfg *config.Config) *Watcher {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	watcher := &Watcher{
		Config:      cfg,
		IgnoreDirs:  utils.SliceToSet(cfg.IgnoreDirs),
		IgnoreFiles: utils.SliceToSet(cfg.IgnoreFiles),
		IgnoreRegex: []*regexp.Regexp{},
		WatchFiles:  utils.SliceToSet(cfg.WatchFiles),
		WatchExts:   utils.SliceToSet(cfg.WatchExts),
		Watcher:     w,
	}

	// set ignored regex
	for _, regex := range cfg.IgnoreRegex {
		watcher.IgnoreRegex = append(watcher.IgnoreRegex, regexp.MustCompile(regex))
	}

	return watcher
}

// ShouldIgnoreDir returns true if the given path should be ignored, otherwise false.
func (w *Watcher) ShouldIgnoreDir(path string) bool {
	// account for absolute paths
	path = filepath.Clean(strings.TrimPrefix(path, w.Config.Root+string(filepath.Separator)))
	if path == "" {
		return true
	}

	if _, ok := w.IgnoreDirs[path]; ok {
		return true
	}

	for _, regex := range w.IgnoreRegex {
		if regex.MatchString(path) {
			return true
		}
	}

	return false
}
