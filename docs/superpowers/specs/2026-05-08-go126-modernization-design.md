# Go 1.26 Upgrade & Codebase Modernization — Design

**Status:** Draft, awaiting user review
**Date:** 2026-05-08
**Scope:** Operation Cache Controller (`github.com/Azure/operation-cache-controller`)
**Target Go version:** 1.26
**Delivery:** 7 sequential, stacked PRs

---

## 1. Problem & Constraints

### Problem

The Operation Cache Controller currently targets Go 1.24 and uses a mix of legacy and modern Go patterns. Over the last several Go releases (1.21–1.26) the language and standard library have grown substantial new features — `slices`/`maps`/`cmp` packages, range-over-int, type-safe atomics, `errors.Join`, `errors.AsType[T]`, `wg.Go`, `new(val)`, structured iterators, `t.Context()`, `omitzero`, `b.Loop()`, `SplitSeq`, and more — that the codebase doesn't take advantage of. This produces three concrete problems:

1. **Code is harder to read and maintain than it needs to be** — hand-rolled loops, manual error wrapping, `interface{}` in mocks, and `for i := 0; i < n; i++` patterns add friction that modern equivalents eliminate.
2. **The codebase will drift further** — without a linter that enforces modern style, every new contribution risks reintroducing legacy patterns.
3. **The toolchain is one major version behind** — Go 1.26 is shipping, and pinning to 1.24 leaves performance and stdlib improvements on the table.

### Goals

- Bump the Go toolchain to **1.26** across `go.mod`, `Dockerfile`, and CI.
- Modernize hand-written code to idiomatic Go 1.26 (mechanical rewrites + judicious library swaps).
- Regenerate mocks via an upgraded `go.uber.org/mock` so generated code is also modern.
- Bump targeted direct dependencies that materially benefit modernization.
- Upgrade `golangci-lint` and enable modernization linters (`intrange`, `copyloopvar`, `usestdlibvars`, `perfsprint`) so the codebase stays modern.
- Ship as **7 sequential, reviewable PRs**, each independently buildable and bisect-safe.

### Non-Goals (explicitly out of scope)

- **No structural / architectural changes.** No splitting files, no reshaping handler interfaces, no reducing duplication across reconcilers, no API/CRD changes.
- **No behavioral changes.** Reconcile loops, finalizer logic, cache hit/miss flows, requeue intervals — all unchanged.
- **No transitive-dep cleanup.** Only direct deps are touched; `go mod tidy` runs but we don't chase indirect upgrades.
- **No Kubebuilder / controller-runtime major upgrade** unless Go 1.26 forces it. Major upgrades are their own project.
- **No new tests.** Existing tests must keep passing; no test-coverage push as part of this work.
- **No CRD / kubectl manifest restructuring.** `make manifests` runs, but we don't reorganize `config/`.

### Constraints

- **Compatibility:** All 4 reconcilers must continue to satisfy `controller-runtime`'s `Reconciler` interface unchanged.
- **Concurrency:** The 50–100 concurrent-reconciles configuration must not regress.
- **Test gates:** Each PR must pass `make fmt vet lint test test-integration` locally before merge. e2e validated by CI.
- **Code generation:** `make manifests generate` and `go generate ./...` must be re-runnable without producing dirty diffs after the work lands.
- **PR sequencing:** Only one PR in this stack open at any time; next PR opens only after the previous merges to `main`.
- **Bisect safety:** Every PR (and ideally every commit) leaves the project in a buildable, test-passing state.

---

## 2. Architecture & PR Sequencing

This isn't a runtime architecture (no new components are introduced). The "architecture" here is the **delivery architecture** — how the 7 PRs stack and what each one owns.

### The 7-PR Stack

