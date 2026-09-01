# Quickstart Validation: Upgrade to Go 1.27

Use this guide after implementation to validate the contracts in `contracts/`.

## Prerequisites

- Repository root: `github.com/sedmess/go-ctx`
- Go 1.27.x and the current stable toolchain (CI may provide the second line)
- An intentionally reviewed working tree
- No new module dependency

On Windows, use isolated writable `GOCACHE` and `GOTMPDIR` directories if the default build
cache is unavailable.

## 1. Confirm the declared baseline and current policy

```text
go version
go env GOVERSION
```

Expected outcome:

- The minimum validation toolchain is Go 1.27.x.
- `go.mod`, CI, constitution, current architecture, AGENTS, README, Spec Kit templates, and
  v0.13.0 migration guidance name Go 1.27.
- `docs/migration-v0.12.0.md` and `specs/001-fix-runtime-contracts/` still describe Go 1.26.

## 2. Compile and validate typed transforms

```text
go test -run '^$' ./...
go test -count=1 ./utils/channels
```

Expected outcome:

- Direct Map/FlatMap calls infer concrete result stream types.
- Explicit method values/expressions compile.
- Concrete, `any`, ordered, empty, source-error, nested-error, `FlatMap`, and deprecated
  `FlapMap` cases pass.

## 3. Review only the new Go 1.27 modernizers

```text
go fix -diff -atomictypes -embedlit -slicesbackward -unsafefuncs ./...
```

Validated 2026-09-01 with Go 1.27.0: exit success with no output. Any future output must be
reviewed and either implemented with validation or documented before completion.

## 4. Run repository gates

```text
go build ./...
go test -count=1 ./...
go vet ./...
go test -race -count=1 ./...
```

Expected outcome in Ubuntu CI: every command passes on Go 1.27.x and stable.

Known local constraints at planning time:

- The Windows full suite has a pre-existing failure in
  `TestConfigurationCasingAndPrecedence` because Windows environment keys are
  case-insensitive and the test expects distinct lowercase/uppercase keys.
- The local Windows race gate is unavailable when `CGO_ENABLED=0`; Linux CI is authoritative.

Report these gates separately; do not claim local success when they remain blocked.

## 5. Smoke-test public examples

```text
go run ./examples/task
$env:EXAMPLE_RUNTIME='1s'
go run ./examples/application
```

Expected outcome: both examples start and finish successfully. The environment override
keeps validation short without modifying the user's existing 120-second default edit.

## 6. Verify scope and formatting

```text
gofmt -w utils/channels/streaming.go utils/channels/streaming_test.go
git diff --check
git diff -- examples/application/main_example.go
```

Expected outcome:

- No formatting errors.
- The example diff contains only the pre-existing one-second to 120-second user change.
- No files under `specs/001-fix-runtime-contracts/` or `docs/migration-v0.12.0.md` changed.
- No external dependency or `go.sum` was introduced.

## Validation Record: 2026-09-01

| Gate | Result | Evidence |
|------|--------|----------|
| Go toolchain | PASS | `go1.27.0 windows/amd64` |
| Compile-only | PASS | `go test -run '^$' ./...` |
| Build | PASS | `go build ./...` |
| Focused channels | PASS | `go test -count=1 ./utils/channels` |
| Vet | PASS | `go vet ./...` |
| Full test | BLOCKED (pre-existing) | Windows-only `TestConfigurationCasingAndPrecedence`: exact process value resolves to `canonical-process` because environment keys are case-insensitive |
| Race | BLOCKED locally | `CGO_ENABLED=0`; `go test -race` requires cgo; Ubuntu CI remains authoritative |
| Task example | PASS | `go run ./examples/task` |
| Application example | PASS | `EXAMPLE_RUNTIME=1s go run ./examples/application` |
| Go 1.27 modernizers | PASS | Combined targeted `go fix -diff` returned no output |
| Diff hygiene | PASS | `git diff --check`; protected historical paths unchanged; example retains only its pre-existing runtime edit |

CI was configured for Go 1.27.x and stable with compile, build, full test, vet, and race
steps, but CI execution is not available from this local workflow. Do not treat the local
Windows full-test or race blockers as confirmed CI success.
