package notify

import (
	"regexp"

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
