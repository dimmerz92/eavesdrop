package notify

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dimmerz92/eavesdrop/internal/config"
	"github.com/dimmerz92/eavesdrop/internal/utils"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

type Watcher struct {
	Config      *config.Config
	IgnoreDirs  map[string]struct{}
	IgnoreFiles map[string]struct{}
	IgnoreRegex []*regexp.Regexp
	WatchFiles  map[string]struct{}
	WatchExts   map[string]struct{}
	Watched     map[string]struct{}
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
		Watched:     make(map[string]struct{}),
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

// Start runs the watcher event loop. Call as part of a goroutine and handle cleanup using w.Close().
func (w *Watcher) Start() {
	HandleNewDirectory(w.Config.Root, w)
	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return
			}

			// ignore chmod events
			if event.Has(fsnotify.Chmod) {
				continue
			}

			// directory check
			f, err := os.Stat(event.Name)

			if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
				if err == nil && f.IsDir() {
					HandleNewDirectory(event.Name, w)
				} else {
					// TODO: handle file events
				}
			} else if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				if _, ok := w.Watched[event.Name]; ok {
					HandleRemoveDirectory(event.Name, w)
				} else {
					// TODO: handle file events
				}
			}

		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			color.Red("error: %v", err)
		}
	}
}
