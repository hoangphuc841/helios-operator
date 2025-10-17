# 🛠️ Makefile Guide

This guide explains the optimized Makefile for Helios Operator, designed specifically for Cloud-Native Internal Developer Platform (IDP) workflows.

## 🎯 **Design Philosophy**

The Helios Makefile follows these principles:

- **Focused on IDP Domain**: Targets are tailored for GitOps, Tekton, and ArgoCD workflows
- **Simple and Clear**: Easy to understand and use for both developers and operators
- **Best Practices**: Follows modern Makefile conventions and Cloud-Native standards
- **Minimal Dependencies**: Only essential tools required for the core workflow

## 📋 **Available Commands**

### Core Commands

| Command          | Description                            |
| ---------------- | -------------------------------------- |
| `make help`      | Display help with organized categories |
| `make build`     | Build the Helios operator binary       |
| `make run`       | Run the operator locally               |
| `make manifests` | Generate CRDs and RBAC manifests       |
| `make generate`  | Generate DeepCopy code                 |

### Development Commands

| Command          | Description                            |
| ---------------- | -------------------------------------- |
| `make dev-setup` | Complete development environment setup |
| `make fmt`       | Format Go code                         |
| `make vet`       | Run go vet                             |
| `make lint`      | Run linter                             |
| `make test`      | Run unit tests                         |

### GitOps & Deploy Commands

| Command             | Description                  |
| ------------------- | ---------------------------- |
| `make install`      | Install Helios CRDs          |
| `make deploy`       | Deploy Helios operator       |
| `make deploy-local` | Deploy to local Kind cluster |
| `make undeploy`     | Undeploy Helios operator     |

### Testing Commands

| Command              | Description                              |
| -------------------- | ---------------------------------------- |
| `make test-e2e`      | Run end-to-end tests for GitOps workflow |
| `make test-coverage` | Run tests with coverage                  |

### Security Commands

| Command                  | Description                                |
| ------------------------ | ------------------------------------------ |
| `make validate-security` | Validate Pod Security Standards compliance |

### Container Commands

| Command             | Description                            |
| ------------------- | -------------------------------------- |
| `make docker-build` | Build Docker image for Helios operator |

### Helm Commands

| Command             | Description                             |
| ------------------- | --------------------------------------- |
| `make helm-package` | Package Helm chart for IDP distribution |
| `make helm-lint`    | Lint Helm chart                         |
| `make helm-deploy`  | Install Helm chart                      |

### Utility Commands

| Command       | Description                  |
| ------------- | ---------------------------- |
| `make status` | Check Helios operator status |
| `make clean`  | Clean build artifacts        |
| `make tools`  | Install essential tools      |

## 🚀 **Quick Start Workflow**

### 1. Development Setup

```bash
# Complete development environment setup
make dev-setup

# This will:
# - Install required tools (kustomize, controller-gen, golangci-lint)
# - Generate manifests and code
# - Build the operator binary
```

### 2. Local Development

```bash
# Run operator locally
make run

# Or build and run in cluster
make docker-build
make deploy-local
```

### 3. Testing

```bash
# Run unit tests
make test

# Run E2E tests
make test-e2e

# Run with coverage
make test-coverage
```

### 4. Deployment

```bash
# Deploy to cluster
make deploy

# Or via Helm
make helm-deploy
```

## 🔧 **Configuration**

### Environment Variables

| Variable       | Default                                   | Description          |
| -------------- | ----------------------------------------- | -------------------- |
| `VERSION`      | `$(git describe --tags --always --dirty)` | Version tag          |
| `COMMIT`       | `$(git rev-parse --short HEAD)`           | Git commit hash      |
| `IMG`          | `helios.dev/helios-operator:$(VERSION)`   | Container image      |
| `NAMESPACE`    | `helios-system`                           | Kubernetes namespace |
| `KIND_CLUSTER` | `helios-test`                             | Kind cluster name    |

### Custom Configuration

```bash
# Use custom image
make docker-build IMG=my-registry/helios-operator:v1.0.0

# Deploy to custom namespace
make deploy NAMESPACE=my-namespace

# Use custom Kind cluster
make deploy-local KIND_CLUSTER=my-cluster
```

## 🛠️ **Tool Management**

The Makefile automatically manages essential tools:

- **kustomize**: Kubernetes configuration management
- **controller-gen**: Generate CRD and RBAC manifests
- **golangci-lint**: Go code linting

Tools are installed to `bin/` directory and versioned for reproducibility.

## 📊 **Best Practices**

### 1. Use Semantic Versioning

```bash
# Tag releases properly
git tag v1.0.0
make docker-build  # Uses tag as version
```

### 2. Clean Development

```bash
# Clean before major changes
make clean
make dev-setup
```

### 3. Test Before Deploy

```bash
# Always test locally first
make test
make test-e2e
make deploy-local
```

### 4. Use Help System

```bash
# Get organized help
make help

# Get help for specific category
make help | grep "Development"
```

## 🔄 **CI/CD Integration**

The Makefile is designed for seamless CI/CD integration:

```yaml
# Example GitHub Actions workflow
- name: Build and Test
  run: |
    make dev-setup
    make test
    make test-e2e
    make docker-build
```

## 🧹 **Cleanup**

```bash
# Clean build artifacts
make clean

# Remove all generated files
rm -rf bin/ dist/ coverage.out coverage.html
```

## 🆘 **Troubleshooting**

### Common Issues

1. **Tool Installation Fails**

   ```bash
   # Check Go version (requires 1.25+)
   go version

   # Clean and retry
   rm -rf bin/
   make tools
   ```

2. **Docker Build Fails**

   ```bash
   # Check Docker daemon
   docker info

   # Try with different image name
   make docker-build IMG=localhost/helios-operator:dev
   ```

3. **Deploy Fails**

   ```bash
   # Check cluster connection
   kubectl cluster-info

   # Check namespace
   kubectl get namespace helios-system
   ```

## 📚 **Additional Resources**

- [Makefile Best Practices](https://docs.cloudposse.com/best-practices/developer/makefile/)
- [Kubebuilder Documentation](https://book.kubebuilder.io/)
- [Helios Architecture Guide](01-architecture.md)

---

**Need help?** Check our [Troubleshooting Guide](../user-guide/04-troubleshooting.md) or [Development Setup Guide](02-development-setup.md).
