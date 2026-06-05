# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Operation Cache Controller is a Kubernetes operator built with kubebuilder v4 framework that manages operations and caches the outcome of those operations. It provisions resources, pre-provisions resources for future use (caching), and manages the lifecycle of resources and cache.

## Architecture

The controller manages four main Custom Resource Definitions (CRDs) in the `controller.azure.github.com` API group:

1. **Requirement** - User-defined specifications for required applications with optional caching
2. **Operation** - Represents the deployment operation based on requirements
3. **AppDeployment** - Manages provision and teardown jobs for actual resource deployment
4. **Cache** - Stores pre-provisioned resources with cache duration and auto-count features

### Controller Architecture Pattern

The project follows a **handler-based reconciliation pattern** with clear separation of concerns:

```
Controller (Reconciler) → Handler Interface → Handler Implementation
                                              ↓
                                    Helper Utilities
```

The main reconcilers are located in `internal/controller/`:
- `RequirementReconciler` - Creates Cache and Operation CRs from Requirement specs
- `OperationReconciler` - Creates/deletes AppDeployment CRs from Operation specs
- `AppDeploymentReconciler` - Creates provision/teardown Kubernetes Jobs
- `CacheReconciler` - Manages cached resource lifecycle

Each controller uses:
- **Sequential operations**: Reconciliation operations are executed in order using `internal/utils/reconciler/operations.go`
- **Handler interfaces**: Business logic is separated into mockable handler interfaces in `internal/handler/`
- **High concurrency**: 50-100 concurrent reconciles per controller
- **Finalizers**: All CRDs except Cache use finalizers for cleanup
- **Field indexers**: Efficient queries for owned resources

### Key Relationships and Data Flow

- **Requirements** own **Operations** (unless acquired from cache)
- **Operations** own multiple **AppDeployments**
- **AppDeployments** own Kubernetes **Jobs**
- **Caches** own multiple **Operations** (for pre-provisioning)

**Cache Hit Flow**: Requirement → finds cached Operation → acquires it → Ready immediately
**Cache Miss Flow**: Requirement → creates Operation → creates AppDeployments → creates Jobs → provisions resources

## Common Development Commands

### Building and Testing
- `make build` - Build the manager binary (runs manifests, generate, fmt, vet first)
- `make run` - Run the controller locally
- `make test` - Run unit tests with coverage (excludes e2e/integration)
- `make test-integration` - Run integration tests with envtest K8s API server
- `make test-e2e` - Run end-to-end tests (requires Kind cluster)

### Running Specific Tests
- `go test ./internal/controller/...` - Run controller tests only
- `go test ./internal/handler/...` - Run handler tests only
- `go test -run TestRequirementReconciler_Reconcile` - Run specific test
- `go test -v ./internal/controller/requirement_controller_test.go` - Run with verbose output

### Code Quality
- `make fmt` - Format Go code with gofmt
- `make vet` - Run go vet to catch suspicious constructs
- `make lint` - Run golangci-lint with project configuration (18 linters enabled)
- `make lint-fix` - Auto-fix linting issues where possible

### Code Generation and Mocks
- `make generate` - Generate DeepCopy methods using controller-gen v0.17.2
- `make manifests` - Generate CRDs, RBAC, and webhook configurations
- `go generate ./...` - Generate mocks for handler interfaces using mockgen

### Docker and Deployment
- `make docker-build` - Build Docker image (use `IMG=name:tag` to specify image)
- `make docker-push` - Push Docker image to registry
- `make install` - Install CRDs into Kubernetes cluster using kustomize v5.6.0
- `make deploy` - Deploy controller to Kubernetes cluster
- `make build-installer` - Generate consolidated YAML in `dist/install.yaml`

## Key Files and Directories

- `api/v1alpha1/` - CRD type definitions with kubebuilder markers
- `internal/controller/` - Controller reconciler implementations (4 controllers)
- `internal/handler/` - Business logic handlers for each CRD type with mockable interfaces
- `internal/utils/controller/` - Controller helper utilities for CRDs
- `internal/utils/reconciler/` - Common reconciliation patterns and operation sequencing
- `internal/utils/rand/` - Random string generation utilities
- `internal/utils/ptr/` - Pointer helper utilities
- `internal/log/` - Logging constants and structured logging keys
- `config/` - Kubernetes manifests and kustomization files
- `test/e2e/` - End-to-end tests using sigs.k8s.io/e2e-framework
- `test/integration/` - Integration tests using envtest
- `test/utils/` - Common test utilities and resource builders
- `doc/arch/` - Architecture documentation with detailed diagrams (6 files)

