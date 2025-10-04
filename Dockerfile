# syntax=docker/dockerfile:1

# Build the manager binary
# Using golang:1.25-alpine3.22 for latest stable Go 1.25.x with Alpine 3.22
# alpine3.22 is the latest stable Alpine Linux (June 2024)
FROM --platform=${BUILDPLATFORM:-linux/amd64} docker.io/golang:1.25-alpine3.22 AS builder

# Build arguments for cross-platform builds and versioning
ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE

# Install build dependencies
# ca-certificates: for HTTPS requests
# git: for go mod download from private repos
# Using --no-cache to keep the layer small
RUN apk add --no-cache ca-certificates git

WORKDIR /workspace

# Copy go.mod and go.sum first for better layer caching
# This layer will only be rebuilt when dependencies change
COPY go.mod go.sum ./

# Download and verify dependencies
# Using cache mount for go modules to speed up rebuilds
# go mod verify ensures integrity of dependencies (supply chain security)
RUN --mount=type=cache,target=/go/pkg/mod \
  go mod download && \
  go mod verify

# Copy source code
# Organized by frequency of changes (api < internal < cmd)
COPY api/ api/
COPY internal/ internal/
COPY cmd/ cmd/

# Build the binary with full optimizations
# Build flags explained:
#   -trimpath: removes file system paths from binary (reproducible builds)
#   -ldflags:
#     -w: omit DWARF symbol table (smaller binary)
#     -s: omit symbol table and debug info (smaller binary)
#     -extldflags '-static': fully static binary (no dynamic linking)
#     -X: inject version info at build time
RUN --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg/mod \
  CGO_ENABLED=0 \
  GOOS=${TARGETOS} \
  GOARCH=${TARGETARCH} \
  go build \
  -trimpath \
  -ldflags="-w -s -extldflags '-static' \
  -X main.version=${VERSION} \
  -X main.commit=${COMMIT} \
  -X main.date=${BUILD_DATE}" \
  -o manager \
  cmd/main.go

# Runtime stage: minimal distroless image
# Using multi-arch distroless image (supports linux/amd64, linux/arm64, linux/arm/v7)
FROM gcr.io/distroless/static-debian12:nonroot

# OCI image labels for metadata (helps with image management and scanning)
LABEL org.opencontainers.image.title="Helios Operator" \
  org.opencontainers.image.description="Kubernetes operator for managing Helios applications" \
  org.opencontainers.image.url="https://github.com/hoangphuc841/helios-operator" \
  org.opencontainers.image.source="https://github.com/hoangphuc841/helios-operator" \
  org.opencontainers.image.vendor="Helios" \
  org.opencontainers.image.licenses="Apache-2.0" \
  org.opencontainers.image.documentation="https://github.com/hoangphuc841/helios-operator/blob/main/README.md"

WORKDIR /

# Copy only the compiled binary from builder
COPY --from=builder /workspace/manager .

# Run as non-root user for security
# User 65532 is "nonroot" user in distroless images
USER 65532:65532

ENTRYPOINT ["/manager"]
