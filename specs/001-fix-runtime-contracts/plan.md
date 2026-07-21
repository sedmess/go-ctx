# Implementation Plan: Restore Runtime Contract Safety

**Branch**: `001-fix-runtime-contracts` | **Date**: 2026-07-21 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-fix-runtime-contracts/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command; its definition describes the execution workflow.

## Summary

Repair the reviewed lifecycle, dependency-wiring, health, configuration, test-support, and
exported-helper defects while retaining the single-active-context model and intentionally
raising the minimum supported toolchain from Go 1.21 to Go 1.26. The design will replace
one-shot blocking shutdown paths with application-owned
completion signaling, move callbacks outside container write-lock critical sections,
derive deterministic shutdown from the dependency graph, snapshot diagnostics, normalize
health severity and configuration lookup, and add focused regression and race coverage.
The exported stream channel correction is an intentional v0.12.0 receive-only signature
repair with an explicit migration contract.

## Technical Context

**Language/Version**: Go 1.26

**Primary Dependencies**: Go standard library only (`context`, `errors`, `os`, `os/signal`,
`reflect`, `slices`, `sync`, `syscall`, `time` as already used); no new dependency proposed

**Storage**: N/A unless the feature explicitly introduces persistence

**Testing**: Go `testing` package; actual Go 1.26 toolchain compile/build/test/vet gate;
current stable toolchain gate; race detector for lifecycle or concurrency work

**Target Platform**: Go 1.26-supported platforms

**Project Type**: Reusable Go library with runnable examples

**Performance Goals**: Repeated stop returns without waiting for an unavailable receiver;
100 start/stop/restart cycles complete in bounded test time; health and shutdown-order
validation remains practical at 1,000 iterations; no unbounded goroutine, ticker, channel,
or signal-subscription growth across restart loops

**Constraints**: Preserve service names, `ctx`/`env` tag grammar, configuration-source
precedence, single-active-context behavior, and Go 1.26 compatibility; invalid wiring must
remain visible; callbacks must not run while holding a lock they can re-enter through public
context APIs; every asynchronous resource must terminate; Go 1.26 standard-library APIs such
as `reflect.TypeFor` may be used; no external dependency

**Scale/Scope**: Core changes span `ctx` lifecycle, registration, health, stats,
configuration, timers, connectors, and test support; helper changes span `u/nopanic` and
`utils/channels`; regression coverage includes 100 restart cycles and 1,000 randomized or
repeated deterministic-result checks

**Public API / Compatibility Impact**: Corrective behavior changes make `Stop` idempotent,
make missing typed lookup return `(zero, false)`, and make panic classification reliable.
`StreamingChan.ToChan` changes from send-only results to `(<-chan T, <-chan error)`. Ship the
intentional pre-v1 signature repair as v0.12.0 with `contracts/migration-v0.12.0.md` and
consumer release notes. The Go baseline increase from 1.21 to 1.26 is also an intentional
breaking compatibility change documented for v0.12.0; the module path is unchanged.

**Reflection / Configuration Impact**: Tag grammar, reflected naming, injected types, and
source precedence remain unchanged. Duplicate explicit names must be detected before a
package map can overwrite them. Configuration keeps exact-case process lookup first and
adds a documented canonical-uppercase fallback only when the exact key is absent.

**Lifecycle / Concurrency Impact**: Startup callback access, state transitions, idempotent
stop/join, dependency-safe stop order, initialization-failure cleanup, state synchronization,
connector listener termination, timer replacement/stop, buffered signal delivery,
subscription cleanup, and restart are all affected. Use short callback-free critical
sections, one shutdown coordinator, close-once stop/completion channels, stable dependency
topological ordering, and joined per-run timer/connector generations.

**Documentation Impact**: Update `readme.md`, `docs/architecture.md`, `utils.md`, comments on
affected exported methods, `examples/application/main_example.go`, and add release migration
guidance for any incompatible exported signature.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] **Stable contracts**: Exported APIs, service naming, tag semantics, configuration
      precedence, failure behavior, and Go 1.26 compatibility are preserved or have an
      explicit migration and versioning plan. The stream signature correction and Go
      baseline increase are both explicitly scoped to v0.12.0 with consumer guidance.
- [x] **Package direction**: Work stays within the documented package boundaries; generic
      helpers do not depend on `ctx`; every new dependency is justified.
      All work remains in existing packages and uses only the standard library.
- [x] **Wiring and configuration**: Reflection, `unsafe`, service resolution, cycles,
      environment loading, and invalid-input behavior are identified and testable when
      affected. Tag grammar and reflection setting are unchanged; duplicate validation and
      configuration casing have explicit regression scope.
