## ADDED Requirements

### Requirement: Requirement CR creates an Operation when caching is disabled

The controller SHALL, for every `Requirement` whose `spec.enableCache` is `false`, create one owned `Operation` named after the `Requirement` whose `spec` is a copy of `requirement.spec.template`, set `status.operationName` and `status.operationId` accordingly, and transition `status.phase` from `""` → `Operating` → `Ready` as the child `Operation` becomes `Reconciled`.

#### Scenario: Cache-disabled requirement provisions a fresh Operation

- **WHEN** a `Requirement` is created with `spec.enableCache=false` and a valid `spec.template`
- **THEN** the controller records the `Requirement` name in `status.operationName`, creates an `Operation` owned by the `Requirement`, and reports `status.phase=Operating`
- **AND** when the child `Operation` reaches `status.phase=Reconciled`, the `Requirement` reports `status.phase=Ready` with `status.operationId` copied from the child `Operation`

### Requirement: Requirement CR consults the cache when caching is enabled

The controller SHALL, for every `Requirement` whose `spec.enableCache` is `true`, derive a deterministic cache key from `spec.template`, look up a `Cache` CR by that key, and either acquire an existing cached `Operation` (cache hit) or create a new `Operation` (cache miss).

#### Scenario: Cache hit acquires a pre-provisioned Operation

- **WHEN** a `Requirement` with `enableCache=true` is reconciled and a matching `Cache` CR exists with at least one available cached `Operation`
- **THEN** the controller records one cached `Operation` name in `status.operationName`, transfers ownership of that `Operation` to the `Requirement`, stamps the annotation `operation.controller.azure.com/acquired` on it with the acquisition timestamp, sets `status.phase=Ready`, and records condition `OperationReady=True` with reason `CacheHit`

#### Scenario: Cache miss creates a new Operation and a Cache CR if missing

- **WHEN** a `Requirement` with `enableCache=true` is reconciled and either the `Cache` CR is absent or has no available cached `Operation`
- **THEN** the controller ensures a `Cache` CR exists for the derived cache key (creating it from `spec.template` if necessary), creates a new owned `Operation` named after the `Requirement`, and records cache-miss status with reason `CacheMiss` when an existing `Cache` has no available entries

### Requirement: Requirement CR does not run finalizer cleanup

The controller SHALL NOT add the declared finalizer `finalizer.requirement.devinfra.goms.io` during `Requirement` reconciliation; cleanup of owned `Operation` resources relies on Kubernetes `ownerReferences` cascade deletion.

#### Scenario: Deletion is not blocked by a Requirement finalizer

- **WHEN** a user deletes a `Requirement` that owns an `Operation`
- **THEN** the controller does not add a `Requirement` finalizer or drive a `Deleting` phase, and Kubernetes garbage collection removes owned `Operation` resources via `ownerReferences`

### Requirement: Requirement reconciliation is periodically requeued

The controller SHALL requeue every `Requirement` for re-reconciliation at most every 10 minutes even when no watch event fires, so that `spec.expireAt` and cache state changes are eventually observed.

#### Scenario: Idle requirement is re-checked

- **WHEN** a `Requirement` is in `status.phase=Ready` and no events occur
- **THEN** the controller re-evaluates it within 10 minutes
