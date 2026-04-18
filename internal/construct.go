package internal

import (
	"context"
	"sync"
	"time"

	"github.com/dimmerz92/eavesdrop"
)

func ConstructEventEmitter(ctx context.Context, config Config) eavesdrop.EventEmitter {
	return *eavesdrop.NewEmitter(
		config.RootDir,
		eavesdrop.WithGlobalExcluder(eavesdrop.NewExcluder(
			config.RootDir,
			eavesdrop.WithDirs(config.GlobalExclude.Dirs...),
			eavesdrop.WithFiles(config.GlobalExclude.Files...),
			eavesdrop.WithRegex(config.GlobalExclude.Regex...),
		)),
	)
}

func ConstructProxy(ctx context.Context, config ProxyConfig) eavesdrop.Proxy {
	if !config.Enabled {
		return nil
	}

	return eavesdrop.NewProxy(ctx,
		eavesdrop.WithAppPort(config.AppPort),
		eavesdrop.WithProxyPort(config.ProxyPort),
	)
}

func ConstructWatcher(ctx context.Context, root string, mu *sync.Mutex, proxy eavesdrop.Proxy, config WatcherConfig) eavesdrop.Watcher {
	return eavesdrop.NewWatcher(ctx, config.Name,
		mu,
		eavesdrop.WithWatchedFiletypes(config.Filetypes...),
		eavesdrop.WithWatchedDirs(config.Dirs...),
		eavesdrop.WithWatchedFiles(config.Files...),
		eavesdrop.WithWatcherExcluder(eavesdrop.NewExcluder(
			root,
			eavesdrop.WithDirs(config.Exclude.Dirs...),
			eavesdrop.WithFiles(config.Exclude.Files...),
			eavesdrop.WithRegex(config.Exclude.Regex...),
		)),
		eavesdrop.WithTasks(config.Shell.Tasks...),
		eavesdrop.WithService(config.Shell.Service),
		eavesdrop.WithTriggerRefresh(config.TriggerRefresh),
		eavesdrop.WithRefreshDelay(time.Millisecond*time.Duration(config.RefreshDelay)),
		eavesdrop.WithShell(eavesdrop.NewShell(ctx,
			eavesdrop.WithTaskTimeout(time.Millisecond*time.Duration(config.Shell.TaskTimeout)),
			eavesdrop.WithServiceTimeout(time.Millisecond*time.Duration(config.Shell.ServiceShutdownTimeout)),
		)),
		eavesdrop.WithDebouncer(eavesdrop.NewDebouncer(time.Millisecond*time.Duration(config.Shell.DebounceDelay))),
		eavesdrop.WithProxy(proxy),
	)
}
