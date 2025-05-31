# go-ctx: Contextualized Services for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/sedmess/go-ctx.svg)](https://pkg.go.dev/github.com/sedmess/go-ctx)

The `go-ctx` library provides a framework for building modular applications with dependency injection, lifecycle management, and environment configuration.

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
	app.Join()
}
```

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

## Running the Example
```bash
go run main_example.go
```

## License
MIT
