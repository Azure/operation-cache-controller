## MODIFIED Requirements

### Requirement: Cache maintains a pool of pre-provisioned Operations

The controller SHALL, for every `Cache`, set `status.keepAlive` to the controller's fixed keep-alive count, list owned `Operation`s, publish the names of owned `Operation`s with `status.phase=Reconciled` in `status.availableCaches`, and create additional owned `Operation`s from `spec.operationTemplate` only when the total owned `Operation` count is below `status.keepAlive`. Every `Operation` created by the `Cache` controller for the pool SHALL carry the annotation `operation-cache-controller.azure.github.com/cache-mode=true`.

#### Scenario: Empty Cache provisions Operations to fill the pool

- **WHEN** a `Cache` is created with `spec.operationTemplate` set and no owned `Operation`s exist
- **THEN** the controller creates owned `Operation`s from the template until the total owned `Operation` count reaches `status.keepAlive`
- **AND** after owned `Operation`s reach `status.phase=Reconciled`, `status.availableCaches` lists their names

#### Scenario: Cached Operations carry cache-mode annotation

- **WHEN** the controller creates an `Operation` to fill a `Cache` pool
- **THEN** the created `Operation` has annotation `operation-cache-controller.azure.github.com/cache-mode` set to `true`
