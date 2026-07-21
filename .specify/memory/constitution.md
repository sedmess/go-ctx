<!--
Sync Impact Report
- Version change: 1.0.0 -> 1.1.0
- Modified principles:
  - I. Stable Library Contracts: raised the supported Go baseline from 1.21 to 1.26 and
    made future baseline increases explicit breaking changes with synchronized release
    documentation requirements
- Modified sections:
  - Architecture and Technical Constraints: production compatibility now targets Go 1.26
  - Development Workflow and Quality Gates: validation must use the declared Go baseline
- Added sections: none
- Removed sections: none
- Templates requiring updates:
  - ✅ updated: .specify/templates/plan-template.md
  - ✅ updated: .specify/templates/tasks-template.md
  - ✅ reviewed: .specify/templates/spec-template.md; no Go baseline guidance present
- Runtime guidance:
  - ✅ updated: AGENTS.md, readme.md, docs/architecture.md, docs/migration-v0.12.0.md
  - ✅ updated: go.mod and .github/workflows/ci.yml
  - ✅ updated: specs/001-fix-runtime-contracts planning, contracts, tasks, and quickstart
  - ✅ updated: generic reflection helpers to use the Go 1.26-supported reflect.TypeFor API
- Command guidance:
  - ✅ reviewed: all .agents/skills/speckit-*/SKILL.md files; no Go baseline guidance present
- Follow-up TODOs: none
-->
# go-ctx Constitution

## Core Principles

### I. Stable Library Contracts

`go-ctx` MUST remain a reusable Go library whose exported identifiers, service naming,
reflection tags, configuration precedence, lifecycle callbacks, and observable failure
behavior are treated as public contracts. Changes MUST be additive by default. A breaking
change requires explicit scope, migration guidance, updated examples, and an intentional
semantic-version release decision. The Go 1.26 language and toolchain baseline MUST remain
supported until a future baseline increase is explicitly approved and documented. Any
future baseline increase MUST be treated as a breaking compatibility change and synchronized
across module metadata, CI, consumer migration guidance, and the release decision.

Rationale: consumers compile the module into their own processes, so a small behavioral
change can break applications without any repository-local signal.

### II. Explicit Package Boundaries

Core dependency injection, configuration, and lifecycle orchestration MUST remain in
`ctx`. Focused `ctx` subpackages MAY provide logging, health, application metadata,
automatic registration, and testing support without creating import cycles. Generic
helpers in `u`, `it`, and `utils` MUST NOT depend on the application container. Examples
MUST consume public APIs instead of internal implementation details. New packages and
external dependencies MUST have a specific responsibility and a documented reason.

Rationale: acyclic, purpose-driven packages keep the module small, reusable, and easy to
adopt without forcing unrelated dependencies on consumers.

### III. Wiring and Configuration Are Contracts

Service registration MUST reject duplicate or reserved names and MUST resolve dependencies
consistently by explicit name or reflected type. Changes to `ctx` and `env` struct-tag
grammar, supported injected types, service-name derivation, environment-source precedence,
or dependency-cycle handling MUST be documented and covered by regression tests. Services
MUST NOT rely on map iteration or concurrent callback order. Invalid wiring and invalid
configuration MUST fail visibly; they MUST NOT be silently ignored.

Rationale: reflection removes compile-time visibility, so strict runtime validation and
compatibility tests are the safety boundary.

### IV. Lifecycle and Concurrency Safety

Every lifecycle or concurrency change MUST define startup, steady-state, shutdown,
cancellation, error, and restart behavior. Every goroutine, channel, ticker, signal
subscription, and blocking operation MUST have an identifiable owner and termination path.
Shared mutable state MUST use appropriate synchronization. Service cleanup MUST remain
observable, every container-owned cleanup path MUST have a termination mechanism, and code
MUST NOT depend on unspecified ordering among concurrent callbacks. Concurrency-sensitive
changes MUST pass the race detector.

Rationale: the container owns process-level resources; leaks, deadlocks, and ordering
assumptions can prevent an embedding application from starting or stopping safely.

### V. Behavioral Verification and Documentation

Public behavior changes MUST include meaningful colocated tests. Lifecycle-related work
MUST exercise relevant startup, shutdown, initialization failure, callback failure,
cancellation, restart, and concurrency paths. Tests MUST assert behavior rather than
implementation structure. User-facing API, tag, configuration, or lifecycle changes MUST
update `readme.md`, `docs/architecture.md`, package comments, or runnable examples as
appropriate. Logs, health, and context statistics MUST remain usable for diagnosing the
container without exposing secrets.

Rationale: executable tests protect behavior while concise documentation lets consumers
use reflection-heavy APIs without reading implementation code.

## Architecture and Technical Constraints

- The module path MUST remain `github.com/sedmess/go-ctx` unless a separately approved
  migration changes the public import path.
- Production code MUST remain compatible with the declared Go 1.26 baseline and SHOULD
  prefer the standard library. Any external dependency requires a plan-level justification.
- `ctx` is the orchestration boundary. Its focused subpackages and generic helper packages
  MUST follow the dependency direction documented in `docs/architecture.md`.
- The supported runtime is an in-process service container with a process-wide active
  application context. Proposals for multiple simultaneous contexts MUST define isolation,
  global-access behavior, logging behavior, and compatibility before implementation.
- Configuration MUST remain local and deterministic. Tracked `.env` fixtures MUST contain
  no secrets; machine-specific or credential-bearing values belong in process environment
  overrides or ignored local files.
- Reflection or `unsafe` changes require focused compatibility tests because they can bypass
  ordinary compile-time guarantees.

## Development Workflow and Quality Gates

1. A feature specification MUST identify affected public contracts, lifecycle behavior,
   configuration or tag semantics, concurrency risks, and user-facing documentation.
2. An implementation plan MUST name real package and file boundaries, justify added
   dependencies, and pass every Constitution Check before implementation begins.
3. Edited Go files MUST be formatted with `gofmt`. The repository MUST pass
   `go build ./...`, `go test ./...`, and `go vet ./...` using the declared Go baseline
   before completion.
4. Work affecting goroutines, channels, synchronization, application state, timers, or
   lifecycle callbacks MUST also pass `go test -race ./...`.
5. Regression tests MUST accompany behavior changes unless the change is documentation-only;
   any exception MUST be recorded in the plan or pull request with a concrete rationale.
6. Reviews MUST verify package direction, public compatibility, cleanup ownership, failure
   visibility, and documentation impact. Unjustified complexity is a gate failure.

## Governance

This constitution is the highest-authority engineering document in the repository. When a
specification, plan, task list, runtime guide, or local convention conflicts with it, that
artifact MUST be corrected before implementation proceeds.

Amendments require an explicit documentation change that explains the motivation, migration
impact, affected principles, and synchronized templates or guides. Constitution versions use
semantic versioning: MAJOR for incompatible principle removals or redefinitions, MINOR for
new principles or materially expanded obligations, and PATCH for non-semantic clarification.

Every Spec Kit plan MUST perform the Constitution Check before research and again after
design. Every implementation review MUST cite the validation commands run and account for
all applicable MUST rules. A temporary exception requires written justification in the
plan's Complexity Tracking section; an exception cannot waive a principle silently.
`docs/architecture.md` is the implementation-oriented architecture reference and MUST remain
consistent with this constitution.

**Version**: 1.1.0 | **Ratified**: 2026-07-19 | **Last Amended**: 2026-07-21
