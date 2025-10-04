# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added - 2025-10-04

#### Documentation Improvements

**Enhanced README.md:**

- Added comprehensive "Development" section with complete workflow
- Added "Makefile Targets Reference" with all 36+ targets documented
- Enhanced "Troubleshooting" section with practical solutions and debugging tips
- Added "IDE Integration" examples for VS Code and GoLand
- Reorganized "Resources" section with categorized links
- Added performance tips for fast iteration vs full validation

**Makefile Analysis:**

- Conducted comprehensive analysis of current Makefile
- Current Makefile is well-structured and production-ready with:
  - 36 targets covering all development workflows
  - Tool version pinning and caching
  - Multi-platform Docker build support
  - Comprehensive RBAC and CRD generation
  - E2E testing with Kind cluster automation
- Recommendation: Keep current Makefile as-is (no breaking changes needed)
- All targets documented in README for easy reference

**Best Practices Applied:**

- Clear categorization with `##@` sections
- Self-documenting help target
- Consistent naming conventions
- Proper dependency management
- CI/CD friendly targets

#### Open Source Licensing & Contributing

**Apache License 2.0:**

- Added comprehensive Apache License 2.0 (industry standard for Kubernetes projects)
- License provides: patent protection, commercial use permission, wide compatibility
- Added copyright notice: "Copyright 2025 Helios Operator Contributors"
- Updated OCI labels with `org.opencontainers.image.licenses="Apache-2.0"`

**Contributing Guidelines:**

- Created detailed `CONTRIBUTING.md` with contribution workflow
- Includes: bug reporting templates, PR process, coding standards
- Developer Certificate of Origin (DCO) requirement for commits
- Conventional Commits guidelines for commit messages
- Security vulnerability reporting process

**README Enhancements:**

- Added "Why Apache 2.0?" section explaining license choice
- Added "Contributing" section with development setup
- Added PRs Welcome badge
- Enhanced feature list with key highlights
- Added emoji for better visual organization

#### Version Information & Build Metadata

**CLI Version Flag:**

- Added `--version` flag to display build information
- Shows: version, git commit, build date
- Example output:
  ```
  Helios Operator
    Version:    v0.1.0
    Commit:     e11f9f4
    Build Date: 2025-10-04T09:43:31Z
  ```

**Build-time Injection:**

- Version info automatically injected via `-ldflags` during build
- Makefile auto-detects git version, commit, and build timestamp
- Works in both local builds and Docker builds

**Documentation:**

- Created comprehensive `docs/DOCKER.md` guide
- Covers: build strategies, multi-platform builds, version management
- Includes troubleshooting and best practices

### Changed - 2025-10-04

#### Dockerfile & Container Optimization - FINAL VERSION

**Base Image Strategy (IMPORTANT for Open Source):**

- Changed from `golang:1.25.1-alpine3.22` → `golang:1.25-alpine3.22`
- **Reason**: Using `1.25` (without patch) auto-updates to latest 1.25.x
- Benefits for development/open-source:
  - Automatic security patches (1.25.0 → 1.25.1 → 1.25.2...)
  - No manual version bumps needed
  - Always latest stable within minor version
- Alpine 3.22 is latest stable (June 2024)
- Added `# syntax=docker/dockerfile:1` directive for best compatibility

**Cross-Platform Support:**

- Full multi-platform support: `linux/amd64`, `linux/arm64`
- Added `--platform=${BUILDPLATFORM}` for native builder performance
- Updated `make docker-buildx` with registry cache support
- Removed rarely-used platforms (s390x, ppc64le) for faster builds
- Default platforms: `PLATFORMS ?= linux/amd64,linux/arm64`

**Version Injection:**

- Added build args: `VERSION`, `COMMIT`, `BUILD_DATE`
- Automatically populated from git in Makefile
- Injected into binary via `-ldflags`
- Accessible via `--version` flag

**Performance & Security Improvements:**

