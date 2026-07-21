# Migrating from v0.11.x to v0.12.0

v0.12.0 keeps the module path `github.com/sedmess/go-ctx` and the standard-library-only
dependency policy. It intentionally raises the minimum supported toolchain from Go 1.21 to
Go 1.26, corrects one exported signature, and repairs several defective observable
behaviors.

## Required toolchain migration: Go 1.26

Upgrade local development, CI, release, and deployment build environments to Go 1.26 or
later before adopting v0.12.0. Environments that cannot provide a Go 1.26 toolchain are not
supported by this release. No module-path or third-party dependency migration is required.

## Required source migration: `StreamingChan.ToChan`

The old result was incorrectly send-only:

```go
func (ch StreamingChan[T]) ToChan(outBufSize int) (chan<- T, chan<- error)
```

The corrected result is receive-only:

```go
func (ch StreamingChan[T]) ToChan(outBufSize int) (<-chan T, <-chan error)
```

Change explicit `chan<- T` and `chan<- error` declarations, copied function types, and
interfaces to `<-chan T` and `<-chan error`. Inferred assignments require no declaration
change. Consumers receive or range over the channels and must not send to or close them.
These intentional pre-v1 compatibility changes are why the release is v0.12.0 rather than
a v0.11.x patch.

## Corrective behavior changes

- `Stop` is non-blocking and idempotent; `Join` waits for all owned cleanup and signal
  unregistration. Manual serialization of stop calls is no longer necessary.
- `AfterStart` and `BeforeStop` may query services, state, stats, and health. `BeforeStop`
  uses stable consumer-before-dependency order; `AfterStart` remains concurrent.
- Missing typed lookup returns zero/false. An incompatible value under the requested name
  still panics with the expected and actual types.
- Duplicate names no longer overwrite inside a package. Give distinct instances unique
  names and use an explicit testing substitution for overrides.
- Health retains the worst normalized severity and unknown statuses fail safe as application
  `DOWN`. Statistics maps and dependency slices are snapshots.
- Exact process configuration keys win, including present-empty values. Uppercase process
  and canonical non-process keys are fallbacks; source precedence is unchanged.
- Timer and connector replacement/stop joins the prior generation. Connector sends unblock
  when either endpoint terminates. In-flight user callbacks must still return.
- `nopanic.IsPanicWrapperError` recognizes direct and `%w`-wrapped captured panics; use
  `errors.As` when wrapper methods are needed.
- Testing environment overrides restore absent, present-empty, and present-value states
  exactly.

## Generic reflection at the new baseline

`Typed` now uses the generic reflection API available at the declared baseline. This does
not require consumer source changes beyond providing the Go 1.26 toolchain.

## Release validation

Before adopting or tagging v0.12.0, run the repository build, test, vet, and race gates on
an actual Go 1.26.x toolchain and the current stable toolchain, then run both examples. See
the feature [quickstart validation](../specs/001-fix-runtime-contracts/quickstart.md) for
the complete command list.
