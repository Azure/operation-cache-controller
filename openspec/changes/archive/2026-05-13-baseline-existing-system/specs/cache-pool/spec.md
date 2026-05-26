## ADDED Requirements

### Requirement: Cache maintains a pool of pre-provisioned Operations

The controller SHALL, for every `Cache`, set `status.keepAlive` to the controller's fixed keep-alive count, list owned `Operation`s, publish the names of owned `Operation`s with `status.phase=Reconciled` in `status.availableCaches`, and create additional owned `Operation`s from `spec.operationTemplate` only when the total owned `Operation` count is below `status.keepAlive`.

#### Scenario: Empty Cache provisions Operations to fill the pool

- **WHEN** a `Cache` is created with `spec.operationTemplate` set and no owned `Operation`s exist
- **THEN** the controller creates owned `Operation`s from the template until the total owned `Operation` count reaches `status.keepAlive`
- **AND** after owned `Operation`s reach `status.phase=Reconciled`, `status.availableCaches` lists their names

### Requirement: Cache CR is keyed by a deterministic cache key

The controller SHALL compute `status.cacheKey` deterministically from `spec.operationTemplate` and SHALL apply the label `operation-cache-controller.azure.github.com/cache-key=<key>` to every `Operation` it creates for the cache, truncating the label value to the Kubernetes label-value limit when required. The controller SHALL NOT add this cache-key label to the `Cache` resource itself.

#### Scenario: Two Caches with identical templates compute the same key

- **WHEN** two `Cache` CRs with byte-identical `spec.operationTemplate` are created
- **THEN** both report the same `status.cacheKey`, and cached `Operation`s created for either `Cache` carry the matching `cache-key` label value

### Requirement: Cache surfaces available entries in status

The controller SHALL keep `status.availableCaches` synchronized with the names of owned `Operation`s that are `Reconciled`. When a cached `Operation` is acquired, ownership transfers away from the `Cache`, so the next cache reconcile SHALL omit that `Operation` from `status.availableCaches`.

#### Scenario: Acquired cached Operation disappears from status

- **WHEN** a `Requirement` acquires a cached `Operation`
- **THEN** that `Operation`'s name is removed from the `Cache`'s `status.availableCaches` within one reconcile cycle

### Requirement: Cache expires entries past spec.expireTime

The controller SHALL delete the `Cache` CR once the wall-clock time exceeds `Cache.spec.expireTime` (when set) and SHALL stop processing additional cache-pool operations in that reconcile. Cleanup of owned cached `Operation`s relies on Kubernetes `ownerReferences` cascade deletion.

#### Scenario: Past-expiry Cache is deleted

- **WHEN** the current time is after `Cache.spec.expireTime`
- **THEN** the controller deletes the `Cache` CR and does not create replacement cached `Operation`s in that reconcile

### Requirement: Cache reconciliation runs at least every 60 seconds

The controller SHALL re-reconcile every `Cache` CR within 60 seconds even with no watch events, so pool depletion and expiry are eventually observed.

#### Scenario: Idle Cache is re-evaluated

- **WHEN** a `Cache` exists and no events fire against it
- **THEN** the controller re-reconciles it within 60 seconds

### Requirement: Cache does not use a finalizer

The controller SHALL NOT register a finalizer on `Cache` CRs; cleanup of owned `Operation`s relies on Kubernetes `ownerReferences` cascade deletion alone.

#### Scenario: Deleting a Cache cascades via owner references

- **WHEN** a `Cache` with owned cached `Operation`s is deleted
- **THEN** Kubernetes garbage-collects those `Operation`s via `ownerReferences` without any finalizer interaction
