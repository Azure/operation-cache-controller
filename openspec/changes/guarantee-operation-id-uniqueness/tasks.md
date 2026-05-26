## 1. Operation ID Assignment

- [ ] 1.1 Add or refactor an operation ID helper that normalizes `Operation.metadata.uid` into the existing lowercase UUID-without-hyphens format
- [ ] 1.2 Update `OperationHandler.EnsureAllAppsAreReady` so empty `status.operationId` values are initialized from the Operation UID before AppDeployments are created
- [ ] 1.3 Ensure reconciliation requeues without publishing status if an initializing Operation has an empty UID
- [ ] 1.4 Preserve non-empty existing `status.operationId` values during later reconciles
- [ ] 1.5 Remove or rename `OperationHelper.NewOperationId()` if it is no longer used for contract-level ID assignment

## 2. Tests

- [ ] 2.1 Add helper coverage for UID normalization and empty UID handling
- [ ] 2.2 Add handler or controller coverage proving two Operations in different namespaces receive distinct non-empty `status.operationId` values
- [ ] 2.3 Add coverage proving an existing non-empty `status.operationId` is not rewritten
- [ ] 2.4 Add coverage that child AppDeployments still receive the parent `status.operationId` in `spec.opId`

## 3. Validation

- [ ] 3.1 Run `go test ./...`
- [ ] 3.2 Run `openspec validate --change guarantee-operation-id-uniqueness`
- [ ] 3.3 Review the archived baseline requirement and this delta for wording consistency before applying