```
                main
                 │
        ┌────────▼────────────────────────┐
PR #1   │ Go 1.26 toolchain bump          │  go.mod, Dockerfile, CI matrix
        └────────┬────────────────────────┘
                 │
        ┌────────▼────────────────────────┐
PR #2   │ mockgen bump + regenerate       │  internal/handler/mocks/*.go
        └────────┬────────────────────────┘
                 │
        ┌────────▼────────────────────────┐
PR #3   │ Targeted direct dep upgrades    │  go.mod, go.sum
        └────────┬────────────────────────┘
                 │
        ┌────────▼────────────────────────┐
PR #4   │ golangci-lint bump + new linters│  .golangci.yml, Makefile
        │ (suppressions added so green)   │  + //nolint markers as needed
        └────────┬────────────────────────┘
                 │
        ┌────────▼────────────────────────┐
PR #5   │ Mechanical modernization        │  any, range-int, t.Context,
        │                                 │  omitzero, b.Loop, SplitSeq
        └────────┬────────────────────────┘
                 │
        ┌────────▼────────────────────────┐
PR #6   │ Idiomatic library swaps         │  slices, maps, cmp, errors.Join,
        │                                 │  sync.OnceValue, atomic.Bool
        └────────┬────────────────────────┘
                 │
        ┌────────▼────────────────────────┐
PR #7   │ Go 1.26-specific features       │  wg.Go, new(val), errors.AsType[T]
        │ + remove all //nolint markers   │  cleanup of PR #4's suppressions
        └────────┬────────────────────────┘
                 ▼
                main (fully modernized, lint-clean, no suppressions)
```

### What each PR owns

| PR | Owns | Doesn't touch |
|---|---|---|
| **#1** Go bump | `go.mod` (`go 1.26`, `toolchain go1.26.0`), `Dockerfile` (base image), `.github/workflows/*.yml` (Go matrix), `go.sum` if needed for transitive forced bumps | Any `.go` file content |
| **#2** mockgen | `go.mod`/`go.sum` for `go.uber.org/mock`, regenerated `internal/handler/mocks/*.go` | Any hand-written `.go` |
| **#3** Direct deps | `go.mod`/`go.sum` for explicitly chosen direct deps | Any `.go` file content |
| **#4** Lint config | `.golangci.yml`, possibly `Makefile` (lint version pin), targeted `//nolint:linter` directives in files that will be fixed in #5–#7 | Logic/style fixes — only suppressions |
| **#5** Mechanical refactor | Hand-written `.go` files: `interface{}`→`any`, `for i := 0; i < n; i++`→`for i := range n`, test contexts→`t.Context()`, JSON tags `omitempty`→`omitzero` where appropriate, benchmarks→`b.Loop()`, split-iteration→`SplitSeq` | Library swaps; semantic changes |
| **#6** Library swaps | Hand-written `.go` files: hand-rolled loops→`slices.*`/`maps.*`, multi-error returns→`errors.Join`, lazy init→`sync.OnceValue`, primitive atomics→`atomic.Bool`/etc., zero-value defaults→`cmp.Or` | 1.26-only features (saved for #7) |
| **#7** Go 1.26 features | `sync.WaitGroup` use→`wg.Go`, `x := v; &x`→`new(v)`, `errors.As(err, &target)`→`errors.AsType[T]`, removal of all `//nolint` markers from PR #4 | New functionality |

### Critical sequencing rules

- **PR #2 depends on PR #1** because mockgen output may differ under the new toolchain.
- **PR #4 depends on PR #1** because some new linters require Go 1.26 awareness.
- **PRs #5–#7 depend on PR #4** because the linter is what tells us *what* to modernize and *where*.
- **PR #7 must come last** because reverting it preserves a fully-modernized state minus the 1.26-specific syntax — useful if a 1.26 dep regression is discovered post-merge.
- **Only one PR open at a time** to keep rebase cost zero.

### Suppression strategy in PR #4

When PR #4 lands, lint must stay green even though no code has been modernized yet. Strategy:

1. Enable each new linter at `severity: error`.
2. Run lint, capture every violation.
3. Add file-level `//nolint:intrange,copyloopvar,usestdlibvars,perfsprint` directives to every flagged file, with a TODO comment: `// TODO(modernization): remove after PR #5/#6/#7`.
4. Each subsequent PR removes the suppressions for the linters it addresses.

This guarantees PR #4 is mergeable, makes the modernization debt visible (grep-able), and forces #5–#7 to actually clean it up.

---

## 3. Per-PR Mechanics & Change Flow

For a refactor, "data flow" is really *change flow* — what concrete edits happen in each PR, how they're discovered, and how they're verified.

### PR #1 — Go 1.26 toolchain bump

**Discovery:** None needed; the change set is fixed.

