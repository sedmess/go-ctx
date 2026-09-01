# Feature Specification: Upgrade to Go 1.27

**Feature Branch**: `002-go-1-27-upgrade`

**Created**: 2026-09-01

**Status**: Draft

**Input**: User description: "Upgrade to Go 1.27 and rewrite selected code to use new Go 1.27 features through the full Spec Kit flow."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Adopt the Go 1.27 Baseline (Priority: P1)

As a maintainer, I can build, test, analyze, and release the module against Go 1.27 as the
declared minimum so contributors and consumers have one accurate compatibility contract.

**Why this priority**: The baseline declaration governs every later source rewrite and is a
breaking consumer requirement that must be consistent across project policy, automation,
architecture, and migration guidance.

**Independent Test**: Validate a clean checkout with an actual Go 1.27 toolchain and inspect
all current compatibility documents and automation to confirm they consistently name Go
1.27 as the minimum supported version.

**Acceptance Scenarios**:

1. **Given** a consumer or contributor using Go 1.27, **When** they build and validate the
   module, **Then** every package and runnable example compiles and the required validation
   gates pass.
2. **Given** a consumer using an older Go version, **When** they review the release migration
   guidance before upgrading, **Then** the new minimum and release impact are stated
   explicitly.
3. **Given** project policy, module metadata, automation, and architecture guidance, **When**
   a maintainer compares their compatibility statements, **Then** all current statements
   identify the same Go 1.27 minimum without rewriting historical release records.

---

### User Story 2 - Preserve Stream Transformation Types (Priority: P2)

As a library consumer, I can transform a typed stream with method syntax and receive a
stream of the mapper's result type without losing that type to `any`.

**Why this priority**: Go 1.27 generic methods remove the language limitation that forced
the current type-erasing method workaround, improving compile-time safety while keeping the
existing package boundary and stream behavior.

**Independent Test**: Map and flat-map integer streams to string streams using method syntax,
assign the results directly to typed stream variables, collect the outputs, and verify their
types, order, values, and error propagation at compile time and runtime.

**Acceptance Scenarios**:

1. **Given** a typed stream and a mapper returning a different type, **When** a consumer calls
   the stream's map method, **Then** the returned stream retains the mapper result type and
   yields values in source order.
2. **Given** a typed stream and a mapper returning typed substreams, **When** a consumer calls
   the stream's flat-map method, **Then** the returned stream retains the substream element
   type and yields flattened values in deterministic source order.
3. **Given** a consumer using the existing misspelled package-level flat-map helper, **When**
   they adopt the new release, **Then** their source continues to compile while a correctly
   spelled alternative is available and documented.
4. **Given** a source stream that reports an error, **When** it is transformed, **Then** the
   error remains observable under the existing stream contract.

---

### User Story 3 - Apply Only Relevant Modernization (Priority: P3)

As a maintainer, I can review Go 1.27's stable modernization opportunities and apply only
changes that improve this module without adding experimental dependencies, unrelated APIs,
or behavior churn.

**Why this priority**: A focused upgrade is easier to review and safer for consumers than a
broad rewrite whose changes are unrelated to the module's responsibilities.

**Independent Test**: Run the Go 1.27 modernizers in review mode and confirm every proposed
change is either implemented with validation or recorded as inapplicable; confirm the module
still has no non-standard-library dependency.

**Acceptance Scenarios**:

1. **Given** the Go 1.27 modernizers, **When** they are run against the repository, **Then**
   every applicable change is reviewed rather than applied blindly.
2. **Given** an experimental or domain-irrelevant Go 1.27 feature, **When** maintainers assess
   it, **Then** it remains out of scope unless it directly serves an existing contract.
3. **Given** the completed upgrade, **When** consumers inspect release and utility guidance,
   **Then** they can identify the baseline change, typed stream transformations, compatibility
   alias, and validation expectations.

### Edge Cases

- A mapper returns `any`; the generic method must still infer and return `StreamingChan[any]`.
- A source or mapped substream is empty; transformation completes without inventing values.
- A source or mapped substream reports an error; transformation stops and exposes the error
  under the existing contract.
- A consumer references `FlapMap`; the deprecated alias remains functional and forwards to
  the correctly spelled helper.
- A consumer stores a stream method as a method value or expression; migration guidance
  identifies any source adjustment required by the generic method signature.
- A consumer declares an interface containing the old non-generic map or flat-map method;
  migration guidance identifies that generic concrete methods cannot satisfy that interface.
- Go 1.27 modernizers produce no relevant diff; the upgrade records that result and does not
  introduce cosmetic rewrites.
- The existing user edit in `examples/application/main_example.go` remains unchanged by this
  feature.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The module MUST declare Go 1.27 as its minimum supported language and toolchain
  baseline.
- **FR-002**: Project policy, current architecture guidance, contributor guidance, release
  documentation, and continuous validation MUST consistently reflect the Go 1.27 baseline.
- **FR-003**: Continuous validation MUST use an actual Go 1.27 toolchain and a current stable
  toolchain rather than treating a metadata declaration as sufficient proof.
