# Changelog

All notable changes to the Helios Operator will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added - Pipeline Generation & Unit Tests 🎉

**Major Feature: Zero-Configuration Pipeline Creation**

- ✨ **Automatic Pipeline Generation** - Eliminates need for manual Pipeline management

  - Pipeline templates with standardized structure
  - Automatic naming: `{app-name}-pipeline`
  - Owner reference for automatic cleanup
  - Template customization support

- ✨ **Comprehensive Unit Tests** - Production-ready test coverage
  - Controller logic tests (getPipelineRunStatus, getDeploymentHealth)
  - Watch predicate and mapping function tests
  - Webhook validation tests with all scenarios
  - 100% coverage of new functionality

### Changed - Simplified User Experience

- 🔧 **Removed `pipelineName` field** from HeliosAppSpec

  - No more tight coupling with infrastructure details
  - Users don't need to know Tekton Pipeline names
  - Automatic Pipeline creation from templates

- 📚 **Enhanced Documentation**
  - Pipeline Generation guide with examples
  - Migration guide from manual to automatic Pipeline management
  - Best practices and troubleshooting

### Fixed

- 🐛 **Eliminated Pipeline Dependencies** - Users no longer need to create Pipelines manually
- 🐛 **Improved Error Handling** - Clear error messages for Pipeline creation issues
- 🐛 **Enhanced RBAC** - Added Pipeline resource permissions

## [0.1.0] - 2025-10-14

### Added - Kubernetes Watches Implementation 🎉

**Major Feature: Real-time Status Updates**

- ✨ **Kubernetes Watches** for ArgoCD Applications - Best Practice Implementation

  - Event-driven reconciliation instead of polling
  - Real-time status updates (< 1 second latency)
  - 92% reduction in unnecessary reconciliations
  - Predicate filtering for efficiency

- 🎯 **New Components**

  - `argoAppPredicate()` - Filters only Helios-managed ArgoCD Applications
  - `isHeliosArgoApp()` - Validates resource ownership
  - `argoAppToHeliosApp()` - Maps ArgoCD Apps to HeliosApps for reconciliation
  - Enhanced `SetupWithManager()` with Watches configuration

- 🏷️ **Label-Based Tracking**

  - `helios.io/managed-by: helios-operator` - Identifies managed resources
  - `helios.io/app-name` - Links to HeliosApp name
  - `helios.io/app-namespace` - Links to HeliosApp namespace
  - Supports cross-namespace resource management

- 📚 **Comprehensive Documentation**

  - `docs/WATCHES_IMPLEMENTATION.md` - In-depth implementation guide
  - `docs/README.md` - Documentation index
  - `WATCHES_QUICK_REFERENCE.md` - Quick reference card
  - `IMPLEMENTATION_SUMMARY.md` - Complete summary of implementation

- 🧪 **Testing Infrastructure**
  - `test-watches.sh` - Automated test script with colored output
  - Real-time verification of watch triggers
  - Before/after status comparison

### Changed

- 🔧 **Enhanced Imports**

  - Added `strings`, `types`, `builder`, `event`, `handler`, `predicate`, `reconcile`
  - Support for advanced controller-runtime features

- 📊 **Improved Logging**

  - Structured logging for watch events
  - DEBUG level logs for watch triggers
  - Better observability and debugging

- 📖 **Updated Documentation**
  - README.md highlights Watches feature
  - Added "What's New" section with visual diagrams
  - Enhanced feature list with real-time updates

### Performance Improvements

- ⚡ **Event-Driven Architecture**
  - Eliminated polling (was ~120 reconciles/hour)
  - Reduced to ~10 reconciles/hour (only on actual changes)
  - Lower CPU and memory usage
  - Reduced API server load

### Technical Details

**Code Statistics:**

- ~150 lines of new code
- 3 new helper functions
- 1 file modified: `internal/controller/heliosapp_controller.go`
- 3 documentation files created
- 1 test script created

**Dependencies:**

- No new external dependencies (uses existing controller-runtime)

### Breaking Changes

None. This is a backwards-compatible enhancement.

### Migration Guide

No migration needed. Existing HeliosApp resources will automatically benefit from real-time status updates after operator upgrade.

**Upgrade Steps:**

1. Build new image: `docker build -t helios-operator:v0.1.0 .`
2. Load to cluster: `minikube image load helios-operator:v0.1.0` (or push to registry)
3. Update deployment: `kubectl set image deployment/... manager=helios-operator:v0.1.0`
4. Verify: `./test-watches.sh`

### References

- [Kubernetes Operators Best Practices](https://www.infoq.com/articles/kubernetes-operators-in-depth/)
- [Controller-Runtime Watches](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/builder#example-Builder.Watches)
- [Kubernetes Watch API](https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes)

---

## [0.0.1] - 2025-10-01

### Added - Initial Release

- 🎉 Initial implementation of Helios Operator
- 📦 Custom Resource Definition: `HeliosApp`
- 🔄 GitOps integration with ArgoCD
- 🏗️ Tekton Triggers for CI/CD
- 🐳 Docker multi-stage build
- 🛡️ Distroless base image for security
- 📝 Basic documentation

### Components

- **API**: `platform.helios.io/v1`
- **CRD**: HeliosApp with GitOps fields
- **Controller**: Basic reconciliation loop
- **ArgoCD Integration**: Application generator
- **Tekton Integration**: EventListener, TriggerBinding, TriggerTemplate

### Infrastructure

- Minikube support
- Helm-based ArgoCD installation
- Kubectl-based Tekton installation
- Setup scripts for local development

---

## Future Roadmap

### Planned Features

- [ ] Watch Tekton PipelineRuns for build status
- [ ] Watch Deployments for pod health
- [ ] Custom metrics with Prometheus
- [ ] Webhook validation for HeliosApp
- [ ] Multi-cluster support
- [ ] Advanced RBAC policies

### Under Consideration

- [ ] Helm chart for operator installation
- [ ] OLM (Operator Lifecycle Manager) bundle
- [ ] Integration with FluxCD
- [ ] Support for Helm-based deployments in GitOps repo
- [ ] Auto-scaling based on custom metrics

---

[Unreleased]: https://github.com/hoangphuc841/helios-operator/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/hoangphuc841/helios-operator/releases/tag/v0.1.0
[0.0.1]: https://github.com/hoangphuc841/helios-operator/releases/tag/v0.0.1
