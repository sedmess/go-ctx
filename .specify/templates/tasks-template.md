---

description: "Task list template for feature implementation"
---

# Tasks: [FEATURE NAME]

**Input**: Design documents from `/specs/[###-feature-name]/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Behavior-changing work MUST include regression tests. Add test tasks before the
corresponding implementation tasks. Documentation-only work may omit tests with an explicit
rationale. Lifecycle or concurrency work MUST include a `go test -race ./...` validation task.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Core container**: `ctx/`, with tests beside code as `ctx/*_test.go`
- **Focused context packages**: `ctx/appinfo/`, `ctx/autoctx/`,
  `ctx/ctx_testing/`, `ctx/health/`, and `ctx/logger/`
- **Generic helpers**: `u/`, `it/`, and `utils/`; these must not import `ctx`
- **Runnable examples**: `examples/application/` and `examples/task/`
- **Architecture and consumer docs**: `docs/`, `readme.md`, and `utils.md`
- Every generated task must use the real Go package path selected in plan.md, never a
  generic placeholder path from another project type.

<!--
  ============================================================================
  IMPORTANT: The tasks below are SAMPLE TASKS for illustration purposes only.

  The /speckit-tasks command MUST replace these with actual tasks based on:
  - User stories from spec.md (with their priorities P1, P2, P3...)
  - Feature requirements from plan.md
  - Entities from data-model.md
  - Public library contracts from contracts/
  - Constitution-required regression, lifecycle, concurrency, and documentation work

  Tasks MUST be organized by user story so each story can be:
  - Implemented independently
  - Tested independently
  - Delivered as an MVP increment

  DO NOT keep these sample tasks in the generated tasks.md file.
  ============================================================================
-->

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish the planned Go package and validation structure

- [ ] T001 Create the planned package and file structure at [exact Go paths]
- [ ] T002 Confirm Go 1.27 compatibility and justify dependencies in go.mod
- [ ] T003 [P] Create colocated test scaffolding in [package]/[feature]_test.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

Examples of foundational tasks (adjust based on the feature):

- [ ] T004 Define shared public contracts in ctx/[feature]_types.go
- [ ] T005 [P] Implement shared wiring or configuration parsing in ctx/[feature]_config.go
- [ ] T006 [P] Implement cancellation and resource ownership in ctx/[feature]_lifecycle.go
- [ ] T007 Integrate shared behavior with the application context in ctx/application_context.go
- [ ] T008 Add shared failure logging, health, or stats behavior in ctx/[feature].go
- [ ] T009 Document foundational architecture decisions in docs/architecture.md

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - [Title] (Priority: P1) 🎯 MVP

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 1 (REQUIRED for behavior changes)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T010 [P] [US1] Add public contract regression tests in ctx/[feature]_contract_test.go
- [ ] T011 [P] [US1] Add lifecycle and error-path tests in ctx/[feature]_lifecycle_test.go

### Implementation for User Story 1

- [ ] T012 [P] [US1] Define public feature types in ctx/[feature]_types.go
- [ ] T013 [P] [US1] Implement isolated feature helpers in ctx/[feature]_helpers.go
- [ ] T014 [US1] Implement the feature behavior in ctx/[feature].go (depends on T012, T013)
- [ ] T015 [US1] Integrate the feature with ctx/application_context.go
- [ ] T016 [US1] Add validation and visible failure handling in ctx/[feature].go
- [ ] T017 [US1] Add logging, health, or stats signals in ctx/[feature].go

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently

---

## Phase 4: User Story 2 - [Title] (Priority: P2)

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 2 (REQUIRED for behavior changes)

- [ ] T018 [P] [US2] Add public contract regression tests in [package]/[feature]_contract_test.go
- [ ] T019 [P] [US2] Add lifecycle and error-path tests in [package]/[feature]_lifecycle_test.go

### Implementation for User Story 2

- [ ] T020 [P] [US2] Define story-specific types in [package]/[feature]_types.go
- [ ] T021 [US2] Implement story behavior in [package]/[feature].go
- [ ] T022 [US2] Integrate story behavior through the affected public API in [exact path]
- [ ] T023 [US2] Integrate with User Story 1 without changing its independent behavior

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 5: User Story 3 - [Title] (Priority: P3)

**Goal**: [Brief description of what this story delivers]

**Independent Test**: [How to verify this story works on its own]

### Tests for User Story 3 (REQUIRED for behavior changes)

- [ ] T024 [P] [US3] Add public contract regression tests in [package]/[feature]_contract_test.go
- [ ] T025 [P] [US3] Add lifecycle and error-path tests in [package]/[feature]_lifecycle_test.go

### Implementation for User Story 3

- [ ] T026 [P] [US3] Define story-specific types in [package]/[feature]_types.go
- [ ] T027 [US3] Implement story behavior in [package]/[feature].go
- [ ] T028 [US3] Integrate story behavior through the affected public API in [exact path]

**Checkpoint**: All user stories should now be independently functional

---

[Add more user story phases as needed, following the same pattern]

---

## Phase N: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] TXXX [P] Update consumer guidance in readme.md and docs/architecture.md
- [ ] TXXX [P] Update runnable usage in examples/task/run_example.go or examples/application/main_example.go
- [ ] TXXX Format all edited Go files with gofmt
- [ ] TXXX Run go build ./..., go test ./..., and go vet ./...
- [ ] TXXX Run go test -race ./... for lifecycle or concurrency changes
- [ ] TXXX Validate the feature quickstart and relevant runnable example

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can then proceed in parallel (if staffed)
  - Or sequentially in priority order (P1 → P2 → P3)
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - May integrate with US1 but should be independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - May integrate with US1/US2 but should be independently testable

### Within Each User Story

- Regression tests for behavior changes MUST be written before implementation and must
  demonstrate the missing or incorrect behavior
- Public types and interfaces before their implementations
- Isolated helpers before container integration
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (within Phase 2)
- Once Foundational phase completes, all user stories can start in parallel (if team capacity allows)
- All tests for a user story marked [P] can run in parallel
- Independent Go files within a story marked [P] can run in parallel
- Different user stories can be worked on in parallel by different team members

---

## Parallel Example: User Story 1

```text
# Launch independent test files for User Story 1 together:
Task: Add public contract regression tests in ctx/[feature]_contract_test.go
Task: Add lifecycle and error-path tests in ctx/[feature]_lifecycle_test.go

# Launch independent implementation files for User Story 1 together:
Task: Define public feature types in ctx/[feature]_types.go
Task: Implement isolated feature helpers in ctx/[feature]_helpers.go
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1
4. **STOP and VALIDATE**: Test User Story 1 independently
5. Validate the public API through the quickstart or runnable example

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Validate example (MVP!)
3. Add User Story 2 → Test independently → Validate example
4. Add User Story 3 → Test independently → Validate example
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: User Story 1
   - Developer B: User Story 2
   - Developer C: User Story 3
3. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify new regression tests expose the missing or incorrect behavior before implementing
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
