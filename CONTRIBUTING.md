# Contributing to Helios Operator

Thank you for your interest in contributing to Helios Operator! 🎉

We welcome contributions from everyone, whether you're fixing a typo, reporting a bug, or proposing a major feature.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Commit Message Guidelines](#commit-message-guidelines)
- [License](#license)

## Code of Conduct

This project adheres to a simple code of conduct:

- **Be respectful**: Treat everyone with respect and kindness
- **Be collaborative**: Help others and ask for help when needed
- **Be constructive**: Provide helpful feedback and accept it gracefully
- **Be inclusive**: Welcome contributions from all backgrounds and skill levels

## How Can I Contribute?

### Reporting Bugs

Before creating a bug report, please check existing issues to avoid duplicates.

**Great bug reports include:**

- Clear, descriptive title
- Steps to reproduce the issue
- Expected vs actual behavior
- Environment details (Go version, Kubernetes version, OS)
- Logs or screenshots if applicable

**Template:**

```markdown
**Description:**
Brief description of the issue

**Steps to Reproduce:**

1. Step one
2. Step two
3. ...

**Expected Behavior:**
What you expected to happen

**Actual Behavior:**
What actually happened

**Environment:**

- Go version: 1.25.1
- Kubernetes version: v1.34.1
- OS: Ubuntu 22.04
- Helios Operator version: v0.1.0

**Additional Context:**
Any other relevant information
```

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues.

**Great enhancement suggestions include:**

- Clear use case and motivation
- Detailed description of the proposed solution
- Alternative solutions you've considered
- Impact on existing functionality

### Contributing Code

We love code contributions! Here's how to get started:

1. **Fork the repository**
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Make your changes**
4. **Write/update tests**
5. **Run tests and linter** (`make test lint`)
6. **Commit your changes** (follow commit guidelines below)
7. **Push to your fork** (`git push origin feature/amazing-feature`)
8. **Open a Pull Request**

### Improving Documentation

Documentation improvements are always welcome:

- Fixing typos or grammatical errors
- Adding examples or clarifications
- Translating documentation (if applicable)
- Creating tutorials or guides

## Development Setup

### Prerequisites

- Go 1.25.1+
- Docker 28.5.0+
- kubectl v1.34.1+
- Kubernetes cluster (kind, minikube, or cloud provider)

### Clone and Build

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/helios-operator.git
cd helios-operator

# Add upstream remote
git remote add upstream https://github.com/hoangphuc841/helios-operator.git

# Install dependencies and build
make build

# Run tests
make test

# Run linter
make lint
```

### Running Locally

```bash
# Install CRDs into your cluster
make install

# Run the controller locally (against the cluster in ~/.kube/config)
make run

# In another terminal, create a sample resource
kubectl apply -f config/samples/heliosapp_v1_heliosapp.yaml
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test
cat cover.out  # View coverage report

# Run specific test
go test -v ./internal/controller -run TestHeliosAppReconciler
```

### Building Docker Image

```bash
# Build for your architecture
make docker-build IMG=myregistry/helios-operator:v0.1.0

# Build for multiple platforms
make docker-buildx IMG=myregistry/helios-operator:v0.1.0
```

## Pull Request Process

### Before Submitting

- [ ] Code builds successfully (`make build`)
- [ ] All tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Documentation updated (if needed)
- [ ] CHANGELOG.md updated (if applicable)
- [ ] Commits follow commit message guidelines

### PR Description

**Good PR descriptions include:**

```markdown
## What does this PR do?

Brief description of changes

## Why is this needed?

Context and motivation

## Related Issues

Fixes #123
Relates to #456

## Testing

How you tested the changes

## Screenshots (if applicable)

Before/after screenshots

## Checklist

- [ ] Tests added/updated
- [ ] Documentation updated
- [ ] CHANGELOG updated
```

### Review Process

1. **Automated checks**: CI/CD will run tests and linters
2. **Code review**: Maintainers will review your code
3. **Feedback**: Address any requested changes
4. **Approval**: Once approved, maintainers will merge

**What we look for:**

- Code quality and clarity
- Test coverage
- Documentation
- Backward compatibility
- Performance impact

## Coding Standards

### Go Style

Follow standard Go conventions:

- Use `gofmt` for formatting (automated by `make fmt`)
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use meaningful variable names
- Add comments for exported functions
- Keep functions focused and small

### Kubebuilder Markers

Use appropriate kubebuilder markers:

```go
// +kubebuilder:rbac:groups=platform.helios.io,resources=heliosapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=platform.helios.io,resources=heliosapps/status,verbs=get;update;patch
```

### Error Handling

Always handle errors appropriately:

```go
// Good
if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
    if apierrors.IsNotFound(err) {
        return ctrl.Result{}, nil
    }
    return ctrl.Result{}, err
}

// Bad
_ = r.Get(ctx, req.NamespacedName, &app)
```

### Testing

Write table-driven tests:

```go
func TestReconcile(t *testing.T) {
    tests := []struct {
        name    string
        app     *heliosappv1.HeliosApp
        want    ctrl.Result
        wantErr bool
    }{
        {
            name: "successful reconciliation",
            app:  makeApp("test-app"),
            want: ctrl.Result{},
            wantErr: false,
        },
        // more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

## Commit Message Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/):

### Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding/updating tests
- `chore`: Maintenance tasks
- `perf`: Performance improvements
- `ci`: CI/CD changes

### Examples

```bash
# Feature
feat(controller): add support for custom labels

# Bug fix
fix(reconciler): handle nil pointer in deployment creation

# Documentation
docs(readme): update installation instructions

# Multiple changes
feat(api): add new field to HeliosApp spec

Add imageRegistry field to allow custom image registries.
Update CRD and regenerate manifests.

Fixes #123
```

### Best Practices

- Use imperative mood ("add feature" not "added feature")
- Keep subject line under 50 characters
- Wrap body at 72 characters
- Reference issues in footer

## License

By contributing to Helios Operator, you agree that your contributions will be licensed under the [Apache License 2.0](LICENSE).

### Developer Certificate of Origin (DCO)

All commits must be signed off, indicating that you agree to the [DCO](https://developercertificate.org/):

```bash
git commit -s -m "feat: add amazing feature"
```

This adds a `Signed-off-by` line to your commit message:

```
feat: add amazing feature

Signed-off-by: Your Name <your.email@example.com>
```

## 🎉 Recognition

Contributors will be recognized in:

- GitHub contributors page
- Release notes (for significant contributions)
- Project README (for major features)

## 📞 Getting Help

Need help? Here's where to reach us:

- **GitHub Issues**: For bugs and feature requests
- **GitHub Discussions**: For questions and discussions
- **Documentation**: Check [docs/](docs/) folder

## 🙏 Thank You!

Your contributions make Helios Operator better for everyone. We appreciate your time and effort!

Happy coding! 🚀
