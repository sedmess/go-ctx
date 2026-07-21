# Migration Contract: v0.11.x to v0.12.0

v0.12.0 keeps the module path and standard-library-only dependency policy. It intentionally
raises the minimum supported toolchain from Go 1.21 to Go 1.26, corrects one exported
signature, and repairs several defective observable behaviors.

## Required toolchain migration: Go 1.26

Upgrade local development, CI, release, and deployment build environments to Go 1.26 or
later before adopting v0.12.0. Environments that cannot provide a Go 1.26 toolchain are not
supported by this release. No module-path or third-party dependency migration is required.

## Required source migration: `StreamingChan.ToChan`

### Old signature

```go
func (ch StreamingChan[T]) ToChan(outBufSize int) (chan<- T, chan<- error)
```

### New signature

```go
func (ch StreamingChan[T]) ToChan(outBufSize int) (<-chan T, <-chan error)
```

### Consumer action

- Change explicit `chan<- T` and `chan<- error` declarations to `<-chan T` and `<-chan error`.
- Update function types and interfaces that copied the old exact signature.
- Receive or range over returned channels; do not send to or close them.
- Inferred assignments such as `values, errs := stream.ToChan(n)` require no declaration
  change and now permit the intended receive operations.

These are intentional pre-v1 compatibility changes and are the reason for selecting v0.12.0
rather than another v0.11.x patch.

## Corrective behavior changes

### Application stop and join

- `Stop` no longer blocks on repeated, concurrent, late, or post-signal calls.
- `Join` now means all container-owned cleanup and signal unregister have completed.
- Remove consumer workarounds that serialize stop calls manually.

### Lifecycle callback access and order

- `AfterStart` and `BeforeStop` may directly query services, stats, health, and state.
- `BeforeStop` is deterministic and dependency-safe; do not depend on the old incidental map
  or traversal order.
- `AfterStart` remains concurrent and unordered.

### Typed lookup

- Missing typed service now returns zero/false instead of panicking.
- Keep checking the existing boolean result.
- An incompatible value under the expected name still fails visibly because it is invalid
  wiring, not ordinary absence.

### Duplicate service registration

- Duplicate names previously overwritten inside one `ServicePackage` now fail startup.
- Remove duplicate registrations or give intentional multiple instances unique names.
- Testing overrides must use the testing substitution API rather than duplicate base entries.

### Health and statistics

- Health always retains the worst normalized component severity; malformed statuses fail safe
  as application `DOWN`.
- `Services()` results are snapshots. Mutating a returned map/slice no longer changes later
  diagnostics; consumers must keep their own mutable model if desired.

### Configuration key casing

- Exact process keys remain first and present-empty still suppresses fallback.
- When exact is absent, uppercase process and canonical non-process keys are considered.
- If an application intentionally sets both lower/mixed-case and uppercase process keys,
  request the intended exact spelling explicitly.
- Non-process source precedence is unchanged.

### Timers, connectors, and signals

- Repeated timer/connector stop is safe; replacement/restart joins the old generation.
- Connector sends unblock when sender or peer stops instead of hanging indefinitely.
- `Join` waits for in-flight timer actions and connector handlers; ensure callbacks return.
- Signal subscriptions no longer persist after application completion.

### Panic classification

- `IsPanicWrapperError` now recognizes direct and wrapped captured-panic errors.
- Code that needs wrapper methods should traverse the error chain rather than directly
  asserting a possibly wrapped error value.

## Generic reflection at the new baseline

`Typed` uses `reflect.TypeFor`, which is available at the declared Go 1.26 baseline.
Consumers need no related source change beyond providing the required toolchain.

## Release gate

Before tagging v0.12.0:

- Confirm the old/new stream signatures and all behavior changes above appear in release
  notes.
- Run the quickstart and all pinned Go 1.26 plus stable-toolchain gates.
- Confirm no new module dependency or module-path change.
- Verify both runnable examples use supported configuration/tag/lifecycle forms.
