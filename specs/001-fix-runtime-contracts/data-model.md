# Phase 1 Data Model: Restore Runtime Contract Safety

This feature persists no data. Its model consists of in-memory lifecycle, dependency,
resource-ownership, diagnostic, and configuration states whose transitions are observable to
library consumers.

## 1. Application Runtime

Represents one invocation of application creation through completion.

### Fields

- **identity**: Unique pointer identity used when publishing and clearing the process-wide
  active context.
- **public state**: One of `not_initialized`, `initialization`, `initialized`, or `used`, with
  existing numeric codes unchanged.
- **internal stopping guard**: Prevents more than one shutdown execution without exposing a
  new public state.
- **root cancellation**: Fresh context and cancel function owned by this runtime.
- **stop request**: Close-once notification owned by the returned application handle.
- **completion**: Close-once notification closed only after all runtime cleanup.
- **services**: Registered service nodes indexed by validated name.
- **stop order**: Stable dependency-topological sequence calculated after initialization.
- **signal subscription**: Optional subscription owned by the shutdown coordinator.

### Relationships

- Owns zero or more **Service Nodes**.
- Owns exactly one **Shutdown Coordinator** after successful startup.
- Owns one root cancellation context shared with initialized service wrappers.
- Is the process-wide active context only between successful publication and final cleanup.

### State transitions

| From | Trigger | To | Required effects |
|------|---------|----|------------------|
| `not_initialized` | Start initialization | `initialization` | Reject further registration; initialize sorted service roots without holding callback-reentrant locks |
| `initialization` | Every service succeeds | `initialized` | Calculate stop order, publish global context, then invoke `AfterStart` |
| `initialization` | Any startup failure | `used` | Stop traversal, cancel root, dispose all successfully initialized services, release state, then report fatal startup failure |
| `initialized` | Explicit stop or signal | `initialized` plus internal stopping guard | Run one dependency-safe `BeforeStop` sequence while context lookup/health/stats remain valid |
| `initialized` plus stopping | Stop callbacks complete | `used` | Close context access, cancel root, dispose initialized services, release maps |
| `used` | Cleanup completes | `used` plus completion closed | Unregister signals, clear matching global identity, allow restart |
| Any stop-request state | Repeated/concurrent stop | Unchanged | Return without blocking or starting another shutdown |

### Validation rules

- Only one runtime may be globally active.
- No consumer callback executes while an application-context or global-pointer lock is held.
- `Join` completion implies all callbacks, disposal, resource joins, signal cleanup, and
  global clearing have completed.
- A callback may call `Stop`; a callback must not synchronously call `Join` on the runtime
  whose completion depends on that callback.

## 2. Service Node

Represents one registered service instance in the dependency graph.

### Fields

- **name**: Explicit or derived unique name; `CTX` remains reserved.
- **instance/type**: Reflected pointer-to-struct instance and immutable reflected type.
- **initialization state**: Not initialized, initializing, initialized, or used.
- **dependencies**: Deduplicated set of service names requested during injection or
  initialization; `CTX` is not a service-ordering edge.
- **descriptor**: Name, type, lifecycle capabilities, and dependency snapshot.
- **health reporter**: Optional health-reporting capability.

### Relationships

- An edge `consumer -> dependency` means the consumer must receive `BeforeStop` before the
  dependency.
- Multiple consumers may share one dependency.
- A service belongs to exactly one application runtime at a time, even when its instance is
  reused for a later runtime.

### Validation rules

- Duplicate explicit, derived, mixed, and cross-package names fail visibly.
- Cycles and missing dependencies fail initialization visibly.
- Only successfully initialized nodes participate in disposal.
- Stable Kahn ordering sorts ready nodes and adjacency by name and must emit every initialized
  node; otherwise the graph is invalid.

## 3. Shutdown Coordinator

Represents the single owner of shutdown initiation and completion.

### Fields

- **stop source**: Explicit close-once request or one delivered process signal.
- **signal channel**: Capacity at least one; only catchable configured signals.
- **completion owner**: Sole closer of the application completion channel.
- **global identity**: Runtime identity that may be cleared only on exact match.

