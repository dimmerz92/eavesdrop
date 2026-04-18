package eavesdrop

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

const (
	DefaultRefreshDelay  = 100 * time.Millisecond
	DefaultDebounceDelay = 100 * time.Millisecond
)

type Watcher interface {
	Watch(events <-chan Event)
	RunJobs()
}

type watcher struct {
	ctx            context.Context
	name           string
	filetypes      Set[string]
	dirs           Set[string]
	files          Set[string]
	tasks          []string
	service        string
	triggerRefresh bool
	refreshDelay   time.Duration
	debouncer      Debouncer
	excluder       Excluder
	proxy          Proxy
	shell          Shell
	mu             *sync.Mutex
}

type WatcherOption func(*watcher)

func WithWatchedFiletypes(filetypes ...string) WatcherOption {
	return func(w *watcher) { w.filetypes = ToSet(filetypes...) }
}

func WithWatchedDirs(dirs ...string) WatcherOption {
	return func(w *watcher) { w.dirs = ToSet(dirs...) }
}

func WithWatchedFiles(files ...string) WatcherOption {
	return func(w *watcher) { w.files = ToSet(files...) }
}

func WithTasks(tasks ...string) WatcherOption {
	return func(w *watcher) { w.tasks = tasks }
}

func WithService(service string) WatcherOption {
	return func(w *watcher) { w.service = service }
}

func WithRefreshDelay(d time.Duration) WatcherOption {
	return func(w *watcher) { w.refreshDelay = d }
}

func WithTriggerRefresh(b bool) WatcherOption {
	return func(w *watcher) { w.triggerRefresh = b }
}

func WithDebouncer(debouncer Debouncer) WatcherOption {
	return func(w *watcher) { w.debouncer = debouncer }
}

func WithWatcherExcluder(excluder Excluder) WatcherOption {
	return func(w *watcher) { w.excluder = excluder }
}

func WithProxy(proxy Proxy) WatcherOption {
	return func(w *watcher) { w.proxy = proxy }
}

func WithShell(shell Shell) WatcherOption {
	return func(w *watcher) { w.shell = shell }
}

func NewWatcher(ctx context.Context, name string, mu *sync.Mutex, opts ...WatcherOption) *watcher {
	if name = strings.TrimSpace(name); name == "" {
		panic("watcher requires a name")
	}

	watcher := &watcher{
		ctx:          ctx,
		name:         name,
		refreshDelay: DefaultRefreshDelay,
		mu:           mu,
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

func (w *watcher) Watch(events <-chan Event) {
	for {
		select {
		case <-w.ctx.Done():
			return

		case event := <-events:
			if w.watched(event) && !w.excluder.ShouldIgnore(event.file) {
				color.Green("%s changed", event.file.Name())
				w.debouncer.Do(func() {
					w.RunJobs()
					if w.triggerRefresh && w.proxy != nil {
						time.Sleep(w.refreshDelay)
						w.proxy.RefreshBrowser()
					}
				})
			}
		}
	}
}

func (w *watcher) watched(event Event) bool {
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

func (w *watcher) RunJobs() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.shell.KillProcessGroup(); err != nil {
		color.Red("%s: failed to kill previous service: %v", w.name, err)
	}

	for _, task := range w.tasks {
		fmt.Printf("%s: running task: %s\n", color.CyanString(w.name), task)
		if err := w.shell.ExecAndWait(task); err != nil {
			color.Red("%s: failed to run task: %v", w.name, err)
		}
	}

	if w.service != "" {
		fmt.Printf("%s: running service: %s\n", color.BlueString(w.name), w.service)
		if err := w.shell.ExecAndReturn(w.service); err != nil {
			color.Red("%s: failed to run service: %v", w.name, err)
		}
	}
}
