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

<p>Eavesdrop is a command line tool for automatic project recompilation and browser reloading. It was made with Go projects in mind, however, it is completely project agnostic and can be adapted for use with any project.</p>

<p>Eavesdrop is small, lightweight, fast, and comes with a proxy option for browser reloading.</p>

<h3>Features</h3>

- Live reloading
- Supports json, toml, or yaml config files
- Multiple watcher profiles to isolate different tasks and services
- Optional proxy server for browser refreshing (very helpful for web development)

<p align="center">
    <img src="/assets/eavesdrop_running.png" alt="eavesdrop running in terminal"/>
</p>

<h2>Quick Start</h2>

1. Download Eavesdrop with Go

```bash
go install github.com/dimmerz92/eavesdrop/cmd/eavesdrop@latest
```

2. Create an eavesdrop config

```bash
eavesdrop init
```

**Note:**

- Optionally add `-ext` to specify the desired config type (.json, .toml, .yaml)
- Optionally add `-out` to specify the output directory for the config

3. Run Eavesdrop

```bash
eavesdrop
```

**Note:**

Eavesdrop will automatically look in the current working directory for a `.eavesdrop.json` config. If you specified a different format or directory, you will need to use the `-config` flag to specify the location.

<h2>Config</h2>

Eavesdrop supports configs written in either `.json`, `.toml`, or `.yaml`, so pick which ever works for you.

Configs can be generated using:

```bash
eavesdrop init
```

To change the output directory you can use the `-out` flag followed by the directory you want it to go to. If the output directory is not specified, the config will be generated to the directory the command was run from.

Similarly, to change the config file type, you can use the `-ext` flag followed by the extension you want. `.json`, `.toml`, or `.yaml` will work. If the extension is not specified, the default config will generate as a json file.

Example configs can be found in the [config examples](/examples) folder.

If your config is located in the same directory that Eavesdrop is run from, you don't have to do anything, you can simply run standalone:

```bash
eavesdrop
```

If your config is located in a different directory, you can use the `-config` flag to specify the directory it is in.

<h2>Contributing</h2>

If you want to contribute, please check out the [contribution guidelines](/CONTRIBUTING.md).

<h2>License</h2>

Provided under the MIT License - [License](/LICENSE)
