# Implementation Plan: Upgrade to Go 1.27

**Branch**: `002-go-1-27-upgrade` | **Date**: 2026-09-01 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/002-go-1-27-upgrade/spec.md`

## Summary

Raise the module's declared minimum from Go 1.26 to Go 1.27 for a v0.13.0 release, amend
the repository constitution and current compatibility guidance in sync, and use Go 1.27
generic methods to make `StreamingChan.Map` and `StreamingChan.FlatMap` preserve their
result element type. Keep package-level `Map`, add the correctly spelled `FlatMap`, and
retain `FlapMap` as a deprecated forwarding alias. Preserve stream ordering, errors,
backpressure, goroutine ownership, all `ctx` contracts, package direction, and the
standard-library-only dependency policy.

## Technical Context

**Language/Version**: Go 1.27; local minimum validation toolchain is Go 1.27.0

**Primary Dependencies**: Go standard library only; no added module dependency

**Storage**: N/A; all affected values are in-memory typed channels and release metadata

**Testing**: Go `testing`; compile-oriented generic-method tests; focused channel behavior
tests; `go build ./...`, `go test ./...`, `go vet ./...`, and `go test -race ./...`; both
runnable examples as smoke tests

**Target Platform**: Go 1.27-supported platforms; Windows/amd64 local validation and Ubuntu
CI minimum/stable validation

**Project Type**: Reusable Go library with runnable examples

**Performance Goals**: Add no transform goroutine, buffer, scheduling, or traversal beyond
the existing one-producer sequential map/flat-map implementation; preserve outer and inner
value order and backpressure

**Constraints**: Preserve package-level `Map`, `FlapMap` source compatibility, stream
completion/error behavior, service names, `ctx`/`env` tags, configuration precedence,
lifecycle, concurrency ownership, module path, and zero external dependencies. Do not edit
historical v0.12.0 records or the user's existing `examples/application/main_example.go`
change.

**Scale/Scope**: Two generic stream methods, one canonical package helper, one compatibility
alias, one channel test file, module/CI metadata, constitution/templates, four current guides,
one migration guide, and this feature's Spec Kit artifacts

**Public API / Compatibility Impact**: Minimum Go rises from 1.26 to 1.27. Direct method calls
gain result-type inference. Context-free method values/expressions need explicit method type
arguments or a target function type, and generic methods cannot satisfy interfaces containing
the former non-generic method signatures. The release is intentionally v0.13.0 with migration
guidance. Package `Map` remains unchanged; `FlatMap` is additive; `FlapMap` remains deprecated
but functional.

**Reflection / Configuration Impact**: None. No reflection, `unsafe`, tag, naming, injected
type, configuration source, casing, or precedence change.

**Lifecycle / Concurrency Impact**: No lifecycle or ownership change. Generic methods are
thin delegates to existing helpers; mapping remains sequential in one transform producer
goroutine. Channel/goroutine code is touched, so race validation remains mandatory.

**Documentation Impact**: Amend `.specify/memory/constitution.md`; update `AGENTS.md`,
`.specify/templates/plan-template.md`, `.specify/templates/tasks-template.md`, `readme.md`,
`utils.md`, and `docs/architecture.md`; add `docs/migration-v0.13.0.md`. Keep
`docs/migration-v0.12.0.md` and `specs/001-fix-runtime-contracts/` unchanged.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-checked after Phase 1 design.*

- [x] **Stable contracts**: The user's explicit baseline-upgrade approval satisfies the
      constitution's amendment trigger. The breaking minimum-version and generic-method
      surfaces have a v0.13.0 release decision, preserved compatibility entry points, tests,
      and migration guidance.
- [x] **Package direction**: Work remains in `utils/channels` and current metadata/docs;
      generic helpers still do not import `ctx`; no dependency is added.
- [x] **Wiring and configuration**: Service naming, reflection, tags, dependency resolution,
      configuration, and visible wiring failures are explicitly unchanged.
- [x] **Lifecycle and concurrency**: The existing single transform producer owns output
      completion. No new goroutine/channel is introduced, and existing sequential ordering,
      error, and backpressure behavior is regression-tested.
- [x] **Verification**: Type and behavior regressions precede implementation. Build, test,
      vet, race, modernizer-review, and example gates are defined in `quickstart.md`.
- [x] **Documentation**: Exact current policy, consumer, architecture, utility, template,
      automation, and migration files are named; historical records and the unrelated dirty
      example file are excluded.

**Pre-research gate result**: PASS. The baseline increase is an explicitly authorized
constitution amendment, not a temporary exception.

**Post-design gate result**: PASS. `research.md`, `data-model.md`, `contracts/`, and
`quickstart.md` cover the public break, compatibility alias, ownership invariants, exact
file boundaries, migration paths, and required validation without a constitution exception.

## Project Structure

### Documentation (this feature)

```text
specs/002-go-1-27-upgrade/
├── spec.md
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── checklists/requirements.md
├── contracts/public-api.md
├── contracts/migration-v0.13.0.md
└── tasks.md
```

### Source Code and Current Project Documents

```text
go.mod                                      # Go 1.27 minimum
.github/workflows/ci.yml                    # Go 1.27.x and stable gates
.specify/memory/constitution.md             # Explicit 1.2.0 baseline amendment
.specify/templates/plan-template.md         # Future Go 1.27 plans
.specify/templates/tasks-template.md        # Future Go 1.27 tasks
AGENTS.md                                    # Contributor baseline
utils/channels/streaming.go                  # Typed methods and FlatMap alias path
utils/channels/streaming_test.go             # Compile/runtime regression coverage
readme.md                                    # v0.13.0 consumer summary
utils.md                                     # Stream API reference
docs/architecture.md                         # Current baseline and typed helper contract
docs/migration-v0.13.0.md                    # Published migration guide
```

**Structure Decision**: Keep all implementation within the existing `utils/channels`
package and all policy/release changes in existing metadata and documentation locations.
No new Go package, import direction, lifecycle component, or external dependency is needed.

## Phase 0 Research Outcome

[`research.md`](research.md) records the decisions and alternatives. Key outcomes are:

- Go 1.27.0 is a stable release and generic methods are the directly relevant language
  feature.
- Direct calls infer the result type, while context-free method values/expressions and old
  non-generic interface conformance form the documented breaking surface.
- The correctly spelled `FlatMap` becomes canonical and `FlapMap` remains a deprecated alias.
- Go 1.27's four new modernizers currently produce no repository diff; unrelated or
  experimental features remain out of scope.
- No transform concurrency behavior changes; the Linux CI race job remains authoritative
  because the local Windows toolchain has cgo disabled.

## Phase 1 Design Outcome

- [`data-model.md`](data-model.md) defines the baseline, stream transform, alias, and
  modernization-audit contracts and their invariants.
- [`contracts/public-api.md`](contracts/public-api.md) fixes exact public signatures and
  compatibility behavior.
- [`contracts/migration-v0.13.0.md`](contracts/migration-v0.13.0.md) defines consumer source
  and toolchain migration.
- [`quickstart.md`](quickstart.md) provides minimum/stable validation, typed method tests,
  targeted modernizer review, example smokes, dirty-tree protection, and known local blockers.

## Complexity Tracking

No constitution violation or complexity exception is required.
