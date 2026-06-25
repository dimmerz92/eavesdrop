package cli

import (
	"context"
	"sync"

	"github.com/dimmerz92/eavesdrop"
	"github.com/dimmerz92/eavesdrop/internal/components"
	"github.com/dimmerz92/eavesdrop/internal/config"
)

func ConstructEventEmitter(ctx context.Context, config config.Config) *ev.EventEmitter {
	ops := make([]ev.Op, 0, len(config.GlobalExclude.Ops))
	for _, op := range config.GlobalExclude.Ops {
		ops = append(ops, ev.OpFromString(op))
	}

	return ev.NewEmitter(config.RootDir).
		WithExcluder(ev.NewExcluder(config.RootDir).
			WithOps(ops...).
			WithDirs(config.GlobalExclude.Dirs...).
			WithFiles(config.GlobalExclude.Files...).
			WithRegex(config.GlobalExclude.Regex...),
		)
}

func ConstructProxy(ctx context.Context, config config.ProxyConfig) (ev.Proxy, error) {
	if !config.Enabled {
		return nil, nil
	}

	proxy, err := components.NewProxy(ctx, config.AppPort, config.ProxyPort)
	if err != nil {
		return nil, err
	}

	return proxy, nil
}

func ConstructWatcher(
	ctx context.Context,
	root string,
	mu *sync.Mutex,
	proxy ev.Proxy,
	config config.WatcherConfig,
) *ev.Watcher {
	shell := ev.NewShell(ctx, config.Shell.TaskTimeout, config.Shell.ServiceShutdownTimeout)

	onChange := NewShellRunner(shell, config.Name, mu, config.Shell.Tasks, config.Shell.Service)

	ops := make([]ev.Op, 0, len(config.Exclude.Ops))
	for _, op := range config.Exclude.Ops {
		ops = append(ops, ev.OpFromString(op))
	}

	return ev.NewWatcher(ctx, config.Name, root).
		WithFiletypes(config.Filetypes...).
		WithDirs(config.Dirs...).
		WithFiles(config.Files...).
		WithOnChange(onChange).
		WithProxy(proxy, config.RefreshDelay).
		WithDebounceDelay(config.Shell.DebounceDelay).
		WithExcluder(ev.NewExcluder(root).
			WithOps(ops...).
			WithDirs(config.Exclude.Dirs...).
			WithFiles(config.Exclude.Files...).
			WithRegex(config.Exclude.Regex...),
		)
}
