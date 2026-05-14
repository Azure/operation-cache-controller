## Context

`CacheHandler.initOperationFromCache` stamps `operation-cache-controller.azure.github.com/cache-mode=true` on every `Operation` it creates from `Cache.spec.operationTemplate`. The baseline `cache-pool` spec already documents cache-owned Operation creation and the cache-key label, but it does not mention this annotation.

Because annotations are visible to cluster operators and automation, this change decides whether the emitted cache-mode marker is part of the controller contract.

## Goals / Non-Goals

**Goals:**
- Document the cache-mode annotation as observable metadata on cache-created `Operation`s.
- Add tests that fail if cached Operation creation drops or changes the annotation.
- Keep the annotation scoped to Operations created by the Cache controller.

**Non-Goals:**
- No CRD schema changes.
- No change to cache hit acquisition, ownerReference transfer, or `status.availableCaches` behavior.
- No requirement that acquired Operations keep or remove the annotation after ownership transfer unless a later change specifies that lifecycle.

## Decisions

### Decision: Treat cache-mode as an operator-facing marker on cache-created Operations

The delta updates `cache-pool` because the annotation is created by Cache reconciliation and identifies Operations that originated from a cache pool. The requirement is limited to creation time, matching current behavior without inventing lifecycle semantics for acquired Operations.

Alternative considered: omit the annotation from specs as internal. Rejected for this proposal because operators can already observe and select on annotations, and the baseline policy documents observable constants that carry integration value.

### Decision: Keep the value as the existing string literal `true`

The spec names both the key and value so future refactors cannot silently change the contract. This follows the baseline approach for finalizers, labels, and acquisition annotations.

Alternative considered: specify only that the annotation exists. Rejected because a marker annotation with changing values is harder for operators to consume consistently.

## Risks / Trade-offs

- **Risk:** Maintainers may decide cache-mode should be an internal implementation detail. -> **Mitigation:** Resolve before apply; if internal, change this proposal to remove the annotation from code instead of documenting it.
- **Risk:** Documenting the annotation freezes a currently undocumented constant. -> **Mitigation:** The scope is narrow: only `Operation`s created by `Cache` reconciliation must carry it at creation.
- **Risk:** Existing tests use mocks and may not inspect created Operation metadata. -> **Mitigation:** Capture the `Create` argument in the cache handler test or add a fake-client controller test.

## Migration Plan

1. Add a `cache-pool` spec delta for the annotation on cached Operation creation.
2. Add focused coverage around `CacheHandler.AdjustCache` or `initOperationFromCache` proving the annotation key and value.
3. Run the existing Go test suite and OpenSpec validation.

## Open Questions

- Should acquired cached Operations keep `operation-cache-controller.azure.github.com/cache-mode=true` after the ownerReference transfers to a Requirement, or should a later acquisition lifecycle delta clear it?