**Edits:**
- `go.mod`: `go 1.24.0` → `go 1.26.0`; add `toolchain go1.26.0`; remove the `godebug default=go1.24` line.
- `Dockerfile`: bump `FROM golang:1.24` → `FROM golang:1.26` (same Alpine/Debian variant as today).
- `.github/workflows/*.yml`: bump `go-version` (and matrix entries if any) to `1.26.x`.
- `Makefile`: if there's a `GO_VERSION` variable, bump it.
- `go.sum`: regenerate via `go mod tidy` if any transitive bumps are forced.

**Verification gate:** `make build && make fmt vet test test-integration`. No code changes, so all must pass unchanged.

**Rollback:** Single-PR revert restores Go 1.24.

### PR #2 — mockgen bump + regenerate

**Discovery:**
- Read current `go.uber.org/mock` version in `go.mod`.
- Check the latest stable release.

**Edits:**
- `go.mod`/`go.sum`: bump `go.uber.org/mock` to latest stable.
- Run `go install go.uber.org/mock/mockgen@<pinned>` and `go generate ./...`.
- Commit only the regenerated files in `internal/handler/mocks/*.go` (4 files: `mock_requirement.go`, `mock_operation.go`, `mock_cache.go`, `mock_appdeployment.go`).

**Verification gate:** `make test test-integration` — mocks must still satisfy the handler interfaces and existing tests must pass without edits.

**Rollback:** Single revert. Mocks regenerate from old version.

### PR #3 — Targeted direct dependency upgrades

**Discovery process:**
1. `go list -m -u all` → list all direct deps with available updates.
2. For each direct dep, check whether the available update has a Go 1.25+ or 1.26 awareness note in its changelog.
3. **Skip** any dep whose update would be a major-version bump (those belong in their own PRs).

**Concrete candidates** (final list determined during execution):
- `sigs.k8s.io/controller-runtime` — minor bumps only.
- `github.com/onsi/ginkgo/v2`, `github.com/onsi/gomega` — modernization-friendly minor bumps.
- `k8s.io/api`, `k8s.io/apimachinery`, `k8s.io/client-go` — patch/minor bumps if available without forcing K8s major-version churn.
- `golang.org/x/*` — patch bumps.

**Explicitly deferred:**
- Any dep requiring a major-version bump.
- `controller-gen` (lives in tools, separate concern).
- `kustomize` (build tool, not a Go import).

**Edits:** `go.mod`, `go.sum` only.

**Verification gate:** `make build fmt vet lint test test-integration`. Critical — this is where dep-induced regressions surface.

**Rollback:** Single revert.

### PR #4 — golangci-lint bump + new linters + suppressions

**Discovery:**
1. Bump `golangci-lint` version in `Makefile`.
2. Edit `.golangci.yml` to add: `intrange`, `copyloopvar`, `usestdlibvars`, `perfsprint`.
3. Run `make lint` — capture every violation.
4. For each flagged file, prepend a file-level `//nolint:<linters> // TODO(modernization): remove after PR #5/#6/#7` directive.

**Edits:**
- `Makefile`: pin new `GOLANGCI_LINT_VERSION`.
- `.golangci.yml`: add 4 linters under `linters.enable`.
- Hand-written `.go` files: insert file-level `//nolint` markers in flagged files only. **No logic edits.**

**Verification gate:** `make lint test test-integration` — must be green with suppressions in place.

**The TODO marker pattern:**

```go
//nolint:intrange,perfsprint // TODO(modernization): remove after PR #5
package controller
```

This makes the remaining debt grep-able: `grep -rn "TODO(modernization)" .` produces the worklist for PRs #5–#7.

### PR #5 — Mechanical modernization

**Discovery:** `grep -rn "TODO(modernization).*PR #5" .` + linter output for `intrange`, `copyloopvar`, `usestdlibvars`.

**Per-pattern change list:**

| Pattern | Find | Replace | Notes |
|---|---|---|---|
| `interface{}` | `interface{}` | `any` | Already mostly clean; verify only 2 files affected. |
| Index loops | `for i := 0; i < n; i++` | `for i := range n` | Only when `i` is unused after the loop or used purely as an index. |
| Test contexts | `ctx, cancel := context.WithCancel(context.Background())` in test funcs | `ctx := t.Context()` | Drop the matching `defer cancel()`. |
| JSON tags | `omitempty` on `time.Duration`, `time.Time`, structs, slices, maps | `omitzero` | **Only** for these types; leave `omitempty` on strings/ints/bools. |
| Benchmarks | `for i := 0; i < b.N; i++` | `for b.Loop()` | Project may have zero benchmarks; if so, no-op. |
| Split iteration | `for _, x := range strings.Split(s, sep)` | `for x := range strings.SplitSeq(s, sep)` | Only in for-range; not when the slice is otherwise used. |
| Loop variable capture | `x := x` shadow inside loops | Remove the shadow | `copyloopvar` lint will flag. |

