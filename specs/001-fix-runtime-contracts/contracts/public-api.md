# Public Library Contract: Runtime Safety Repairs

**Target release**: v0.12.0
**Module path**: `github.com/sedmess/go-ctx` (unchanged)
**Minimum supported toolchain**: Go 1.26.x

## Application lifecycle

The exported interface remains:

```go
type Application interface {
	Stop() Application
	Join()
}
```

### `Stop`

- Records a stop request without waiting for a receiver or cleanup.
- Is safe before the shutdown coordinator begins waiting, concurrently from multiple
  goroutines, after a process signal, and after completion.
- Repeated calls return the same application handle and execute no second shutdown.
- May be called from consumer code or a lifecycle callback.

### `Join`

- Waits until stop callbacks, root cancellation, disposal, connector/timer joins, signal
  unregister, service-map release, and matching global-context clearing have completed.
- Is safe repeatedly and concurrently, including after completion.
- Must not be called synchronously from a `BeforeStop`, `Dispose`, or in-flight timer/connector
  callback whose own return is required for that completion.

### Callback phase access

| Phase | Public state | `GetService` | `Stats` / `Health` | Root context |
|-------|--------------|--------------|--------------------|--------------|
| Service initialization | `initialization` | Use the supplied `ServiceProvider` | Not yet valid | Active |
| `AfterStart` | `initialized` | Valid, including global typed/name lookup | Valid | Active |
| Steady state | `initialized` | Valid | Valid | Active |
| `BeforeStop` | `initialized` | Valid | Valid | Active |
| Disposal | `used` | Invalid | Invalid | Canceled |
| After `Join` | `used` | No active global context | Invalid | Canceled |

No consumer callback executes while the container holds a context/global lock that the
callback can re-enter.

### Callback ordering and failure

- `AfterStart` remains concurrent; services must not assume ordering between callbacks.
- `BeforeStop` is sequential and stable. Every consumer precedes each dependency recorded
  during initialization; unrelated ready services are ordered by service name.
- A panic in `AfterStart`, `BeforeStop`, `Dispose`, or a timer/connector callback is captured,
  logged, and does not skip remaining container cleanup.
- Startup failure runs no `AfterStart` or `BeforeStop`. Successfully initialized services are
  canceled, disposed, and joined before the existing fatal startup report.

## Service registration and dependency ordering

- Every registration is preserved until container validation; package assembly never silently
  overwrites a named service.
- Duplicate explicit, derived, mixed, within-package, and cross-package names fail visibly.
- `CTX` remains reserved and fails through the same startup validation boundary.
- Initialization-time edges `consumer -> dependency` define the dependency-safe stop order.
- Root initialization and unrelated stop ordering use service-name order, not map iteration.
- Existing `ctx` tag grammar, reflected service-name derivation, missing/cycle failure, and
  pointer-to-struct registration requirements remain unchanged.

## Typed lookup

The signature remains:

```go
func GetTypedService[T any]() (T, bool)
```

| Condition | Result |
|-----------|--------|
| Active context; matching service absent | Zero `T`, `false` |
| Active context; service present and assignable | Typed service, `true` |
| Active context; resolved value incompatible with `T` | Visible panic identifying service name, expected type, and actual type |
| No active context or invalid lifecycle phase | Existing visible lifecycle failure |

## Application diagnostics

### Health

- Every call returns a fresh component map.
- Aggregate severity is order-independent and follows the table in `data-model.md`.
- An unknown component status fails safe as application `DOWN` while the original component
  report remains present.

### Statistics

- `AppContextStats.Services()` retains its existing signature.
- Every call returns a consumer-owned map, descriptor values, and dependency slices.
- Consumer mutation never alters a later snapshot or container state.

## Timers

`TimerTask.StartTimer` and `TimerTask.StopTimer` retain their signatures.

- At most one generation is active.
- Starting again stops and joins the prior generation before the replacement becomes active.
- Stop before first tick, repeated stop, concurrent stop, and stop while inactive are safe.
- Stop waits for an in-flight action to return and for the worker to acknowledge exit.
- A later start creates a fresh generation.
- An action must not synchronously start or stop the same timer because that would wait on
  itself.

## Service connectors

Existing public connector constructors and `Send` signatures remain unchanged.

- Each application run wires a fresh connection/listener generation.
- `BeforeStop` closes the listener's stop notification once and waits for listener exit.
- A closed inbound endpoint ends the listener; an in-flight handler finishes first.
- Active sends retain channel backpressure but return without delivery when sender or peer
  termination makes delivery impossible.
- Data channels are not closed while concurrent senders may use them.
- An `OnMessage` handler must return; shutdown waits for it and cannot forcibly terminate
  consumer code.

## Panic classification

`nopanic.IsPanicWrapperError(error)` returns true for wrappers produced by `Run`, `RunE`, or
`RunResult`, including through standard `%w` wrapping. It returns false for nil and ordinary
errors. Internal users extract the matching wrapper through standard error type traversal.

## Stream conversion

The corrected signature is:

```go
func (ch StreamingChan[T]) ToChan(outBufSize int) (<-chan T, <-chan error)
```

- The conversion worker alone sends and closes the returned channels.
- Values preserve source order.
- The error channel emits at most one non-nil source error and otherwise closes empty.
- Both channels close after conversion terminates.
- Slow consumers apply backpressure. Consumers must drain outputs; the existing method has no
  cancellation parameter for an abandoned stream.

## Configuration casing and precedence

For `GetEnv(name)`, presence includes an empty value and lookup is:

1. exact process-environment `name`;
2. uppercase process-environment name only when exact is absent;
3. uppercase canonical non-process properties.

Non-process properties retain precedence:

1. command arguments;
2. `.env_custom`;
3. `.env`;
4. defaults registered with `SetEnv`.

Every non-process key is canonicalized to uppercase when ingested. If exact and uppercase
process variables coexist, the exact spelling requested by the caller wins. Property loading
remains process-wide and lazy; `SetEnv` must run before the first lookup.

## Testing environment overrides

A testing application restores both environment presence and value:

- originally absent -> absent via unset;
- originally present-empty -> present-empty;
- originally present-with-value -> the original value.

Overrides and other process-global context/configuration/signal tests are serial by contract.

## Signal ownership

- One application coordinator owns one signal channel with capacity at least one.
- Only catchable configured signals are registered.
- Explicit stop and signal stop converge on the same one-time shutdown.
- The subscription is unregistered before `Join` returns and a restart registers a fresh one.
