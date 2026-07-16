# Story 8.1: Upgrade Go from 1.22 to 1.26

Status: done

## Story

As an operator developer,
I want to upgrade the Go version from 1.22 to 1.26,
So that we can use the latest controller-runtime (v0.24, which requires Go 1.26) and benefit from Go language improvements, the Green Tea GC, and security fixes.

## Acceptance Criteria

1. **Given** the project uses Go 1.22, **When** `go.mod` is updated to `go 1.26` and all source files are adapted, **Then** `go build ./...` succeeds, `go vet ./...` passes, `go test ./...` passes.

2. **Given** the Dockerfile builder stage references `golang:1.22`, **When** the base image is updated to `golang:1.26`, **Then** the container builds and the operator binary runs correctly.

3. **Given** CI workflows (pr.yaml, push.yaml) reference `GO_VERSION: ~1.22`, **When** both are updated to `GO_VERSION: ~1.26`, **Then** all CI jobs pass.

4. **Given** the Dockerfile hardcodes `GOARCH=amd64`, **When** the build is updated to use `TARGETARCH` build arg for multi-arch support, **Then** the image can be built for both amd64 and arm64.

## Tasks / Subtasks

- [x] Task 1: Update `go.mod` (AC: #1)
  - [x] 1.1 Change `go 1.22.0` to `go 1.26`
  - [x] 1.2 Add `toolchain go1.26.4` directive (current latest patch)
  - [x] 1.3 Run `go mod tidy` to update `go.sum` and resolve any dependency conflicts
  - [x] 1.4 Verify `go build ./...` succeeds
  - [x] 1.5 Verify `go vet ./...` passes
  - [x] 1.6 Run `go test ./...` (unit tests via envtest — may need ENVTEST fix from Story 8.3, but should work with current envtest)
- [x] Task 2: Update Dockerfile for Go 1.26 + multi-arch (AC: #2, #4)
  - [x] 2.1 Change `FROM golang:1.22 AS builder` to `FROM --platform=$BUILDPLATFORM golang:1.26 AS builder`
  - [x] 2.2 Add `ARG TARGETOS` and `ARG TARGETARCH` declarations
  - [x] 2.3 Replace `RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go` with `RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o manager main.go`
  - [x] 2.4 Verify `podman build .` or `docker build .` succeeds
- [x] Task 3: Update CI workflows (AC: #3)
  - [x] 3.1 In `.github/workflows/pr.yaml`: change `GO_VERSION: ~1.22` to `GO_VERSION: ~1.26`
  - [x] 3.2 In `.github/workflows/push.yaml`: change `GO_VERSION: ~1.22` to `GO_VERSION: ~1.26`
- [x] Task 4: Update project-context.md (AC: #1)
  - [x] 4.1 Update `_bmad-output/project-context.md` to reflect Go 1.26 (Language, Container image, and any related references)
- [x] Task 5: Verify compilation and tests (AC: #1)
  - [x] 5.1 Run `make fmt && make vet`
  - [x] 5.2 Run `make test` (unit tests)
  - [x] 5.3 Verify no new lint warnings from `go vet`

## Dev Notes

### Scope Boundary

This story covers ONLY the Go version bump and Dockerfile multi-arch change. It does NOT upgrade controller-runtime, K8s libs, controller-gen, envtest, Kind, or kubectl — those are Stories 8.2, 8.3, and 8.4. The existing controller-runtime v0.17.3 and K8s libs v0.29.2 must continue to compile and work with Go 1.26.

### Files to Modify

| File | Change |
|------|--------|
| `go.mod` | `go 1.22.0` → `go 1.26`, add `toolchain go1.26.4` |
| `go.sum` | No change (go mod tidy produced no hash updates) |
| `Dockerfile` | Go 1.26 base image + multi-arch `TARGETARCH` pattern |
| `.github/workflows/pr.yaml` | `GO_VERSION: ~1.22` → `GO_VERSION: ~1.26` |
| `.github/workflows/push.yaml` | `GO_VERSION: ~1.22` → `GO_VERSION: ~1.26` |
| `_bmad-output/project-context.md` | Go version references |

Files NOT modified in this story (but referenced for awareness):
- `ci.Dockerfile` — does not reference Go (copies pre-built binary); no change needed
- `bundle.Dockerfile` — does not reference Go; no change needed
- `Makefile` — Go version is not set there (ENVTEST_VERSION, CONTROLLER_TOOLS_VERSION etc. are Story 8.3)
- `Tiltfile` — no Go version references

### Go 1.22 → 1.26 Breaking Changes Summary

Four Go minor versions are spanned (1.23, 1.24, 1.25, 1.26). Go maintains backward compatibility; the changes below are the only ones that could affect this codebase:

**Go 1.23:**
- `range` over integers and iterator functions — additive, no breaking changes
- `time.Timer`/`time.Ticker` channels became unbuffered when `go.mod` targets 1.23+ — the operator does not use timer channels directly, but verify with `go vet`

**Go 1.24:**
- Generic type aliases — additive, no action needed
- `go tool` subcommand semantics changed — operator code doesn't use `go tool` programmatically
- Bootstrap now requires Go 1.22.6+ — only relevant if building the Go toolchain itself (CI downloads binary)
- `for` loop copylock diagnostic — verify `go vet` doesn't flag any `sync.Locker` in 3-clause for loops in controllers

**Go 1.25:**
- Compiler fix: nil-pointer panics now trigger correctly for deferred uses of error-returning function results (bug in Go 1.21–1.24) — audit any `f, err := os.Open(); f.Name()` patterns where `f` is used before `err` is checked. This project's code follows the `if err != nil { return }` pattern, so LOW RISK
- `GOMAXPROCS` auto-adjusts to container CPU quotas when `go.mod` targets 1.25+ — BENEFICIAL for operator running in pods with CPU limits
- DWARF v5 debug info by default — no impact on runtime behavior
- macOS 12 required — CI runs on Linux, no impact

**Go 1.26:**
- Green Tea GC enabled by default — 10-40% GC overhead reduction, no code changes needed. BENEFICIAL
- `net/url.Parse` rejects malformed URLs with unbracketed colons — the operator constructs Vault URLs from `spec.connection` which uses well-formed URLs. LOW RISK. `GODEBUG=urlstrictcolons=0` available as escape hatch
- `net/http.ServeMux` trailing-slash redirects use 307 instead of 301 — operator does not define HTTP routes via ServeMux (webhooks use controller-runtime's webhook server). NO IMPACT
- `cmd/doc` / `go tool doc` deleted — CI and Makefile do not invoke these. NO IMPACT
- `go mod init` defaults to lower version — irrelevant (we're upgrading an existing module)
- JPEG encoder/decoder output differs bit-for-bit — no image processing in this project. NO IMPACT

**Conclusion: No source code changes are expected.** The upgrade should be a drop-in Go version bump + `go mod tidy`.

### Dockerfile Multi-Arch Pattern

The current Dockerfile hardcodes `GOARCH=amd64`. The acceptance criteria require multi-arch support. Use the canonical BuildKit cross-compilation pattern:

```dockerfile
FROM --platform=$BUILDPLATFORM golang:1.26 AS builder

WORKDIR /workspace
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -o manager main.go

FROM registry.access.redhat.com/ubi9/ubi-minimal
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
```

Key points:
- `--platform=$BUILDPLATFORM` on the builder stage ensures the Go toolchain runs natively (no QEMU emulation)
- `ARG TARGETOS` and `ARG TARGETARCH` are auto-injected by BuildKit when building with `docker buildx --platform`
- Do NOT set default values for `TARGETOS`/`TARGETARCH` — defaults shadow BuildKit injection and cause incorrect cross-compilation
- When building without `--platform` (single-arch), BuildKit defaults `TARGETOS`/`TARGETARCH` to the host values, preserving backward compatibility

### CI Workflow Notes

Both `pr.yaml` and `push.yaml` delegate to `redhat-cop/github-workflows-operators` reusable workflows (pinned at commit SHA for v1.1.6). The `GO_VERSION` input is passed to the reusable workflow which uses it in `actions/setup-go`. The `~1.26` semver range will match Go 1.26.x (currently 1.26.4).

The reusable workflow pin (`@5b0493408637600ef9c71d2321337adac65153a3 # v1.1.6`) is NOT changed in this story — updating the reusable workflow reference is Story 8.4.

### `toolchain` Directive

Go 1.21+ introduced the `toolchain` directive in `go.mod`. Since we're jumping from `go 1.22.0` (which did not have a `toolchain` line), add `toolchain go1.26.4` to declare the minimum patch version for builds. When `GOTOOLCHAIN=auto` (the default), Go will automatically download this toolchain if the local version is older; with `GOTOOLCHAIN=local`, builds require the system Go to already satisfy the directive.

After editing `go.mod`, run `go mod tidy` which will validate the directive and update `go.sum`.

### Testing Standards

- Run `go build ./...` to verify compilation
- Run `go vet ./...` to verify static analysis
- Run `make test` (which runs envtest-based unit tests with build tag `!integration`)
- Integration tests (`make integration`) require a Kind cluster with Vault — verify locally if feasible, but these are also covered by CI. Note: ENVTEST_K8S_VERSION is still 1.29.0 (updated in Story 8.3), but envtest should still function with Go 1.26 as the test binary compiler

### Project Structure Notes

- No new files created; only existing files modified
- All paths align with the established project structure
- In practice, `go mod tidy` produced no `go.sum` changes for this version bump (no new or changed dependencies)

### References

- [Source: _bmad-output/planning-artifacts/epics.md — Epic 8, Story 8.1]
- [Source: _bmad-output/project-context.md — Technology Stack & Versions]
- [Source: go.mod — current Go 1.22.0 and dependency versions]
- [Source: Dockerfile — current golang:1.22 base image and GOARCH=amd64]
- [Source: .github/workflows/pr.yaml — GO_VERSION: ~1.22]
- [Source: .github/workflows/push.yaml — GO_VERSION: ~1.22]
- [Source: Go 1.26 Release Notes — https://go.dev/doc/go1.26]
- [Source: Go 1.25 Release Notes — https://go.dev/doc/go1.25]
- [Source: Go 1.24 Release Notes — https://go.dev/doc/go1.24]

## Dev Agent Record

### Agent Model Used

claude-4.6-opus (Cursor Agent)

### Debug Log References

- GOTOOLCHAIN=local prevented automatic toolchain download; resolved by using GOTOOLCHAIN=auto for go mod tidy and build commands
- Kind cluster container was stopped; recreated cluster before baseline integration tests

### Completion Notes List

- Upgraded Go from 1.22.0 to 1.26 with toolchain go1.26.4 in go.mod
- go mod tidy ran clean — no dependency conflicts (existing controller-runtime v0.17.3 and K8s libs v0.29.2 compile and work with Go 1.26)
- go build, go vet, and go test all pass with zero issues — no source code changes needed (as predicted in Dev Notes)
- Dockerfile updated to golang:1.26 with --platform=$BUILDPLATFORM + TARGETOS/TARGETARCH for multi-arch support
- Container build verified with podman build (golang:1.26 image pulled, binary compiled, image tagged successfully)
- CI workflows (pr.yaml, push.yaml) updated from GO_VERSION: ~1.22 to GO_VERSION: ~1.26
- project-context.md updated to reflect Go 1.26 for Language and Container image references
- make fmt, make vet, make test all pass — full unit test suite green with envtest on Go 1.26

### Change Log

- 2026-07-16: Upgraded Go 1.22.0 → 1.26 with toolchain go1.26.4, Dockerfile multi-arch, CI workflows updated

### File List

- go.mod (modified: go 1.22.0 → go 1.26, added toolchain go1.26.4)
- go.sum (unchanged — go mod tidy produced no hash changes for this version bump)
- Dockerfile (modified: golang:1.22 → golang:1.26, added --platform=$BUILDPLATFORM, ARG TARGETOS/TARGETARCH, GOOS/GOARCH parameterized)
- .github/workflows/pr.yaml (modified: GO_VERSION: ~1.22 → ~1.26)
- .github/workflows/push.yaml (modified: GO_VERSION: ~1.22 → ~1.26)
- _bmad-output/project-context.md (modified: Go 1.22.0 → Go 1.26, golang:1.22 → golang:1.26)