**Per-file workflow:**
1. Apply the mechanical edits.
2. Remove the relevant `//nolint` directives from PR #4.
3. Re-run `make lint test test-integration` per file batch.

**Verification gate:** Full `make fmt vet lint test test-integration`. No `intrange`/`copyloopvar`/`usestdlibvars` suppressions remain.

### PR #6 — Idiomatic library swaps

**Discovery:** Manual reading + `perfsprint` lint output + targeted `grep`s:
- `grep -rn "for .* range" --include="*.go"` for `slices.Contains`, `slices.IndexFunc`, etc.
- `grep -rn "fmt.Errorf" --include="*.go"` for `errors.Join`.
- `grep -rn "sync.Once" --include="*.go"` for `sync.OnceValue`.
- `grep -rn "atomic.Store\|atomic.Load" --include="*.go"` for `atomic.Bool`/`atomic.Int64`.

**Per-pattern change list (apply with judgment, not mechanically):**

| Replace | With | When |
|---|---|---|
| Manual contains loop | `slices.Contains` | Loop's only purpose is membership test |
| Manual index search | `slices.Index` / `slices.IndexFunc` | Loop returns first match |
| Manual sort wrappers | `slices.SortFunc` + `cmp.Compare` | Custom comparators |
| Manual map clone | `maps.Clone` | Whole-map copy with no transformation |
| Manual map filter delete | `maps.DeleteFunc` | Conditional deletion |
| `if x == ""; x = "default"` | `cmp.Or(x, "default")` | Zero-value fallback chains |
| `fmt.Errorf("a: %w; b: %w", e1, e2)` | `errors.Join(e1, e2)` | Combining errors |
| `sync.Once` + value var | `sync.OnceValue` | Lazy single-value init |
| `atomic.StoreInt32` family | `atomic.Bool`/`atomic.Int64`/`atomic.Pointer[T]` | Type-safe equivalent exists |
| `time.Now().Sub(x)` | `time.Since(x)` | Already idiomatic; sweep for stragglers |

**Verification gate:** Full `make fmt vet lint test test-integration`. `perfsprint` suppressions removed.

### PR #7 — Go 1.26-specific features + suppression cleanup

**Discovery:** Manual reading + `grep`s:
- `grep -rn "wg.Add\|sync.WaitGroup" --include="*.go"` for `wg.Go` candidates.
- `grep -rn "errors.As(" --include="*.go"` for `errors.AsType[T]` candidates.
- Hand-search for `x := val; ptr = &x` patterns that should become `new(val)`.

**Per-pattern change list:**

| Replace | With |
|---|---|
| `wg.Add(1); go func(){ defer wg.Done(); ... }()` | `wg.Go(func(){ ... })` |
| `x := val; ptr = &x` (struct field init pattern) | `ptr = new(val)` |
| `var t *T; if errors.As(err, &t) { ... }` | `if t, ok := errors.AsType[*T](err); ok { ... }` |

**Final cleanup:**
- `grep -rn "TODO(modernization)" .` must return zero results before this PR is opened for review.
- All `//nolint` markers added in PR #4 must be gone.

**Verification gate:** Full `make fmt vet lint test test-integration` + manual confirmation of the grep checks above.

---

## 4. Interfaces, Contracts, and Verification

### Contract 1 — Public Go API surface

**Rule:** No exported identifier in `api/v1alpha1/` or `internal/` may change name, signature, or visibility across the entire 7-PR stack.

**Verification:**
- `go vet ./...` after each PR.
- Manual diff check: `git diff main -- 'api/**/*.go' 'internal/handler/*.go' | grep -E '^[+-](func|type|var|const) [A-Z]'`. If anything appears, justify or revert.
- Mocks regenerate cleanly in PR #2 — if they don't, an interface changed.

**Allowed exception:** PR #6 may swap a return value from a custom error to `errors.Join(...)` — this is a value-level change, not a signature change.

### Contract 2 — Operator runtime behavior

