package eavesdrop

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

const (
	DefaultRefreshDelay  = 300 * time.Millisecond
	DefaultDebounceDelay = 300 * time.Millisecond
)

type Watcher struct {
	ctx            context.Context
	name           string
	filetypes      Set[string]
	dirs           Set[string]
	files          Set[string]
	tasks          []string
	service        string
	refreshDelay   time.Duration
	triggerRefresh bool
	debouncer      *Debouncer
	excluder       *Excluder
	proxy          *Proxy
	shell          *Shell
}

type WatcherOption func(*Watcher)

func WithWatchedFiletypes(filetypes ...string) WatcherOption {
	return func(w *Watcher) { w.filetypes = ToSet(filetypes...) }
}

func WithWatcherDirs(dirs ...string) WatcherOption {
	return func(w *Watcher) { w.dirs = ToSet(dirs...) }
}

func WithWatchedFiles(files ...string) WatcherOption {
	return func(w *Watcher) { w.files = ToSet(files...) }
}

func WithTasks(tasks ...string) WatcherOption {
	return func(w *Watcher) { w.tasks = tasks }
}

func WithService(service string) WatcherOption {
	return func(w *Watcher) { w.service = service }
}

func WithRefreshDelay(d time.Duration) WatcherOption {
	return func(w *Watcher) { w.refreshDelay = d }
}

func WithTriggerRefresh(b bool) WatcherOption {
	return func(w *Watcher) { w.triggerRefresh = b }
}

func WithDebouncer(debouncer *Debouncer) WatcherOption {
	return func(w *Watcher) { w.debouncer = debouncer }
}

func WithWatcherExcluder(excluder *Excluder) WatcherOption {
	return func(w *Watcher) { w.excluder = excluder }
}

func WithProxy(proxy *Proxy) WatcherOption {
	return func(w *Watcher) { w.proxy = proxy }
}

func WithShell(shell *Shell) WatcherOption {
	return func(w *Watcher) { w.shell = shell }
}

func NewWatcher(ctx context.Context, name string, opts ...WatcherOption) *Watcher {
	if name = strings.TrimSpace(name); name == "" {
		panic("watcher requires a name")
	}

	watcher := &Watcher{
		ctx:          ctx,
		name:         name,
		refreshDelay: DefaultRefreshDelay,
	}

	for _, opt := range opts {
		opt(watcher)
	}

	if watcher.filetypes == nil {
		watcher.filetypes = Set[string]{}
	}

	if watcher.dirs == nil {
		watcher.dirs = Set[string]{}
	}

	if watcher.files == nil {
		watcher.files = Set[string]{}
	}

	if len(watcher.filetypes)+len(watcher.dirs)+len(watcher.files) == 0 {
		panic(fmt.Sprintf("watcher %s has nothing to watch", watcher.name))
	}

	if len(watcher.tasks) == 0 && watcher.service == "" {
		panic(fmt.Sprintf("watcher %s has no tasks or servies to run", watcher.name))
	}

	if watcher.debouncer == nil {
		watcher.debouncer = NewDebouncer(DefaultDebounceDelay)
	}

	if watcher.shell == nil {
		watcher.shell = NewShell(ctx)
	}

	return watcher
}

func (w *Watcher) Watch(events <-chan Event) {
	w.runJobs()
	for {
		select {
		case <-w.ctx.Done():
			return

		case event := <-events:
			if w.watched(event) {
				w.runJobs()
			}
		}
	}
}

func (w *Watcher) watched(event Event) bool {
	if _, hasExt := w.filetypes[filepath.Ext(event.file.Name())]; hasExt {
		return true
	}

	if _, watchedFile := w.files[event.file.Name()]; watchedFile {
		return true
	}

	for dir := range w.dirs {
		if IsRelative(dir, event.file.Name()) {
			return true
		}
	}

	return false
}

func (w *Watcher) runJobs() {
	if err := w.shell.KillProcessGroup(); err != nil {
		color.Red("%s: failed to kill previous service: %v", w.name, err)
	}

	for _, task := range w.tasks {
		if err := w.shell.ExecAndWait(task); err != nil {
			color.Red("%s: failed to run task: %v", w.name, err)
		}
	}

	if w.service != "" {
		if err := w.shell.ExecAndReturn(w.service); err != nil {
			color.Red("%s: failed to run service: %v", w.name, err)
		}
	}
}
