# operation-orchestration Specification

## Purpose
Specify Operation reconciliation for AppDeployment fan-out, status aggregation, cache acquisition metadata, unique operation IDs, and deletion lifecycle.
## Requirements
### Requirement: Operation fans out to one AppDeployment per application

The controller SHALL, for every `Operation`, create exactly one owned `AppDeployment` per entry in `spec.applications`, copying the application's `provision`, `teardown`, and `dependencies` fields and stamping `spec.opId` with the `Operation`'s unique `status.operationId`.

#### Scenario: Multi-application Operation creates matching AppDeployments

- **WHEN** an `Operation` is created with two `ApplicationSpec` entries named `app-a` and `app-b`
- **THEN** the controller creates two `AppDeployment` resources owned by the `Operation`, each carrying the parent `operationId` in `spec.opId`, and the `Operation` reports `status.phase=Reconciling`

### Requirement: Operation status aggregates child AppDeployment phases

The controller SHALL set `status.phase=Reconciled` on an `Operation` only when every owned `AppDeployment` reports `status.phase=Ready`, and SHALL keep `status.phase=Reconciling` otherwise.

#### Scenario: Operation becomes Reconciled when all children are Ready

- **WHEN** all `AppDeployment`s owned by an `Operation` report `status.phase=Ready`
- **THEN** the `Operation` reports `status.phase=Reconciled`

#### Scenario: One pending child keeps the Operation reconciling

- **WHEN** at least one owned `AppDeployment` is not yet `Ready`
- **THEN** the `Operation` continues to report `status.phase=Reconciling`

### Requirement: Operation acquisition is recorded via annotation

When an `Operation` is acquired from a cache by a `Requirement`, the controller SHALL stamp the annotation `operation.controller.azure.com/acquired` on the `Operation` with an RFC3339 timestamp and SHALL transfer the `ownerReference` from the `Cache` to the acquiring `Requirement`.

#### Scenario: Acquired Operation carries the timestamp annotation

- **WHEN** a `Requirement` acquires a cached `Operation`
- **THEN** the `Operation` has annotation `operation.controller.azure.com/acquired` set to the acquisition time and its sole controller `ownerReference` points to the `Requirement`

### Requirement: Operation uses a finalizer to record deletion lifecycle

The controller SHALL add the finalizer `finalizer.operation.controller.azure.com` to every `Operation`, transition a deleting `Operation` through `status.phase=Deleting` and `status.phase=Deleted`, and then remove the finalizer. Deletion of owned `AppDeployment`s is delegated to Kubernetes `ownerReferences` and the `AppDeployment` controller; the `Operation` controller does not wait for every child to report `Deleted`.

#### Scenario: Operation deletion records Deleting and Deleted phases

- **WHEN** an `Operation` is deleted
- **THEN** the `Operation` enters `status.phase=Deleting`, then `status.phase=Deleted`, after which the controller removes `finalizer.operation.controller.azure.com` and the API server may remove the `Operation`

### Requirement: Operation IDs are unique within the cluster

The controller SHALL assign every `Operation` a `status.operationId` value that is unique across all `Operation` resources in the cluster.

#### Scenario: Two independently created Operations get distinct IDs

- **WHEN** two `Operation` CRs are created in any namespaces
- **THEN** their `status.operationId` values differ
