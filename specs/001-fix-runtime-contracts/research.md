# Phase 0 Research: Restore Runtime Contract Safety

## Decision 1: Keep lifecycle locks out of consumer callbacks

**Decision**: Refactor `ctx/application_context.go` so lifecycle methods use short lock
sections only to validate or transition state and snapshot container-owned collections.
Service `Init`, `AfterStart`, `BeforeStop`, `Dispose`, health reporting, and logging execute
without the application-context or global-context lock held. Transition the public state to
`initialized` before `AfterStart`; retain public `initialized` during `BeforeStop`; transition
to `used` after stop callbacks and before cancellation/disposal, preserving existing public
state codes. Protect `State()` with `RLock`.

Publish the initialized context under the global pointer lock before `AfterStart` so start
callbacks may use global lookup. The global lock protects only pointer read/compare/replace;
callers copy the pointer, release the lock, and then invoke context methods. Never hold it
across initialization, callbacks, shutdown, or lookup.

**Rationale**: The current write lock spans callbacks and `WaitGroup.Wait`, while public
context methods attempt `RLock`, causing self-deadlock. Short state transitions retain
synchronization and allow documented context access without exposing a new public stopping
state.

**Alternatives considered**:

- Use an atomic state: rejected because maps and lifecycle snapshots still require a mutex.
- Add a public `stopping` state: rejected as an unnecessary observable contract change.
- Hold a read lock during callbacks: rejected because callbacks can re-enter container APIs
  and container code must not wait for consumer code while holding its locks.
- Cancel before `BeforeStop`: rejected because it changes the existing dependency-usable
  shutdown phase.

## Decision 2: Use one shutdown coordinator with close-once signals

**Decision**: Replace the mutex latch and boolean stop channel in
`ctx/application_context_singleton.go` with an application-owned `sync.Once`, a
`chan struct{}` stop request closed once, and a `chan struct{}` completion closed once.
`Stop()` records the request and returns immediately; `Join()` waits on completion and is
safe repeatedly. One coordinator handles explicit stop or a process signal, calls
`appContext.stop()` exactly once, unregisters the signal channel, clears only the matching
global context, and closes completion after all cleanup.

Remove the unused `eventBus`, event loop, event payload types, and stop-event hop from
`ctx/application_context.go`, `ctx/application_context_singleton.go`, and `ctx/events.go`.
They have no panic-event producers and create additional blocking ownership paths.

