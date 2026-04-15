package eavesdrop

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

type Event struct {
	file fs.FileInfo
}

type EventEmitter struct {
	rootdir  string
	subs     map[chan Event]struct{}
	cache    map[string]fs.FileInfo
	excluder *Excluder
	watcher  *fsnotify.Watcher
	mu       sync.RWMutex
}

type EventEmitterOption func(*EventEmitter)

func WithExcluder(excluder *Excluder) EventEmitterOption {
	return func(ee *EventEmitter) { ee.excluder = excluder }
}

func NewEmitter(rootdir string, opts ...EventEmitterOption) *EventEmitter {
	if rootdir == "" {
		panic("root directory cannot be blank")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	emitter := &EventEmitter{
		rootdir: rootdir,
		subs:    make(map[chan Event]struct{}),
		cache:   make(map[string]fs.FileInfo),
		watcher: watcher,
	}

	for _, opt := range opts {
		opt(emitter)
	}

	return emitter
}

func (e *EventEmitter) Start(ctx context.Context) {
	signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer stop()
		<-signalCtx.Done()

		err := e.watcher.Close()
		if err != nil {
			slog.Error("EventManager.Run", slog.Any("error", err))
		}

		for ch := range e.subs {
			close(ch)
		}
	}()

	err := e.RecursiveWatch(e.rootdir)
	if err != nil {
		color.Red("failed to watch: %s", e.rootdir)
		return
	}

	for {
		select {
		case event, ok := <-e.watcher.Events:
			if !ok {
				return
			}

			if event.Has(fsnotify.Chmod) {
				continue
			}

			file, ok := e.cache[event.Name]
			if !ok {
				var err error
				file, err = os.Stat(event.Name)
				if err != nil {
					color.Red("failed to read file: %s", event.Name)
					continue
				}
				e.cache[event.Name] = file
			}

			if e.excluder != nil && e.excluder.ShouldIgnore(file) {
				continue
			}

			if file.IsDir() {
				if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {
					err := e.RecursiveWatch(file.Name())
					if err != nil {
						color.Red("failed to watch recursively: %v", err)
						continue
					}
					color.Magenta("watching: %s", event.Name)
				}

				if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
					err := e.RecursiveUnwatch(file.Name())
					if err != nil {
						color.Red("failed to unwatch recursively: %v", err)
						continue
					}
					color.Magenta("unwatched: %s", event.Name)
				}
			}

			e.publish(Event{file: file})

		case err, ok := <-e.watcher.Errors:
			if !ok {
				return
			}
			slog.Error("EventManager", slog.Any("event loop error", err))
		}
	}
}

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

		if e.excluder != nil && e.excluder.ShouldIgnore(file) {
			return fs.SkipDir
		}

		err = e.watcher.Add(path)
		if err != nil {
			color.Red("failed to watch: %s", path)
		} else {
			color.Magenta("watching: %s", path)
		}

		return nil
	})
}

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
			color.Red("failed to unwatch: %s\nerror: %v", err)
		}

		return nil
	})
}

func (e *EventEmitter) Subscribe() <-chan Event {
	e.mu.Lock()
	defer e.mu.Unlock()

	ch := make(chan Event)
	e.subs[ch] = struct{}{}

	return ch
}

func (e *EventEmitter) publish(event Event) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	for ch := range e.subs {
		select {
		case ch <- event:
		default:
		}
	}
}
