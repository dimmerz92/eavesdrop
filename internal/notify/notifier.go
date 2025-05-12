package notify

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

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
		if _, ok := n.WatchExts[filepath.Ext(path)]; !ok {
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

// HandleBuild runs the Exec.Build command with the Config.Build directive with logging.
func (n *Notifier) HandleBuild() {
	utils.PrintBuild("building...")
	if out, err := n.Exec.Build(n.Config.Build); err != nil {
		utils.PrintError(out)
	} else if out != "" {
		println(out)
	}
}

// HandleRun runs the Exec.Run command with the Config.Run directive with logging.
func (n *Notifier) HandleRun() {
	utils.PrintRun("running...")
	if err := n.Exec.Run(n.Config.Run); err != nil {
		utils.PrintError("%v", err)
	}
}

// HandleReset kills any existing processes and runs the Exec.BuildExec.Run command with the Config.Build and Config.Run
// directives with logging.
func (n *Notifier) HandleReset() {
	if err := n.Exec.Kill(); err != nil {
		utils.PrintError("%v", err)
		return
	}

	utils.PrintBuild("building...")
	if out, err := n.Exec.Build(n.Config.Build); err != nil {
		utils.PrintError(out)
		return
	} else if out != "" {
		println(out)
	}

	utils.PrintRun("running...")
	if err := n.Exec.Run(n.Config.Run); err != nil {
		utils.PrintError("%v", err)
	}
}

func (n *Notifier) Start() {
	// initial set up
	n.HandleNewDir(n.Config.Root)
	n.HandleBuild()
	n.HandleRun()

	// main event loop
	for {
		select {
		case event, ok := <-n.Events:
			// return on channel closure
			if !ok {
				return
			}

			// ignore chmod events
			if event.Has(fsnotify.Chmod) {
				continue
			}

			// handle directory events
			if f, err := os.Stat(event.Name); err == nil && f.IsDir() {
				if n.ShouldIgnore(event.Name, true) {
					continue
				}
				if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
					n.HandleNewDir(event.Name)
				} else {
					n.HandleRemovedDir(event.Name)
				}

				continue
			}

			// handle file events
			if n.ShouldIgnore(event.Name, false) {
				continue
			}

			utils.PrintFileChange("%s changed", event.Name)
			n.Debouncer.Run(200*time.Millisecond, n.HandleReset)

		case err, ok := <-n.Errors:
			// return on channel closure
			if !ok {
				return
			}
			utils.PrintError("%v", err)
		}
	}
}

// Stop stops any running processes and closes the watcher.
func (n *Notifier) Stop() {
	n.Debouncer.timer.Stop()
	n.Exec.Kill()
	n.Close()
}
