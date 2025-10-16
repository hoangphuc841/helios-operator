# Helios Operator - Refactoring Checklist

> This is a practical, actionable checklist for implementing the refactoring plan. Mark items as you complete them.

## 📋 How to Use This Checklist

1. Work through items in order (they build on each other)
2. Check off `[ ]` → `[x]` when complete
3. Test after each section before moving to the next
4. Commit changes frequently

---

## Phase 1: Critical Fixes (Week 1) 🚨

### 1.1 Fix Compilation Errors

- [ ] Create `internal/resources/tekton/` package
- [ ] Move pipeline generation logic to `internal/resources/tekton/pipeline.go`
- [ ] Move trigger generation logic to `internal/resources/tekton/trigger.go`
- [ ] Create `internal/resources/argocd/` package
- [ ] Move ArgoCD generation to `internal/resources/argocd/application.go`
- [ ] Update imports in `internal/controller/heliosapp_controller.go`
- [ ] Run `go build ./cmd/main.go` - should succeed
- [ ] Run `make build` - should succeed

### 1.2 Fix Helm Chart

- [ ] Open `helm/helios-operator/values.yaml`
- [ ] Find and remove duplicate `podDisruptionBudget` key (around line 68)
- [ ] Run `helm lint helm/helios-operator` - should pass
- [ ] Run `helm template helm/helios-operator | kubectl apply --dry-run=client -f -`

### 1.3 Fix Tests

- [ ] Run `make test` and note failures
- [ ] Fix any test import errors
- [ ] Ensure all tests pass
- [ ] Check test coverage: `go test ./... -coverprofile=coverage.out`

### 1.4 Fix Linting

- [ ] Run `make lint` and note issues
- [ ] Fix all critical linting errors
- [ ] Run `make vet` - should pass
- [ ] Run `make fmt` to format code

### Validation Checkpoint

```bash
make manifests generate
make build
make test
make lint
make vet
```

All commands should succeed before proceeding.

---

## Phase 2: Code Quality (Week 2) 🔧

### 2.1 Refactor Controller

- [ ] Create `internal/controller/reconcile_helpers.go`
- [ ] Extract `fetchHeliosApp()` method
- [ ] Extract `reconcilePVC()` method
- [ ] Extract `reconcilePipeline()` method
- [ ] Extract `reconcileTriggers()` method
- [ ] Extract `reconcileArgoCD()` method
- [ ] Extract `updateStatus()` method
- [ ] Update main `Reconcile()` to call these methods
- [ ] Run tests after each extraction

### 2.2 Improve Error Handling

- [ ] Create `internal/common/errors.go`
- [ ] Define custom error types
- [ ] Update controller to use custom errors
- [ ] Add error wrapping with context
- [ ] Add retry logic for transient errors

### 2.3 Enhance Logging

- [ ] Review all log statements
- [ ] Add structured fields to logs
- [ ] Use appropriate log levels
- [ ] Add correlation IDs
- [ ] Test log output readability

### 2.4 Add Input Validation

- [ ] Review `api/v1/heliosapp_webhook.go`
- [ ] Add Git URL validation
- [ ] Add image repository validation
- [ ] Add ServiceAccount existence check
- [ ] Add webhook secret validation
- [ ] Add unit tests for validation

### 2.5 Improve Resource Management

- [ ] Add owner references to all created resources
- [ ] Implement proper finalizer cleanup
- [ ] Test resource deletion
- [ ] Test garbage collection

### Validation Checkpoint

```bash
make test
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
# Should show >70% coverage
```

---

## Phase 3: Testing (Week 2-3) 🧪

### 3.1 Unit Tests

- [ ] Identify functions without tests
- [ ] Add tests for all public functions
- [ ] Add tests for error paths
- [ ] Add table-driven tests for complex logic
- [ ] Achieve >80% coverage for `internal/` packages

### 3.2 Integration Tests

- [ ] Review existing integration tests
- [ ] Add tests for complete reconciliation flow
- [ ] Add tests for resource ownership
- [ ] Add tests for status updates
- [ ] Add tests for error recovery

### 3.3 E2E Tests

- [ ] Review `test/e2e/e2e_test.go`
- [ ] Add complete workflow test
- [ ] Add multi-application test
- [ ] Add failure recovery test
- [ ] Document E2E test setup

### Validation Checkpoint

```bash
make test
make test-e2e
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
# Should show >80% coverage
```

---

## Phase 4: Documentation (Week 3) 📚

### 4.1 Code Documentation

