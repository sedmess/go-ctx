# Tasks: Restore Runtime Contract Safety

**Input**: Design documents from `/specs/001-fix-runtime-contracts/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/`, `quickstart.md`

**Tests**: Every behavior change has a regression task before implementation. Lifecycle,
channel, timer, signal, and shared-state work requires the race detector.

**Organization**: Tasks are grouped by user story and executed phase by phase. Tasks marked
`[P]` touch independent files and may run concurrently after their phase prerequisites.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish repository hygiene and the real minimum-toolchain gate.

- [X] T001 Verify and add missing Go/universal ignore patterns in `.gitignore`
- [X] T002 Set the Go 1.26 baseline in `go.mod` and create a Go 1.26.x/stable validation matrix in `.github/workflows/ci.yml`
- [X] T003 Record the v0.12.0 compatibility target and unchanged dependency policy in `specs/001-fix-runtime-contracts/plan.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Restore compilation and establish the approved minimum-version baseline before behavioral TDD.

**⚠️ CRITICAL**: No user-story implementation begins until this phase completes.

- [X] T004 Repair the two-result typed lookup use and unsupported injection tag in `ctx/application_context_singleton_test.go`
- [X] T005 Use the Go 1.26 `reflect.TypeFor` API for generic service construction in `ctx/service_package.go`
- [X] T006 Validate the compile-only gate against `go.mod` with actual Go 1.26.x and stable toolchains

**Checkpoint**: Production and test packages compile on the declared baseline; behavioral tests may now be added.

---

## Phase 3: User Story 1 - Start and Stop Applications Reliably (Priority: P1) 🎯 MVP

**Goal**: Application startup, callback access, stop, cleanup, and restart are deadlock-free,
race-free, deterministic for dependencies, and terminate every owned asynchronous resource.

**Independent Test**: Services directly query context APIs in lifecycle callbacks; 64
concurrent stop callers and repeated joins complete one shutdown; 100 restarts cleanly finish;
dependency permutations are stable; timer, connector, and signal generations acknowledge exit.

### Tests for User Story 1

- [X] T007 [P] [US1] Add callback-access, concurrent stop/join, state-race, restart, callback-failure, and startup-failure cleanup tests in `ctx/application_context_lifecycle_test.go`
- [X] T008 [P] [US1] Add fixed-seed dependency DAG ordering and end-to-end stop-order tests in `ctx/dependency_order_test.go`
- [X] T009 [P] [US1] Add listener exit, in-flight handler, blocked send, repeated stop, and restart tests in `ctx/connectors_test.go`
- [X] T010 [P] [US1] Add manual-ticker replacement, stop-before-tick, repeated stop, in-flight action, and restart tests in `ctx/timer_task_test.go`
- [X] T011 [P] [US1] Add buffered registration, explicit/signal stop, unregister, and restart tests in `ctx/application_signal_test.go`

### Implementation for User Story 1

- [X] T012 [US1] Implement stable consumer-before-dependency topological ordering in `ctx/dependency_order.go`
- [X] T013 [US1] Refactor state transitions, callback-safe locking, dependency recording, ordered stop, snapshots, cleanup, and synchronized state reads in `ctx/application_context.go`
- [X] T014 [US1] Replace mutex/send shutdown with close-once `Stop`, completion-based `Join`, identity-safe global publication, and one signal-owning coordinator in `ctx/application_context_singleton.go`
- [X] T015 [US1] Propagate startup wiring/configuration/init failures internally and extract wrapped panic errors safely in `ctx/reflective.go`
- [X] T016 [US1] Remove unused shutdown/panic event infrastructure from `ctx/events.go`
- [X] T017 [US1] Implement joined self/peer-aware connector generations and Go 1.26 generic type reflection in `ctx/connectors.go`
- [X] T018 [US1] Implement replaceable joined timer generations with idempotent start/stop in `ctx/timer_task.go`
- [X] T019 [US1] Run focused lifecycle and race validation for `ctx/application_context_lifecycle_test.go`, `ctx/dependency_order_test.go`, `ctx/connectors_test.go`, `ctx/timer_task_test.go`, and `ctx/application_signal_test.go`

**Checkpoint**: User Story 1 is independently usable; stop/join/restart and owned-resource termination pass focused race validation.

---

## Phase 4: User Story 2 - Receive Deterministic Wiring and Diagnostics (Priority: P1)

**Goal**: Duplicate wiring fails visibly and diagnostic health/statistics are deterministic,
order-independent, and consumer-owned.

**Independent Test**: Every duplicate/reserved-name path fails before steady state; 1,000
mixed health aggregations agree; mutations of returned statistics never affect later calls.

### Tests for User Story 2

- [X] T020 [P] [US2] Add within/cross-package explicit, derived, mixed, testing-override, and reserved-name tests in `ctx/service_package_test.go`
- [X] T021 [P] [US2] Add exhaustive permutation, unknown-status, and 1,000-iteration health tests in `ctx/application_context_health_test.go`
- [X] T022 [P] [US2] Add deep map/slice snapshot and concurrent mutation tests in `ctx/application_context_stats_test.go`

### Implementation for User Story 2

- [X] T023 [US2] Preserve ordered package entries through container duplicate validation in `ctx/service_package.go`
- [X] T024 [US2] Reject duplicate base/testing registrations while allowing one explicit substitution in `ctx/ctx_testing/application_context_helper.go`
- [X] T025 [US2] Implement order-independent worst-severity health reduction in `ctx/application_context_health.go`
- [X] T026 [US2] Return deep consumer-owned diagnostic snapshots in `ctx/application_context_stats.go`
- [X] T027 [US2] Run focused deterministic wiring, health, stats, and race validation for `ctx/service_package_test.go`, `ctx/application_context_health_test.go`, and `ctx/application_context_stats_test.go`

**Checkpoint**: User Story 2 is independently usable; wiring and diagnostics no longer depend on map order or shared mutable collections.

---

## Phase 5: User Story 3 - Use Exported Helpers as Their Contracts Promise (Priority: P2)

**Goal**: Typed lookup, panic classification, stream conversion, configuration casing, and
testing overrides match their exported contracts.

**Independent Test**: External consumers receive stream outputs, missing typed lookup returns
zero/false, direct/wrapped panic errors classify, all configuration sources preserve precedence,
and absent/empty/value environment states restore exactly.

### Tests for User Story 3

- [X] T028 [P] [US3] Add missing and incompatible typed-service lookup tests in `ctx/application_context_singleton_test.go`
- [X] T029 [P] [US3] Add direct, wrapped, nil, and ordinary panic-classification tests in `u/nopanic/safe_run_test.go`
- [X] T030 [P] [US3] Add external receive-only compile, ordered value/error, closure, and slow-drain tests in `utils/channels/streaming_test.go`
- [X] T031 [P] [US3] Add child-process source precedence, exact/canonical casing, and present-empty tests in `ctx/config_test.go`
- [X] T032 [P] [US3] Add absent, present-empty, and present-value override restoration tests in `ctx/ctx_testing/application_context_helper_test.go`

### Implementation for User Story 3

- [X] T033 [US3] Return zero/false for absent typed service and diagnose incompatible values in `ctx/application_context_singleton.go`
- [X] T034 [US3] Implement direct/wrapped panic-wrapper classification with standard error traversal in `u/nopanic/safe_run.go`
- [X] T035 [US3] Change `StreamingChan.ToChan` to producer-owned receive-only outputs in `utils/channels/streaming.go`
- [X] T036 [US3] Canonicalize non-process keys and implement exact-then-uppercase environment lookup without changing precedence in `ctx/config.go`
- [X] T037 [US3] Restore exact environment presence/value after testing overrides in `ctx/ctx_testing/application_context_helper.go`
- [X] T038 [US3] Run focused helper, configuration, test-support, and race validation for `ctx`, `ctx/ctx_testing`, `u/nopanic`, and `utils/channels`

**Checkpoint**: User Story 3 is independently usable and consumer compile contracts match v0.12.0.

---

## Phase 6: User Story 4 - Validate and Adopt the Corrected Release (Priority: P2)

**Goal**: Maintainers can validate a clean checkout and consumers can migrate deliberately to v0.12.0.

**Independent Test**: Pinned Go 1.26/stable build, test, vet, race, task example, and
signal-driven application example all pass; migration guidance accounts for every changed contract.

### Validation and documentation for User Story 4

- [X] T039 [P] [US4] Publish lifecycle, ordering, ownership, health, stats, lookup, and configuration contracts in `docs/architecture.md`
- [X] T040 [P] [US4] Publish v0.12.0 usage and compatibility guidance in `readme.md`
- [X] T041 [P] [US4] Publish panic and stream helper contracts in `utils.md`
- [X] T042 [P] [US4] Copy and adapt the reviewed migration contract into `docs/migration-v0.12.0.md`
- [X] T043 [US4] Correct configuration casing and bounded lifecycle usage in `examples/application/main_example.go`
- [X] T044 [US4] Correct runnable task configuration behavior in `examples/task/run_example.go`
- [X] T045 [US4] Validate both examples using the scenarios in `specs/001-fix-runtime-contracts/quickstart.md`

**Checkpoint**: User Story 4 provides a validated and documented v0.12.0 adoption path.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Complete formatting, governance, and all repository quality gates.

- [X] T046 Format every edited Go file and verify no output from `gofmt -l` for `ctx/`, `u/`, `utils/`, and `examples/`
- [X] T047 Run pinned Go 1.26.x compile, build, test, and vet gates against `go.mod`
- [X] T048 Run current-toolchain `go build ./...`, `go test -count=1 ./...`, and `go vet ./...`
- [X] T049 Run current and pinned `go test -race -count=1 ./...` for lifecycle/concurrency validation
- [X] T050 Reconcile implementation behavior with `specs/001-fix-runtime-contracts/spec.md`, `plan.md`, `contracts/`, and `quickstart.md`
- [X] T051 Re-check `.specify/extensions.yml` post-implementation hooks and mark every completed task in `specs/001-fix-runtime-contracts/tasks.md`

---

## Dependencies & Execution Order

### Phase dependencies

- **Setup (Phase 1)**: Starts immediately.
- **Foundational (Phase 2)**: Depends on Setup and blocks every story.
- **US1 (Phase 3)**: Depends on Foundational and establishes lifecycle ownership used by all later stories.
- **US2 (Phase 4)**: Depends on Foundational and integrates with US1's dependency graph/snapshots.
- **US3 (Phase 5)**: Depends on Foundational; configuration startup propagation integrates after US1.
- **US4 (Phase 6)**: Depends on US1, US2, and US3 behavior being complete.
- **Polish (Phase 7)**: Depends on all stories.

### Within each story

- Write the listed regression tests before implementation and observe the baseline failure or
  compile mismatch where safe.
- Child processes isolate fatal/deadlock paths; process-global tests remain serial.
- Implement pure helpers before container integration.
- Run the story checkpoint before advancing.

### Parallel opportunities

- T007–T011 can be authored in parallel because they use separate test files.
- T020–T022 can be authored in parallel.
- T028–T032 can be authored in parallel across packages/files.
- T039–T042 can be authored in parallel after behavior stabilizes.
- Full user stories are not run in parallel in this implementation because US2/US3 share core
  files and the process-global application/configuration test state.

## Parallel Example: User Story 1

```text
Task T007: Add lifecycle callback/stop/restart tests in ctx/application_context_lifecycle_test.go
Task T008: Add dependency-order tests in ctx/dependency_order_test.go
Task T009: Add connector termination tests in ctx/connectors_test.go
Task T010: Add timer generation tests in ctx/timer_task_test.go
Task T011: Add signal ownership tests in ctx/application_signal_test.go
```

## Implementation Strategy

### MVP first

1. Complete Setup and Foundational compilation.
2. Complete US1 lifecycle tests and implementation.
3. Validate US1 with focused unit and race tests before changing diagnostics/helpers.

### Incremental delivery

1. US1: lifecycle ownership and termination.
2. US2: deterministic wiring and diagnostics.
3. US3: exported helpers and configuration contracts.
4. US4: examples, migration, and release validation.
5. Polish: pinned/current gates and final constitution reconciliation.

## Notes

- `[P]` means independent files after phase prerequisites, not permission to mutate shared
  process-global state concurrently.
- Every task names an exact repository path or validation target.
- No external dependency or module-path change is permitted.
- Commit creation is outside this task unless explicitly requested.
