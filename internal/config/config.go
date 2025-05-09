package config

import (
	"fmt"
	"path/filepath"
)

type Config struct {
	// The location of the config file.
	cfg string

	// The project root directory.
	// Defaults to ".".
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
	// Defaults to []string{"node_modules", "tmp", "vendor"}
	IgnoreDirs []string `json:"ignore_dirs" yaml:"ignore_dirs" toml:"ignore_dirs"`

	// Ignores named files.
	IgnoreFiles []string `json:"ignore_files" yaml:"ignore_files" toml:"ignore_files"`

	// Ignores regular expression patterns.
	// Defaults to []string{`^\.{1}.*$`}
	IgnoreRegex []string `json:"ignore_regex" yaml:"ignore_regex" toml:"ignore_regex"`

	// Overrides any possible Ignore instructions and explicitly watches the file.
	Watch []string `json:"watch" yaml:"watch" toml:"watch"`

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
	return &Config{
		cfg:         configPath,
		Root:        ".",
		Tmp:         "tmp",
		Build:       "go build -o ./tmp/main ./main.go",
		Run:         "tmp/main",
		IgnoreDirs:  []string{"node_modules", "tmp", "vendor"},
		IgnoreFiles: []string{},
		IgnoreRegex: []string{`^\.{1}.*$`},
		Watch:       []string{},
		Proxy:       false,
		AppPort:     8000,
		ProxyPort:   8001,
	}
}
