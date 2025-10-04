# Helios Operator

[![Go Version](https://img.shields.io/badge/Go-1.25.1-blue.svg)](https://golang.org/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-v1.34.1-blue.svg)](https://kubernetes.io/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

A Kubernetes Operator for automating application deployment and management through custom resources.

## 📋 Overview

Helios Operator simplifies application deployment on Kubernetes by watching `HeliosApp` custom resources and automatically creating and managing the corresponding Deployments and Services.

**Key Features:**

- 🚀 Automated deployment management
- ⚡ Declarative application configuration
- 🔄 GitOps-ready workflow
- 🛡️ Production-ready security (distroless, non-root)
- 🌐 Multi-platform support (amd64, arm64)
- 📦 Minimal footprint (~55MB container image)

## Prerequisites

- Go 1.25.1+
- Docker 28.5.0+
- kubectl v1.34.1+
- Kubernetes cluster v1.34.1+

## Quick Start

### Install the Operator

```bash
# Install CRDs
make install

# Deploy the operator
make deploy IMG=<your-registry>/helios-operator:latest
```

### Create Your First Application

```yaml
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: my-app
spec:
  gitRepo: https://github.com/example/my-app
  imageRepo: nginx:latest
  port: 80
  replicas: 3
```

```bash
kubectl apply -f my-app.yaml
```

The operator automatically creates:

- ✅ Kubernetes Deployment with your image
- ✅ NodePort Service to expose the application

## HeliosApp Specification

| Field       | Type   | Required | Description                |
| ----------- | ------ | -------- | -------------------------- |
| `gitRepo`   | string | Yes      | Git repository URL         |
| `imageRepo` | string | Yes      | Container image (repo:tag) |
| `port`      | int32  | Yes      | Application port (1-65535) |
| `replicas`  | int32  | No       | Replica count (default: 1) |

## Development

### First Time Setup

```bash
# Install development tools (controller-gen, kustomize, etc.)
make setup-envtest

# Verify installation
make version
```

### Development Workflow

```bash
# 1. Format code
make fmt

# 2. Run linter
make lint

# 3. Run tests (fast)
make test

# 4. Build binary
make build

# 5. Run controller locally (outside cluster)
make run
```

### Makefile Targets Reference

#### Code Quality

- `make fmt` - Format Go code with gofmt
- `make vet` - Run go vet for static analysis
- `make lint` - Run golangci-lint (comprehensive linting)
- `make lint-fix` - Auto-fix linting issues where possible

#### Building

- `make build` - Build manager binary (`bin/manager`)
- `make manifests` - Generate CRD, RBAC manifests
- `make generate` - Generate deepcopy code
- `make docker-build` - Build Docker image
- `make docker-push` - Push Docker image to registry
- `make docker-buildx` - Multi-platform build (amd64, arm64)

#### Testing

- `make test` - Run all tests with coverage
- `make test-e2e` - Run end-to-end tests (requires Kind)
- `make setup-envtest` - Download Kubernetes test binaries

#### Deployment

- `make install` - Install CRDs to cluster
- `make uninstall` - Remove CRDs from cluster
- `make deploy` - Deploy controller to cluster
- `make undeploy` - Remove controller from cluster

#### Utilities

- `make help` - Show all available targets
- `make build-installer` - Generate single YAML installer

### Build Docker Image

```bash
# Build for local platform
make docker-build IMG=<your-registry>/helios-operator:tag

# Build and push for multiple platforms
make docker-buildx IMG=<your-registry>/helios-operator:tag PLATFORMS=linux/amd64,linux/arm64

# Push image
make docker-push IMG=<your-registry>/helios-operator:tag
```

### Performance Tips

**Fast Iteration:**

```bash
# Quick check before commit
make fmt vet lint-fix

# Fast tests only
make test
```

**Full Validation:**

```bash
# Complete checks (use before PR)
make manifests generate fmt vet lint test
```

### IDE Integration

**VS Code:**

```json
// .vscode/tasks.json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "make build",
      "type": "shell",
      "command": "make build",
      "group": {
        "kind": "build",
        "isDefault": true
      }
    },
    {
      "label": "make test",
      "type": "shell",
      "command": "make test",
      "group": {
        "kind": "test",
        "isDefault": true
      }
    }
  ]
}
```

**GoLand/IntelliJ:**

- Right-click `Makefile` → "Run 'make <target>'"
- Or use Terminal: `make <target>`

## Common Operations

### Manage Applications

```bash
# List applications
kubectl get heliosapp

# Describe an application
kubectl describe heliosapp my-app

# Scale application
kubectl patch heliosapp my-app -p '{"spec":{"replicas":5}}'

# Update application
kubectl edit heliosapp my-app

# Delete application
kubectl delete heliosapp my-app
```

### Monitor Operator

```bash
# View operator logs
kubectl logs -n helios-operator-system deployment/helios-operator-controller-manager -f

# Check operator status
kubectl get deployment -n helios-operator-system
```

## Uninstall

```bash
# Delete all applications
kubectl delete heliosapp --all

# Uninstall operator
make undeploy

# Remove CRDs
make uninstall
```

## Architecture

The operator implements the following workflow:

1. **Watch** `HeliosApp` custom resources
2. **Create** Kubernetes Deployments with specified configuration
3. **Manage** NodePort Services for application exposure
4. **Update** resources when specifications change
5. **Scale** applications by adjusting replica counts

### RBAC Permissions

- `platform.helios.io/heliosapps` - Full access
- `apps/deployments` - Full access
- `core/services` - Full access
- `core/events` - Create, patch
- `core/pods` - Read-only

## Troubleshooting

### Common Issues

#### Tests fail with "no such file or directory"

```bash
# Download Kubernetes test binaries
make setup-envtest
```

#### CRD not found when deploying

```bash
# Install CRDs first
make install

# Then deploy operator
make deploy IMG=<your-image>
```

#### Build failures

```bash
# Clean and rebuild
go mod tidy
go mod verify
make build
```

#### CRD changes not applying

```bash
# Regenerate manifests and code
make manifests generate

# Reinstall CRDs
make uninstall
make install
```

#### Linter issues

```bash
# Auto-fix common issues
make lint-fix

# Manual fixes may be needed for some issues
make lint
```

#### Docker build fails

```bash
# Enable BuildKit (recommended)
export DOCKER_BUILDKIT=1
make docker-build

# Or use podman
make docker-build CONTAINER_TOOL=podman
```

#### E2E tests fail

```bash
# Ensure Kind is installed
kind version

# Setup test cluster
make setup-test-e2e

# Run E2E tests
make test-e2e
```

### Debugging Tips

**Controller not reconciling:**

```bash
# Check controller logs
kubectl logs -n helios-operator-system deployment/helios-operator-controller-manager -f

# Check RBAC permissions
kubectl auth can-i --list --as=system:serviceaccount:helios-operator-system:helios-operator-controller-manager

# Describe the custom resource
kubectl describe heliosapp <name>
```

**Resource not being created:**

```bash
# Check events
kubectl get events --sort-by='.lastTimestamp'

# Check controller conditions
kubectl get heliosapp <name> -o yaml | grep conditions -A 10
```

**Performance issues:**

```bash
# Check resource limits
kubectl top pods -n helios-operator-system

# Enable profiling (development only)
# Add --profiler-address=:8082 to manager args
kubectl port-forward -n helios-operator-system deployment/helios-operator-controller-manager 8082:8082
# Visit http://localhost:8082/debug/pprof/
```

## Contributing

We welcome contributions! Please follow these steps:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes and add tests
4. Ensure all checks pass: `make test lint`
5. Commit: `git commit -m 'Add my feature'`
6. Push: `git push origin feature/my-feature`
7. Open a Pull Request

### Development Guidelines

- Follow Go best practices
- Add tests for new features
- Run `make fmt` before committing
- Ensure `make lint` shows 0 issues
- Update documentation as needed

## Project Status

| Component | Status            |
| --------- | ----------------- |
| Build     | ✅ Passing        |
| Tests     | ✅ 26.6% coverage |
| Linter    | ✅ 0 issues       |
| Docker    | ✅ Working        |

## 📄 License

This project is licensed under the **Apache License 2.0** - see the [LICENSE](LICENSE) file for details.

### Why Apache 2.0?

Apache License 2.0 is the industry standard for cloud-native and Kubernetes projects:

- ✅ **Permissive**: Allows commercial and private use
- ✅ **Patent protection**: Explicit patent grant from contributors
- ✅ **Widely adopted**: Used by Kubernetes, Docker, Terraform, and most CNCF projects
- ✅ **Enterprise-friendly**: Compatible with corporate policies
- ✅ **Community-focused**: Encourages contributions while protecting contributors

### Copyright Notice

```
Copyright 2025 Helios Operator Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

## 🤝 Contributing

We welcome contributions! This is an open-source project and we appreciate:

- 🐛 Bug reports and feature requests via [GitHub Issues](https://github.com/hoangphuc841/helios-operator/issues)
- 💡 Code contributions via Pull Requests
- 📝 Documentation improvements
- ⭐ Stars on GitHub to show your support!

Please read our [Contributing Guide](CONTRIBUTING.md) before submitting a PR.

### Development Setup

```bash
# Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/helios-operator.git
cd helios-operator

# Install dependencies
make build

# Run tests
make test

# Run linter
make lint
```

## 📚 Resources

## Resources

### Documentation

- [Changelog](CHANGELOG.md) - Version history and changes
- [Contributing Guide](CONTRIBUTING.md) - How to contribute
- [Kubebuilder Book](https://book.kubebuilder.io/) - Operator development guide
- [Controller Runtime](https://pkg.go.dev/sigs.k8s.io/controller-runtime) - Framework documentation

### API Reference

- [Kubernetes API](https://kubernetes.io/docs/reference/kubernetes-api/) - K8s API reference
- [HeliosApp CRD](config/crd/bases/platform.helios.io_heliosapps.yaml) - Custom Resource Definition

### Tools & Dependencies

- [Go 1.25.1](https://golang.org/) - Programming language
- [Docker](https://docs.docker.com/) - Container runtime
- [Kind](https://kind.sigs.k8s.io/) - Kubernetes in Docker (for testing)
- [Kustomize](https://kustomize.io/) - Kubernetes configuration management
- [golangci-lint](https://golangci-lint.run/) - Go linter aggregator

### Related Projects

- [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) - SDK for building K8s operators
- [Operator SDK](https://sdk.operatorframework.io/) - Alternative operator framework
- [CNCF Operators](https://operatorhub.io/) - Community operator hub

---

**Made with ❤️ using Kubebuilder**
