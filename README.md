<p align="center">
    <a href="https://goreportcard.com/report/github.com/dimmerz92/eavesdrop">
        <img src="https://goreportcard.com/badge/github.com/dimmerz92/eavesdrop" alt="Go Report Card" />
    </a>
    <a href="https://pkg.go.dev/github.com/dimmerz92/eavesdrop">
        <img src="https://pkg.go.dev/badge/github.com/dimmerz92/eavesdrop" alt="Go Reference" />
    </a>
    <a href="https://opensource.org/licenses/MIT">
        <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="MIT License" />
    </a>
</p>
<p align="center">
    <img src="/assets/eavesdrop.png" alt="eavesdrop logo"/>
</p>

<h2>Introduction</h2>

Eavesdrop is yet another command line tool for live reloading. It was made with Go projects in mind, but could relatively easily be adapted for use with other projects.


Eavesdrop is small, lightweight, fast, and comes with a proxy option for browser reloading.

<h3>Features</h3>

- Live reloading
- Supports json, toml, or yaml config files
- Optional proxy server for browser refreshing (very helpful for web development)

<h2>Motivation</h2>

I have been using Go to build simple webapps for roughly a year now (as of 2025-05). Throughout my journey, I have discovered a range of tools that have really helped my workflow, especially a project named [Air](https://github.com/air-verse/air), a mature live reloader for Go projects.


The Air project inspired me to write my own version, and so you could say this is a copy of it. Rather than forking it, I wanted to have a project that was reasonably challenging but not impossible to write on my own. For that reason, I chose to start from scratch rather than forking it.

<p align="center">
    <img src="/assets/eavesdrop_running.png" alt="easedrop running in terminal"/>
</p>

<h2>Quick Start</h2>

1. Download Eavesdrop with Go
```sh
go install github.com/dimmerz92/eavesdrop@latest
```

2. Run Eavesdrop
```sh
eavesdrop
```

This runs Eavesdrop without a config, providing quick and easy live reloading on file changes for `.go`, `.html`, `.templ`, `.tmpl`, and `.tpl` files.

<h2>Config</h2>

Eavesdrop supports configs written in either `.json`, `.toml`, or `.yaml`, so pick which ever works for you.

Configs can be generated using:
```sh
eavesdrop init
```

To change the output directory you can use the `--out` or `-o` flag followed by the directory you want it to go to. If the output directory is not specified, the config will be generated to the directory the command was run from.

Similarly, to change the config file type, you can use the `--ext` or `-e` flag followed by the extension you want. `json`, `.json`, `toml`, `.toml`, `yaml`, or `.yaml` will work. If the extension is not specified, the default config will generate as a json file.

Example configs can be found in the [config examples](/examples/configs) folder.

If your config is located in the same directory that Eavesdrop is run from, you don't have to do anything, you can simply run standalone:
```sh
eavesdrop
```

If your config is located in a different directory, you can use the `--config` or `-c` flag to specify the directory it is in.

<h2>Contributing</h2>

If you want to contribute, please check out the [contribution guidelines](/CONTRIBUTING.md).

<h2>License</h2>

Provided under the MIT License - [License](/LICENSE)
