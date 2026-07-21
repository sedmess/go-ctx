# Feature Specification: Restore Runtime Contract Safety

**Feature Branch**: `001-fix-runtime-contracts`

**Created**: 2026-07-20

**Status**: Draft

**Input**: User description: "Fix the concrete correctness, lifecycle, wiring, configuration, testing, and exported-helper defects identified by the repository code review."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Start and Stop Applications Reliably (Priority: P1)

As an application author, I can start, inspect, stop, and restart an application without deadlocks, data races, leaked asynchronous work, or shutdown behavior that depends on registration-map iteration.

**Why this priority**: Application lifecycle failures can prevent a consuming process from starting or exiting and can leave process-wide resources active across restarts.

**Independent Test**: Assemble services that query application diagnostics during lifecycle callbacks, depend on one another during shutdown, use connectors and timers, and stop repeatedly. Verify every operation completes, dependent services stop before their dependencies, all owned asynchronous work terminates, and restart remains available.

**Acceptance Scenarios**:

1. **Given** an initialized service with access to application diagnostics and service lookup, **When** it uses those capabilities during its start or stop callback, **Then** the callback and lifecycle transition complete without blocking indefinitely.
2. **Given** a running application, **When** one or many callers request stop repeatedly before, during, or after shutdown, **Then** all calls return safely and exactly one shutdown sequence completes.
3. **Given** a dependent service and its dependency registered in any order, **When** the application stops, **Then** the dependent receives its stop callback before the dependency on every run.
4. **Given** services that own connector listeners, timers, or signal subscriptions, **When** the application stops and restarts repeatedly, **Then** every resource from the prior run terminates or unregisters before completion.
5. **Given** a concurrent lifecycle transition, **When** a consumer reads application state, **Then** the read is consistent and free from races.

---

### User Story 2 - Receive Deterministic Wiring and Diagnostics (Priority: P1)

As an application author, I receive explicit wiring failures and stable diagnostic results so the same service graph produces the same observable behavior on every run.

**Why this priority**: Silent duplicate replacement or nondeterministic health and shutdown results can select the wrong service or conceal a critical outage.

**Independent Test**: Register duplicate names through each supported package form and aggregate mixed health states repeatedly. Verify duplicates fail visibly before startup and every aggregation reports the same worst applicable severity.

**Acceptance Scenarios**:

1. **Given** two services with the same explicit name in one package, **When** the package is assembled, **Then** registration fails visibly instead of retaining one service silently.
2. **Given** health reporters with different severities including a critical failure, **When** health is aggregated repeatedly, **Then** every result reports the same worst applicable aggregate severity.
3. **Given** a consumer reading service statistics, **When** it inspects the result, **Then** it cannot mutate container-owned diagnostic state accidentally.

---

### User Story 3 - Use Exported Helpers as Their Contracts Promise (Priority: P2)

As a library consumer, I can rely on lookup results, panic classification, stream conversion, configuration resolution, and test environment restoration matching their documented contracts.

**Why this priority**: These defects are individually narrower than lifecycle failures but currently cause compilation errors, unexpected panics, incorrect classification, or state contamination between tests.

**Independent Test**: Compile and run consumer-level scenarios for missing typed services, captured panics, stream conversion, mixed-case configuration keys, and environment overrides. Verify returned values, channel direction, precedence, and post-test environment presence exactly match the documented contract.

**Acceptance Scenarios**:

1. **Given** an active application without the requested typed service, **When** a consumer performs typed lookup, **Then** it receives the type's zero value and a false presence result without a panic.
2. **Given** a panic captured by a panic-safe helper, **When** a consumer tests whether the returned error represents a captured panic, **Then** classification succeeds, including when the error is wrapped.
3. **Given** a stream converted into ordinary channels, **When** a consumer receives values and errors, **Then** the channels are usable in the documented receive direction and close after delivery.
4. **Given** configuration supplied through supported sources with case variants, **When** a consumer requests a key, **Then** resolution follows one documented deterministic rule without changing source precedence.
5. **Given** an environment key that was absent, present-empty, or present-with-value before a testing override, **When** the test application finishes, **Then** both the original presence and value are restored exactly.

---

### User Story 4 - Validate and Adopt the Corrected Release (Priority: P2)

