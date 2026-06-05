## Context

`OperationHandler.EnsureAllAppsAreReady` initializes an empty `Operation` by clearing conditions and assigning `status.operationId` from `OperationHelper.NewOperationId()`. That helper returns a random UUID string with hyphens removed. The baseline `operation-orchestration` spec now treats the ID as cluster-unique, but random generation only makes collision probability small.

The ID is externally observable through `Operation.status.operationId`, copied into `Requirement.status.operationId`, and stamped into child `AppDeployment.spec.opId` and generated resource names.

## Goals / Non-Goals

**Goals:**
- Make `status.operationId` strictly unique across `Operation` resources before child resources consume it.
- Keep the current string shape compatible with generated names where practical.
- Add tests that prove collision handling or deterministic uniqueness.

**Non-Goals:**
- No CRD schema changes.
- No migration of existing non-empty `status.operationId` values.
- No change to cache acquisition, AppDeployment fan-out semantics, or Requirement status copying beyond using the guaranteed ID.

## Decisions

### Decision: Prefer Kubernetes UID-derived IDs for new Operations

Use the `Operation` object's Kubernetes `metadata.uid` as the uniqueness source and normalize it to the existing 32-character lowercase UUID-without-hyphens format. Kubernetes assigns UIDs at object creation and uses them as the stable identity for namespaced resources, which avoids the race inherent in list-then-generate checks under concurrent reconciles.

Alternative considered: keep random IDs and list all `Operation`s before status update. Rejected as the primary mechanism because two reconcilers can list the same pre-update state and still publish the same candidate. A list check is useful as a defensive assertion, but it is not a strict uniqueness primitive by itself.

### Decision: Preserve existing non-empty IDs

Only initialize `status.operationId` when it is empty. Existing Operations with a populated ID continue through reconciliation unchanged so this change does not rename already-created AppDeployments or invalidate Requirement status references.

Alternative considered: rewrite all IDs to UID-derived values. Rejected because it would churn child names and externally visible status for live resources.

### Decision: Keep the random helper available only for non-contract use or remove it after refactor

If no production path needs `OperationHelper.NewOperationId()` after UID-derived assignment, remove it and its narrow test. If the helper remains, rename or document it so callers do not mistake it for a uniqueness guarantee.

Alternative considered: leave `NewOperationId()` unchanged and wrap it with collision retries. Rejected unless UID-derived IDs prove incompatible, because retry logic adds code while still needing concurrency-safe proof.

## Risks / Trade-offs

- **Risk:** Some consumer expects operation IDs to be random rather than UID-derived. -> **Mitigation:** The API only exposes the value as an opaque ID; keep the same normalized UUID format.
- **Risk:** Fake-client or unit-test objects may lack `metadata.uid`. -> **Mitigation:** Test helper setup should assign UIDs for initialization cases; production reconcilers should treat an empty UID as a transient error and requeue instead of publishing a fallback random ID.
- **Risk:** Existing duplicate IDs could remain if they were already published before this change. -> **Mitigation:** This change guarantees new assignments only; add a defensive test/documentation note rather than mutating live IDs.

## Migration Plan

1. Refactor operation ID initialization to derive from `Operation.UID` when `status.operationId` is empty.
2. Ensure status is updated only after a non-empty, normalized ID is available.
3. Add handler/controller tests for two Operations in different namespaces receiving distinct IDs and for preserving existing IDs.
4. Run the existing Go test suite and OpenSpec validation.

## Open Questions

- Should the implementation also list existing Operations and emit an event if it detects a legacy duplicate ID? That would aid diagnosis but is not required for strict uniqueness of new assignments.
