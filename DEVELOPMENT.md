# Development Guide

This guide covers setting up your development environment and contributing to the Helios Operator.

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Development Environment](#development-environment)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Testing](#testing)
- [CI/CD Pipeline](#cicd-pipeline)
- [Release Process](#release-process)

## Prerequisites

### Required Tools

- **Go**: 1.25.2 or later
- **Docker**: 28.5.0 or later
- **kubectl**: v1.34.1 or later
- **Kind**: v0.25.0 or later
- **Kubebuilder**: v4.3.1 or later

### Optional Tools

- **golangci-lint**: v2.3.0 for linting
- **VS Code**: Recommended IDE with Dev Containers extension

## Development Environment

### Option 1: VS Code Dev Container (Recommended)

The project includes a fully configured Dev Container with all necessary tools pre-installed.

1. **Prerequisites**:

   - Install [Docker](https://docs.docker.com/get-docker/)
   - Install [VS Code](https://code.visualstudio.com/)
   - Install [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)

2. **Open in Dev Container**:

   ```bash
   # Clone the repository
   git clone https://github.com/hoangphuc841/helios-operator.git
   cd helios-operator

   # Open in VS Code
   code .

   # Command Palette (Ctrl+Shift+P / Cmd+Shift+P)
   # Select: "Dev Containers: Reopen in Container"
   ```

3. **What's Included**:

   - ✅ Go 1.25.2 (Alpine-based)
   - ✅ Docker-in-Docker support
   - ✅ Kind for local Kubernetes clusters
   - ✅ Kubebuilder scaffolding tools
   - ✅ kubectl
   - ✅ golangci-lint
   - ✅ All Go development tools (gopls, dlv, ginkgo, etc.)
   - ✅ Zsh with Oh My Zsh
   - ✅ Persistent Go module and build caches

4. **Verify Installation**:
   ```bash
   # Inside the Dev Container
   make help
   go version
   kind version
   kubectl version --client
   ```

### Option 2: Local Setup

If you prefer not to use Dev Containers:

1. **Install Go**:

   ```bash
   # Download and install Go 1.25.2
   wget https://go.dev/dl/go1.25.2.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.25.2.linux-amd64.tar.gz
   export PATH=$PATH:/usr/local/go/bin
   ```

2. **Install Docker**:

   ```bash
   # Follow instructions at https://docs.docker.com/get-docker/
   ```

3. **Install Kind**:

   ```bash
   curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.25.0/kind-linux-amd64
   chmod +x ./kind
   sudo mv ./kind /usr/local/bin/kind
   ```

4. **Install Kubebuilder**:

   ```bash
   curl -L -o kubebuilder https://github.com/kubernetes-sigs/kubebuilder/releases/download/v4.3.1/kubebuilder_linux_amd64
   chmod +x kubebuilder
   sudo mv kubebuilder /usr/local/bin/
   ```

5. **Install kubectl**:

   ```bash
   curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
   chmod +x kubectl
   sudo mv kubectl /usr/local/bin/
   ```

6. **Install golangci-lint**:

   ```bash
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.3.0
   ```

7. **Install Go Tools**:
   ```bash
   go install golang.org/x/tools/gopls@latest
   go install github.com/go-delve/delve/cmd/dlv@latest
   go install golang.org/x/tools/cmd/goimports@latest
   go install github.com/onsi/ginkgo/v2/ginkgo@latest
   go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
   go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
   ```

## Project Structure

```
helios-operator/
├── api/v1/              # API definitions (CRDs)
│   ├── heliosapp_types.go
│   └── ...
├── cmd/                 # Main application entry point
│   └── main.go
├── config/              # Kubernetes manifests & kustomize
│   ├── crd/            # CRD definitions
│   ├── rbac/           # RBAC configurations
│   ├── manager/        # Manager deployment
│   └── samples/        # Sample CRs
├── internal/            # Private application code
│   └── controller/     # Controller implementations
├── test/                # Test files
│   ├── e2e/            # End-to-end tests
│   └── utils/          # Test utilities
├── .devcontainer/      # Dev Container configuration
├── .github/            # GitHub workflows & configs
│   ├── workflows/      # CI/CD pipelines
│   └── dependabot.yml  # Dependency automation
├── Dockerfile          # Multi-stage container build
├── Makefile            # Build & development tasks
└── go.mod              # Go module dependencies
```

## Development Workflow

### 1. Make Changes

```bash
# Create a new feature branch
git checkout -b feature/my-feature

# Make your changes
vim internal/controller/heliosapp_controller.go
```

### 2. Format & Lint

```bash
# Format code
make fmt

# Run linter
make lint

# Auto-fix linting issues
make lint-fix
```

### 3. Generate Code

```bash
# Generate DeepCopy methods
make generate

# Generate CRD manifests
make manifests
```

### 4. Run Tests

```bash
# Run unit tests
make test

# Run with race detector
go test -race ./...

# Run E2E tests (requires Kind cluster)
make test-e2e
```

### 5. Build & Run Locally

```bash
# Build binary
make build

# Run against your K8s cluster
make run

# Build Docker image
make docker-build IMG=helios-operator:dev
```

## Testing

### Unit Tests

Located in `internal/controller/*_test.go`:

```bash
# Run all unit tests
make test

# Run specific test
go test -v ./internal/controller -run TestHeliosAppReconciler

# Generate coverage report
make test
go tool cover -html=cover.out
```

### E2E Tests

Located in `test/e2e/`:

```bash
# Run E2E tests (creates temporary Kind cluster)
make test-e2e

# Run against existing cluster
go test ./test/e2e/ -v -ginkgo.v

# Cleanup test cluster
make cleanup-test-e2e
```

### Test Coverage

- Unit tests: Target **80%+** coverage
- E2E tests: Cover critical user workflows
- Integration tests: Validate K8s API interactions

## CI/CD Pipeline

### GitHub Actions Workflows

Our CI/CD pipeline consists of multiple workflows:

#### 1. **Lint Workflow** (`.github/workflows/lint.yml`)

- Runs on: Every push and PR
- Jobs:
  - `golangci-lint`: Code quality checks
  - `formatting`: Go fmt & vet validation
- Duration: ~5 minutes

#### 2. **Unit Tests** (`.github/workflows/test.yml`)

- Runs on: Every push and PR
- Jobs:
  - `test`: Unit tests with coverage (Go 1.25.2 & 1.25.x)
  - `test-race`: Race condition detection
- Features:
  - ✅ Matrix testing across Go versions
  - ✅ Coverage reports uploaded to Codecov
  - ✅ Artifacts saved for 30 days
- Duration: ~10 minutes

#### 3. **E2E Tests** (`.github/workflows/test-e2e.yml`)

- Runs on: Every push, PR, and daily schedule
- Jobs:
  - `test-e2e`: End-to-end tests on Kind (K8s 1.34.1 & 1.33.0)
  - `test-e2e-summary`: Aggregate results
- Features:
  - ✅ Multi-version Kubernetes testing
  - ✅ Automatic cluster cleanup
  - ✅ Log artifacts on failure
- Duration: ~20 minutes

#### 4. **Security Scanning** (`.github/workflows/security.yml`)

- Runs on: Push to main, PRs, and daily schedule
- Jobs:
  - `govulncheck`: Go vulnerability scanning
  - `trivy-repo`: Repository vulnerability scan
  - `trivy-config`: Configuration security scan
  - `gosec`: Security code analysis
  - `dependency-review`: Dependency security check (PRs only)
  - `secret-scanning`: Secret leak detection
- Duration: ~15 minutes

#### 5. **Container Security** (`.github/workflows/container-security.yml`)

- Runs on: Changes to Dockerfile or code
- Jobs:
  - `build-and-scan`: Build & scan container image
  - `hadolint`: Dockerfile linting
- Features:
  - ✅ Multi-layer image scanning with Trivy
  - ✅ Image size validation (<100MB)
  - ✅ Dockle best practices check
- Duration: ~10 minutes

#### 6. **Release** (`.github/workflows/release.yml`)

- Runs on: Version tags (`v*.*.*`)
- Jobs:
  - `release`: Build multi-arch images & create GitHub release
  - `notify`: Release status notification
- Features:
  - ✅ Multi-platform builds (amd64, arm64)
  - ✅ Automated changelog generation
  - ✅ Release artifacts (install.yaml, deploy.yaml)
  - ✅ Container registry publishing
- Duration: ~20 minutes

### Caching Strategy

All workflows use aggressive caching:

- **Go modules**: `~/.cache/go-mod`
- **Go build cache**: `~/.cache/go-build`
- **Docker layers**: GitHub Actions cache
- **golangci-lint cache**: `~/.cache/golangci-lint`

### Workflow Optimizations

1. **Concurrency Control**: Auto-cancel duplicate runs
2. **Conditional Execution**: Run only when relevant files change
3. **Parallel Jobs**: Independent jobs run concurrently
4. **Minimal Permissions**: Principle of least privilege

## Release Process

### Semantic Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking API changes
- **MINOR**: New features (backwards compatible)
- **PATCH**: Bug fixes

### Creating a Release

1. **Update Version**:

   ```bash
   # Update VERSION in Makefile if needed
   vim Makefile
   ```

2. **Create and Push Tag**:

   ```bash
   # Ensure main branch is up to date
   git checkout main
   git pull origin main

   # Create annotated tag
   git tag -a v1.0.0 -m "Release v1.0.0"

   # Push tag to trigger release workflow
   git push origin v1.0.0
   ```

3. **Automated Release**:

   - GitHub Actions workflow builds multi-arch images
   - Pushes to GitHub Container Registry
   - Creates GitHub Release with:
     - Auto-generated changelog
     - Installation manifests
     - Container image details

4. **Verify Release**:

   ```bash
   # Pull the released image
   docker pull ghcr.io/hoangphuc841/helios-operator:v1.0.0

   # Install in cluster
   kubectl apply -f https://github.com/hoangphuc841/helios-operator/releases/download/v1.0.0/install.yaml
   ```

### Hotfix Process

For urgent production fixes:

1. Create hotfix branch from tag:

   ```bash
   git checkout -b hotfix/v1.0.1 v1.0.0
   ```

2. Make fix and test:

   ```bash
   # Fix the issue
   vim internal/controller/heliosapp_controller.go

   # Test thoroughly
   make test
   make test-e2e
   ```

3. Create new patch version:
   ```bash
   git tag -a v1.0.1 -m "Hotfix: Fix critical bug"
   git push origin v1.0.1
   ```

## Best Practices

### Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting (automatically done by `make fmt`)
- Run `make lint` before committing

### Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `ci`

**Examples**:

```
feat(controller): add support for custom annotations
fix(webhook): validate port range correctly
docs: update installation guide
ci: add codecov integration
```

### Pull Requests

1. Create feature branch from `main`
2. Make atomic commits with clear messages
3. Ensure all tests pass locally
4. Update documentation if needed
5. Create PR with description of changes
6. Request review from maintainers

### Security

- Never commit secrets or credentials
- Use environment variables for sensitive data
- Run security scans before releasing
- Keep dependencies up to date

## Troubleshooting

### Dev Container Issues

```bash
# Rebuild container
# Command Palette -> "Dev Containers: Rebuild Container"

# Check container logs
docker logs <container-id>
```

### Test Failures

```bash
# Verbose test output
go test -v ./...

# Run specific test with race detector
go test -race -run TestHeliosAppReconciler ./internal/controller
```

### Kind Cluster Issues

```bash
# List clusters
kind get clusters

# Delete cluster
kind delete cluster --name helios-operator-test-e2e

# View cluster logs
kind export logs /tmp/kind-logs
```

## Additional Resources

- [Kubebuilder Book](https://book.kubebuilder.io/)
- [Kubernetes Operator Pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
- [Controller Runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime)
- [Contributing Guidelines](CONTRIBUTING.md)

## Need Help?

- 📝 [Open an issue](https://github.com/hoangphuc841/helios-operator/issues)
- 💬 [Start a discussion](https://github.com/hoangphuc841/helios-operator/discussions)
- 📧 Contact maintainers
