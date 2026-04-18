package internal

import (
	"fmt"
	"path/filepath"

	"github.com/dimmerz92/eavesdrop"
)

type Config struct {
	RootDir       string          `json:"root_dir" toml:"root_dir" yaml:"root_dir"`
	Tmp           bool            `json:"tmp" toml:"tmp" yaml:"tmp"`
	CleanupTmp    bool            `json:"cleanup_tmp" toml:"cleanup_tmp" yaml:"cleanup_tmp"`
	GlobalExclude ExcluderConfig  `json:"global_exclude" toml:"global_exclude" yaml:"global_exclude"`
	Watchers      []WatcherConfig `json:"watchers" toml:"watchers" yaml:"watchers"`
	Proxy         ProxyConfig     `json:"proxy" toml:"proxy" yaml:"proxy"`
}

type ExcluderConfig struct {
	Dirs  []string `json:"dirs" toml:"dirs" yaml:"dirs"`
	Files []string `json:"files" toml:"files" yaml:"files"`
	Regex []string `json:"regex" toml:"regex" yaml:"regex"`
}

type WatcherConfig struct {
	Name           string         `json:"name" toml:"name" yaml:"name"`
	Filetypes      []string       `json:"filetypes" toml:"filetypes" yaml:"filetypes"`
	Dirs           []string       `json:"dirs" toml:"dirs" yaml:"dirs"`
	Files          []string       `json:"files" toml:"files" yaml:"files"`
	Exclude        ExcluderConfig `json:"exclude" toml:"exclude" yaml:"exclude"`
	Shell          ShellConfig    `json:"shell" toml:"shell" yaml:"shell"`
	RunOnStart     bool           `json:"run_on_start" toml:"run_on_start" yaml:"run_on_start"`
	TriggerRefresh bool           `json:"trigger_refresh" toml:"trigger_refresh" yaml:"trigger_refresh"`
	RefreshDelay   uint           `json:"refresh_delay" toml:"refresh_delay" yaml:"refresh_delay"`
}

type ShellConfig struct {
	Tasks                  []string `json:"tasks" toml:"tasks" yaml:"tasks"`
	TaskTimeout            uint     `json:"task_timeout" toml:"task_timeout" yaml:"task_timeout"`
	Service                string   `json:"service" toml:"service" yaml:"service"`
	ServiceShutdownTimeout uint     `json:"service_shutdown_timeout" toml:"service_shutdown_timeout" yaml:"service_shutdown_timeout"`
	DebounceDelay          uint     `json:"debounce_delay" toml:"debounce_delay" yaml:"debounce_delay"`
}

type ProxyConfig struct {
	Enabled   bool   `json:"enabled" toml:"enabled" yaml:"enabled"`
	AppPort   uint16 `json:"app_port" toml:"app_port" yaml:"app_port"`
	ProxyPort uint16 `json:"proxy_port" toml:"proxy_port" yaml:"proxy_port"`
}

func DefaultConfig() Config {
	return Config{
		RootDir: ".",
		GlobalExclude: ExcluderConfig{
			Dirs:  []string{"data", "dist", "node_modules", "tmp"},
			Files: []string{},
			Regex: []string{
				`^\.?(\/?|\\?)(?:\w+(\/|\\))*(\.\w+)$`, // dotfiles on windows or unix at any hierarchy
				`^.+\.sqlite$`, `^.+\.wal$`, `^.+\.shm$`,
			},
		},
		Watchers: []WatcherConfig{{
			Name:      "watcher",
			Filetypes: []string{},
			Dirs:      []string{},
			Files:     []string{},
			Exclude: ExcluderConfig{
				Dirs:  []string{},
				Files: []string{},
				Regex: []string{},
			},
			Shell: ShellConfig{
				Tasks:                  []string{},
				TaskTimeout:            uint(eavesdrop.DefaultTaskRunTimeout.Milliseconds()),
				Service:                "",
				ServiceShutdownTimeout: uint(eavesdrop.DefaultServiceShutdownTimeout.Milliseconds()),
				DebounceDelay:          uint(eavesdrop.DefaultDebounceDelay.Milliseconds()),
			},
			RunOnStart:     true,
			TriggerRefresh: false,
			RefreshDelay:   uint(eavesdrop.DefaultRefreshDelay.Milliseconds()),
		}},
		Proxy: ProxyConfig{
			Enabled:   false,
			AppPort:   eavesdrop.DefaultAppPort,
			ProxyPort: eavesdrop.DefaultProxyPort,
		},
	}
}

func GetConfig(path string) (Config, error) {
	var (
		err    error
		config Config
	)

	switch ext := filepath.Ext(path); ext {
	case ".json":
		config, err = ReadJsonConfig(path)

	case ".yaml":
		config, err = ReadYamlConfig(path)

	case ".toml":
		config, err = ReadTomlConfig(path)

	default:
		err = fmt.Errorf("please use .json, .yaml, or .toml, not %s", ext)
	}

	return config, err
}

func GenerateConfig(path, ext string) error {
	var err error
	switch ext {
	case ".json", "json", "":
		err = GenerateJsonConfig(path)

	case ".yaml", "yaml", ".yml", "yml":
		err = GenerateYamlConfig(path)

	case ".toml", "toml":
		err = GenerateTomlConfig(path)

	default:
		err = fmt.Errorf("invalid extension: %s", ext)
	}

	return err
}
