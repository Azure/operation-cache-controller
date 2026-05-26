## 1. Contract Confirmation

- [ ] 1.1 Confirm maintainers want `operation-cache-controller.azure.github.com/cache-mode=true` to be an operator-facing contract
- [ ] 1.2 If the annotation should remain internal, revise this change before implementation to remove or stop relying on the emitted annotation

## 2. Test Coverage

- [ ] 2.1 Add cache handler coverage that captures created `Operation`s and verifies the cache-mode annotation key and value
- [ ] 2.2 Keep existing cache-key label assertions passing while adding the annotation assertion
- [ ] 2.3 Add or update coverage so the test fails if cached Operation creation drops annotations while preserving labels and ownerReferences

## 3. Validation

- [ ] 3.1 Run `go test ./...`
- [ ] 3.2 Run `openspec validate --change document-cache-mode-annotation`
- [ ] 3.3 Review the `cache-pool` delta for consistency with the existing cache-key label requirement
