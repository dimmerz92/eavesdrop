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

## Introduction

Eavesdrop is a lightweight, fast file watcher built for automatic project recompilation and browser reloading.

Eavesdrop can be used in two ways:

- **As a CLI tool** - driven by a config file (JSON, TOML, or YAML), it watches your project and runs shell tasks and services whenever files change.
- **As a Go library** - import the `ev` package and supply your own callback to react to file system events however you like.

### Features

- Live reloading with configurable debounce
- Supports JSON, TOML, and YAML config files
- Multiple named watcher profiles to isolate different tasks
- Optional reverse proxy with Server-Sent Events for automatic browser refresh
- Modular Go library API - no shell required

<p align="center">
    <img src="/assets/eavesdrop_running.png" alt="eavesdrop running in terminal"/>
</p>

---

## CLI Usage

### Installation

```bash
go install github.com/dimmerz92/eavesdrop/cmd/eavesdrop@latest
```

### Create a config

```bash
eavesdrop init
```

| Flag   | Description                                 | Default |
|--------|---------------------------------------------|---------|
| `-ext` | Config format: `.json`, `.toml`, or `.yaml` | `.json` |
| `-out` | Output directory for the generated config   | `.`     |

### Run

```bash
eavesdrop
```

By default eavesdrop probes for `eavesdrop.json`, `eavesdrop.toml`, and `eavesdrop.yaml` in the current directory (in that order). Use `-config` to point at a specific file:

```bash
eavesdrop -config path/to/eavesdrop.yaml
```

### Config reference

Example configs are in the [examples](/examples) folder. Full field reference below.

#### Top-level fields

| Field            | Type   | Description                                              |
|------------------|--------|----------------------------------------------------------|
| `root_dir`       | string | Root directory to watch. Defaults to `.`.                |
| `tmp`            | bool   | Create a `tmp/` directory at startup.                    |
| `cleanup_tmp`    | bool   | Delete `tmp/` on shutdown.                               |
| `global_exclude` | object | Exclude rules applied before any watcher sees events.    |
| `watchers`       | array  | One or more named watcher profiles.                      |
| `proxy`          | object | Optional reverse proxy for browser live-reload.          |

#### `global_exclude` / watcher `exclude` fields

| Field   | Type     | Description                                                                                                       |
|---------|----------|-------------------------------------------------------------------------------------------------------------------|
| `ops`   | string[] | File operations to ignore: `"CHMOD"`, `"CREATE"`, `"REMOVE"`, `"RENAME"`, `"WRITE"`.                            |
| `dirs`  | string[] | Directory paths relative to `root_dir` to skip. `"tmp"` skips only `./tmp`, not `./src/tmp`. Use regex for name-based matching at any depth. |
| `files` | string[] | File paths relative to `root_dir` to skip. `"go.sum"` skips only `./go.sum`. Use regex for name-based matching at any depth.               |
| `regex` | string[] | Regular expressions matched against each file's full path — the permissive option for matching at any depth.      |

#### Watcher fields

| Field             | Type     | Description                                                                              |
|-------------------|----------|------------------------------------------------------------------------------------------|
| `name`            | string   | Unique label for this watcher (shown in log output).                                     |
| `filetypes`       | string[] | File extensions to react to, e.g. `[".go", ".html"]`.                                   |
| `dirs`            | string[] | Directory paths to watch, relative to `root_dir`. `"cmd"` watches `./cmd` only, not `./src/cmd`. |
| `files`           | string[] | File paths to watch, relative to `root_dir`. `"go.sum"` watches `./go.sum` only.                |
| `exclude`         | object   | Per-watcher exclude rules, layered on top of `global_exclude`.                           |
| `run_on_start`    | bool     | Run tasks/service once immediately when eavesdrop starts.                                 |
| `trigger_refresh` | bool     | Signal the proxy to reload the browser after each onChange.                               |
| `refresh_delay`   | uint     | Milliseconds to wait after onChange before triggering a browser refresh. Default: `100`. |
| `shell`           | object   | Shell execution settings.                                                                 |

#### Shell fields

| Field                      | Type     | Description                                                                         |
|----------------------------|----------|-------------------------------------------------------------------------------------|
| `tasks`                    | string[] | Commands run sequentially before the service starts.                                |
| `task_timeout`             | uint     | Milliseconds before a task is forcibly killed. Default: `2000`.                     |
| `service`                  | string   | Long-running command started after tasks complete (e.g. your compiled binary).      |
| `service_shutdown_timeout` | uint     | Milliseconds to wait for the service to exit before force-killing. Default: `5000`. |
| `debounce_delay`           | uint     | Quiet period in milliseconds before reacting to file changes. Default: `100`.       |

#### Proxy fields

| Field        | Type   | Description                                          |
|--------------|--------|------------------------------------------------------|
| `enabled`    | bool   | Enable the reverse proxy.                            |
| `app_port`   | uint16 | Port your application listens on. Default: `8000`.  |
| `proxy_port` | uint16 | Port the proxy server listens on. Default: `8001`.  |

When the proxy is enabled, browse to `http://localhost:<proxy_port>` instead of your app's port directly. The proxy automatically refreshes the browser whenever eavesdrop detects a change.

---

## Library Usage

The import path ends in `eavesdrop` but the package identifier is `ev`:

```go
import ev "github.com/dimmerz92/eavesdrop"
```

