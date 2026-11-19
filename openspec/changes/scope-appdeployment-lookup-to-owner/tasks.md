## 1. Owner Lookup

- [x] 1.1 Add `OperationOwnerIndexKey(namespace, name)` and use it in the `AppDeployment` owner indexer
- [x] 1.2 Scope `listCurrentAppDeployments` with `client.InNamespace` and the namespaced index key
- [x] 1.3 Filter listed `AppDeployment`s with `metav1.IsControlledBy` before diffing

## 2. Tests

- [x] 2.1 Update the indexer test to expect the namespaced key
- [x] 2.2 Give handler fixtures a controller owner reference to the test `Operation`
- [x] 2.3 Add handler coverage proving a same-named `Operation`'s `AppDeployment` in another namespace is not listed or deleted

## 3. Validation

- [x] 3.1 Run `go test ./internal/handler/... ./internal/utils/...` and `go test ./internal/controller -run IndexerFunc`
- [ ] 3.2 Run `make test` and `make test-integration` in CI
- [ ] 3.3 Run `openspec validate --change scope-appdeployment-lookup-to-owner`
