## 1. Verify spec accuracy against existing code

- [x] 1.1 Cross-check `requirement-management` scenarios against `internal/controller/requirement_controller.go` and `internal/handler/requirement.go`; note any divergence in PR review thread
- [x] 1.2 Cross-check `operation-orchestration` scenarios against `internal/controller/operation_controller.go` and `internal/handler/operation.go`
- [x] 1.3 Cross-check `appdeployment-execution` scenarios against `internal/controller/appdeployment_controller.go` and `internal/handler/appdeployment.go`
- [x] 1.4 Cross-check `cache-pool` scenarios against `internal/controller/cache_controller.go` and `internal/handler/cache.go`
- [x] 1.5 Confirm constant names quoted in specs (finalizers, annotation keys, label keys) match `api/v1alpha1/*_types.go` exactly

## 2. Validate OpenSpec artifacts

- [x] 2.1 Run `openspec status --change baseline-existing-system` and confirm all four artifacts are `done`
- [x] 2.2 Run `openspec validate --change baseline-existing-system` (or equivalent) and resolve any schema errors
- [x] 2.3 Verify each spec file lives under `openspec/changes/baseline-existing-system/specs/<capability>/spec.md`

## 3. Reconcile with existing tests

- [x] 3.1 For each requirement, identify at least one existing Ginkgo test in `internal/controller/*_test.go` or `internal/handler/*_test.go` that exercises it; note coverage gaps as follow-up issues (do NOT add tests in this change)
- [x] 3.2 Record any requirement that has no covering test as an open question for the next change

## 4. Project documentation hookup

- [x] 4.1 Add a one-paragraph "Specifications" section to `README.md` (or `doc/arch/`) pointing at `openspec/specs/` and explaining that future PRs must ship deltas
- [x] 4.2 Optionally seed `openspec/config.yaml` `context:` with the controller's tech-stack summary (kubebuilder v4, Go 1.26, controller-runtime, Ginkgo) so future LLM runs have project context

## 5. Archive

- [x] 5.1 Run `/opsx:verify` and address any findings
- [x] 5.2 Run `/opsx:archive` to promote the four capability specs from `openspec/changes/baseline-existing-system/specs/` into `openspec/specs/`
- [x] 5.3 Confirm `openspec/specs/{requirement-management,operation-orchestration,appdeployment-execution,cache-pool}/spec.md` exist after archive