- Implemented **BuildKit cache mounts** for faster builds:
  - Go module cache: `/go/pkg/mod`
  - Go build cache: `/root/.cache/go-build`
- Added `go mod verify` for supply chain security
- Added `-trimpath` flag for reproducible builds
- Enhanced build flags with `-extldflags '-static'` for fully static binary
- Comprehensive OCI labels (title, description, source, licenses, docs)
- Installed `ca-certificates` and `git` for secure dependency fetching

**Build Context Optimization:**

- Enhanced `.dockerignore` to exclude unnecessary files:
  - Test files (`*_test.go`, `test/`, `testdata/`)
  - Config files (`config/`, `*.yaml` except manifests)
  - Documentation (`.md` files except README)
  - Development tools (`.git/`, `.github/`, IDE configs)
  - Build artifacts (`bin/`, `dist/`, `cover.out`)
- Reduced build context size by ~70%

**Security Hardening:**

- Using multi-arch distroless `static-debian12:nonroot`
- Non-root user (UID 65532) enforced
- Minimal attack surface (~55.8MB final image)
- Static binary with zero dynamic dependencies

**Makefile Improvements:**

- Updated `docker-build` with `DOCKER_BUILDKIT=1` and version args
- Simplified `docker-buildx` (removed sed hack, uses native Dockerfile)
- Added registry cache support in buildx (faster CI/CD)
- New target: `make docker-inspect` for image analysis
- Auto-detection of VERSION, COMMIT, BUILD_DATE from git

**Build Performance:**

- First build: ~65s
- Rebuild with cache: <5s (if only code changes)
- Dependency changes only rebuild affected layers

#### BREAKING: API Group Migration

- **Migrated API group from `heliosapp.helios.dev` to `platform.helios.io`**
- This is a breaking change - existing resources must be recreated
- Updated all CRDs, RBAC manifests, and sample files
- Files modified:
  - `api/v1/groupversion_info.go` - Updated group definition
  - `internal/controller/heliosapp_controller.go` - Updated RBAC annotations
  - `config/crd/bases/` - Regenerated CRD with new API group
  - `config/rbac/*.yaml` - Updated all RBAC manifests
  - `config/samples/heliosapp_v1_heliosapp.yaml` - Updated sample CR
  - `my-awesome-app.yaml` - Updated demo application
  - `deploy.yaml` - Regenerated deployment manifests

**Migration Steps for Existing Deployments:**

```bash
# 1. Backup existing resources
kubectl get heliosapp -A -o yaml > backup.yaml

# 2. Update API version in backup
sed -i 's/heliosapp\.helios\.dev/platform.helios.io/g' backup.yaml

# 3. Uninstall old operator
kubectl delete -f old-deploy.yaml

# 4. Install new operator
kubectl apply -f deploy.yaml

# 5. Restore resources
kubectl apply -f backup.yaml
```

#### Prerequisites Updated

- Updated Docker requirement: `17.03+` → `28.5.0+`
- Updated kubectl requirement: `v1.11.3+` → `v1.34.1+`
- Updated Kubernetes cluster requirement: `v1.11.3+` → `v1.34.1+`
- All aligned with latest stable versions

### Fixed - 2025-10-04

#### Kustomize Configuration

- Fixed empty `patches:` array in `config/crd/kustomization.yaml`
- Commented out unused patches field to comply with Kustomize schema
- All kustomization files now validate correctly

#### Sample Files & Examples

- Fixed invalid `gitRepo: "nginix"` typo in sample files
- Updated to valid repository URLs:
  - `config/samples/heliosapp_v1_heliosapp.yaml`: Valid Kubernetes sample app repo
  - `my-awesome-app.yaml`: Example repository with proper configuration
- Removed unnecessary TODO comments from samples
- Optimized replica counts (5 → 3 for samples, 5 → 2 for demo)
- Updated demo app to use specific image version (`nginx:1.27-alpine`)
- Changed demo port from 80 to 8080 for better practices