- [x] **Lifecycle and concurrency**: Startup, shutdown, cancellation, error, restart, and
      ownership of every goroutine/channel/ticker/signal path are defined when affected.
      Phase 0 must select concrete ownership and coordination patterns before design closes.
- [x] **Verification**: Regression tests cover changed behavior; `go test -race ./...` is
      planned for concurrency-sensitive work.
      Build, unit, vet, race, and runnable-example gates are required.
- [x] **Documentation**: User-facing API and architectural changes name the exact guide,
      comment, or example that will be updated.
      Exact documentation targets are listed in Technical Context.

**Pre-research gate result**: PASS. No constitution exception is required. Phase 0 has now
resolved the two delegated design decisions in `research.md`.

## Project Structure

### Documentation (this feature)

```text
specs/001-fix-runtime-contracts/
├── plan.md              # This file (/speckit-plan command output)
├── research.md          # Phase 0 output (/speckit-plan command)
├── data-model.md        # Phase 1 output (/speckit-plan command)
├── quickstart.md        # Phase 1 output (/speckit-plan command)
├── contracts/           # Phase 1 output (/speckit-plan command)
└── tasks.md             # Phase 2 output (/speckit-tasks command - NOT created by /speckit-plan)
```

### Source Code (repository root)

```text
ctx/
├── application_context.go                 # State, init/stop order, callbacks, snapshots
├── application_context_singleton.go       # Stop/Join, global context, signal ownership
├── dependency_order.go                    # Stable dependency-topological ordering
├── reflective.go                          # Startup error propagation and wrapper extraction
├── application_context_health.go          # Order-independent severity reduction
├── application_context_stats.go           # Defensive diagnostic snapshots
├── application_context_singleton_test.go  # Core regression and compile repairs
├── application_context_lifecycle_test.go  # Callback access, stop, restart, failure cleanup
├── application_context_health_test.go     # Severity permutations and unknown status
├── application_context_stats_test.go      # Deep snapshot ownership
├── application_signal_test.go             # Buffered registration and unregister ownership
├── dependency_order_test.go               # 1,000 fixed-seed graph permutations
├── service_package.go                     # Duplicate rejection and Go 1.26 reflection
├── service_package_test.go                # Duplicate/reserved-name paths
├── config.go                              # Compatible casing resolution
├── config_test.go                         # Source/casing precedence in child processes
├── connectors.go                          # Listener termination
├── connectors_test.go                     # Send/listener stop, join, restart
├── timer_task.go                          # Timer replacement and idempotent stop
├── timer_task_test.go                     # Manual ticker generations
├── events.go                              # Removed unused internal event infrastructure
└── ctx_testing/
    ├── application_context_helper.go       # Exact environment restoration
    └── application_context_helper_test.go  # Presence/value regression coverage

u/nopanic/
├── safe_run.go                            # Wrapped panic classification
└── safe_run_test.go                       # Direct/wrapped classification coverage

utils/channels/
├── streaming.go                           # Receive-capable conversion contract
└── streaming_test.go                      # Consumer compile/close/error coverage

examples/application/main_example.go       # Supported configuration/lifecycle usage
readme.md                                  # Consumer contract and migration summary
utils.md                                   # Helper contract documentation
docs/architecture.md                       # Lifecycle, ownership, ordering, configuration
docs/migration-v0.12.0.md                  # Published consumer migration guidance
.github/workflows/ci.yml                   # Actual Go 1.26.x and stable-toolchain gates
specs/001-fix-runtime-contracts/            # Feature design and migration contract
```

**Structure Decision**: Keep all changes in the existing package boundaries. `ctx` retains
container orchestration and lifecycle ownership; `ctx/ctx_testing` remains a public consumer
of `ctx`; `u/nopanic` and `utils/channels` remain generic and do not import `ctx`. Tests are
colocated with each affected package, and examples exercise only public APIs.

## Phase 0 Research Outcome

The complete decisions, rationale, and rejected alternatives are recorded in
[`research.md`](research.md). Key outcomes are:

- callbacks execute outside container/global locks, with `initialized` published before
  `AfterStart` and retained through `BeforeStop`;
- one coordinator owns explicit/signal shutdown, signal unregister, global clearing, and the
  final completion close;
- close-once stop plus completion channels replace blocking sends and mutex latches;
- a stable `consumer -> dependency` topological order drives sequential `BeforeStop`;
- startup errors propagate internally so successful dependencies clean up before the existing
  fatal creation boundary;
- timer and connector work uses joined replaceable generations with self/peer termination;
- duplicate entries remain intact until validation; health uses maximum severity; stats are
  deep snapshots;
- exact process configuration is compatible, uppercase fallback is deterministic, and
  testing restoration preserves presence;
- `ToChan` becomes receive-only and all repairs ship as v0.12.0;
- a real Go 1.26.x gate validates the declared minimum while a stable-toolchain gate catches
  forward compatibility; generic type queries use `reflect.TypeFor` at the supported baseline.

