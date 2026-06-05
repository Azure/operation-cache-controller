## MODIFIED Requirements

### Requirement: Operation IDs are unique within the cluster

The controller SHALL assign every `Operation` a non-empty `status.operationId` value that is unique across all `Operation` resources in the cluster. For newly initialized `Operation`s, the controller SHALL derive or validate the ID before publishing status so two `Operation`s cannot receive the same `status.operationId` even when reconciled concurrently.

#### Scenario: Two independently created Operations get distinct IDs

- **WHEN** two `Operation` CRs are created in any namespaces
- **THEN** their `status.operationId` values differ

#### Scenario: Concurrent initialization does not publish duplicate IDs

- **WHEN** two empty `Operation` CRs are reconciled at the same time
- **THEN** each `Operation` publishes a non-empty `status.operationId`
- **AND** no `status.operationId` value is shared by both `Operation`s

#### Scenario: Existing Operation ID is preserved

- **WHEN** an `Operation` already has a non-empty `status.operationId`
- **THEN** reconciliation does not replace that value while the `Operation` continues normal fan-out and status aggregation
