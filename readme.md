# go-ctx: Contextualized Services for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/sedmess/go-ctx.svg)](https://pkg.go.dev/github.com/sedmess/go-ctx)

The `go-ctx` library provides a framework for building modular applications with dependency injection, lifecycle management, and environment configuration. v0.13.0 requires Go 1.27 or later and adds no non-standard-library dependencies.

See [Architecture](docs/architecture.md) for package boundaries, dependency injection,
configuration precedence, and the application lifecycle.

## DeepWiki Documentation
[![DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/sedmess/go-ctx)

## Features

- **Service Lifecycle**: Services can implement `Initializable`, `StartAware`, `StopAware`, and `Disposable` interfaces
- **Dependency Injection**: By name, by type, or via reflection tags
- **Environment Configuration**: Inject environment variables with support for default values and complex types
- **Logging**: Integrated with `slog` and loggers injection
- **Health Checks**: Services can report health status
- **Timers**: Run periodic tasks with `TimerTask`
- **Inter-Service Communication**: Use `ServiceConnector` for message passing
- **Reflection-Based Wiring**: Automatic dependency injection using struct tags
- **Panic Recovery**: Utilities to handle panics gracefully

## Installation

Use Go modules:
```bash
go get github.com/sedmess/go-ctx
```

## Quick Start

Define a service:
```go
package main

import (
	"context"
	"github.com/sedmess/go-ctx/ctx"
	"github.com/sedmess/go-ctx/ctx/logger"
)

type HelloService struct {
	ctx.AppContext `ctx:"CTX"`
	l              logger.Logger `ctx:""`
}

func (h *HelloService) AfterStart() {
	h.l.Info("Hello, World!")
}

func main() {
	app := ctx.CreateContextualizedApplication(
		ctx.PackageOf(&HelloService{}),
	)
	app.Stop().Join()
}
```

`Stop` is immediate, idempotent, and safe from concurrent callers. `Join` waits until stop
callbacks, cancellation, disposal, signal cleanup, and global context cleanup finish.

## Detailed Examples

### Service with Dependencies
```go
type aService struct {
	paramA int
}

func (instance *aService) Init(_ ctx.ServiceProvider) {
	instance.paramA = ctx.GetEnv(paramAName).AsIntDefault(5)
	logger.Info(instance.Name(), "initialized")
}

func (instance *aService) Name() string {
	return "a_service"
}
```

### Lifecycle Management
```go
type timedService struct {
	ctx.TimerTask
	l logger.Logger `ctx:""`
}

func (instance *timedService) AfterStart() {
	instance.l.Info("Starting timer")
	instance.StartTimer(2*time.Second, func() {
		logger.Warn("timer", "onTimer!")
	})
}

func (instance *timedService) BeforeStop() {
	instance.StopTimer()
}
```

### Environment Configuration
```go
type envInjectDemoService struct {
	Duration time.Duration `env:"DURATION"`
}
```

### Reflection-Based Injection
```go
type reflectiveSingletonServiceImpl struct {
	L logger.Logger `ctx:"singleton"`
	A *aService     `ctx:"a_service"`
}
```

### Logging
```go
_ = os.Setenv("SLOG_LEVEL", "debug")
_ = os.Setenv("SLOG_ADD_SOURCE", "true")
ctx.SetSlogWriter(os.Stdout, os.Stderr)
```

### Health Checks
```go
func (instance *aService) Health() health.ServiceHealth {
	return health.Status(health.Up)
}
```

## Advanced Topics

### v0.13.0 Go 1.27 Typed Streams

v0.13.0 raises the minimum supported toolchain to Go 1.27. Stream transformations now use
generic methods, so result types are preserved without `any` wrappers:

```go
strings := channels.SliceToChannel([]int{1, 2, 3}).Map(strconv.Itoa)
nested := strings.FlatMap(func(value string) channels.StreamingChan[int] {
	return channels.SingleElemChannel(len(value))
})
```

Package-level `FlatMap` is now the canonical spelling. The shipped `FlapMap` helper remains
as a deprecated compatibility alias. Context-free method values/expressions may need an
explicit result type argument, and generic methods no longer satisfy interfaces containing
the former non-generic method signatures. See the
[v0.13.0 migration guide](docs/migration-v0.13.0.md).

### v0.12.0 Runtime Contracts

- `BeforeStop` is deterministic: consumers stop before dependencies, with service-name
  ordering for unrelated services. Lifecycle callbacks may query services and diagnostics.
- `GetTypedService[T]` returns `(zero, false)` for ordinary absence; always check the boolean.
- Duplicate service names are rejected, health keeps the worst normalized severity, and
  statistics are deep consumer-owned snapshots.
- Configuration resolves exact process spelling, uppercase process fallback, then uppercase
  canonical arguments/files/defaults without changing source precedence.
- Timer and connector shutdown joins the active generation; application `Join` includes
  signal unregistration and all owned cleanup.

`StreamingChan.ToChan` now returns receive-only outputs:

```go
values, failures := stream.ToChan(16)
for value := range values {
	_ = value
}
if err := <-failures; err != nil {
	// handle the source error
}
```

This receive-direction correction is intentionally source-incompatible for callers that
copied the old exact `chan<-` signature. See the [v0.12.0 migration guide](docs/migration-v0.12.0.md).

### Custom Service Tags
```go
type newTags struct {
	l1 logger.Logger `ctx:""`
	l2 logger.Logger `ctx:"named_logger_2"`
}
```

### Structured Logging with slog
```go
type slogExample struct {
	l1 *slog.Logger  `ctx:""`
	l2 *slog.Logger  `ctx:"named_slog1"`
}
```

## Running the Examples
```bash
go run ./examples/task
go run ./examples/application
```

## License
MIT