Use a signal channel buffered to one, register only catchable `os.Interrupt` and
`syscall.SIGTERM`, and call `signal.Stop` before completion. The standard library specifies
that signal delivery is non-blocking and callers must provide sufficient buffer space; a
single expected signal needs capacity one ([`os/signal.Notify`](https://pkg.go.dev/os/signal#Notify)).

**Rationale**: Closing records a stop request even before the coordinator waits and cannot
block late or concurrent callers. A completion channel is the correct reusable join
primitive. One coordinator gives every channel and signal subscription one owner.

**Alternatives considered**:

- Buffered one-shot sends: workable but less direct than a close-once state transition.
- Retain `RWMutex` as a join latch: rejected because it is not a completion primitive and
  permits incorrect cross-phase unlock behavior.
- `signal.NotifyContext`: valid, but the explicit buffered channel makes delivery capacity,
  logging, and unregister verification clearer.
- Keep the internal event loop: rejected because no remaining event producer needs it.

## Decision 3: Derive stop order from the dependency graph

**Decision**: Record each initialization-time service dependency as a deduplicated directed
edge `consumer -> dependency`. After initialization, calculate one stable stop order with
Kahn topological sorting: nodes with zero incoming edges are eligible first, and ready sets
and adjacency traversal are sorted by service name. This yields dependents before their
dependencies and deterministic ordering for unrelated nodes. Sort root service names before
initialization as well. `BeforeStop` remains sequential; `AfterStart` remains concurrent and
unordered as documented.

**Rationale**: Reversing the current pre-order traversal can stop a dependency first, and
moving the append alone does not make shared or independent node order stable across graph
discovery order. The explicit graph already corresponds to dependencies recorded for stats.

**Alternatives considered**:

- Append only after successful initialization and reverse: dependency-safe for simple DFS,
  but independent/shared ordering can still depend on discovery order.
- Stable DFS preorder: rejected for shared dependency graphs.
- Concurrent stop callbacks: rejected because dependency order would become unspecified.

## Decision 4: Clean successful initialization before fatal startup reporting

**Decision**: Make recursive initialization and internal configuration/injection paths
propagate internal errors instead of calling `logger.Fatal` mid-traversal. On missing
dependency, cycle, injection/configuration failure, initialization error, or captured panic:
stop initialization, mark the context unusable, cancel its root context, dispose and join
every successfully initialized service, release container state/global reservation, and only
then retain the existing fatal behavior at the exported creation boundary. Do not call
`AfterStart` or `BeforeStop` for a context that never completed initialization.

**Rationale**: `logger.Fatal` exits immediately, so deferred container cleanup cannot run.
The exported creation APIs have no error result; an internal error path plus the current
outer fatal boundary restores cleanup without changing their signatures.

**Alternatives considered**:

- Change creation APIs to return errors: cleaner but a broad breaking redesign outside this
  corrective feature.
- Dispose a service whose initializer was entered but did not complete: rejected because its
  disposer may require completed initialization; cancellation is the partial-work signal.
- Recover only root panics: insufficient for ordinary missing/cycle/returned-error paths.

## Decision 5: Give timers replaceable, joined generations

**Decision**: In `ctx/timer_task.go`, serialize lifecycle operations and store a per-run
generation with close-once stop and completion channels. `StartTimer` stops and joins any
prior generation before installing a fresh ticker worker. The worker owns `ticker.Stop`,
closes completion on exit, and preserves panic suppression for actions. `StopTimer` detaches,
closes, and joins the active generation; inactive and repeated calls return immediately.

**Rationale**: The current mutable closer can orphan an earlier ticker and blocks after the
only receiver exits. Joined generations make shutdown and restart observable.

**Alternatives considered**:

- Cancellation without waiting: rejected because application completion would not prove
  worker termination.
- One permanent worker: rejected as extra state for a replaceable periodic task.
- `time.AfterFunc`: rejected because it is not periodic and still needs join coordination.

An in-flight action must return before stop completes. Calling synchronous timer lifecycle
methods from that same action is unsupported to avoid self-wait and must be documented.

## Decision 6: Give connectors joined listener/connection generations

**Decision**: In `ctx/connectors.go`, create a synchronized connection generation containing
message endpoints, close-once self/peer stop signals, and listener completion. Connection
wiring prepares fresh generations on each application initialization. The listener uses
`return` on stop, detects closed inputs with the two-result receive, lets an in-flight handler
finish, and closes completion. `BeforeStop` closes once and waits. `Send` selects among
delivery, sender stop, and peer stop so shutdown cannot strand an unbuffered sender. Do not
close data channels that concurrent senders may still use. Use
`reflect.TypeOf((*T)(nil)).Elem()` for generic endpoint types, including interface types.

**Rationale**: Changing the current `break` to `return` fixes only one leak; it does not make
repeated stop, restart, blocked sends, or handler completion observable. Self and peer
termination paths satisfy the blocking-operation ownership rule.

**Alternatives considered**:

- Fix only the loop `break`: rejected as incomplete termination behavior.
- Close message channels: rejected because concurrent senders can panic and closed channels
  cannot be reused on restart.
- Let sends block after peer stop: rejected because the blocked operation has no termination
  path.

An `OnMessage` callback that never returns cannot be forcibly stopped; shutdown waits for
consumer callback completion and documentation must state that limitation.

## Decision 7: Preserve all registrations until container validation

**Decision**: Replace `ServicePackage`'s named-service map with an ordered internal entry
collection while preserving exported constructors and iteration methods. Pass every entry to
`appContext.register`, which retains the existing visible duplicate/reserved-name rejection.
In `ctx_testing`, reject duplicate base or repeated testing registrations while allowing one
explicit testing service to replace one base registration.

Use `reflect.TypeFor[T]()` for generic service and connector type queries. The API is
available within the approved Go 1.26 baseline and directly expresses the requested generic
type, including interface types.

**Rationale**: A map destroys the evidence needed for duplicate validation. An ordered
collection is internal, additive, and removes incidental map ordering. The reflection repair
uses the clearest standard-library form available at the declared baseline.

**Alternatives considered**:

- Panic immediately in `PackageOf`: rejected because it changes failure timing.
- Return an error from `PackageOf`: rejected as an unnecessary public signature break.
- Retain the map and check before assignment: catches only duplicates within one package and
  still makes iteration order unstable.
- Retain `reflect.TypeOf((*T)(nil)).Elem()`: compatible but less direct now that every
  supported toolchain provides `reflect.TypeFor`.

## Decision 8: Reduce health severity and snapshot statistics

**Decision**: Use an explicit order-independent health rank: `Up` contributes application
`Up`; component `Partially` or noncritical `Down` contributes application `Partially`;
component `DownCritical` contributes application `Down`. Select the maximum rank and preserve
component reports. Treat an unknown status conservatively as application `Down`.

Keep `AppContextStats.Services()`'s public map signature but return a fresh map and clone each
descriptor's `Dependencies` slice on every call.

**Rationale**: Maximum reduction cannot be downgraded by map iteration. Unknown malformed
health must not appear healthy. Deep snapshots prevent consumer mutation or races without a
signature change.

**Alternatives considered**:

- Last-write health aggregation: rejected as nondeterministic.
- Ignore unknown status: rejected because it can conceal malformed/down health.
- Replace stats maps/slices with new immutable types: rejected as an unnecessary break.

## Decision 9: Correct helper contracts and release as v0.12.0

**Decision**:

- Keep `GetTypedService[T]() (T, bool)`: missing service returns zero/false; assignable service
  returns value/true; incompatible dynamic type panics with requested, expected, and actual
  types. No-active-context and invalid-state failures remain visible.
- Implement `nopanic.IsPanicWrapperError` with `errors.As` so direct and wrapped panic errors
  classify correctly. Internal consumers needing the wrapper also extract it with
  `errors.As`, not a direct assertion after wrapped classification.
- Change `StreamingChan.ToChan` to return `(<-chan T, <-chan error)`. The producer retains send
  and close ownership; the error channel emits at most one non-nil error and otherwise closes
  empty.

Ship the set as **v0.12.0**. `ToChan` has shipped since v0.11.7 and its direction correction
is an intentional pre-v1 exported signature break. Migration guidance must show old/new
signatures, require explicit send-only declarations/function types to become receive-only,
and tell consumers never to send to or close returned channels.

**Rationale**: Typed lookup already advertises absence through `bool`; panic wrappers require
type classification rather than sentinel identity; stream conversion returns producer-owned
outputs and must encode receive ownership. v0.12.0 makes the incompatibility explicit without
changing the module path.

**Alternatives considered**:

- Return bidirectional stream channels: rejected because it grants unsafe send/close rights.
- Add a differently named method and leave `ToChan`: rejected because the misleading contract
  remains.
- Treat typed mismatch as absence: rejected because it hides corrupt wiring.
- Add `Is` to panic wrappers: rejected because wrapper type is not sentinel identity.

## Decision 10: Canonicalize non-process configuration while preserving exact environment keys

**Decision**: Resolve configuration in this order:

1. exact requested process key, including present-empty;
2. canonical-uppercase process key only when exact is absent;
3. canonical-uppercase non-process properties.

Canonicalize keys from `SetEnv`, `.env`, `.env_custom`, and command arguments when ingesting
them. Preserve non-process merge order: arguments, custom file, default file, then defaults.
If exact and uppercase process variables both exist, the spelling requested by the caller
wins. Continue lazy one-time non-process loading and require `SetEnv` before first lookup.

In `ctx_testing`, back up each override with `os.LookupEnv` as value plus presence. Restore a
present key with `Setenv` (including empty) and an absent key with `Unsetenv`; register cleanup
before applying overrides.

**Rationale**: Exact process behavior stays compatible, uppercase fallback makes file/argument
normalization coherent, and canonicalizing defaults prevents a lower-precedence case variant
from bypassing precedence. Presence-aware backup prevents cross-test contamination.

**Alternatives considered**:

- Uppercase before process lookup: rejected because it breaks exact-case variables.
- Scan `os.Environ` case-insensitively: rejected because collisions become order/platform
  dependent.
- Restore absence as empty: rejected because presence itself affects precedence/defaults.

## Decision 11: Validate ownership, not ambient goroutine counts

**Decision**: Tests use barriers, acknowledgements, internal completion channels, atomic
counters, manual ticker/signal seams, fixed-seed graph permutations, and bounded child
processes. Timeouts guard failure only. Deadlock and fatal paths run through the standard
self-subprocess pattern so a regression cannot poison global locks or exit the main test
process. Tests touching active context, configuration, environment, logger, or signals remain
serial.

Run the full 1,000-case ordering criterion against the pure topological-order function and a
smaller representative end-to-end callback set. Verify each timer/listener/coordinator exits
through its owned completion acknowledgement; do not assert `runtime.NumGoroutine`.

**Rationale**: Ambient goroutine counts and sleeps are flaky under the race detector and do
not prove a specific resource terminated. Owner acknowledgements directly test the lifecycle
contract.

**Alternatives considered**:

- Goroutine-count deltas: rejected because runtime/test goroutines are unrelated noise.
- Sleep then assert no event: rejected as scheduler-dependent negative timing.
- Third-party leak/fake-clock packages: rejected because no dependency is needed.

## Decision 12: Raise and gate the minimum toolchain at Go 1.26

**Decision**: Set `go.mod` to Go 1.26 and validate compilation, build, tests, vet, and race
coverage with an actual Go 1.26.x toolchain as well as the current stable toolchain. Run both
examples in bounded scenarios, require no new module dependency, and keep formatting clean.

**Rationale**: The baseline increase is explicitly approved and lets the library use the
current standard-library surface while making the consumer toolchain requirement honest in
module metadata, CI, and migration guidance. The installed Go 1.26.5 provides the actual
minimum-toolchain validation environment.

**Alternatives considered**:

- Validate only with whatever `stable` means at execution time: rejected because a later
  stable release would not prove compatibility with the declared Go 1.26 minimum.
- Retain the Go 1.21 baseline: rejected after explicit approval to make Go 1.26 the release
  requirement and document the resulting breaking compatibility impact.
- Add a compatibility-analysis dependency: rejected because the actual baseline compiler is
  authoritative.
