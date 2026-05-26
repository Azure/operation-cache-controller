## Why

The Operation Cache Controller has been built and is running in production, but it has no machine-readable specification. New contributors must reverse-engineer behavior from Go code, and future OpenSpec change proposals have nothing to diff against. This baseline change captures the *current* behavior of the four CRDs and their reconcilers as OpenSpec capability specs so that subsequent work can propose true deltas.

## What Changes

- Document the existing system as four capability specs derived from the current `internal/controller/` and `internal/handler/` implementations.
- No source code, CRDs, RBAC, or runtime behavior change. This is a documentation-only baseline.
- Establish naming conventions for future capability deltas (one capability per top-level CRD).
- Record the cache hit/miss data flow and finalizer/ownership rules as testable requirements.

## Capabilities

### New Capabilities

- `requirement-management`: Reconciliation of `Requirement` CRs into `Operation` and (optionally) `Cache` resources, including cache-key derivation and cache-hit acquisition.
- `operation-orchestration`: Lifecycle of `Operation` CRs — fan-out to per-app `AppDeployment` children, status aggregation, finalizer-driven teardown, and acquisition annotations.
- `appdeployment-execution`: Translation of `AppDeployment` CRs into provision/teardown Kubernetes `Job`s, job-status reconciliation, and finalizer cleanup.
- `cache-pool`: Pre-provisioning of `Operation`s under a `Cache`, auto-count maintenance, cache-duration expiry, and label-based cache-key indexing.

### Modified Capabilities

<!-- None: this is the initial baseline; no prior specs exist. -->

## Impact

- Affected code: none (read-only documentation pass over `api/v1alpha1/`, `internal/controller/`, `internal/handler/`).
- Affected APIs: none. The four CRDs in `controller.azure.github.com/v1alpha1` are described, not changed.
- Affected processes: future changes must now ship spec deltas against these baselines instead of free-form proposals.
- Risk: low — if the spec mis-describes current behavior, it is corrected in a follow-up change; runtime is unaffected.
