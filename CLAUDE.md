# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

The Operation Cache Controller is a Kubernetes operator built with kubebuilder framework that manages operations and caches the outcome of those operations. It provisions resources, pre-provisions resources for future use (caching), and manages the lifecycle of resources and cache.

## Architecture

The controller manages four main Custom Resource Definitions (CRDs):

1. **Requirement** - User-defined specifications for required applications and their configurations
2. **Operation** - Represents the deployment operation based on requirements
3. **AppDeployment** - Manages provision and teardown jobs for actual resource deployment
4. **Cache** - Stores pre-provisioned resources with cache duration and auto-count features

The main reconcilers are located in `internal/controller/`:
- `RequirementReconciler` - Creates Cache and Operation CRs from Requirement specs
- `OperationReconciler` - Creates/deletes AppDeployment CRs from Operation specs  
- `AppDeploymentReconciler` - Creates provision/teardown Kubernetes Jobs
- `CacheReconciler` - Manages cached resource lifecycle

## Common Development Commands

### Building and Testing
- `make build` - Build the manager binary (runs manifests, generate, fmt, vet first)
- `make run` - Run the controller locally
- `make test` - Run unit tests with coverage
- `make test-integration` - Run integration tests  
- `make test-e2e` - Run end-to-end tests (requires Kind cluster)

### Code Quality
- `make fmt` - Format Go code
- `make vet` - Run go vet
- `make lint` - Run golangci-lint with project configuration
- `make lint-fix` - Auto-fix linting issues where possible

### Code Generation
- `make generate` - Generate DeepCopy methods using controller-gen
- `make manifests` - Generate CRDs, RBAC, and webhook configurations

### Docker and Deployment
- `make docker-build` - Build Docker image (use `IMG=name:tag` to specify image)
- `make docker-push` - Push Docker image to registry
- `make install` - Install CRDs into Kubernetes cluster
- `make deploy` - Deploy controller to Kubernetes cluster
- `make build-installer` - Generate consolidated YAML in `dist/install.yaml`

## Key Files and Directories

- `api/v1alpha1/` - CRD type definitions
- `internal/controller/` - Controller reconciler implementations
- `internal/handler/` - Business logic handlers for each CRD type
- `internal/utils/controller/` - Controller helper utilities
- `config/` - Kubernetes manifests and kustomization files
- `doc/arch/` - Architecture documentation with detailed diagrams

## Testing Requirements

- Unit tests must be written for new functionality
- Integration tests validate controller behavior with Kubernetes API
- E2E tests require a running Kind cluster
- All tests must pass before submitting changes
- Coverage reports are generated in `cover.out`

## Linting Configuration

The project uses golangci-lint with custom configuration in `.golangci.yml`. Key enabled linters include errcheck, gofmt, goimports, govet, staticcheck, and others. Some linters are disabled for specific paths (api/ and internal/ directories have relaxed line length requirements).