### Basic example

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"

    ev "github.com/dimmerz92/eavesdrop"
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()

    // Create an emitter rooted at the current directory with a global excluder.
    emitter := ev.NewEmitter(".").
        WithExcluder(ev.NewExcluder(".").
            WithDirs("vendor", "node_modules", ".git", "tmp"),
        )

    // Create a watcher that reacts to .go file changes.
    watcher := ev.NewWatcher(ctx, "go-watcher", ".").
        WithFiletypes(".go").
        WithDebounceDelay(200).
        WithOnChange(func(e ev.Event) { fmt.Printf("changed: %s\n", e.Path()) })

    emitter.Subscribe(watcher)
    emitter.Start(ctx)

    watcher.Trigger() // run once on startup before waiting for changes

    <-ctx.Done()
}
```

### API overview

**`NewEmitter(root string) *EventEmitter`** — creates an emitter rooted at `root`.

| Method | Description |
|--------|-------------|
| `.WithExcluder(e *Excluder)` | Attach a global excluder; matching paths are skipped before any watcher sees them. |
| `.Subscribe(s Subscriber)` | Register a `Subscriber` to receive events. Must be called before `Start`. |
| `.Start(ctx context.Context)` | Begin watching and dispatching events. Stops when `ctx` is cancelled. |

**`NewWatcher(ctx context.Context, name, root string) *Watcher`** — creates a named watcher rooted at `root`. `name` must be unique across the process.

| Method | Description |
|--------|-------------|
| `.WithFiletypes(ext ...string)` | React to files with these extensions, e.g. `".go"`, `".html"`. |
| `.WithDirs(dir ...string)` | React to files under these directories (relative to `root`). |
| `.WithFiles(file ...string)` | React to these specific files (relative to `root`). |
| `.WithOnChange(fn func(Event))` | Handler called on each matching event after debounce. |
| `.WithDebounceDelay(ms uint)` | Quiet period before firing onChange. Default: `100` ms. |
| `.WithExcluder(e *Excluder)` | Per-watcher excluder, applied after the emitter's global excluder. |
| `.WithProxy(p Proxy, delayMs uint)` | Trigger `p.RefreshBrowser()` after each onChange with an optional delay. |
| `.Trigger()` | Manually invoke onChange immediately, bypassing filters and debounce. |

**`NewExcluder(root string) *Excluder`** — creates an excluder rooted at `root`.

| Method | Description |
|--------|-------------|
| `.WithOps(op ...Op)` | Exclude events by operation: `ev.CHMOD`, `ev.CREATE`, `ev.REMOVE`, `ev.RENAME`, `ev.WRITE`. |
| `.WithDirs(dir ...string)` | Exclude exact directory paths (relative to `root`) and their contents. |
| `.WithFiles(file ...string)` | Exclude exact file paths (relative to `root`). |
| `.WithRegex(pattern ...string)` | Exclude files whose full path matches any of these regular expressions. |

### Multiple watchers

Each watcher subscribes independently and runs concurrently:

```go
cssWatcher := ev.NewWatcher(ctx, "css", ".").
    WithFiletypes(".css").
    WithOnChange(func(e ev.Event) { buildCSS() })

goWatcher := ev.NewWatcher(ctx, "go", ".").
    WithFiletypes(".go").
    WithOnChange(func(e ev.Event) { buildBinary() })

emitter.Subscribe(cssWatcher)
emitter.Subscribe(goWatcher)
emitter.Start(ctx)
```

### Browser-refresh proxy

`WithProxy` accepts any value that satisfies the `ev.Proxy` interface:

```go
type Proxy interface {
    RefreshBrowser()
}
```

Pass your implementation to a watcher and eavesdrop will call `RefreshBrowser()` after each onChange:

```go
watcher := ev.NewWatcher(ctx, "web", ".").
    WithFiletypes(".html", ".css", ".js").
    WithProxy(myProxy, 150). // 150 ms delay before refresh
    WithOnChange(func(e ev.Event) { buildAssets() })
```

Browse to your proxy's port — it injects the SSE script and reloads the browser on each change.

---

## Real-world example

`WithOnChange` is a plain Go function — call whatever you need. This example watches `.go` and `.html` files with separate watchers running concurrently:

```go
package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"

    ev "github.com/dimmerz92/eavesdrop"
)

func main() {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
    defer stop()

    emitter := ev.NewEmitter(".").
        WithExcluder(ev.NewExcluder(".").
            WithDirs("vendor", "node_modules", ".git"),
        )

    goWatcher := ev.NewWatcher(ctx, "go watcher", ".").
        WithFiletypes(".go").
        WithOnChange(func(e ev.Event) {
            fmt.Printf("go file changed: %s\n", e.Path())
        })

    htmlWatcher := ev.NewWatcher(ctx, "html watcher", ".").
        WithFiletypes(".html").
		WithExcluder(ev.NewExcluder(".").
			WithOps(ev.CHMOD).
			WithDirs("tests").
			WithRegex("_test\\.go"),
		).
        WithOnChange(func(e ev.Event) {
            fmt.Printf("template changed: %s\n", e.Path())
        })

    emitter.Subscribe(goWatcher)
    emitter.Subscribe(htmlWatcher)
    emitter.Start(ctx)

    <-ctx.Done()
}
```

---

## Shell helper

For running shell commands or managing a subprocess, eavesdrop exposes `ev.Shell`. Create one with `ev.NewShell(ctx, taskTimeoutMs, serviceTimeoutMs)`:

| Method | Description |
|--------|-------------|
| `ExecAndWait(task string) error` | Run a command and block until it exits or the task timeout elapses. |
| `ExecAndReturn(service string) error` | Start a long-running process in the background and return immediately. |
| `Stop() error` | Send SIGTERM to the running service; force-kill after the service timeout. |

---

## Contributing

See [CONTRIBUTING.md](/CONTRIBUTING.md).

## License

Provided under the MIT License — [LICENSE](/LICENSE).
