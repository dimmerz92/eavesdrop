package ev

import (
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/dimmerz92/eavesdrop/internal/components"
)

// DefaultDebounceDelay is the default debounce delay in milliseconds applied to file change events.
const DefaultDebounceDelay = 100

var watcherRegistry = map[string]struct{}{}

type Proxy interface {
	RefreshBrowser()
}

// Watcher is a profile that defines which file system events to respond to and how.
// Add it to an EventEmitter to begin receiving events. Configure it with the With* builder methods.
type Watcher struct {
	name           string
	root           string
	filetypes      components.Set[string]
	dirs           components.Set[string]
	files          components.Set[string]
	onChange       func(Event)
	triggerRefresh bool
	refreshDelay   time.Duration
	proxy          Proxy
	debouncer      *components.Debouncer
	excluder       *Excluder
}

// NewWatcher returns a new Watcher profile rooted at root. name must be unique across all watchers
// in the process. If root is empty, it defaults to the current directory.
// Panics if name is empty or already registered.
func NewWatcher(name, root string) *Watcher {
	name = strings.TrimSpace(name)
	if name == "" {
		panic("watcher requires a non empty name")
	}

	if _, ok := watcherRegistry[name]; ok {
		panic("watcher requires a unique name: " + name)
	}

	watcherRegistry[name] = struct{}{}

	if strings.TrimSpace(root) == "" {
		root = "."
	}

	return &Watcher{
		name:      name,
		root:      root,
		filetypes: make(components.Set[string]),
		dirs:      make(components.Set[string]),
		files:     make(components.Set[string]),
		onChange:  func(_ Event) { slog.Warn("default handler", slog.String("watcher", name)) },
		debouncer: components.NewDebouncer(DefaultDebounceDelay),
	}
}

// Handle processes an event, calling the onChange handler if the event is watched and not excluded.
// Called by the EventEmitter; not intended for direct use.
func (w *Watcher) Handle(event Event) {
	if !w.Watched(event) {
		return
	}
	if w.excluder != nil && w.excluder.ShouldIgnore(event) {
		return
	}

	w.debouncer.Do(func() {
		slog.Info("file changed", slog.String("watcher", w.name), slog.String("path", event.Path()))
		w.onChange(event)
		if w.triggerRefresh {
			time.Sleep(w.refreshDelay)
			w.proxy.RefreshBrowser()
		}
	})
}

// Watched reports whether the event matches this watcher's root, filetypes, files, or dirs.
// Events with nil Info are always considered watched (e.g. manual triggers).
func (w *Watcher) Watched(event Event) bool {
	if event.Info() == nil {
		return true // for testing or manual triggering
	}

	rel, err := filepath.Rel(w.root, event.Path())
	if err != nil || strings.HasPrefix(rel, "..") {
		return false
	}

	if _, hasExt := w.filetypes[filepath.Ext(event.Path())]; hasExt {
		return true
	}

	if _, watchedFile := w.files[rel]; watchedFile {
		return true
	}

	for dir := range w.dirs {
		if components.IsRelative(dir, rel) {
			return true
		}
	}

	return false
}

// Trigger manually invokes the onChange handler with an empty event, bypassing filters and debounce.
func (w *Watcher) Trigger() {
	w.onChange(Event{})
}

// WithFiletypes adds file extensions to watch (e.g. ".go", ".html").
func (w *Watcher) WithFiletypes(filetypes ...string) *Watcher {
	for _, ftype := range filetypes {
		w.filetypes[ftype] = struct{}{}
	}
	return w
}

// WithDirs adds directories to watch, relative to the watcher root. Events from files inside
// these directories are matched; the directory path itself is not.
func (w *Watcher) WithDirs(dirs ...string) *Watcher {
	for _, dir := range dirs {
		w.dirs[dir] = struct{}{}
	}
	return w
}

// WithFiles adds specific file paths to watch, relative to the watcher root.
func (w *Watcher) WithFiles(files ...string) *Watcher {
	for _, file := range files {
		w.files[file] = struct{}{}
	}
	return w
}

// WithOnChange sets the handler called when a matching event is received.
func (w *Watcher) WithOnChange(fn func(Event)) *Watcher {
	w.onChange = fn
	return w
}

// WithProxy configures a Proxy to trigger a browser refresh after each onChange call.
// refreshDelayMs is the delay in milliseconds to wait before refreshing. A nil proxy is a no-op.
func (w *Watcher) WithProxy(proxy Proxy, refreshDelayMs uint) *Watcher {
	if proxy == nil {
		return w
	}
	w.triggerRefresh = true
	w.refreshDelay = time.Duration(refreshDelayMs) * time.Millisecond
	w.proxy = proxy
	return w
}

// WithDebounceDelay overrides the default debounce delay (DefaultDebounceDelay) in milliseconds.
func (w *Watcher) WithDebounceDelay(delayMs uint) *Watcher {
	w.debouncer.UpdateDelay(delayMs)
	return w
}

// WithExcluder attaches an Excluder that filters out events before they reach the onChange handler.
func (w *Watcher) WithExcluder(excluder *Excluder) *Watcher {
	w.excluder = excluder
	return w
}
