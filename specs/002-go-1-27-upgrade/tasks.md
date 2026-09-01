# Tasks: Upgrade to Go 1.27

**Input**: Design documents from `/specs/002-go-1-27-upgrade/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Tests**: The generic stream method change requires compile/runtime regression tests before
implementation. Channel/goroutine code requires `go test -race ./...` in Linux CI.

**Organization**: Tasks are grouped by user story so the baseline, typed transforms, and
bounded modernization audit remain independently reviewable.

## Phase 1: Setup (Shared Scope Guards)

**Purpose**: Lock the active feature and protect unrelated or historical work.

- [X] T001 Verify `.specify/feature.json` targets `specs/002-go-1-27-upgrade/`, snapshot the existing `examples/application/main_example.go` diff, and confirm `docs/migration-v0.12.0.md` plus `specs/001-fix-runtime-contracts/` are excluded

---

## Phase 2: Foundational (Constitution Amendment)

**Purpose**: Authorize the new baseline before source code begins using Go 1.27 syntax.

**⚠️ CRITICAL**: Source implementation must not proceed until the explicit amendment and
future planning guidance agree on Go 1.27.

- [X] T002 Amend constitution 1.1.0 to 1.2.0 for Go 1.27 with synchronized impact accounting in `.specify/memory/constitution.md`
- [X] T003 [P] Update future Go compatibility and validation guidance in `.specify/templates/plan-template.md`, `.specify/templates/tasks-template.md`, and `AGENTS.md`

**Checkpoint**: The governing baseline is Go 1.27 and no temporary exception is required.

---

## Phase 3: User Story 1 - Adopt the Go 1.27 Baseline (Priority: P1) 🎯 MVP

**Goal**: Maintainers and consumers have one declared, validated, documented Go 1.27 minimum.

**Independent Test**: Inspect current metadata/policy for Go 1.27, compile every package with
Go 1.27.0, and confirm CI retains minimum/stable build, test, vet, and race jobs.

### Implementation for User Story 1

- [X] T004 [US1] Raise the minimum and CI matrix from Go 1.26 to Go 1.27 in `go.mod` and `.github/workflows/ci.yml`
- [X] T005 [P] [US1] Update the current baseline and v0.13.0 release context in `docs/architecture.md` and `readme.md` without relabeling the existing v0.12.0 section
- [X] T006 [P] [US1] Publish the approved baseline and release migration from `specs/002-go-1-27-upgrade/contracts/migration-v0.13.0.md` to `docs/migration-v0.13.0.md`
- [X] T007 [US1] Run the Go 1.27 compile, build, and vet gates against `go.mod` and verify the two-line gate matrix in `.github/workflows/ci.yml`

**Checkpoint**: Go 1.27 consumers can compile the repository and understand the v0.13.0
minimum-toolchain migration independently of typed stream adoption.

---

## Phase 4: User Story 2 - Preserve Stream Transformation Types (Priority: P2)

**Goal**: Method-style map and flat-map retain their concrete result element type while old
package entry points remain available.

**Independent Test**: Compile direct inferred calls, explicit method values/expressions, and
typed assignments; validate ordered, empty, `any`, source-error, nested-error, `FlatMap`, and
deprecated `FlapMap` runtime cases.

### Tests for User Story 2 (REQUIRED)

- [X] T008 [US2] Add failing compile/runtime regressions for typed Map/FlatMap, explicit method references, `any`, ordering, empty streams, source/nested errors, and both flat-map spellings in `utils/channels/streaming_test.go`

### Implementation for User Story 2

- [X] T009 [US2] Add canonical package `FlatMap` and retain `FlapMap` as a deprecated forwarding alias in `utils/channels/streaming.go`
- [X] T010 [US2] Replace type-erasing `StreamingChan.Map` and `StreamingChan.FlatMap` with Go 1.27 generic methods in `utils/channels/streaming.go`
- [X] T011 [US2] Run focused compile and behavior validation for `utils/channels/streaming.go` and `utils/channels/streaming_test.go`
- [X] T012 [P] [US2] Document typed method signatures, ordering/error invariants, alias status, and method-reference/interface migration in `utils.md` and `docs/migration-v0.13.0.md`
- [X] T013 [US2] Add v0.13.0 typed-stream usage and migration links in `readme.md`

**Checkpoint**: Typed method calls work without `any`, compatibility aliases remain, and all
specified stream behavior passes focused tests.

---

## Phase 5: User Story 3 - Apply Only Relevant Modernization (Priority: P3)

**Goal**: The upgrade contains reviewed Go 1.27 modernization only, with no experimental or
unrelated churn.

**Independent Test**: Run the four new Go 1.27 modernizers in non-mutating mode and verify no
unreviewed patch, external dependency, or package-direction change remains.

### Implementation for User Story 3

- [X] T014 [US3] Run the targeted Go 1.27 modernizer audit and record the final no-diff or reviewed decision in `specs/002-go-1-27-upgrade/research.md` and `specs/002-go-1-27-upgrade/quickstart.md`
- [X] T015 [US3] Verify zero external dependencies and unchanged generic-helper package direction in `go.mod`, `utils/channels/streaming.go`, and `docs/architecture.md`

**Checkpoint**: Every new stable modernizer is accounted for and all unrelated Go 1.27
features remain out of scope.

---

## Phase 6: Polish & Cross-Cutting Validation

**Purpose**: Format, validate, and account for every acceptance gate and protected file.

- [X] T016 Format edited Go files with `gofmt` in `utils/channels/streaming.go` and `utils/channels/streaming_test.go`
- [X] T017 Run `go build ./...`, `go test -count=1 ./...`, and `go vet ./...` against `go.mod`, recording the known Windows-only test blocker separately from passing gates
- [X] T018 Run `go test -race -count=1 ./...` where cgo is available and smoke both `examples/task/run_example.go` and `examples/application/main_example.go` with `EXAMPLE_RUNTIME=1s`, recording Linux CI as authoritative when local race is unavailable
- [X] T019 Run `git diff --check`, verify `examples/application/main_example.go` retains only its pre-existing change, confirm `docs/migration-v0.12.0.md` and `specs/001-fix-runtime-contracts/` are unchanged, and validate all acceptance steps in `specs/002-go-1-27-upgrade/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: Starts immediately and establishes protected paths.
- **Foundational (Phase 2)**: Depends on Setup and blocks Go 1.27 source syntax.
- **US1 (Phase 3)**: Depends on the constitution amendment.
- **US2 (Phase 4)**: Depends on the Go 1.27 module baseline; tests precede implementation.
- **US3 (Phase 5)**: Depends on the final Go 1.27 source and metadata state.
- **Polish (Phase 6)**: Depends on all user stories.

### User Story Dependencies

```text
Setup -> Constitution amendment -> US1 baseline -> US2 typed streams -> US3 audit -> Polish
```

US1 is the MVP. US2 consumes the new baseline. US3 is deliberately last because modernizer
results may depend on the `go 1.27` directive and final syntax.

### Parallel Opportunities

- T003 can update future workflow guidance independently of T002's constitution text.
- T005 and T006 touch separate current documentation after T004 establishes the release.
- T012 can document the frozen US2 contract while focused code validation runs.

## Implementation Strategy

1. Establish and govern the minimum version.
2. Write type/behavior regressions before changing stream methods.
3. Implement only thin generic delegates and the compatibility alias.
4. Rerun the exact Go 1.27 modernizers after the baseline change.
5. Validate locally where supported and keep Windows/full-suite and cgo/race results distinct
   from authoritative Ubuntu CI gates.

## Format Validation

All 19 tasks use the required checkbox, sequential task ID, optional parallel marker, user
story label where applicable, imperative description, and exact repository path format.
