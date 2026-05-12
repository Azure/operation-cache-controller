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
# GOEXPERIMENT=ms_nocgo_opensslcrypto: the Microsoft Go 1.26+ base image defaults
# to GOEXPERIMENT=systemcrypto, which routes crypto/* through OpenSSL via cgo and
# requires CGO_ENABLED=1. ms_nocgo_opensslcrypto keeps the OpenSSL backend (FIPS-
# friendly) but resolves libssl via dlopen at runtime, so we can keep CGO_ENABLED=0
# and ship a static binary into the distroless final stage. The distroless base
# below must provide libssl.so.3 at runtime.
RUN GOEXPERIMENT=ms_nocgo_opensslcrypto CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o manager cmd/main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM mcr.microsoft.com/azurelinux/distroless/minimal:3.0
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
