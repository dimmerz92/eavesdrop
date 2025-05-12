package config

import (
	"fmt"
	"os"
)

type Config struct {
	// The location of the config file.
	cfg string

	// The project root directory.
	// Defaults to the current working directory.
	Root string `json:"root" yaml:"root" toml:"root"`

	// The temp directory that the binary is compiled to.
	// Defaults to "./tmp".
	Tmp string `json:"tmp" yaml:"tmp" toml:"tmp"`

	// The command to build the project.
	// Defaults to "go build -o ./tmp/main ./main.go".
	Build string `json:"build" yaml:"build" toml:"build"`

	// The command to run the project after successful build.
	// Defaults to "./tmp/main"
	Run string `json:"run" yaml:"run" toml:"run"`

	// Ignores named directories.
	// Defaults to []string{"assets", "data", "node_modules", "testdata", "tmp", "vendor"}
	IgnoreDirs []string `json:"ignore_dirs" yaml:"ignore_dirs" toml:"ignore_dirs"`

	// Ignores named files.
	IgnoreFiles []string `json:"ignore_files" yaml:"ignore_files" toml:"ignore_files"`

	// Ignores regular expression patterns.
	// Defaults to []string{`^\.{1}.*$`, `^.*_templ\.go$`, `^.*_test\.go$`}
	IgnoreRegex []string `json:"ignore_regex" yaml:"ignore_regex" toml:"ignore_regex"`

	// Overrides any possible Ignore instructions and explicitly watches the file.
	WatchFiles []string `json:"watch_files" yaml:"watch_files" toml:"watch_files"`

	// Only watches files with the specified extensions
	// Defaults to []string{".go", ".html", ".templ", ".tmpl", ".tpl"}
	WatchExts []string `json:"watch_exts" yaml:"watch_exts" toml:"watch_exts"`

	// Sets up a proxy server for browser reloading.
	// Defaults to false (off)
	Proxy bool `json:"proxy" yaml:"proxy" toml:"proxy"`

	// The port for your project's server.
	// Defaults to 8000.
	AppPort int `json:"app_port" yaml:"app_port" toml:"app_port"`

	// The port that will serve your project's server with browser reloading.
	// Defaults to 8001.
	ProxyPort int `json:"proxy_port" yaml:"proxy_port" toml:"proxy_port"`
}

// DefaultConfig returns the default config.
func DefaultConfig(configPath string) *Config {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return &Config{
		cfg:         configPath,
		Root:        cwd,
		Tmp:         "tmp",
		Build:       "go build -o ./tmp/main ./main.go",
		Run:         "tmp/main",
		IgnoreDirs:  []string{"assets", "data", "node_modules", "testdata", "tmp", "vendor"},
		IgnoreFiles: []string{},
		IgnoreRegex: []string{`^\.{1}.*$`, `^.*_templ\.go$`, `^.*_test\.go$`},
		WatchFiles:  []string{},
		WatchExts:   []string{".go", ".html", ".templ", ".tmpl", ".tpl"},
		Proxy:       false,
		AppPort:     8000,
		ProxyPort:   8001,
	}
}

// Validate checks that required fields are set, otherwise returns an error.
func (c *Config) Validate() error {
	if c.Root == "" {
		return fmt.Errorf("a project root directory is required")
	}
	if c.Build == "" {
		return fmt.Errorf("a build command is required")
	}
	if c.Run == "" {
		return fmt.Errorf("a run command is required")
	}
	if c.Proxy && (c.AppPort == 0 || c.ProxyPort == 0) {
		return fmt.Errorf("an application and proxy port is required if proxy is true")
	}
	return nil
}
