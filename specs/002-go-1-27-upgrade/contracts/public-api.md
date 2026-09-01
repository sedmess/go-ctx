# Public Library Contract: Go 1.27 Typed Streams

**Target release**: v0.13.0
**Module path**: `github.com/sedmess/go-ctx` (unchanged)
**Minimum supported toolchain**: Go 1.27.x

## Typed method transforms

```go
func (ch StreamingChan[T]) Map[Q any](mapper func(T) Q) StreamingChan[Q]
func (ch StreamingChan[T]) FlatMap[Q any](mapper func(T) StreamingChan[Q]) StreamingChan[Q]
```

- Ordinary calls infer `Q` from the mapper.
- Explicit `[Q]` method instantiation is supported.
- Concrete results do not pass through `any`.
- A mapper explicitly returning `any` produces `StreamingChan[any]`.
- Map preserves source order. Flat-map preserves source order and each nested stream's order.
- Empty streams remain empty, and source/nested errors remain observable.
- The methods add no goroutine or buffering beyond the existing package helpers.

## Package transforms

The existing helper remains unchanged:

```go
func Map[P, Q any](ch StreamingChan[P], mapper func(P) Q) StreamingChan[Q]
```

The canonical flat-map helper is:

```go
func FlatMap[P, Q any](ch StreamingChan[P], mapper func(P) StreamingChan[Q]) StreamingChan[Q]
```

The legacy typo remains source-compatible:

```go
// Deprecated: use FlatMap.
func FlapMap[P, Q any](ch StreamingChan[P], mapper func(P) StreamingChan[Q]) StreamingChan[Q]
```

`FlapMap` forwards to `FlatMap`; both have identical type inference and runtime behavior.

## Method references and interfaces

These explicit forms are supported:

```go
mapValue := stream.Map[string]
mapExpression := StreamingChan[int].Map[string]
```

A known target function type may allow inference without `[string]`. A context-free
`mapValue := stream.Map` or uninstantiated method expression cannot infer `Q` and must be
updated.

Generic methods cannot implement non-generic interface methods. A consumer interface that
contains the former exact `Map(func(T) any) StreamingChan[any]` or corresponding `FlatMap`
method requires an adapter, a package-level helper, or interface redesign.

## Unchanged contracts

- `StreamingChan`, `ChanElem`, `ToChan`, creation helpers, collection, and iteration retain
  their signatures and documented behavior.
- Service naming, DI tags, configuration, lifecycle, logging, health, statistics, timers,
  connectors, module path, and dependency policy are unchanged.
