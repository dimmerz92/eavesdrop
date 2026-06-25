package cli

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/dimmerz92/eavesdrop/internal/config"
)

func RunEavesdrop(ctx context.Context) {
	println(Splash)

	path := flag.String("config", "eavesdrop.json", "the path to the config file")
	flag.Parse()

	config, err := config.GetConfig(*path)
	if err != nil {
		panic(err)
	}

	proxy, err := ConstructProxy(ctx, config.Proxy)
	if err != nil {
		panic(err)
	}

	emitter := ConstructEventEmitter(ctx, config)

	var mu sync.Mutex
	for _, watcherConfig := range config.Watchers {
		watcher := ConstructWatcher(ctx, config.RootDir, &mu, proxy, watcherConfig)
		emitter.Subscribe(watcher)
		if watcherConfig.RunOnStart {
			watcher.Trigger()
		}
	}

	if config.Tmp {
		err := os.MkdirAll(filepath.Join(config.RootDir, "tmp"), 0755)
		if err != nil {
			panic(err)
		}

		if config.CleanupTmp {
			defer func() {
				err := os.RemoveAll(filepath.Join(config.RootDir, "tmp"))
				if err != nil {
					slog.Error("tmp cleanup", slog.Any("error", err))
				}
			}()
		}
	}

	emitter.Start(ctx)

	<-ctx.Done()
}
