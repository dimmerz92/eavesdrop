package eavesdrop

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fatih/color"
)

type WatcherConfig struct {
	Name              string         `json:"name" toml:"name" yaml:"name"`
	FileTypes         []string       `json:"file_types" toml:"file_types" yaml:"file_types"`
	FileNames         []string       `json:"file_names" toml:"file_names" yaml:"file_names"`
	Exclude           ExcluderConfig `json:"exclude" toml:"exclude" yaml:"exclude"`
	Tasks             []string       `json:"tasks" toml:"tasks" yaml:"tasks"`
	Service           string         `json:"service" toml:"service" yaml:"service"`
	RunOnStart        bool           `json:"run_on_start" toml:"run_on_start" yaml:"run_on_start"`
	MaxTaskTime       int64          `json:"max_task_time" toml:"max_task_time" yaml:"max_task_time"`
	MaxServiceTimeout int64          `json:"max_service_timeout" toml:"max_service_timeout" yaml:"max_service_timeout"`
	DebounceDelay     int64          `json:"debounce_delay" toml:"debounce_delay" yaml:"debounce_delay"`
	TriggerRefresh    bool           `json:"trigger_refresh" toml:"trigger_refresh" yaml:"trigger_refresh"`
}

// Validate checks to make sure the WatcherConfig fields are valid.
func (w *WatcherConfig) Validate() error {
	if w.Name == "" {
		return fmt.Errorf("watcher requires a name")
	}

	if len(w.FileTypes)+len(w.FileNames) == 0 {
		return fmt.Errorf("%s: at least one file type or file name is required", w.Name)
	}

	if len(w.Tasks) == 0 && w.Service == "" {
		return fmt.Errorf("%s: at least one task or service is required", w.Name)
	}

	if w.MaxTaskTime < 0 {
		return fmt.Errorf("%s: max task time cannot be negative", w.Name)
	}

	if w.MaxServiceTimeout < 0 {
		return fmt.Errorf("%s: service kill timeout cannot be negative", w.Name)
	}

	if w.DebounceDelay < 0 {
		return fmt.Errorf("%s: debounce delay cannot be negative", w.Name)
	}

	return nil
}

// ToWatcher returns an initialised and ready watcher.
// Runs the tasks and service if the RunOnStart flag is true.
func (w *WatcherConfig) ToWatcher(root string, proxy *Proxy) (*Watcher, error) {
	excluder, err := w.Exclude.ToExcluder(root)
	if err != nil {
		return nil, err
	}

	watcher := &Watcher{
		Name:           w.Name,
		Exts:           ToSet(w.FileTypes),
		Files:          ToSet(w.FileNames),
		Excluder:       excluder,
		Tasks:          w.Tasks,
		Service:        w.Service,
		Shell:          NewShell(time.Millisecond*time.Duration(w.MaxTaskTime), time.Millisecond*time.Duration(w.MaxServiceTimeout)),
		Debouncer:      &Debouncer{Delay: time.Millisecond * time.Duration(w.DebounceDelay)},
		Proxy:          proxy,
		TriggerRefresh: w.TriggerRefresh,
	}

	if w.RunOnStart {
		watcher.Debouncer.Run(func() {
			err := watcher.RunTasks()
			if err != nil {
				color.Red("%s error: %v", w.Name, err)
				return
			}

			err = watcher.RunService()
			if err != nil {
				color.Red("%s error: %v", w.Name, err)
				return
			}

			if watcher.TriggerRefresh {
				watcher.Proxy.Refresh()
			}
		})
	}

	return watcher, nil
}

type Watcher struct {
	Name           string
	Exts           map[string]struct{}
	Files          map[string]struct{}
	Excluder       *Excluder
	Tasks          []string
	Service        string
	Shell          *Shell
	Debouncer      *Debouncer
	Proxy          *Proxy
	TriggerRefresh bool
}

// Notify passes the file change event path to watcher to handle.
func (w *Watcher) Notify(path string) {
	_, watchedExt := w.Exts[filepath.Ext(path)]
	_, watchedFiles := w.Files[path]
	if !watchedExt && !watchedFiles {
		return
	}

	if w.Excluder.ShouldIgnore(path, false) {
		return
	}

	w.Debouncer.Run(func() {
		color.Green("%s changed", path)

		err := w.Shell.Kill()
		if err != nil {
			color.Red("%s kill error: %v", w.Name, err)
			return
		}

		err = w.RunTasks()
		if err != nil {
			color.Red("%s task error: %v", w.Name, err)
			return
		}

		err = w.RunService()
		if err != nil {
			color.Red("%s service error: %v", w.Name, err)
			return
		}

		if w.Proxy != nil && w.TriggerRefresh {
			w.Proxy.Refresh()
		}
	})
}

// RunTasks loops through the task list if any and executes them.
func (w *Watcher) RunTasks() error {
	if len(w.Tasks) > 0 {
		color.Magenta("%s: running tasks", w.Name)

		for _, task := range w.Tasks {
			output, err := w.Shell.Exec(task)
			if err != nil {
				return err
			}

			if output != "" {
				println(color.CyanString("%s:", w.Name), output)
			}
		}
	}

	return nil
}

// RunService runs the long/infinite running service if one exists in a detached process without waiting.
func (w *Watcher) RunService() error {
	if w.Service != "" {
		color.Blue("%s: running service", w.Name)

		err := w.Shell.Run(w.Service)
		if err != nil {
			return err
		}
	}

	return nil
}

// Close stops the debounce timer and kills the long running service if it exists.
func (w *Watcher) Close() error {
	if w.Debouncer.timer != nil {
		w.Debouncer.timer.Stop()
	}

	return w.Shell.Kill()
}