**Rule:** No reconciler logic, requeue interval, finalizer behavior, owner reference structure, label/annotation key, or status transition may change.

| Frozen behavior | Source |
|---|---|
| Requeue intervals (10s default, 10m requirement, 60s cache) | `internal/utils/reconciler/` and per-controller |
| Finalizer strings | `internal/log/` constants and per-controller |
| Label/annotation keys (`cache-key`, `acquired`, `cache-mode`) | per-CRD type files |
| Concurrency settings (50–100 concurrent reconciles) | controller `SetupWithManager` |
| Owner reference graph (Requirement→Operation→AppDeployment→Job; Cache→Operation) | handler implementations |
| Cache hit/miss flow (acquisition, pre-provisioning) | `RequirementHandler`, `CacheHandler` |

**Verification:**
- `make test test-integration` passes unchanged.
- Spot-check: `git diff main -- 'internal/controller/*.go' | grep -E 'RequeueAfter|Finalizer|owner'`. Hits require explicit justification.
- e2e on CI must remain green after each PR.

### Contract 3 — Developer build & test commands

**Rule:** Every existing `make` target keeps working with identical semantics.

**Verification gate (per PR), executed in order:**

```bash
# 1. Format & vet
make fmt && git diff --exit-code
make vet

# 2. Generated artifacts in sync
make manifests generate
git diff --exit-code

# 3. Mocks regenerate clean
go generate ./...
git diff --exit-code

# 4. Lint
make lint

# 5. Tests
make test
make test-integration

# 6. Build
make build
make docker-build IMG=local-test:dev
```

Any of steps 1–6 failing blocks PR merge.

### Contract 4 — Per-PR completion criteria

| PR | Exit criteria |
|---|---|
| **#1** | `go.mod` shows `go 1.26.0`. Dockerfile/CI use Go 1.26. All 6 verification steps pass. No `.go` file under `internal/` or `api/` has changed. |
| **#2** | `go.uber.org/mock` bumped. Only `internal/handler/mocks/*.go` changed. All 6 verification steps pass. `go generate ./...` produces clean diff. |
| **#3** | Only `go.mod`/`go.sum` changed. No major-version bumps. All 6 verification steps pass. |
| **#4** | New linters enabled in `.golangci.yml`. `golangci-lint --version` reports the new pinned version. `make lint` is green. Every flagged file has a `//nolint:<linters> // TODO(modernization): remove after PR #N` directive. `grep -rn "TODO(modernization)" . \| wc -l` produces a non-zero count. |
| **#5** | All `intrange`, `copyloopvar`, `usestdlibvars` suppressions added in PR #4 are removed. `grep -rn "TODO(modernization).*PR #5" . \| wc -l == 0`. All 6 verification steps pass. |
| **#6** | All `perfsprint` suppressions removed. `grep -rn "TODO(modernization).*PR #6" . \| wc -l == 0`. All 6 verification steps pass. |
| **#7** | `grep -rn "TODO(modernization)" . \| wc -l == 0`. `grep -rn "//nolint:intrange\|//nolint:copyloopvar\|//nolint:usestdlibvars\|//nolint:perfsprint" .` returns nothing. All 6 verification steps pass. |

### Contract 5 — Bisect safety

**Rule:** `git bisect` between any two commits in the merged history must work — every merge commit on `main` must be buildable and test-passing.

**PR squash policy:** Squash on merge is fine for all PRs in this stack. Commit-level bisect within a PR is not required since the PR-level bisect is sufficient.

---

## 5. Risks, Rollback, and Open Questions

### Risks

| # | Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|---|
| **R1** | A direct dep in PR #3 has a hidden behavior change (e.g., controller-runtime patch tightens validation) | Medium | High — could break reconcile loops | Run `make test-integration` and `make test-e2e` on CI for PR #3. Roll back individual deps via `go mod edit -replace` if needed. |
| **R2** | A linter in PR #4 produces thousands of false positives | Medium | Medium — bloats PR #4 | Cap at the 4 named linters. If suppression count exceeds ~50 files, defer one of the linters to a follow-up PR. |
| **R3** | mockgen regeneration produces mocks incompatible with current Ginkgo setup | Low | High — blocks PR #2 | Pin to a known-good mockgen version. Test locally before opening PR. |
| **R4** | Go 1.26 stdlib behavior change breaks an assumption in handler logic | Low | Medium | Full `make test test-integration` must pass on PR #1. e2e on CI catches anything tests miss. |
| **R5** | Reviewer demands structural changes during PRs #5–#7, pulling scope back in | Medium | Medium — derails schedule | Explicit non-goals in Section 1. Refer reviewer to spec. Open follow-up issue rather than expanding PR. |
| **R6** | Stacked-PR rebases cause merge conflicts when an unrelated PR lands on `main` | Medium | Low | Only one stack PR open at a time. Conflicts will be tiny because each PR has tightly-scoped diff. |
| **R7** | `omitzero` migration in PR #5 changes JSON serialization in a way that breaks an external CRD-status consumer | Low | High — silent data drift | `omitzero` is *more* correct than `omitempty` for `time.Duration`/`time.Time`. Verify CRD status fixtures in tests. If a fixture depends on old `omitempty` behavior, treat as a real bug to fix. |
| **R8** | `errors.AsType[T]` doesn't behave identically to `errors.As(&target)` in some edge case | Low | Medium | Limit usage to patterns shown in `use-modern-go` skill. For unusual error-handling sites, leave `errors.As` in place. |
| **R9** | CI Docker image cache misses on Go 1.26 image | High | Low — annoyance only | Pre-warm the cache by pushing a no-op commit on a throwaway branch after PR #1 merges. |
| **R10** | A `//nolint` directive in PR #4 hides a real bug | Low | Medium | Suppressions are file-scoped and time-limited (removed in #5–#7). Bugs hidden by suppressions surface as suppressions come off. |

### Rollback strategy

| Scenario | Response |
|---|---|
| Production incident traced to a specific PR | `git revert <merge commit>`. No cascade — every PR's contract guarantees buildable state on its own. |
| Production incident, unclear which PR is to blame | Bisect by PR boundaries (7 candidates max). |
| Need to revert the entire modernization | Revert PRs in reverse order (#7, then #6, etc.). |
| Regret a single mid-stack feature | Revert just that PR; PRs after it may need a small rebase. |

**Hard rule:** No PR in this stack may merge if the previous PR isn't already on `main`.

### Open Questions (deferred to execution)

| # | Question | When answered | Default if unresolved |
|---|---|---|---|
| **Q1** | Exact list of direct deps to bump in PR #3 | Beginning of PR #3 (`go list -m -u all`) | Patch updates only; defer minor bumps. |
| **Q2** | Specific `golangci-lint` version to pin in PR #4 | Beginning of PR #4 | Latest stable as of PR open date. |
| **Q3** | Specific `go.uber.org/mock` version to pin in PR #2 | Beginning of PR #2 | Latest stable, validated locally. |
| **Q4** | Are there any benchmarks in the codebase? | First file scan of PR #5 | If zero, omit `b.Loop()` from PR #5. |
| **Q5** | Does the project have a `tools.go` pinning developer tools? | PR #2 | If yes, version bumps go there. If no, ad-hoc `go install` is fine. |
| **Q6** | Does CI use a build matrix? | PR #1 | Single-version matrix is the safe default. |
| **Q7** | Are all 4 modernization linters supported by the chosen `golangci-lint` version? | PR #4 | Drop any unsupported linter; don't downgrade golangci-lint. |

### Success criteria for the whole effort

When all 7 PRs are merged and `main` is fully modernized, these statements must all be true:

- [ ] `go.mod` declares `go 1.26.0` and `toolchain go1.26.0`.
- [ ] `Dockerfile` uses a Go 1.26 base image.
- [ ] CI runs on Go 1.26.
- [ ] `make fmt vet lint test test-integration` is green on `main`.
- [ ] `grep -rn "interface{}" --include="*.go"` returns zero hand-written hits.
- [ ] `grep -rn "TODO(modernization)" .` returns zero hits.
- [ ] No `//nolint` markers added by PR #4 remain.
- [ ] `.golangci.yml` enables `intrange`, `copyloopvar`, `usestdlibvars`, `perfsprint`.
- [ ] `go generate ./...` and `make manifests generate` produce zero diff on a clean checkout.
- [ ] e2e CI on `main` is green.
- [ ] No exported identifier in `api/v1alpha1/` or `internal/handler/` interfaces changed.
- [ ] No reconciler logic, requeue interval, finalizer, or owner-reference structure changed.
