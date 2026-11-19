## Why

`OperationHandler` finds an `Operation`'s current `AppDeployment`s through a field index keyed only by the owner `Operation` name, and the list is not namespace-scoped. Two `Operation`s with the same name in different namespaces therefore share one index key, so reconciling one of them can see, and diff against, `AppDeployment`s that belong to the other. The baseline spec says an `Operation` manages its *owned* `AppDeployment`s, so the lookup should enforce that.

## What Changes

- Key the `AppDeployment` owner index by `<namespace>/<operation-name>` instead of the bare name.
- Scope the `AppDeployment` list to the `Operation`'s namespace.
- Keep only `AppDeployment`s whose controller owner reference matches the reconciling `Operation`'s UID before computing adds, updates and removals.
- Add tests covering a same-named `Operation` in another namespace.

## Capabilities

### New Capabilities

<!-- None. -->

### Modified Capabilities

- `operation-orchestration`: Make explicit that an `Operation` only reconciles `AppDeployment`s it controls in its own namespace.

## Impact

- Affected code: `internal/controller/operation_controller.go`, `internal/handler/operation.go`, `internal/utils/controller/operation_helper.go`, and related tests.
- Affected APIs: no CRD schema changes.
- Affected systems: none. The field index is in-memory and rebuilt at startup, and existing `AppDeployment`s already carry a controller owner reference from `SetControllerReference`, so no migration is needed.
- Risk: low.
