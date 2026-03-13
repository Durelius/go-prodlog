# prodlog

A lightweight Go logging package with [systemd journald](https://www.freedesktop.org/software/systemd/man/systemd.journal-fields.html) priority prefixes, optional file output with automatic daily rotation, and safe concurrent use.

## Installation

```bash
go get github.com/durelius/go-prodlog
```

## Usage

```go
prodlog.Info("server started")
prodlog.Infof("listening on port %d", 8080)

prodlog.Warning("config missing, using defaults")
prodlog.Warningf("retrying in %d seconds", 5)

prodlog.Error("failed to connect to database")
prodlog.Errorf("unexpected status code: %d", 500)

prodlog.Fatal("unrecoverable error")   // logs then calls os.Exit(1)
prodlog.Fatalf("exit code: %d", 1)
```

## Log Levels

| Function | Output | journald priority |
|----------|--------|-------------------|
| `Info` / `Infof` | stdout | `<6>` INFO |
| `Warning` / `Warningf` | stdout | `<4>` WARNING |
| `Error` / `Errorf` | stderr | `<3>` ERROR |
| `Fatal` / `Fatalf` | stderr | `<3>` ERROR + exit |

Each line is prefixed with the journald priority tag so that `systemd-journald` correctly categorises log entries when the process runs as a service.

## File Logging

Call `EnableLogFile` with a folder path to start writing logs to disk:

```go
prodlog.EnableLogFile("/var/log/myapp")
```

Log files are named `log_YYMMDD` and rotate automatically at midnight — when the date changes, the current file is closed and a new one is opened. The folder must already exist.

## Disabling stdout

Useful in production when output is captured by journald and you only want file logs:

```go
prodlog.DisableStdout()
```

`Fatal` and `Fatalf` always write to stderr regardless of this setting.

## Concurrency

All public functions are safe to call from multiple goroutines. File writes are serialised with a mutex and the `disableStdout`/`logFolderPath` state is managed with `sync/atomic`.

## Running tests

```bash
go test -race ./...
```
