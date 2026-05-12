# Build the manager binary
FROM mcr.microsoft.com/oss/go/microsoft/golang:1.26 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/main.go cmd/main.go
COPY api/ api/
COPY internal/ internal/

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
# GOEXPERIMENT=nosystemcrypto: the Microsoft Go 1.26+ base image defaults to
# GOEXPERIMENT=systemcrypto, which routes crypto/* through OpenSSL. The cgo
# variant requires CGO_ENABLED=1; the no-cgo variant (ms_nocgo_opensslcrypto)
# still dlopens libssl at runtime and therefore needs glibc's dynamic linker
# in the final image — which `distroless/minimal:3.0` does not ship, causing
# "exec /manager: no such file or directory" at container start. Disabling the
# experiment selects Go's pure-Go crypto, keeping the binary fully static and
# runnable on minimal distroless.
RUN GOEXPERIMENT=nosystemcrypto CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager cmd/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM mcr.microsoft.com/azurelinux/distroless/minimal:3.0
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