- [ ] Add package docs to all packages in `internal/`
- [ ] Add package docs to `api/v1`
- [ ] Add docs to all exported types
- [ ] Add docs to all exported functions
- [ ] Add examples in docs where helpful

### 4.2 API Documentation

- [ ] Review `api/v1/heliosapp_types.go`
- [ ] Enhance field documentation
- [ ] Add examples for each field
- [ ] Document all status conditions
- [ ] Generate API reference

### 4.3 User Guides

- [ ] Update `docs/user-guide/01-getting-started.md`
  - [ ] Add prerequisites checklist
  - [ ] Add step-by-step installation
  - [ ] Add verification steps
- [ ] Update `docs/user-guide/02-helios-app-spec.md`
  - [ ] Complete field reference
  - [ ] Add real-world examples
  - [ ] Document defaults
- [ ] Update `docs/user-guide/03-troubleshooting.md`
  - [ ] Add common errors and solutions
  - [ ] Add debug commands
  - [ ] Add FAQ

### 4.4 Developer Guides

- [ ] Update `docs/developer-guide/01-architecture.md`
  - [ ] Fix code block language at line 237
  - [ ] Add detailed diagrams
  - [ ] Document data flow
- [ ] Update `docs/developer-guide/02-development-setup.md`
  - [ ] Step-by-step setup
  - [ ] IDE configuration
  - [ ] Debug examples
- [ ] Update `docs/developer-guide/03-testing.md`
  - [ ] Testing best practices
  - [ ] Coverage requirements
- [ ] Update `docs/developer-guide/makefile-guide.md`
  - [ ] Document all targets

### 4.5 Create Missing Docs

- [ ] Create `docs/user-guide/04-examples.md`
- [ ] Create `docs/user-guide/05-monitoring.md`
- [ ] Create `docs/reference/api-reference.md`

### Validation Checkpoint

```bash
# Check all markdown files
find docs -name "*.md" -exec echo "Checking {}" \;
# Manually review each doc for accuracy and completeness
```

---

## Phase 5: CI/CD & Build (Week 3-4) 🚀

### 5.1 Enhance Makefile

- [ ] Add `verify` target (fmt + vet + lint + test)
- [ ] Add `build-all` target (multi-platform)
- [ ] Add `docs` target
- [ ] Add `api-docs` target
- [ ] Improve target documentation
- [ ] Test all new targets

### 5.2 Enhance GitHub Actions

- [ ] Review `.github/workflows/ci.yaml`
- [ ] Add manifest validation job
- [ ] Add Helm chart validation job
- [ ] Add documentation check job
- [ ] Add security scanning (Trivy)
- [ ] Add coverage reporting
- [ ] Test workflow on PR

### 5.3 Add Release Automation

- [ ] Create `.github/workflows/release.yaml`
- [ ] Add semantic versioning
- [ ] Add changelog generation
- [ ] Add GitHub release creation
- [ ] Add Docker image push
- [ ] Test release workflow

### Validation Checkpoint

```bash
make verify
# All checks should pass

# Test locally before pushing
git commit -m "test: trigger CI"
git push origin <branch>
# Verify GitHub Actions pass
```

---

## Phase 6: Configuration & Deployment (Week 4) 🏗️

### 6.1 Improve Helm Chart

- [ ] Fix duplicate keys in `values.yaml`
- [ ] Add comprehensive values documentation
- [ ] Add NOTES.txt with post-install instructions
- [ ] Improve upgrade strategy
- [ ] Add proper CRD handling
- [ ] Test installation: `helm install helios helm/helios-operator`
- [ ] Test upgrade: `helm upgrade helios helm/helios-operator`

### 6.2 Kustomize Configuration

- [ ] Create `config/development/` overlay
- [ ] Create `config/staging/` overlay
- [ ] Create `config/production/` overlay
- [ ] Add resource limits
- [ ] Add network policies
- [ ] Test each overlay

### 6.3 Deployment Scripts

- [ ] Review `infrastructure/setup.sh`
- [ ] Add comprehensive pre-flight checks
- [ ] Implement idempotent operations
- [ ] Add dry-run mode
- [ ] Review `infrastructure/teardown.sh`
- [ ] Test both scripts

### 6.4 Monitoring

- [ ] Create `config/monitoring/servicemonitor.yaml`
- [ ] Create Grafana dashboard
- [ ] Create Prometheus rules
- [ ] Document metrics in `docs/reference/prometheus-metrics.md`

### Validation Checkpoint

