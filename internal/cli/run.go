package cli

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dimmerz92/eavesdrop/internal/config"
)

var defaultConfigNames = []string{"eavesdrop.json", "eavesdrop.toml", "eavesdrop.yaml"}

func findDefaultConfig() (string, error) {
	for _, name := range defaultConfigNames {
		if _, err := os.Stat(name); err == nil {
			return name, nil
		}
	}
	return "", fmt.Errorf("no config file found; expected one of: %s", strings.Join(defaultConfigNames, ", "))
}

func RunEavesdrop(ctx context.Context) {
	println(Splash)

	path := flag.String("config", "", "the path to the config file")
	flag.Parse()

	configPath := *path
	if configPath == "" {
		var err error
		configPath, err = findDefaultConfig()
		if err != nil {
			panic(err)
		}
	}

	config, err := config.GetConfig(configPath)
	if err != nil {
		panic(err)
	}

	proxy, err := ConstructProxy(ctx, config.Proxy)
	if err != nil {
		panic(err)
	}

	emitter := ConstructEventEmitter(ctx, config)

	emitter.Start(ctx)

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

	<-ctx.Done()
}
