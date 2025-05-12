package notify

import (
	"errors"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dimmerz92/eavesdrop/internal/config"
	"github.com/dimmerz92/eavesdrop/internal/utils"
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
		IgnoreDirs:  utils.SliceToSet(cfg.IgnoreDirs),
		IgnoreFiles: utils.SliceToSet(cfg.IgnoreFiles),
		IgnoreRegex: []*regexp.Regexp{},
		WatchFiles:  utils.SliceToSet(cfg.WatchFiles),
		WatchExts:   utils.SliceToSet(cfg.WatchExts),
		WatchedDirs: make(map[string]struct{}),
		Watcher:     watcher,
	}

	for _, regex := range cfg.IgnoreRegex {
		notifier.IgnoreRegex = append(notifier.IgnoreRegex, regexp.MustCompile(regex))
	}

	return notifier
}

// ShouldIgnore checks whether the given event file or directory should be ignored.
// If a file, explicitly watched files take precedence over ignored files and regex.
func (n *Notifier) ShouldIgnore(path string, isDir bool) bool {
	// account for absolute paths
	path = filepath.Clean(strings.TrimPrefix(path, n.Config.Root+string(filepath.Separator)))
	if path == "" {
		return true
	}

	if isDir {
		if _, ok := n.IgnoreDirs[path]; ok {
			return true
		}
	} else {
		if _, ok := n.WatchFiles[path]; ok {
			return false
		}
		if _, ok := n.IgnoreFiles[path]; ok {
			return true
		}
	}

	for _, regex := range n.IgnoreRegex {
		if regex.MatchString(path) {
			return true
		}
	}

	return false
}

// HandleNewDir recursively adds the directories at the given path if not ignored.
func (n *Notifier) HandleNewDir(path string) {
	filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}

		if n.ShouldIgnore(path, true) {
			return fs.SkipDir
		}

		if err := n.Add(path); err != nil {
			utils.PrintError("failed to watch %s with err %v", path, err)
		} else {
			n.WatchedDirs[path] = struct{}{}
			utils.PrintWatching("watching %s", path)
		}

		return nil
	})
}

// HandleRemovedDir recursively removes watch on directories at the given path.
func (n *Notifier) HandleRemovedDir(path string) {
	if err := n.Remove(path); err != nil && !errors.Is(err, fsnotify.ErrNonExistentWatch) {
		utils.PrintError("failed to unwatch %s with error %v", path, err)
	} else {
		utils.PrintWatching("unwatched %s", path)

		delete(n.WatchedDirs, path)

		for dir := range n.WatchedDirs {
			if strings.HasPrefix(dir, path+string(filepath.Separator)) {
				delete(n.WatchedDirs, dir)
			}
		}
	}
}