```bash
# Test Helm installation
helm lint helm/helios-operator
helm install helios-test helm/helios-operator --dry-run

# Test Kustomize overlays
kustomize build config/development
kustomize build config/production
```

---

## Phase 7: Security & Observability (Week 4-5) 🔒

### 7.1 Security

- [ ] Add Pod Security Standards labels
- [ ] Update security contexts
- [ ] Add network policies
- [ ] Review RBAC permissions
- [ ] Document secrets management
- [ ] Run security scan: `trivy image <operator-image>`

### 7.2 Observability

- [ ] Review Prometheus metrics
- [ ] Add business metrics
- [ ] Add SLO/SLI metrics
- [ ] Add health probes
- [ ] Test metrics endpoint

### 7.3 Performance

- [ ] Add caching where appropriate
- [ ] Optimize reconciliation predicates
- [ ] Implement rate limiting
- [ ] Add resource quotas

### Validation Checkpoint

```bash
# Deploy with monitoring enabled
kubectl apply -f config/monitoring/

# Check metrics
kubectl port-forward svc/helios-operator-metrics 8443:8443
curl -k https://localhost:8443/metrics
```

---

## Phase 8: Developer Experience (Week 5) 🛠️

### 8.1 Development Environment

- [ ] Review `.devcontainer/` configuration
- [ ] Add pre-commit hooks
- [ ] Add VS Code settings
- [ ] Add debugging configuration
- [ ] Document setup process

### 8.2 CLI Improvements

- [ ] Add version command
- [ ] Add validation command
- [ ] Add completion scripts
- [ ] Test CLI commands

### Validation Checkpoint

```bash
# Test development setup
./helios-operator version
./helios-operator validate examples/simple-heliosapp.yaml
```

---

## Phase 9: Cleanup (Week 5) 🧹

### 9.1 Remove Duplicates

- [ ] Find duplicate code between controller and reconciler
- [ ] Consolidate duplicate logic
- [ ] Remove unused imports
- [ ] Remove commented code

### 9.2 Remove Unused Files

- [ ] Audit shell scripts
- [ ] Remove backup files
- [ ] Remove empty directories
- [ ] Clean test fixtures

### 9.3 Dependencies

- [ ] Run `go mod tidy`
- [ ] Remove unused dependencies
- [ ] Update outdated dependencies
- [ ] Document all dependencies

### Validation Checkpoint

```bash
go mod tidy
go mod verify
make verify
make test-e2e
```

---

## Final Validation ✅

### Complete System Test

- [ ] Build operator: `make build`
- [ ] Run all tests: `make test`
- [ ] Run E2E tests: `make test-e2e`
- [ ] Check coverage: >80%
- [ ] Run linters: `make lint`
- [ ] Build Docker image: `make docker-build`
- [ ] Install via Helm: `helm install helios helm/helios-operator`
- [ ] Create test HeliosApp
- [ ] Verify Pipeline created
- [ ] Verify EventListener created
- [ ] Verify ArgoCD Application created
- [ ] Verify status updates
- [ ] Delete HeliosApp
- [ ] Verify cleanup

### Documentation Review

- [ ] All code documented
- [ ] All user guides complete
- [ ] All developer guides complete
- [ ] All examples work
- [ ] All diagrams accurate

### Success Criteria

- [ ] ✅ All code compiles
- [ ] ✅ All tests pass
- [ ] ✅ Test coverage >80%
- [ ] ✅ No critical vulnerabilities
- [ ] ✅ All linters pass
- [ ] ✅ Operator creates resources correctly
- [ ] ✅ Status updates work
- [ ] ✅ Cleanup works
- [ ] ✅ Documentation complete
- [ ] ✅ CI/CD working

---

## Completion

Once all items are checked:

1. Create final commit: `git commit -m "refactor: complete project refactoring"`
2. Push changes: `git push origin <branch>`
3. Create PR with summary of changes
4. Run final validation in CI
5. Merge to main branch
6. Tag release: `git tag v2.1.0`
7. Celebrate! 🎉

---

## Notes

- **Estimated Time**: 5 weeks
- **Can be adjusted**: Focus on P0 and P1 items first
- **Test frequently**: After each section
- **Document as you go**: Don't leave it to the end
- **Ask for help**: If stuck on something

## Resources

- Main Plan: `REFACTORING_PLAN.md`
- Summary: `docs/REFACTORING_SUMMARY.md`
- This Checklist: `REFACTORING_CHECKLIST.md`