## Development Environment Requirements

- **Go**: v1.26.0+ (as specified in go.mod and Dockerfile; toolchain pinned to go1.26.3)
- **Docker**: For building container images
- **Kind**: For running E2E tests locally
- **Tools**: controller-gen v0.17.2, kustomize v5.6.0, golangci-lint v2.12.2 (auto-installed via Makefile)

## Testing Strategy

### Test Structure
- **Unit Tests**: 22 test files using Ginkgo v2 + Gomega framework
- **Integration Tests**: Use envtest for lightweight K8s API server testing
- **E2E Tests**: Require Kind cluster, test full deployment lifecycle in `operation-cache-controller-system` namespace
- **Mock Generation**: Uses go.uber.org/mock with `//go:generate mockgen` directives

### Mock Interfaces
All handler interfaces are mocked for testing:
- `internal/handler/mocks/mock_requirement.go`
- `internal/handler/mocks/mock_operation.go`
- `internal/handler/mocks/mock_cache.go`
- `internal/handler/mocks/mock_appdeployment.go`

### Testing Patterns
- Controllers use envtest suite in `internal/controller/suite_test.go`
- Handlers are unit tested with mocked Kubernetes clients
- Coverage reports generated in `cover.out`

## Linting Configuration

The project uses golangci-lint with custom configuration in `.golangci.yml`. Key enabled linters include:
- **Error handling**: errcheck, staticcheck, govet
- **Code quality**: gocyclo, dupl, goconst, revive
- **Style**: gofmt, goimports, misspell
- **Performance**: prealloc, ineffassign
- **Testing**: ginkgolinter
- **Go 1.22+**: copyloopvar

Some linters are disabled for specific paths (api/ and internal/ have relaxed line length requirements).

## Important Constants and Patterns

### Labels and Annotations
- `operation-cache-controller.azure.github.com/cache-key` - Cache key label for grouping
- `operation.controller.azure.com/acquired` - Operation acquisition timestamp annotation
- `operation-cache-controller.azure.github.com/cache-mode` - Cache mode annotation

### Finalizers
- `finalizer.operation.controller.azure.com` - Operation cleanup finalizer
- `finalizer.appdeployment.devinfra.goms.io` - AppDeployment cleanup finalizer

### Resource Naming Patterns
- Max length: 63 characters (Kubernetes limit)
- Format: `{type}-{hash}` for generated names (e.g., `cache-a31acb88`)
- Operation IDs are unique across the cluster

### Requeue Intervals
- Default reconciliation: 10 seconds
- Requirement check: 10 minutes
- Cache check: 60 seconds

## Development Workflow

### Adding New Features
1. **CRD Changes**: Update `api/v1alpha1/*_types.go` with kubebuilder markers
2. **Generate Code**: Run `make manifests generate` to update CRDs and DeepCopy methods
3. **Handler Logic**: Implement business logic in `internal/handler/` with interface-first approach
4. **Controller Updates**: Update reconciler in `internal/controller/` to call new handler methods
5. **Testing**: Add unit tests for handlers and integration tests for controllers
6. **Mocks**: Run `go generate ./...` to update mock interfaces

### Debugging Tips
- Use structured logging with keys from `internal/log/`
- Check controller metrics at `:8443/metrics` endpoint
- Examine finalizers if resources aren't deleting
- Review owner references for cascade deletion issues
- Check field indexes if queries are slow

### RBAC Considerations
- Controllers have full CRUD on all 4 CRDs
- AppDeployment controller has full access to batch/jobs
- Metrics endpoint requires authentication
- Generated RBAC includes admin/editor/viewer roles for each CRD

## Key Kubebuilder Patterns

### Controller Configuration
- High concurrency: 50-100 reconciles per controller
- Field indexers for efficient child resource queries
- Named controllers with clear reconciliation strategies
- Status subresources for optimistic concurrency control

### Resource Management
- Owner references enable cascade deletion
- Finalizers ensure external resource cleanup
- PrintColumns provide useful kubectl output
- Webhooks scaffolded but not currently implemented