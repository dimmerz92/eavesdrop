package notify

import (
	"regexp"

	"github.com/dimmerz92/eavesdrop/internal/config"
	"github.com/fsnotify/fsnotify"
)

type Notifier struct {
	Config      *config.Config
	Debouncer   *Debouncer
	Exec        *Exec
	IgnoreDirs  map[string]struct{}
	IgnoreFiles map[string]struct{}
	IgnoreRegex []*regexp.Regexp
	WatchFiles  map[string]struct{}
	WatchExts   map[string]struct{}
	WatchedDirs map[string]struct{}
	*fsnotify.Watcher
}

// NewNotifier returns a newly constructed Notifier.
func NewNotifier(cfg *config.Config) *Notifier {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	notifier := &Notifier{
		Config:      cfg,
		Debouncer:   &Debouncer{},
		Exec:        &Exec{},
		IgnoreDirs:  SliceToSet(cfg.IgnoreDirs),
		IgnoreFiles: SliceToSet(cfg.IgnoreFiles),
		IgnoreRegex: []*regexp.Regexp{},
		WatchFiles:  SliceToSet(cfg.WatchFiles),
		WatchExts:   SliceToSet(cfg.WatchExts),
		WatchedDirs: make(map[string]struct{}),
		Watcher:     watcher,
	}

	for _, regex := range cfg.IgnoreRegex {
		notifier.IgnoreRegex = append(notifier.IgnoreRegex, regexp.MustCompile(regex))
	}

	return notifier
}