As a maintainer or upgrading consumer, I can run the repository's required quality gates and understand any compatibility impact before adopting the corrected release.

**Why this priority**: The current tests do not compile, and at least one exported channel signature may require migration guidance and an intentional release decision.

**Independent Test**: Run every documented repository gate and consumer quickstart, then follow the migration guide from the previous exported contracts to the corrected ones.

**Acceptance Scenarios**:

1. **Given** a clean checkout, **When** all documented build, test, static-analysis, race, and example checks run, **Then** they complete successfully.
2. **Given** an existing consumer of an affected exported signature or behavior, **When** it reads the release guidance, **Then** it can identify the change, migration action, and intended semantic-version impact.
3. **Given** a consumer build environment older than Go 1.26, **When** the consumer prepares
   to adopt the corrected release, **Then** the guidance identifies the required toolchain
   upgrade before the module is adopted.

### Edge Cases

- A stop request arrives before the shutdown coordinator begins waiting, concurrently with another stop request, after signal-driven shutdown, or after shutdown has completed.
- A lifecycle callback queries services, health, statistics, or state while startup or shutdown is transitioning.
- Initialization fails after some dependencies initialize but before their dependent completes.
- A service starts a timer more than once, stops it before it fires, or stops it repeatedly.
- A connector is stopped before receiving a message, after its peer stops, or while a message handler is running.
- The same dependency graph is registered in different orders and is restarted repeatedly.
- Health contains no reporters, unknown statuses, multiple critical reporters, or a mix of every known severity.
- Duplicate names occur within one package, across packages, through derived names, or through explicit names; the reserved context name remains invalid.
- Configuration contains both an exact-case key and a canonical-case variant, and a test override replaces an absent or present-empty process variable.
- A captured panic error is wrapped before classification, and a typed lookup encounters a name with an incompatible value.
- A stream succeeds with zero values, returns an error, or has a slow/abandoned consumer.
- A newer development toolchain accepts a standard-library API that is unavailable on the
  declared minimum language/toolchain baseline.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The repository's existing test packages MUST compile before behavioral validation begins.
- **FR-002**: Lifecycle callbacks MUST be able to use the application capabilities documented as valid for their phase without deadlocking.
- **FR-003**: Application stop requests MUST be safe and non-blocking when repeated, concurrent, or issued after signal-driven shutdown.
- **FR-004**: Exactly one shutdown sequence MUST run for each application instance.
- **FR-005**: Dependents MUST receive ordered stop notification before the dependencies they consumed, independent of registration order.
- **FR-006**: Every application-owned listener, timer, channel, signal subscription, and goroutine MUST have one identifiable owner and a termination path that completes during shutdown.
- **FR-007**: Timer start, replacement, stop, repeated stop, and restart behavior MUST be explicit and deterministic.
- **FR-008**: Reads and writes of application lifecycle state MUST be synchronized.
- **FR-009**: Application restart after completed shutdown MUST remain supported without retaining resources from a prior run.
- **FR-010**: Duplicate and reserved service names MUST fail visibly for all supported package and naming forms before the application enters steady state.
- **FR-011**: Health aggregation MUST be order-independent and MUST preserve the worst applicable severity from all components.
- **FR-012**: Diagnostic service descriptions returned to consumers MUST not expose mutable container-owned collections.
- **FR-013**: Missing typed-service lookup MUST return a zero value and a false presence result; incompatible registered values MUST fail visibly with a diagnostic that identifies the contract mismatch.
- **FR-014**: Captured panic errors MUST be recognizable directly and through standard error wrapping.
- **FR-015**: Stream conversion MUST expose channels in a direction that permits consumers to receive all produced values and errors.
- **FR-016**: Configuration key casing MUST follow a documented deterministic rule while retaining the existing source-precedence order and compatibility for exact-case process variables.
- **FR-017**: Testing overrides MUST restore both the prior presence and prior value of each environment variable.
- **FR-018**: Signal handling MUST use a delivery-safe subscription and unregister it when the owning application completes.
- **FR-019**: Every reviewed defect MUST have colocated regression coverage that asserts observable behavior, including restart and concurrent paths where applicable.
- **FR-020**: Runnable examples MUST use supported tags, matching configuration keys, and lifecycle patterns that complete successfully.
- **FR-021**: Any corrected exported signature, behavior, or minimum-toolchain requirement that is not backward compatible MUST include migration guidance and an explicit semantic-version release decision before release.
- **FR-022**: The corrected library MUST compile and pass its supported validation on the
  actual declared minimum toolchain, not merely declare that version in module metadata,
  and MUST introduce no new external dependency unless separately justified.

