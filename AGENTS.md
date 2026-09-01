# Repository Guidelines

## Constitutional Authority

The project constitution at `.specify/memory/constitution.md` is the highest-authority
engineering document in this repository and is non-negotiable during ordinary feature work.
Read it together with `docs/architecture.md` before changing public APIs, package boundaries,
dependency injection, configuration, lifecycle, or concurrency behavior. If a specification,
plan, task list, implementation, or this guide conflicts with the constitution, correct the
lower-authority artifact before proceeding. Constitution changes require a separate explicit
amendment; temporary implementation exceptions must be justified in the plan's Complexity
Tracking section and may not silently waive a principle.

## Project Structure & Module Organization

This repository is the Go module `github.com/sedmess/go-ctx`, with its declared language
baseline in `go.mod`. Core application context, dependency injection, lifecycle,
configuration, logging, and health behavior lives in `ctx/`; focused subpackages include
`ctx/appinfo`, `ctx/autoctx`, `ctx/ctx_testing`, `ctx/health`, and `ctx/logger`. General
helpers are grouped under `u/` and `utils/`, while `it/` contains iterator support. Runnable
demonstrations live in `examples/application/` and `examples/task/`. Tests stay beside the
package they cover as `*_test.go` files.

Keep package dependencies acyclic: `ctx` owns orchestration, focused subpackages keep narrow
responsibilities, and generic helpers in `u/`, `it/`, and `utils/` must not depend on `ctx`.
Examples must consume public APIs rather than internal implementation details. Prefer the Go
standard library; every new package or external dependency needs a specific, documented
reason in the implementation plan.

## Change Planning Requirements

Before implementation, feature work must identify impacts to exported APIs, service naming,
`ctx` and `env` tags, configuration precedence, failure behavior, lifecycle, concurrency,
Go compatibility, and consumer documentation. Implementation plans must name real packages
and file paths, justify every dependency, and pass the Constitution Check before research and
again after design. If no Spec Kit artifacts are in scope, record the same analysis in the
task handoff or pull request.

## Build, Test, and Development Commands

- `go build ./...` compiles every library and example package.
- `go test ./...` runs the full test suite; use `go test ./ctx/...` for a focused core pass.
- `go test -race ./...` checks concurrency-sensitive lifecycle and channel code.
- `go vet ./...` performs standard Go static analysis.
- `go run ./examples/task` runs the short task example.
- `go run ./examples/application` runs the longer lifecycle demonstration.

The module currently declares Go 1.27. Keep changes compatible with that baseline unless the
version update is intentional, approved, and documented. Before completing work, run
`go build ./...`, `go test ./...`, and `go vet ./...`; also run `go test -race ./...` for
changes affecting lifecycle, goroutines, channels, timers, synchronization, signals, or
application state. Report any pre-existing blocker rather than silently skipping a gate.

## Coding Style & Naming Conventions

Run `gofmt -w path/to/file.go` on edited files; use Go's tabs and standard import grouping.
Follow established lowercase package names and lowercase snake-case filenames such as
`application_context.go`. Use PascalCase for exported identifiers and camelCase for internal
ones. Treat exported identifiers, service naming, `ctx` and `env` tag grammar, injected
types, configuration precedence, lifecycle callbacks, and observable failure behavior as
public contracts. Changes are additive by default; breaking changes require explicit scope,
migration guidance, updated examples, and a semantic-version release decision. Reflection
or `unsafe` changes require focused compatibility tests. Add comments where exported behavior
or non-obvious lifecycle rules need explanation.

## Lifecycle and Concurrency Rules

For every lifecycle or concurrency change, define startup, steady-state, shutdown,
cancellation, callback failure, initialization failure, restart, and repeated-stop behavior.
Every goroutine, channel, ticker, signal subscription, and blocking operation must have an
identifiable owner and termination mechanism. Synchronize shared mutable state and never rely
on map iteration or unspecified concurrent callback order. Duplicate or reserved service
names, missing dependencies, cycles, invalid tags, unsupported configuration types, and
malformed values must fail visibly rather than being ignored.

## Testing Guidelines

Tests use the standard `testing` package. Name tests `TestXxx`, colocate them with the
implementation, and use package-level `TestMain` only for shared setup such as logging
configuration. Public behavior changes must include meaningful regression coverage;
documentation-only changes may omit tests when the rationale is explicit. Exercise every
affected startup, steady-state, shutdown, initialization-failure, callback-failure,
cancellation, restart, and concurrency path. Assert observable behavior rather than private
implementation structure. No formal coverage threshold is configured; prioritize meaningful
behavioral assertions and run the race detector for concurrent changes.

## Commit & Pull Request Guidelines

Recent commits use short, action-led subjects such as `add slice.Filter`, `fix main_example`,
and `remove deprecated ...`. Keep each commit focused and use a concise imperative subject.
Pull requests must explain the behavior changed, identify affected packages and public
contracts, record constitution compliance or justified complexity, link relevant issues, and
list validation commands run. Update `readme.md`, `docs/architecture.md`, package comments,
or runnable examples whenever user-facing APIs, tags, configuration, lifecycle, or package
boundaries change.

## Configuration Safety

Treat the tracked `.env` and `.env_custom` files as non-secret fixtures. Never commit
credentials or machine-specific values; provide local overrides through the process
environment or ignored local configuration.
