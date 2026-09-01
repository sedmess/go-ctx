# Phase 1 Contract Model: Upgrade to Go 1.27

This feature persists no data. Its model is a set of compatibility and stream-transformation
contracts observable to maintainers and library consumers.

## 1. Toolchain Baseline

### Fields

- **minimum language version**: Go 1.27.
- **minimum validation line**: Latest available Go 1.27.x.
- **forward validation line**: Current stable Go.
- **target release**: v0.13.0.
- **module path**: `github.com/sedmess/go-ctx`, unchanged.
- **dependency set**: Standard library only.

### Validation rules

- Module metadata, CI, constitution, current architecture, contributor, README, templates,
  and v0.13.0 migration guidance agree on Go 1.27.
- Historical v0.12.0 and feature-001 records remain on Go 1.26.
- Both CI toolchain lines execute compile, build, test, vet, and race gates.

## 2. Typed Stream Transformation

### Fields

- **source element type `T`**: The compile-time type carried by the input stream.
- **result element type `Q`**: The method-specific type returned by the mapper or nested
  stream.
- **mapper**: `func(T) Q` for map or `func(T) StreamingChan[Q]` for flat-map.
- **output stream**: `StreamingChan[Q]`.
- **producer owner**: The existing transformation worker created by `CreateChannel`.
- **terminal result**: Normal close or the first observed source/nested error.

### Relationships and state transitions

| Operation | Input | Transition | Output |
|-----------|-------|------------|--------|
| Map | One `T` | Apply mapper once, send one `Q` | Ordered `Q` stream |
| FlatMap | One `T` | Open mapped stream, drain it sequentially | Ordered zero-or-more `Q` values |
| Source error | Error element | Stop transform and return error | Error then close |
| Nested error | Error in mapped stream | Stop flat-map and return error | Error then close |
| Empty source | Closed stream | Perform no mapping | Empty closed stream |

### Invariants

- Direct calls infer `Q`; explicit method instantiation remains available.
- No `any` conversion or assertion is required when `Q` is concrete.
- Outer and nested value order, error visibility, and backpressure match the prior helpers.
- No additional goroutine, buffer, cancellation path, or dependency is introduced.

## 3. FlatMap Compatibility Entry Point

### Fields

- **canonical name**: `FlatMap`.
- **legacy name**: `FlapMap`.
- **legacy status**: Deprecated, still callable.
- **implementation**: One canonical behavior path; the legacy function forwards.

### Validation rules

- Both names infer the same `P` and `Q` and produce equivalent ordered values/errors.
- Documentation uses `FlatMap` for new code and identifies `FlapMap` as compatibility-only.

## 4. Generic Method Reference

### States

- **Direct invocation**: Mapper arguments infer `Q`.
- **Explicit method value/expression**: Consumer supplies `[Q]`.
- **Contextually typed value**: Assignment target may infer `Q`.
- **Context-free method reference**: Invalid until explicitly instantiated.
- **Old interface contract**: No longer implemented because generic methods do not satisfy
  non-generic interface methods.

### Validation rules

- Regression tests compile direct inferred calls plus explicit method values and expressions.
- Migration guidance gives an adapter/explicit-instantiation path for broken forms.

## 5. Modernization Audit

### Fields

- **analyzers**: `atomictypes`, `embedlit`, `slicesbackward`, `unsafefuncs`.
- **mode**: Non-mutating diff.
- **result**: Proposed patch or empty output.
- **decision**: Implemented with validation or classified inapplicable.

### Validation rules

- Run once against the initial code and again after `go 1.27` is declared.
- Empty output creates no source change.
- Suggestions from other modernizers are outside this feature unless separately justified.