## Phase 1 Design

### Lifecycle and global ownership

Split context startup into registration, internal initialization, transition to initialized,
global publication, and callback notification. Only state transitions and snapshots occur
under `appContext` locks. A separate short global-pointer lock enforces the single-active-
context rule and identity-safe clearing; it is never held while invoking context methods.

The returned application owns `stopOnce`, `stopCh`, and `doneCh`. A single coordinator waits
on explicit stop or its capacity-one signal channel, calls one context shutdown, unregisters
signals, clears the matching global, and closes `doneCh`. `Stop` is immediate and idempotent;
`Join` is repeatable and means all cleanup is complete. Remove the unused event loop and event
types rather than retaining a second shutdown owner.

During shutdown, snapshot the stable service/order data, run dependency-safe `BeforeStop`
without locks while public state remains initialized, transition to used, cancel root context,
dispose initialized services, release maps, then finish coordinator cleanup. Initialization
failure skips start/stop callbacks but cancels/disposes every successfully initialized
dependency before fatal reporting.

### Dependency, wiring, and observability

`ServicePackage` uses an internal ordered entry list, preserving every registration for the
existing container duplicate/reserved-name validation. Initialization records deduplicated
dependency edges. `dependency_order.go` applies stable Kahn sorting by service name; a pure
test exercises 1,000 fixed-seed permutations, and representative end-to-end tests confirm
callbacks consume the result.

Health maps known statuses to explicit ranks and takes the maximum; unknown status contributes
application `DOWN`. Stats retain their public map signature but deep-copy the map and every
dependency slice. Typed lookup treats ordinary absence as zero/false and invalid type mismatch
as a visible diagnostic failure.

### Asynchronous service resources

Timer and connector structs gain internal operation synchronization and per-run stop/done
generations. Replacement stops and joins the old generation. Connector endpoints include
self/peer termination so `Send` cannot remain blocked after shutdown; message data channels
are not closed under concurrent senders. Tests use owned completion acknowledgements and
manual ticker/signal seams, never global goroutine counts.

### Helpers, configuration, and compatibility

Panic-wrapper classification and extraction use standard error type traversal. `ToChan`
returns receive-only outputs, preserves ordering/error/closure behavior, and documents that
abandoned streams still require consumer draining because no cancellation input exists.

Configuration checks exact process spelling, canonical process fallback, then the canonical
non-process map. Every non-process source canonicalizes keys before precedence merge.
Testing overrides capture value plus presence and restore with set or unset accordingly.

The public runtime contract is [`contracts/public-api.md`](contracts/public-api.md). The
incompatible stream signature and all corrective behavior are covered by
[`contracts/migration-v0.12.0.md`](contracts/migration-v0.12.0.md), which becomes published
`docs/migration-v0.12.0.md` during implementation.

### Verification design

Use self-subprocess tests for potential deadlock/fatal paths, barriers and atomic counters for
concurrent stop, manual ticker/signal seams for resource ownership, fixed local random seeds
for ordering, and external-package compilation for receive-only stream outputs. Serialize
tests that mutate process-global context, configuration, environment, logging, or signals.
The runnable commands and expected outcomes are in [`quickstart.md`](quickstart.md).

## Post-Design Constitution Check

- [x] **Stable contracts**: v0.12.0, unchanged module path, explicit old/new stream signature,
      migration guidance, behavior notes, the explicit Go 1.26 baseline increase, and actual
      Go 1.26 validation are defined.
- [x] **Package direction**: All implementation remains in existing package responsibilities;
      the one new core file remains in `ctx`; no helper imports `ctx`; no external dependency.
- [x] **Wiring and configuration**: Duplicate/reserved names, stable graph ordering, unchanged
      tag grammar, compatible casing fallback, precedence, and invalid-input behavior have
      explicit contracts and tests.
- [x] **Lifecycle and concurrency**: Startup, callback access, failure cleanup, stop,
      cancellation, disposal, restart, and every connector/timer/signal/goroutine/channel
      owner and termination acknowledgement are defined.
- [x] **Verification**: Colocated regression, child-process, deterministic stress, baseline,
      full build/test/vet, race, and runnable-example gates are specified.
- [x] **Documentation**: `readme.md`, `utils.md`, `docs/architecture.md`, exported comments,
      examples, v0.12.0 migration guidance, and release notes are named.

**Post-design gate result**: PASS. No unresolved clarifications, dependency additions,
package-boundary violations, or temporary constitution exceptions remain.

## Complexity Tracking

No constitution violations or temporary exceptions are planned. The standard-library-only
dependency policy is unchanged. The confirmed stream-conversion signature correction ships
as v0.12.0 with the explicit migration and semantic-version treatment documented above.
