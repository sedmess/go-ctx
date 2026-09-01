# Migration Contract: v0.12.x to v0.13.0

v0.13.0 keeps the module path and standard-library-only dependency policy. It intentionally
raises the minimum supported toolchain from Go 1.26 to Go 1.27 and replaces two type-erasing
stream methods with Go 1.27 generic methods.

## Required toolchain migration

Upgrade development, CI, release, and deployment builders to Go 1.27 or later before adopting
v0.13.0. Older Go versions cannot parse or compile the generic method declarations.

## Direct stream calls

Calls can return concrete stream types without `any` wrappers:

```go
strings := channels.SliceToChannel([]int{1, 2}).Map(strconv.Itoa)
```

Existing mappers explicitly returning `any` continue to infer `StreamingChan[any]`.

## Method values and expressions

Add explicit result type arguments where no assignment context can infer them:

```go
mapValue := stream.Map[string]
mapExpression := channels.StreamingChan[int].Map[string]
```

Contextually typed assignment may infer the same result type:

```go
var mapValue func(func(int) string) channels.StreamingChan[string] = stream.Map
```

## Interfaces containing Map or FlatMap

A generic method does not satisfy a non-generic interface method. Consumers that declared
the former exact methods in an interface must use the package-level `Map`/`FlatMap` helpers,
introduce an adapter with a concrete result type, or redesign the interface around the
operation they need.

## FlatMap spelling

Use `FlatMap` for new package-level calls. The shipped `FlapMap` spelling remains available
as a deprecated alias in v0.13.0, so migration is recommended but not required.

## Unchanged behavior

Stream ordering, nested ordering, completion, error propagation, and backpressure are
unchanged. This release does not add parallel mapping, buffering, cancellation, or new
dependencies, and it does not change application lifecycle, configuration, wiring, tags,
logging, health, or statistics.

## Release validation

Run the feature [quickstart](../quickstart.md), including Go 1.27.x and stable build/test/vet
gates, Linux race validation, typed stream regressions, targeted Go 1.27 modernizers, and both
examples before tagging v0.13.0.
