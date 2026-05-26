## Context

The Operation Cache Controller is a kubebuilder v4 operator already running in production. It manages four CRDs in `controller.azure.github.com/v1alpha1` (`Requirement`, `Operation`, `AppDeployment`, `Cache`) using a handler-based reconciliation pattern in `internal/controller/` and `internal/handler/`. There is no machine-readable spec, so this change introduces OpenSpec as the source of truth going forward without modifying any code paths.

Stakeholders: controller maintainers (need a baseline to diff against), reviewers (want spec-level PRs), new contributors (want to learn intent without reading 22 test files).

## Goals / Non-Goals

**Goals:**
- Capture the *current* externally observable behavior of each CRD as testable scenarios under `openspec/specs/<capability>/spec.md`.
- One capability spec per top-level CRD; nested concepts (Jobs, conditions, finalizers) appear as scenarios within the parent capability rather than separate capabilities.
- Use SHALL/MUST language so each requirement is normative and verifiable against existing Ginkgo tests.

**Non-Goals:**
- No code, CRD field, RBAC, or runtime behavior changes.
- No reorganization of `internal/` packages.
- No documentation of internal helper packages (`internal/utils/...`); these are implementation detail.
- No webhook spec — webhooks are scaffolded but not implemented.

## Decisions

### Decision: One capability per top-level CRD

We map exactly one capability to each of `Requirement`, `Operation`, `AppDeployment`, `Cache`. Alternative considered: a single `cache-controller` capability covering everything. Rejected — it would force every future delta to touch one giant spec and erase the natural CRD-level review boundary the team already uses.

### Decision: Scenarios written in WHEN/THEN against observable cluster state

Each scenario describes what an external observer (kubectl / a watcher) sees, not what handler functions do internally. Alternative: describe internal handler invocations. Rejected — couples specs to implementation, defeats the purpose of being able to refactor handlers without rewriting specs.

### Decision: Document constants (finalizers, annotations, label keys) as part of requirements

Names like `finalizer.operation.controller.azure.com` and `operation.controller.azure.com/acquired` are part of the *contract* with cluster operators and cannot be silently renamed. They appear inline in the relevant scenarios. Alternative: keep them only in code constants. Rejected — anyone integrating with the controller (dashboards, GitOps tooling) needs to discover these from the spec.

### Decision: No delta operations in this change

Because no prior specs exist, every requirement uses `## ADDED Requirements`. Future changes will use `MODIFIED`/`REMOVED`/`RENAMED` against these baselines. This keeps the baseline change reviewable as additive-only.

## Risks / Trade-offs

- **Risk:** Spec drifts from code if it under-describes edge cases (e.g., race conditions on cache acquisition). → **Mitigation:** Treat baseline as v1; correct in follow-up changes when discrepancies are found during real reviews. Do not block this PR on exhaustive coverage.
- **Risk:** Naming a capability after a CRD ties spec churn to API renames. → **Mitigation:** API is `v1alpha1` and stable in practice; if a CRD is ever renamed, that change uses `RENAMED Requirements` plus a folder rename — supported by OpenSpec.
- **Risk:** Cache-hit acquisition logic is genuinely non-trivial (annotation timestamps, ownership transfer). Spec may oversimplify. → **Mitigation:** Scenarios cite the annotation key and ownership-transfer behavior explicitly so future deltas have something concrete to amend.

## Migration Plan

1. Land this change. Archive after `/opsx:verify` passes.
2. Going forward, every PR that touches `api/v1alpha1/` or reconciler behavior MUST include an OpenSpec change with deltas against the baseline.
3. No rollback needed — pure documentation.

## Open Questions

- Should `internal/utils/reconciler/operations.go` (sequential-operations pattern) be promoted to its own capability later? Deferred; it's an implementation detail today.
- Webhook scaffolding exists but is unused — do we add a placeholder capability now or wait until it's wired up? Deferred to when webhooks are activated.
- Follow-up test gap: no existing controller or handler test covers the current absence of `Requirement` finalizer reconciliation, despite the API constant `RequirementFinalizerName`.
- Follow-up test gap: no existing controller or handler test asserts that two independently initialized `Operation` resources receive distinct `status.operationId` values; helper coverage currently checks only that generated IDs are non-empty.
- Follow-up test gap: cache acquisition removal from `Cache.status.availableCaches` is covered indirectly by ownership-transfer behavior, but no end-to-end controller or handler test verifies the acquired operation disappearing from cache status after a cache reconcile.
- Follow-up test gap: no existing controller or handler test explicitly covers the `Cache` no-finalizer contract or garbage-collection behavior after user deletion.