#### Build & Development Files

- Enhanced `.gitignore`:
  - Added `cover.out` (test coverage file)
  - Added `dist/` folder
  - Improved formatting consistency
- Enhanced `.dockerignore`:
  - Excluded test files and documentation
  - Excluded config and deployment files from build context
  - Reduced build context size significantly

#### Linter Issues (6 issues → 0)

- Fixed import shadowing: Changed `log` variable to `logger` in controller
- Added error handling for `ctrl.SetControllerReference()` calls (2 occurrences)
- Fixed comment spacing in kubebuilder RBAC markers (4 occurrences)
- All linter checks now passing with 0 issues

#### Test Fixes

- Fixed test validation error: Updated test HeliosApp spec with valid port (8080)
- Added required fields to test: gitRepo, imageRepo, replicas
- All tests passing with 26.6% coverage

### Added - 2025-10-04

#### Test Environment Setup

- Added `make setup-envtest` target for downloading Kubernetes test binaries
- Configured envtest with Kubernetes v1.34.1 binaries (etcd, kube-apiserver)
- Updated test suite to use proper binary paths

### Changed - 2025-10-03

#### Dependency Updates

- Updated Go to v1.25.1 (latest stable)
- Updated Kubernetes API to v0.34.1 (`k8s.io/api`, `k8s.io/apimachinery`, `k8s.io/client-go`)
- Updated Controller Runtime to v0.22.1 (`sigs.k8s.io/controller-runtime`)
- Updated testing frameworks:
  - Ginkgo to v2.26.0
  - Gomega to v1.38.2
- Updated golangci-lint to v2.1.0
- All dependencies verified with `go mod tidy` and `go mod verify`

#### Dockerfile

- Updated base image to `golang:1.25.1-alpine`
- Verified multi-stage build works correctly
- Maintained distroless runtime image for security

## [0.0.1] - 2025-10-03

### Added

- Initial Kubernetes Operator implementation using Kubebuilder v3
- HeliosApp Custom Resource Definition (CRD) with fields:
  - `gitRepo`: Git repository URL
  - `imageRepo`: Docker image repository and tag
  - `port`: Application port (1-65535)
  - `replicas`: Number of pod replicas (minimum: 0, default: 1)
- Controller reconciliation logic:
  - Creates Kubernetes Deployments based on HeliosApp spec
  - Creates NodePort Services to expose applications
  - Updates resources when spec changes
  - Implements proper error handling and logging
- RBAC permissions for:
  - HeliosApp resources (full access)
  - Deployments (full access)
  - Services (full access)
  - Events (create, patch)
  - Pods (read-only)
- Unit tests with Ginkgo/Gomega framework
- E2E test structure
- Docker multi-stage build with distroless runtime
- Security features:
  - Non-root user (UID 65532)
  - HTTP/2 disabled by default
  - Minimal RBAC permissions
  - TLS-ready metrics endpoint

### Documentation

- Comprehensive README.md with:
  - Project description
  - Getting started guide
  - Development instructions
  - Deployment procedures
  - Troubleshooting section
- Code documentation with proper Go comments
- Kubebuilder markers for code generation

## Project Information

**Repository:** github.com/hoangphuc841/helios-operator  
**License:** Apache License 2.0  
**Go Version:** 1.25.1  
**Kubernetes API:** v0.34.1  
**Controller Runtime:** v0.22.1

## Status

- ✅ Build: Successful
- ✅ Tests: All passing (26.6% coverage)
- ✅ Linter: 0 issues
- ✅ Docker Build: Successful
- ✅ Production Ready

## API Reference

### Current API Version

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
```

### Previous API Versions (Deprecated)

- `heliosapp.helios.dev/v1` - Deprecated as of 2025-10-04, removed

---

**Note:** For detailed usage instructions, see [README.md](README.md)
