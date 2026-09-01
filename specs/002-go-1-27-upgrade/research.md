# Phase 0 Research: Upgrade to Go 1.27

## Decision 1: Adopt stable Go 1.27 as the minimum

**Decision**: Set the language/toolchain baseline to Go 1.27 and validate with Go 1.27.x
plus the current stable toolchain.

**Rationale**: Go 1.27.0 was released on 2026-08-19. The installed toolchain is Go 1.27.0,
so the minimum can be exercised locally rather than declared only in metadata. The official
[release history](https://go.dev/doc/devel/release) and
[release notes](https://go.dev/doc/go1.27) are the sources of truth.

**Alternatives considered**:

- Keep `go 1.26` plus a `toolchain` directive: rejected because it does not enable or declare
  the requested Go 1.27 language minimum.
- Wait for a later patch: rejected because 1.27.0 is stable and locally available.

## Decision 2: Treat the upgrade as v0.13.0 and constitution 1.2.0

**Decision**: Publish the compatibility change as v0.13.0 and amend constitution 1.1.0 to
1.2.0 with synchronized module, CI, templates, architecture, contributor, and migration
statements.

**Rationale**: The constitution explicitly allows a future baseline increase only after
approval and documentation. The user supplied that approval. A pre-v1 minor release clearly
signals that older toolchains and some copied method contracts no longer compile.

**Alternatives considered**:

- v0.12.1: rejected because a patch release understates the compatibility impact.
- Temporary constitution exception: rejected because the requested baseline is permanent.

## Decision 3: Replace type-erasing stream methods with generic methods

**Decision**: Use:

```go
func (ch StreamingChan[T]) Map[Q any](mapper func(T) Q) StreamingChan[Q]
func (ch StreamingChan[T]) FlatMap[Q any](mapper func(T) StreamingChan[Q]) StreamingChan[Q]
```

Both delegate to package-level helpers.

**Rationale**: Go 1.27 permits a method to declare its own type parameters. Mapper arguments
provide enough information to infer `Q` during ordinary calls, eliminating the current
`any` workaround. This matches the official
[generic methods guidance](https://go.dev/blog/generic-methods) and
[language specification](https://go.dev/ref/spec#Method_declarations).

**Alternatives considered**:

- Keep only package functions: rejected because consumers still lose ergonomic typed method
  chaining.
- Add `MapTo`/`FlatMapTo`: rejected because duplicate naming preserves an obsolete workaround
  and makes the new primary API less clear.
- Redesign `it.It`: rejected because interfaces cannot declare generic methods and converting
  that interface to a concrete type creates unrelated compatibility risk.

## Decision 4: Document the complete generic-method break

**Decision**: Migration guidance covers direct calls, explicit and contextually inferred
method values/expressions, and old interface conformance.

**Rationale**: `stream.Map(mapper)` infers `Q`, and a mapper returning `any` still infers
`Q=any`. A context-free `f := stream.Map` or method expression has no result-type inference
context and must be instantiated, such as `stream.Map[string]`. Generic methods also cannot
implement non-generic interface methods. Go has no overloads, so the old and new methods
cannot coexist under one name. See the official
[instantiation](https://go.dev/ref/spec#Instantiations) and
[type inference](https://go.dev/ref/spec#Type_inference) rules.

**Alternatives considered**:

- Describe the change as fully source-compatible: rejected because method values,
  expressions, and interface assertions are real external contracts.

## Decision 5: Canonicalize FlatMap without removing FlapMap

**Decision**: Add package-level `FlatMap[P,Q]` as the canonical implementation. Retain
`FlapMap[P,Q]` with `// Deprecated: use FlatMap.` and forward it to the canonical helper.

**Rationale**: `FlapMap` has shipped since v0.11.6. An alias removes no capability, avoids
duplicated logic, and supplies the expected spelling without another unnecessary break.

**Alternatives considered**:

- Remove `FlapMap`: rejected because it breaks consumers for no functional benefit.
- Keep only the typo: rejected because new documentation and methods should use the correct
  domain term.

## Decision 6: Preserve transform concurrency and failure semantics

**Decision**: Keep exactly one transform producer goroutine and sequential source/nested
consumption. Do not add buffering, parallel mapping, cancellation, or drain behavior.

**Rationale**: Thin generic delegates change types, not lifecycle. Current helpers preserve
outer and inner order and use existing channel backpressure and error propagation. Fixing
early-error upstream blocking would require a separate cancellation/ownership contract and
is outside this upgrade.

**Alternatives considered**:

- Parallel mapping or buffering: rejected because it changes ordering/backpressure.
- Opportunistic upstream cancellation: rejected because the public API has no cancellation
  contract and consumers need separate design and migration work.

## Decision 7: Limit automated modernization to Go 1.27's new fixers

**Decision**: Run `go fix -diff -atomictypes -embedlit -slicesbackward -unsafefuncs ./...`
before and after the baseline edit. Apply only reviewed output.

**Rationale**: These are the four modernizers newly listed in the Go 1.27 release notes.
With Go 1.27.0 they currently return no diff for this repository. The broader default
`go fix` output includes unrelated older modernizers and is not the requested scope.

**Final audit (2026-09-01)**: Re-running the combined command after declaring `go 1.27` and
implementing the generic methods exited successfully with empty output. No modernizer source
change is applicable.

**Alternatives considered**:

- Apply all default `go fix` suggestions: rejected as unrelated churn.
- Manually rewrite atomic code because it resembles a fixer pattern: rejected because the
  targeted fixer intentionally proposed no change.

## Decision 8: Keep current validation gates and state local blockers explicitly

**Decision**: Require compile, build, full test, vet, race, targeted channels, modernizer,
and example gates. Use Ubuntu CI as authoritative for full/race validation.

**Rationale**: Local compile, build, vet, focused channels, and both examples pass on Go
1.27.0. The local full suite has a pre-existing Windows case-insensitive-environment blocker
in `TestConfigurationCasingAndPrecedence`. Local race is unavailable because cgo is disabled.
Neither limitation justifies weakening CI, which already runs all main gates on Ubuntu.

**Alternatives considered**:

- Skip or claim the blocked gates: rejected because project governance requires explicit
  acceptance accounting.
- Change configuration behavior in this feature: rejected as unrelated scope.

## Decision 9: Preserve historical and user-owned changes

**Decision**: Do not edit `docs/migration-v0.12.0.md`, `specs/001-fix-runtime-contracts/`,
or the existing user change in `examples/application/main_example.go`. Add current v0.13.0
material alongside the historical records.

**Rationale**: Historical artifacts describe the earlier Go 1.26 release accurately, and
the example runtime edit predates this feature.

**Alternatives considered**:

- Global replacement of Go 1.26: rejected because it falsifies release history.