### Validation rules

- Executes shutdown exactly once.
- Calls signal unregister exactly once before completion.
- Never blocks a repeated stop caller.
- Does not clear a newer runtime's global identity.

## 4. Asynchronous Resource Generation

A common ownership shape used by timer and connector implementations.

### Fields

- **generation identity**: Distinguishes a fresh run after replacement/restart.
- **stop notification**: Close-once channel.
- **completion acknowledgement**: Closed by the owned worker on exit.
- **running state**: Protected installation/detachment state.

### Timer specialization

- Owns one ticker and periodic action.
- Replacement joins the old generation before activating the new one.
- Stop joins an in-flight action; inactive/repeated stop succeeds immediately.

### Connector specialization

- Owns inbound/outbound message endpoints plus self and peer stop notifications.
- Active send uses backpressure but unblocks if sender or peer terminates.
- Listener exit waits for an in-flight handler and then acknowledges completion.
- A fresh application run receives fresh endpoints and notifications.

### Validation rules

- Data channels are not closed while concurrent senders may use them.
- A generation cannot be reused after its stop notification closes.
- Owner shutdown waits for completion rather than inferring exit from a sent signal.

## 5. Health Aggregate

Represents one pull-based application health snapshot.

### Fields

- **components**: Fresh map of service name to component report.
- **aggregate status**: Maximum normalized severity.

### Severity normalization

| Component status | Aggregate contribution | Rank |
|------------------|------------------------|------|
| `UP` | `UP` | 0 |
| `PARTIALLY` | `PARTIALLY` | 1 |
| `DOWN` | `PARTIALLY` | 1 |
| `DOWN_CRITICAL` | `DOWN` | 2 |
| Unknown | `DOWN` | 2 |

Aggregation selects the maximum rank, so later iteration can never improve an already worse
result.

## 6. Statistics Snapshot

Represents consumer-owned diagnostic output.

### Fields

- **services map**: Fresh map on every `Services()` call.
- **service descriptors**: Value copies.
- **dependency lists**: Fresh slice backing storage for every descriptor.

### Validation rules

- Mutating a returned map, descriptor, or dependency list cannot affect the container or a
  subsequent snapshot.
- Concurrent callers may mutate only their own snapshots without racing.

## 7. Configuration Lookup

Represents a requested key and ordered candidates.

### Fields

- **requested name**: Exact spelling supplied by the consumer.
- **canonical name**: Uppercase form used by non-process sources and fallback.
- **candidate value**: Value plus presence; present-empty is distinct from absent.
- **source**: Exact process, canonical process, arguments, custom file, default file, or
  application default.

### Resolution sequence

1. Exact requested process key.
2. Canonical process key when exact is absent.
3. Canonical property map already merged as arguments over custom file over default file over
   application defaults.

### Validation rules

- Present-empty stops fallback.
- Exact process spelling wins when exact and canonical variants coexist.
- All non-process keys, including `SetEnv`, are canonicalized before merge.
- Lazy property initialization and the requirement to call `SetEnv` first remain unchanged.

## 8. Environment Override Backup

Represents the original process state for one testing override.

### Fields

- **key**: Environment variable name.
- **value**: Original value, including empty.
- **present**: Whether the variable originally existed.

### Transitions

- Capture with presence before applying the override.
- Restore with `Setenv` when previously present.
- Restore with `Unsetenv` when previously absent.
- Cleanup registration precedes applying overrides so partial failure cannot skip restoration.

## 9. Exported Helper Results

### Typed lookup

- Missing: zero value and `false`.
- Present and compatible: value and `true`.
- Present and incompatible: visible mismatch failure containing requested/expected/actual type.

### Panic classification

- Direct or standard-wrapped panic wrapper: recognized.
- Nil or ordinary error: not recognized.

### Stream conversion

- Value and error outputs are receive-only to consumers.
- Producer sends and closes both channels.
- Error output contains at most one non-nil error.
- A caller abandoning an undrained value stream remains responsible for cancellation/draining;
  the existing signature provides no cancellation input.
