# Quickstart Validation: Restore Runtime Contract Safety

Use this guide after implementation. It validates the contracts in
[`contracts/public-api.md`](contracts/public-api.md) and the migration in
[`contracts/migration-v0.12.0.md`](contracts/migration-v0.12.0.md); it does not replace the
implementation task list.

## Prerequisites

- Repository root: `github.com/sedmess/go-ctx`
- A clean or intentionally reviewed working tree
- An actual Go 1.26.x toolchain available locally or in pinned CI
- The current stable Go toolchain for the second validation pass
- No new module dependency

If `GOTOOLCHAIN=go1.26.5` is not already cached and network download is unavailable, run the
pinned commands in the project's Go 1.26 CI image. Do not substitute a newer toolchain and
treat the `go 1.26` directive as equivalent minimum-toolchain validation.

## 1. Compile before running behavior

```text
GOTOOLCHAIN=go1.26.5 go test -run '^$' ./...
```

Expected outcome:

- Every production and test package compiles.
- `ctx/application_context_singleton_test.go` consumes both `GetTypedService` results and uses
  supported `ctx` tags.
- No standard-library API newer than Go 1.26 is referenced.

## 2. Validate lifecycle access and idempotent shutdown

```text
go test -count=1 ./ctx -run 'Test(LifecycleCallbacksCanReenterContext|ApplicationConcurrentStopJoinAndStateReads|ApplicationRestartCycles|LifecycleCallbackFailureDoesNotSkipCleanup|StartupFailureDisposesInitializedDependencies)$'
```

Expected outcome:

- Direct diagnostic/service lookup inside start and stop callbacks completes.
- Concurrent and repeated `Stop` calls return before blocked cleanup is released.
- Exactly one shutdown callback runs.
- Repeated `Join` succeeds.
- One hundred restart cycles reach fresh initialized contexts and fully cleaned terminal
  contexts without deadlock.

Timeouts guard failed completion paths; resource tests use owned acknowledgements rather than
ambient goroutine counts.

## 3. Validate dependency order, duplicates, health, and snapshots

```text
go test -count=1 ./ctx -run 'Test(StableDependencyOrderAcrossPermutations|ApplicationStopsConsumerBeforeDependency|ServicePackagePreservesDuplicateEntriesForValidation|ServiceRegistrationRejectsDerivedCrossPackageMixedAndReservedNames|HealthAggregation|Statistics)'
```

Expected outcome:

- One thousand fixed-seed registration permutations place every dependent before each
  dependency and produce stable unrelated ordering.
- All explicit/derived/mixed/reserved duplicate paths fail visibly.
- One thousand repeated mixed-health aggregations return the same worst severity.
- Mutating a returned stats map or dependency slice cannot affect a later snapshot.

## 4. Validate owned timers, connectors, and signals

```text
go test -count=1 ./ctx -run 'Test(Connector|TimerTask|ApplicationSignal)'
```

Expected outcome:

- Connector stop waits for its listener acknowledgement and safely repeats/restarts.
- A blocked send unblocks when its sender or peer terminates.
- Timer replacement joins the prior worker; stop-before-tick and repeated stop are safe.
- The signal channel is buffered, explicit/signal stop unregister exactly once, and restart
  creates a fresh subscription.

These tests assert resource-owned completion tokens and fake ticker/signal calls rather than
ambient goroutine counts or sleeps.

## 5. Validate helper and configuration contracts

```text
go test -count=1 ./u/nopanic
go test -count=1 ./utils/channels
go test -count=1 ./ctx -run 'Test(GetTypedService|Configuration)'
go test -count=1 ./ctx/ctx_testing
```

Expected outcome:

- Direct and wrapped captured-panic errors classify correctly.
- An external consumer test compiles assignments to `<-chan T` and `<-chan error`, receives
  ordered values/errors, and observes both channels close.
- Missing typed lookup returns zero/false; mismatched wiring fails visibly.
- Configuration preserves source precedence, exact process spelling, present-empty values,
  and uppercase fallback.
- Testing overrides exactly restore absent, present-empty, and present-with-value variables.

## 6. Run the race-sensitive gates

```text
GOTOOLCHAIN=go1.26.5 go test -race -count=1 ./...
go test -race -count=1 ./...
```

Expected outcome:

- No race in application state, global context publication/clearing, stop/join, timer
  generations, connector generations, stats snapshots, or signal ownership.
- Process-global tests remain serial; independent package tests may still run normally.

## 7. Run all repository gates

Pinned baseline:

```text
GOTOOLCHAIN=go1.26.5 go build ./...
GOTOOLCHAIN=go1.26.5 go test -count=1 ./...
GOTOOLCHAIN=go1.26.5 go vet ./...
```

Current toolchain:

```text
go build ./...
go test -count=1 ./...
go vet ./...
```

Formatting:

```text
gofmt -l ctx u utils examples
```

Expected outcome: every command exits zero and `gofmt -l` prints no edited Go file.

## 8. Exercise public examples

Short task example:

```text
A=validation C=validation go run ./examples/task
```

Expected outcome: the task starts, resolves its dependency/configuration, finishes, and the
application exits zero.

Lifecycle application example:

```text
go run ./examples/application
```

The example performs bounded timer and connector work, requests stop after
`EXAMPLE_RUNTIME` (one second by default), completes one orderly shutdown, and returns all
`Join` waiters. Sending an interrupt earlier exercises the same shutdown coordinator.

## 9. Review the v0.12.0 migration

Compile at least one consumer fixture that:

- receives both `ToChan` outputs through receive-only channel variables;
- handles missing typed lookup through the boolean result;
- calls `Stop` and `Join` repeatedly;
- registers no duplicate names;
- requests both exact-case and canonical-fallback configuration keys.

Confirm release notes explicitly identify the `ToChan` signature change and Go 1.26 minimum
toolchain as incompatible changes, select v0.12.0, retain the module path, and link
[`contracts/migration-v0.12.0.md`](contracts/migration-v0.12.0.md).
