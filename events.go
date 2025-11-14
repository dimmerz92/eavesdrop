package eavesdrop

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

const (
	STARTUP_DELAY = 50
	TMP_PERMS     = 0755
)

type EventManager struct {
	*fsnotify.Watcher

	RootDir    string
	CleanupTmp bool
	Watching   map[string]struct{}
	Excluder   *Excluder
	Watchers   []*Watcher
	StatCache  map[string]os.FileInfo
	Proxy      *Proxy
}

func NewEventManager(config Config) (*EventManager, error) {
	excluder, err := config.Exclude.ToExcluder(config.RootDir)
	if err != nil {
		return nil, err
	}

	if config.Tmp {
		err = os.MkdirAll(filepath.Join(config.RootDir, "tmp"), TMP_PERMS)
		if err != nil {
			return nil, fmt.Errorf("error: %v", err)
		}
	}

	manager := &EventManager{
		RootDir:    config.RootDir,
		CleanupTmp: config.CleanupTmp,
		Watching:   make(map[string]struct{}),
		Excluder:   excluder,
		StatCache:  make(map[string]os.FileInfo),
	}

	manager.Watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = manager.Add(config.RootDir)
	if err != nil {
		return nil, err
	}
	manager.HandleNewDir(config.RootDir)

	if config.Proxy.Enabled {
		manager.Proxy = config.Proxy.ToProxy()
		if manager.Proxy != nil {
			go func() {
				err := manager.Proxy.Server.ListenAndServe()
				if err != nil && err != http.ErrServerClosed {
					panic(fmt.Sprintf("proxy error: server failed: %v", err))
				}
			}()
		}
	}

	for _, watcherConfig := range config.Watchers {
		watcher, err := watcherConfig.ToWatcher(config.RootDir, manager.Proxy)
		if err != nil {
			return nil, err
		}
		manager.Watchers = append(manager.Watchers, watcher)
		time.Sleep(STARTUP_DELAY * time.Millisecond) // ensures the watchers run in order on start up
	}

	return manager, nil
}

// Start runs the event manager event loop.
func (e *EventManager) Start() {
	for {
		select {
		case event, ok := <-e.Events:
			if !ok {
				return
			}

			if event.Has(fsnotify.Chmod) {
				continue
			}

			var err error
			f, ok := e.StatCache[event.Name]
			if !ok {
				f, err = os.Stat(event.Name)
				if err == nil {
					e.StatCache[event.Name] = f
				}
			}

			if err != nil || e.Excluder.ShouldIgnore(event.Name, f.IsDir()) {
				continue
			}

			if f.IsDir() {
				if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
					e.HandleNewDir(event.Name)
				}

				if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					e.HandleRemovedDir(event.Name)
				}

				continue
			}

			for _, watcher := range e.Watchers {
				watcher.Notify(event.Name)
			}

		case err, ok := <-e.Errors:
			if !ok {
				return
			}
			panic(err)
		}
	}
}

// HandleNewDir recursively adds directories at the given path if not ignored.
func (e *EventManager) HandleNewDir(path string) {
	err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		f, err := d.Info()
		if err != nil {
			return nil
		}
		e.StatCache[path] = f

		if e.Excluder.ShouldIgnore(path, d.IsDir()) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		err = e.Add(path)
		if err != nil {
			color.Red("failed to watch %s with err %v", path, err)
		} else {
			e.Watching[path] = struct{}{}
			if d.IsDir() {
				color.Magenta("watching %s", path)
			}
		}

		return nil
	})

	if err != nil {
		panic(fmt.Sprintf("file walk error occurred: %v", err))
	}
}

// HandleRemovedDir recursively removes watch on directories at the given path.
func (e *EventManager) HandleRemovedDir(path string) {
	err := e.Remove(path)
	if err != nil && !errors.Is(err, fsnotify.ErrNonExistentWatch) {
		color.Red("failed to unwatch %s with error %v", path, err)
		return
	}

	for watched := range e.Watching {
		if IsChild(path, watched) {
			delete(e.StatCache, watched)
			delete(e.Watching, watched)
		}
	}

	delete(e.StatCache, path)
	delete(e.Watching, path)

	color.Magenta("unwatched %s", path)
}

// Stop calls stop on all watchers, closes the proxy if running, and closes the fsnotify watcher.
func (e *EventManager) Stop() {
	for _, watcher := range e.Watchers {
		err := watcher.Close()
		if err != nil && !strings.Contains(err.Error(), "terminated") {
			color.Red("%s: %v", watcher.Name, err)
		}
	}

	if e.Proxy != nil {
		err := e.Proxy.Server.Close()
		if err != nil {
			color.Red(err.Error())
		}
	}

	if e.CleanupTmp {
		err := os.RemoveAll(filepath.Join(e.RootDir, "tmp"))
		if err != nil {
			color.Red(err.Error())
		}
	}

	err := e.Close()
	if err != nil {
		color.Red(err.Error())
	}
}
