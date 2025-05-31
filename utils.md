# Utilities Documentation

This document details utility packages in the project.

## Package: u/nopanic

### File: safe_run.go

Provides panic recovery utilities to safely run functions that might panic.

```go
// Defines PanicWrapperError interface for captured panics
type PanicWrapperError interface {
	Error() string
	Stack() u.CallStack
}

// Implementations of panicWrapperError store error and stack trace

// Run: Executes a function and recovers any panics as errors
func Run(fn func()) (err PanicWrapperError)

// RunE: Executes a function returning error and recovers panics as errors
func RunE(fn func() error) (err error)

// RunResult: Executes a function returning (T, error) and recovers panics
func RunResult[T any](fn func() (T, error)) (res T, err error)
```

## Package: u

### File: debug.go

Provides utilities for capturing call stacks.

```go
type CallStack struct { ... }

// Captures current call stack skipping specified frames
func CurrentCallStack(skip int) CallStack

// Formats stack as single-line string
func (s CallStack) StringSingeLine() string

// Formats stack as multi-line string
func (s CallStack) StringMultiline() string
```

### File: common.go

Provides common utilities:

```go
// Must: Panics if error exists 
func Must(err error) 

// Must2/Must3: Panic on error for multi-return functions
func Must2[T any](value T, err error) T
func Must3[T1 any, T2 any](val1 T1, val2 T2, err error) (T1, T2)

// First: Returns first argument when unsure about return types
func First[T any](t T, _ ...any) T

// Use: Safely use closable resources with automatic close
func Use[T io.Closer](resource T, block func(it T))

// JoinAsString: Joins array with separator using custom converter
func JoinAsString[T any](arr []T, converter func(val T) string, sep string) string

// CloseOptimistic: Close resource ignoring errors
func CloseOptimistic(resource io.Closer)
```

### File: reflection.go

Provides reflection utilities.

```go
// Gets interface name as string
func GetInterfaceName[T any]() string
```

### File: writer_proxy.go

Provides writer proxy to spy on written data.

```go
// ProxyWriter intercepts writes to delegate writer
type ProxyWriter struct { ... }

// Write captures chunk for spying before delegating
func (proxy ProxyWriter) Write(chunk []byte) (n int, err error)

// Creates writer that duplicates writes to spyWriter
func NewSpyWriter(writer io.Writer, spyWriter io.Writer) io.Writer
```

## Example Application

### File: main_example.go

Demonstrates go-ctx framework features through various services:

a) Basic Services (`aService`, `bService`) - Dependency injection and lifecycle hooks  
b) `timedService` - Periodic timer tasks  
c) `appLCService` - Application start/stop awareness  
d) `connAService`/`connBService` - Service messaging with `ServiceConnector`  
e) `multiInstanceService` - Multiple service instances  
f) `reflectiveSingletonServiceImpl` - Reflection-based dependency injection  
g) `panicService` - Panic recovery  
h) `loggerDemoService` - Logger injection  
i) `envInjectDemoService` - Environment variable injection  
j) `slogExample` - Structured logging with slog  
k) `contextService` - Context handling  
l) `panicExample` - Panic handling examples  

```go
// Environment setup
_ = os.Setenv("SLOG_LEVEL", "debug")

// Application creation with service packages
application := ctx.CreateContextualizedApplication(
	ctx.PackageOf(&aService{}, &bService{}, ...),
	...
)
```
