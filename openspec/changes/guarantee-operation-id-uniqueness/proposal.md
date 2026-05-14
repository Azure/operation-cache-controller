## Why

The baseline spec says `Operation.status.operationId` is unique across the cluster, but the current implementation relies only on random ID generation. That makes collisions unlikely, not impossible, so the implementation should enforce the contract before the spec is treated as strict.

## What Changes

- Add a cluster-wide uniqueness check before assigning a generated `status.operationId`.
- Regenerate candidate IDs when a collision with an existing `Operation` is found.
- Preserve the existing operation ID format unless implementation proves it must change.
- Add tests that force a collision and verify the controller resolves it before publishing status.

## Capabilities

### New Capabilities

<!-- None. -->

### Modified Capabilities

- `operation-orchestration`: Strengthen the existing operation ID requirement so uniqueness is actively enforced, not only probabilistically expected.

## Impact

- Affected code: `internal/handler/operation.go`, `internal/utils/controller/operation_helper.go`, and related tests.
- Affected APIs: no CRD schema changes; `status.operationId` remains the externally visible field.
- Affected systems: controller-runtime client behavior may include an extra `OperationList` or indexed lookup during ID assignment.
- Risk: low to moderate; retry logic must avoid publishing an ID until a non-conflicting value is found.
