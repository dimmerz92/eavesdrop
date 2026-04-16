package internal

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"path/filepath"
)

func RunEavesdrop(ctx context.Context) {
	println(Splash)

	path := flag.String("config", "eavesdrop.json", "the path to the config file")
	flag.Parse()

	config, err := GetConfig(*path)
	if err != nil {
		panic(err)
	}

	emitter := ConstructEventEmitter(ctx, config)
	proxy := ConstructProxy(ctx, config.Proxy)
	watchers := ConstructWatchers(ctx, proxy, config.Watchers)

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

	for _, watcher := range watchers {
		go watcher.Watch(emitter.Subscribe())
	}

	<-ctx.Done()
}
