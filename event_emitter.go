package ev

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
)

// EventEmitter watches a directory tree for file system events and dispatches
// them to registered Subscribers. Add watchers via Subscribe and call Start to begin.
type EventEmitter struct {
	root        string
	cache       map[string]fs.FileInfo
	excluder    *Excluder
	watcher     *fsnotify.Watcher
	subscribers []Subscriber
	mu          sync.RWMutex
}

// NewEmitter returns a new EventEmitter rooted at root. Panics if root is empty.
func NewEmitter(root string) *EventEmitter {
	if root == "" {
		panic("root directory cannot be blank")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	return &EventEmitter{
		root:    root,
		cache:   make(map[string]fs.FileInfo),
		watcher: watcher,
	}
}

// Start begins watching the root directory tree and dispatching events to subscribers.
// Newly created directories are watched automatically; removed directories are unwatched.
// Stops when ctx is cancelled.
func (e *EventEmitter) Start(ctx context.Context) {
	go func() {
		<-ctx.Done()

		err := e.watcher.Close()
		if err != nil {
			slog.Error("EventManager.Run", slog.Any("error", err))
		}
	}()

	err := e.RecursiveWatch(e.root)
	if err != nil {
		slog.Error("failed to watch", slog.String("dir", e.root))
		return
	}

	go func() {
		for {
			select {
			case fevent, ok := <-e.watcher.Events:
				if !ok {
					return
				}

				file, ok := e.cache[fevent.Name]
				if !ok {
					var err error
					file, err = os.Stat(fevent.Name)
					if err != nil {
						// log nothing, this is noisy and usually as a result of temp files.
						continue
					}
					e.cache[fevent.Name] = file
				}

				event := NewEvent(Op(fevent.Op), fevent.Name, file)

				if e.excluder != nil && e.excluder.ShouldIgnore(event) {
					continue
				}

				if file.IsDir() {
					if event.Has(CREATE) || event.Has(WRITE) {
						err := e.RecursiveWatch(event.Path())
						if err != nil {
							slog.Error("failed to watch recursively", slog.Any("error", err))
						}
					}

					if event.Has(REMOVE) || event.Has(RENAME) {
						err := e.RecursiveUnwatch(event.Path())
						if err != nil {
							slog.Error("failed to unwatch recursively", slog.Any("error", err))
						}
						slog.Info("unwatched", slog.String("path", event.Path()))
					}
				}

				e.publish(event)

			case err, ok := <-e.watcher.Errors:
				if !ok {
					return
				}
				slog.Error("EventManager", slog.Any("event loop error", err))
			}
		}
	}()
}

// RecursiveWatch adds dir and all subdirectories to the watch list, populating the file cache.
// Excluded directories (via WithExcluder) are skipped entirely.
func (e *EventEmitter) RecursiveWatch(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		file, err := d.Info()
		if err != nil {
			return nil
		}
		e.cache[path] = file

		if !d.IsDir() {
			return nil
		}

		if e.excluder != nil && e.excluder.ShouldIgnore(Event{path: path, info: file}) {
			return fs.SkipDir
		}

		err = e.watcher.Add(path)
		if err != nil {
			slog.Error("failed to watch", slog.String("path", path))
			return nil
		}

		slog.Info("watching", slog.String("path", path))

		return nil
	})
}

// RecursiveUnwatch removes dir and all subdirectories from the watch list and clears them from the cache.
func (e *EventEmitter) RecursiveUnwatch(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		delete(e.cache, path)

		if !d.IsDir() {
			return nil
		}

		err = e.watcher.Remove(path)
		if err != nil && !errors.Is(err, fsnotify.ErrNonExistentWatch) {
			slog.Error("failed to unwatch", slog.String("path", path), slog.Any("error", err))
		}

		return nil
	})
}

// Subscriber is implemented by any type that can receive file system events from an EventEmitter.
type Subscriber interface {
	Handle(event Event)
}

// Subscribe registers a Subscriber to receive events. Must be called before Start.
func (e *EventEmitter) Subscribe(subscriber Subscriber) {
	e.subscribers = append(e.subscribers, subscriber)
}

func (e *EventEmitter) publish(event Event) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for _, sub := range e.subscribers {
		sub.Handle(event)
	}
}

// WithExcluder attaches an Excluder that filters events and directories before they are watched or dispatched.
func (e *EventEmitter) WithExcluder(excluder *Excluder) *EventEmitter {
	e.excluder = excluder
	return e
}
