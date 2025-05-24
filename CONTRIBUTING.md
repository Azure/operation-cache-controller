# Contributing

This project welcomes contributions and suggestions. Most contributions require you to
agree to a Contributor License Agreement (CLA) declaring that you have the right to,
and actually do, grant us the rights to use your contribution. For details, visit
https://cla.microsoft.com.

When you submit a pull request, a CLA-bot will automatically determine whether you need
to provide a CLA and decorate the PR appropriately (e.g., label, comment). Simply follow the
instructions provided by the bot. You will only need to do this once across all repositories using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/)
or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Development Environment Setup

Before you can build and test this project, you'll need to have the following installed:

* [Go](https://golang.org/dl/) (version 1.23 or higher, as specified in the `Dockerfile`)
* [Docker](https://docs.docker.com/get-docker/) (for building container images)
* [Make](https://www.gnu.org/software/make/)
* `controller-gen`, `golangci-lint`, `kustomize`, `setup-envtest` (these can often be installed via the Makefile, e.g., `make controller-gen golangci-lint`)

## Build Instructions

This project uses `make` to streamline common development tasks.

### Common Makefile Targets

* **`make all` or `make build`**: Builds the manager binary. This also runs code generation, formatting, and vetting.
* **`make run`**: Runs the controller locally on your host.
* **`make docker-build`**: Builds the Docker image for the manager.
  * You can specify the image name and tag using the `IMG` variable: `make docker-build IMG=myregistry/mycontroller:latest`
* **`make docker-push`**: Pushes the built Docker image to a container registry.
  * Ensure `IMG` is set to the desired repository.
* **`make manifests`**: Generates Kubernetes manifest files (CRDs, RBAC, Webhook configurations) using `controller-gen`.
* **`make generate`**: Generates Go code, such as DeepCopy methods, using `controller-gen`.

For a complete list of build targets and their descriptions, run `make help`.

## Coding Conventions

To maintain code quality and consistency, please adhere to the following conventions:

### Formatting and Linting

* **Go Formatting**: Run `make fmt` to format your Go code according to standard Go style.
* **Go Vet**: Run `make vet` to catch suspicious constructs.
* **Linting**: Run `make lint` to check for a wide range of issues using `golangci-lint`. The configuration for `golangci-lint` is located in `.golangci.yml`.
  * Key linters enabled include `errcheck`, `gofmt`, `goimports`, `govet`, `staticcheck`, `unused`, and others. Please refer to `.golangci.yml` for the full list.
  * You can attempt to automatically fix some linting issues with `make lint-fix`.

### Code Generation

* Ensure that any changes to API types (`api/**/*.go`) are followed by running `make generate` and `make manifests` to update generated code and Kubernetes manifests. Commit these generated files along with your changes.

### Licensing

* All Go source files must include the license header provided in `hack/boilerplate.go.txt`. This project uses the Apache License 2.0.

### Testing

* **Unit Tests**: Run `make test` to execute unit tests.
* **Integration Tests**: Run `make test-integration` for integration tests.
* **End-to-End (E2E) Tests**: Run `make test-e2e` for E2E tests. Ensure your Kind cluster is set up as per the project's requirements.

Write new tests for new features and ensure existing tests pass before submitting a pull request.

## Submitting Changes

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make your changes, adhering to the coding conventions and including tests.
4. Ensure all tests pass (`make test test-integration test-e2e`).
5. Ensure the code is formatted and linted (`make fmt lint`).
6. Commit your changes with clear and concise commit messages.
7. Push your branch to your fork.
8. Open a pull request against the main repository.
9. Ensure the CLA bot checks are addressed.