### Contract and Lifecycle Impact *(mandatory)*

- **Public API compatibility**: Mostly corrective behavior changes. The stream-conversion channel direction changes an exported signature, and the minimum supported toolchain rises from Go 1.21 to Go 1.26; both require documented migration and an intentional semantic-version decision. Missing typed lookup changes from panic to its already-advertised false result.
- **Service wiring and configuration**: Duplicate-name validation is extended to package assembly. Tag grammar and service-name derivation remain unchanged. Configuration keeps the existing source-precedence order while defining compatible casing behavior.
- **Lifecycle and concurrency**: Startup and shutdown callback access, deterministic dependency stop order, idempotent stop, synchronized state, timer and connector termination, signal cleanup, failure cleanup, and restart are affected.
- **Failure and observability**: Duplicate names and incompatible typed lookup values fail visibly; health becomes deterministic; statistics are returned as snapshots; captured panic classification becomes reliable.
- **Documentation**: The consumer README, architecture lifecycle/configuration sections, package comments for affected exported helpers, runnable examples, and release migration notes require updates.

### Key Entities

- **Application lifecycle**: One application instance moving through registration, initialization, initialized operation, shutdown, and terminal use, with a single stop outcome shared by all callers.
- **Service dependency graph**: Named services and directed dependency relationships that determine initialization and ordered stop notification.
- **Owned asynchronous resource**: A listener, timer, channel, goroutine, or signal subscription associated with one application or service lifecycle and its termination condition.
- **Health aggregate**: Component health reports and their deterministic worst-severity application result.
- **Configuration key**: A requested key, its exact and canonical representations, source, presence state, and value.
- **Exported helper contract**: Lookup, panic-classification, and stream-conversion results consumed by downstream applications.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: One hundred consecutive start, repeated-stop, join, and restart cycles complete without deadlock, leaked owned work, or inconsistent terminal state.
- **SC-002**: Every tested dependency graph produces the same dependency-safe stop order across at least one thousand randomized registration orders.
- **SC-003**: At least one thousand aggregations of the same mixed health set produce one identical worst-severity result.
- **SC-004**: All supported duplicate-name paths are rejected in 100% of validation scenarios; none silently select a replacement service.
- **SC-005**: All reviewed exported-helper scenarios compile and return the documented value, error, and channel behavior without unexpected panic.
- **SC-006**: Environment overrides restore absent, present-empty, and present-with-value states correctly in 100% of test cases.
- **SC-007**: Every repository build, unit, static-analysis, concurrency, and runnable-example
  validation completes successfully from a clean checkout, including compilation with the
  declared minimum toolchain and validation with the current supported toolchain.
- **SC-008**: Every owned asynchronous resource introduced by the affected features has a regression scenario demonstrating termination on normal stop and restart.
- **SC-009**: Migration and release documentation accounts for 100% of corrected exported signatures and observable behavior changes.

## Assumptions

- All material defects reported in the preceding repository review are in scope; unrelated redesigns and new capabilities are out of scope.
- One process-wide active application context remains the supported operating model; simultaneous isolated contexts are not introduced.
- Existing service names, tag grammar, configuration-source precedence, and initialization callback interfaces remain stable unless a requirement explicitly states otherwise.
- Exact-case process-environment lookup remains compatible; canonical casing may be used only as a deterministic fallback when no exact key exists.
- Dependents may require their dependencies during stop notification, so dependency-safe ordering means dependents are notified before dependencies.
- Timer replacement stops the prior timer before a new timer becomes active, and stopping an inactive timer succeeds without blocking.
- No external dependency is necessary; existing packages and the language standard library are sufficient.
- Correcting the exported stream-conversion signature is treated as a breaking public-contract repair requiring migration guidance and a semantic-version release decision.
- Go 1.26 is the explicitly approved minimum language and toolchain baseline for the corrected release.