- **FR-004**: The baseline increase MUST be documented as an intentional compatibility change
  for a v0.13.0 release while retaining the existing module path and standard-library-only
  dependency policy.
- **FR-005**: `StreamingChan.Map` MUST use a method-specific result type and return a stream of
  that type without requiring conversion through `any`.
- **FR-006**: `StreamingChan.FlatMap` MUST use a method-specific result type and return a
  flattened stream of that type without requiring conversion through `any`.
- **FR-007**: The channels package MUST expose a correctly spelled package-level `FlatMap`
  helper and MUST retain `FlapMap` as a deprecated compatibility alias with equivalent
  behavior.
- **FR-008**: Existing package-level `Map` behavior and existing stream ordering, completion,
  error propagation, and backpressure behavior MUST remain unchanged.
- **FR-009**: Type-preserving map and flat-map behavior MUST have compile-time and runtime
  regression coverage, including empty streams and source or nested-stream errors.
- **FR-010**: New Go 1.27 modernizers MUST be run in non-mutating review mode, and only
  applicable behavior-preserving suggestions MAY be implemented.
- **FR-011**: Experimental Go 1.27 APIs and unrelated standard-library additions MUST remain
  out of scope.
- **FR-012**: The upgrade MUST add no external dependency and MUST preserve package dependency
  direction.
- **FR-013**: Service naming, `ctx` and `env` tag grammar, injected types, configuration
  precedence, application lifecycle, concurrency ownership, and failure behavior MUST remain
  unchanged.
- **FR-014**: Consumer and utility documentation MUST explain typed stream method results, the
  compatibility alias, and any generic-method migration edge cases.
- **FR-015**: Historical v0.12.0 and feature-001 artifacts MUST remain accurate records of the
  Go 1.26 release and MUST NOT be relabeled as current Go 1.27 work.
- **FR-016**: The pre-existing edit in `examples/application/main_example.go` MUST be preserved
  and excluded from this feature's implementation changes.

### Contract and Lifecycle Impact *(mandatory)*

- **Public API compatibility**: The minimum toolchain increases from Go 1.26 to Go 1.27.
  `StreamingChan.Map` and `StreamingChan.FlatMap` become generic methods that retain result
  types; ordinary direct calls remain natural, while exact copied signatures, stored method
  values, method expressions, or interfaces containing the old methods may require migration.
  `FlatMap` is additive and `FlapMap` remains as a deprecated alias. The release target is
  v0.13.0.
- **Service wiring and configuration**: None. Service names, `ctx`/`env` tags, injected types,
  dependency resolution, and configuration precedence are unchanged.
- **Lifecycle and concurrency**: No ownership or lifecycle contract changes are intended.
  Stream transformation continues to use the existing producer-owned goroutine and channel
  termination behavior, so race validation remains required.
- **Failure and observability**: Existing stream errors remain observable; no logging,
  health, statistics, initialization, or callback-failure behavior changes.
- **Documentation**: Update the constitution, contributor guide, architecture guide, README,
  utility guide, CI baseline, planning templates, and a new v0.13.0 migration guide.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Build, test, vet, and race validation pass with Go 1.27.0 or later in the Go
  1.27 release line.
- **SC-002**: CI defines both a Go 1.27.x minimum-toolchain job and a stable-toolchain job,
  with 100% of required repository gates present in each job.
- **SC-003**: Map and flat-map method results assign directly to differently typed stream
  variables in regression tests with zero type assertions or `any` conversions.
- **SC-004**: All map, flat-map, empty-stream, source-error, and nested-error regression
  scenarios preserve documented ordering and error behavior.
- **SC-005**: All current compatibility documents and module metadata agree on Go 1.27 and
  v0.13.0, while all historical v0.12.0 records remain unchanged.
- **SC-006**: Go 1.27 modernizer review completes with every proposed repository change either
  implemented and validated or explicitly classified as inapplicable.
- **SC-007**: The final module dependency graph contains zero new external dependencies and
  generic helper packages retain zero imports of `ctx`.
- **SC-008**: Migration guidance accounts for 100% of intentional compatibility changes:
  minimum toolchain, generic stream method signatures, and the `FlatMap`/`FlapMap` naming path.

## Assumptions

- The baseline increase is explicitly approved by the user's request and therefore authorizes
  the synchronized constitution amendment required by project governance.
- v0.13.0 is the intended pre-v1 semantic-version release because the minimum toolchain and
  generic method signatures change consumer compatibility after v0.12.0.
- The selected code rewrite is intentionally limited to typed stream transformation methods;
  redesigning the `it.It` interface is out of scope because interfaces cannot declare generic
  methods and changing it to a concrete type would create unrelated compatibility risk.
- The new `atomictypes`, `embedlit`, `slicesbackward`, and `unsafefuncs` modernizers have been
  reviewed in non-mutating mode and currently propose no changes in this repository.
- Go 1.27's experimental SIMD APIs, JSON v2 packages, UUID package, cryptographic additions,
  and unrelated networking changes do not serve an existing go-ctx contract and are out of
  scope.
