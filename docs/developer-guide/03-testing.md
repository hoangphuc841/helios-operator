# 🧪 Testing Guide

This guide covers all aspects of testing in the Helios Operator project, from unit tests to end-to-end testing.

## 🎯 **Testing Strategy**

Our testing strategy follows the testing pyramid approach:

```
    /\
   /  \
  /E2E \     <- End-to-End Tests (Few, Slow, High Confidence)
 /______\
/        \
/Integration\ <- Integration Tests (Some, Medium Speed, Medium Confidence)
/____________\
/              \
/    Unit Tests   \ <- Unit Tests (Many, Fast, Low Confidence)
/__________________\
```

### Test Categories

1. **Unit Tests** - Test individual functions and methods in isolation
2. **Integration Tests** - Test component interactions within the operator
3. **End-to-End Tests** - Test complete workflows from HeliosApp creation to deployment

## 🧪 **Unit Testing**

### Test Framework

We use the standard Go testing package with [Ginkgo](https://onsi.github.io/ginkgo/) and [Gomega](https://onsi.github.io/gomega/) for BDD-style testing.

### Available Makefile Targets

The project provides several Makefile targets for testing:

| Target               | Description                                               | Command                                                            |
| -------------------- | --------------------------------------------------------- | ------------------------------------------------------------------ |
| `make test`          | Run all unit tests with race detection and verbose output | `go test ./... -race -v`                                           |
| `make test-e2e`      | Run end-to-end tests with 30-minute timeout               | `go test ./test/e2e/ -v -timeout=30m`                              |
| `make test-coverage` | Run tests with coverage report and HTML output            | `go test ./... -race -coverprofile=coverage.out -covermode=atomic` |

### Running Unit Tests

```bash
# Run all unit tests
make test

# Run tests with coverage
make test-coverage

# Run specific test package
make test

# Run tests with race detection (included in make test)
make test

# Run tests with verbose output (included in make test)
make test
```

### Test Structure

The project follows a clear test structure with tests organized by functionality:

```
├── api/v1/                          # API layer tests
│   ├── heliosapp_types_test.go      # HeliosApp type tests
│   └── heliosapp_webhook_test.go    # Webhook validation tests
├── internal/controller/              # Controller tests
│   ├── heliosapp_controller_test.go # Main controller tests
│   └── suite_test.go                # Test suite setup
├── internal/resources/               # Resource generation tests
│   └── argocd_test.go               # ArgoCD resource tests
└── test/                            # E2E tests
    ├── e2e/
    │   ├── e2e_test.go              # E2E test scenarios
    │   └── e2e_suite_test.go        # E2E test setup
    └── utils/
        └── utils.go                 # Test utilities
```

### Test Framework Example

```go
package controller

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    heliosappv1 "github.com/hoangphuc841/helios-operator/api/v1"
)

var _ = Describe("HeliosApp Controller", func() {
    Context("When creating a HeliosApp", func() {
        It("should generate a Pipeline with correct name", func() {
            // Test implementation
        })
    })
})
```

### Controller Tests

**Location**: `internal/controller/heliosapp_controller_test.go`

**Test Coverage**:

- Pipeline generation logic
- ArgoCD Application creation
- Status updates and conditions
- Error handling scenarios
- Resource cleanup

**Example Test**:

```go
var _ = Describe("Pipeline Generation", func() {
    Context("GeneratePipeline", func() {
        It("should create pipeline with correct name", func() {
            heliosApp := &heliosappv1.HeliosApp{
                ObjectMeta: metav1.ObjectMeta{
                    Name: "test-app",
                },
                Spec: heliosappv1.HeliosAppSpec{
                    GitRepo:   "https://github.com/test/app.git",
                    ImageRepo: "docker.io/test/app",
                },
            }

            pipeline, err := GeneratePipeline(heliosApp)
            Expect(err).NotTo(HaveOccurred())
            Expect(pipeline.GetName()).To(Equal("test-app-pipeline"))
        })
    })
})
```

### Webhook Tests

**Location**: `api/v1/heliosapp_webhook_test.go`

**Test Coverage**:

- Validation rules
- Defaulting logic
- Error messages
- Edge cases

**Example Test**:

```go
var _ = Describe("HeliosApp Webhook", func() {
    Context("Validation", func() {
        It("should reject empty gitRepo", func() {
            heliosApp := &heliosappv1.HeliosApp{
                Spec: heliosappv1.HeliosAppSpec{
                    GitRepo: "", // Invalid
                    ImageRepo: "docker.io/test/app",
                },
            }

            err := webhook.ValidateCreate(ctx, heliosApp)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(ContainSubstring("gitRepo is required"))
        })
    })
})
```

## 🔗 **Integration Testing**

### Test Framework

Integration tests use the same Ginkgo/Gomega framework but run against a real Kubernetes cluster.

### Running Integration Tests

```bash
# Run integration tests (using unit test framework)
make test

# Run specific integration test
make test

# Run with specific focus
make test
```

### Integration Test Structure

```
internal/controller/
├── heliosapp_controller_test.go    # Main controller tests
└── suite_test.go                   # Test suite setup

api/v1/
├── heliosapp_types_test.go         # API types tests
└── heliosapp_webhook_test.go       # Webhook tests

internal/resources/
└── argocd_test.go                  # Resource generation tests
```

### Test Setup

```go
var _ = BeforeSuite(func() {
    // Setup test environment
    // Create test namespace
    // Install dependencies
})

var _ = AfterSuite(func() {
    // Cleanup test environment
    // Delete test resources
})
```

### Example Integration Test

```go
var _ = Describe("Pipeline Integration", func() {
    Context("When creating a HeliosApp", func() {
        It("should create Pipeline in cluster", func() {
            // Create HeliosApp
            heliosApp := createTestHeliosApp()
            Expect(k8sClient.Create(ctx, heliosApp)).To(Succeed())

            // Wait for Pipeline creation
            Eventually(func() bool {
                pipeline := &unstructured.Unstructured{}
                pipeline.SetGroupVersionKind(schema.GroupVersionKind{
                    Group:   "tekton.dev",
                    Version: "v1beta1",
                    Kind:    "Pipeline",
                })
                err := k8sClient.Get(ctx, client.ObjectKey{
                    Name:      "test-app-pipeline",
                    Namespace: "test-namespace",
                }, pipeline)
                return err == nil
            }, timeout, interval).Should(BeTrue())
        })
    })
})
```

## 🎬 **End-to-End Testing**

### Test Framework

E2E tests use Ginkgo/Gomega and run against a complete Kubernetes cluster with all dependencies.

### Running E2E Tests

```bash
# Set project image
export PROJECT_IMAGE="helios-operator:latest"

# Run E2E tests
make test-e2e

# Run specific E2E test
make test-e2e

# Run with timeout (included in make test-e2e)
make test-e2e
```

### E2E Test Structure

```
test/
├── e2e/
│   ├── e2e_test.go            # Main E2E test file
│   └── e2e_suite_test.go      # E2E test suite setup
└── utils/
    └── utils.go              # E2E test utilities
```

### Test Scenarios

1. **Complete Lifecycle Test**

   - Create HeliosApp
   - Verify Pipeline creation
   - Verify Tekton Trigger resources
   - Verify ArgoCD Application
   - Simulate build process
   - Verify deployment
   - Test cleanup

2. **Error Handling Test**

   - Invalid configuration
   - Missing dependencies
   - Network failures
   - Resource conflicts

3. **Performance Test**
   - Multiple concurrent HeliosApps
   - Resource cleanup performance
   - Memory usage patterns

### Example E2E Test

```go
var _ = Describe("HeliosApp E2E", func() {
    Context("Complete Lifecycle", func() {
        It("should handle full application lifecycle", func() {
            By("Creating HeliosApp")
            heliosApp := createTestHeliosApp()
            cmd := exec.Command("kubectl", "apply", "-f", "-")
            cmd.Stdin = strings.NewReader(heliosAppYAML)
            _, err := utils.Run(cmd)
            Expect(err).NotTo(HaveOccurred())

            By("Waiting for Pipeline creation")
            Eventually(func() bool {
                cmd := exec.Command("kubectl", "get", "pipeline", "test-app-pipeline", "-n", "test-namespace")
                _, err := utils.Run(cmd)
                return err == nil
            }, 2*time.Minute).Should(BeTrue())

            By("Verifying ArgoCD Application")
            Eventually(func() bool {
                cmd := exec.Command("kubectl", "get", "application", "test-app-argocd", "-n", "argocd")
                _, err := utils.Run(cmd)
                return err == nil
            }, 2*time.Minute).Should(BeTrue())
        })
    })
})
```

## 🚀 **Test Automation**

### CI/CD Pipeline

Our tests run automatically on every pull request and merge:

```yaml
# .github/workflows/ci.yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v5
      - name: Set up Go
        uses: actions/setup-go@v6
        with:
          go-version: 1.25
      - name: Run unit tests
        run: make test
      - name: Run E2E tests
        run: make test-e2e
      - name: Run tests with coverage
        run: make test-coverage
```

### Test Scripts

**Location**: `Makefile` targets

The project uses Makefile targets for automated testing:

- Prerequisites checking
- Dependency installation
- Operator deployment
- Test execution
- Cleanup

```bash
# Run unit tests
make test

# Run E2E tests
make test-e2e

# Run with coverage
make test-coverage

# Run all available tests
make test && make test-e2e
```

## 📊 **Test Coverage**

### Coverage Reports

```bash
# Generate coverage report
make test-coverage

# View coverage in browser (automatically generated)
open coverage.html

# Check coverage threshold
make test-coverage
```

### Coverage Targets

- **Unit Tests**: >90% coverage
- **Integration Tests**: >80% coverage
- **Critical Paths**: 100% coverage

### Coverage Exclusions

```go
// +build !integration

// Test files are excluded from coverage
// Generated files are excluded
// Vendor dependencies are excluded
```

## 🐛 **Test Debugging**

### Debug Mode

```bash
# Run tests with debug output
make test

# Run specific test with focus
make test

# Run tests with race detection (included in make test)
make test

# Run tests with timeout (included in make test)
make test
```

### Test Logs

```bash
# Enable verbose logging in tests
export LOG_LEVEL=debug
make test

# Capture test logs
make test 2>&1 | tee test.log
```

### Debugging Failed Tests

```bash
# Run single test repeatedly
make test

# Run with race detection (included in make test)
make test

# Run with timeout (included in make test)
make test
```

## 🔧 **Test Utilities**

### Test Helpers

**Location**: `test/utils/utils.go`

```go
// utils.go
func Run(cmd *exec.Cmd) (string, error) {
    output, err := cmd.CombinedOutput()
    return string(output), err
}

func GetNonEmptyLines(output string) []string {
    lines := strings.Split(output, "\n")
    var result []string
    for _, line := range lines {
        if strings.TrimSpace(line) != "" {
            result = append(result, strings.TrimSpace(line))
        }
    }
    return result
}
```

### Test Fixtures

**Location**: `test/utils/utils.go`

```go
// utils.go
func CreateTestHeliosApp() *heliosappv1.HeliosApp {
    return &heliosappv1.HeliosApp{
        ObjectMeta: metav1.ObjectMeta{
            Name:      "test-app",
            Namespace: "test-namespace",
        },
        Spec: heliosappv1.HeliosAppSpec{
            GitRepo:   "https://github.com/test/app.git",
            ImageRepo: "docker.io/test/app",
            // ... other fields
        },
    }
}
```

## 🎯 **Performance Testing**

### Load Testing

```go
var _ = Describe("Performance Tests", func() {
    Context("Multiple HeliosApps", func() {
        It("should handle 100 concurrent HeliosApps", func() {
            var wg sync.WaitGroup
            for i := 0; i < 100; i++ {
                wg.Add(1)
                go func(index int) {
                    defer wg.Done()
                    heliosApp := createTestHeliosApp()
                    heliosApp.Name = fmt.Sprintf("test-app-%d", index)
                    Expect(k8sClient.Create(ctx, heliosApp)).To(Succeed())
                }(i)
            }
            wg.Wait()
        })
    })
})
```

### Memory Testing

```bash
# Run tests with memory profiling
make test

# Run memory profiling
go tool pprof mem.prof
```

### Benchmark Tests

```go
func BenchmarkPipelineGeneration(b *testing.B) {
    heliosApp := createTestHeliosApp()
    for i := 0; i < b.N; i++ {
        _, err := GeneratePipeline(heliosApp)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## 🧹 **Test Cleanup**

### Automatic Cleanup

```go
var _ = AfterEach(func() {
    // Clean up test resources
    cleanupTestResources()
})

func cleanupTestResources() {
    // Delete test namespaces
    // Remove test CRDs
    // Clean up test images
}
```

### Manual Cleanup

```bash
# Clean up test resources
make clean

# Remove test namespaces
kubectl delete namespace test-namespace --ignore-not-found

# Remove test images
docker rmi test-app:latest
```

## 📈 **Test Metrics**

### Test Results

```bash
# Generate test report
make test

# Count test cases
make test

# Measure test execution time
time make test
```

### Quality Gates

- All tests must pass
- Coverage thresholds must be met
- No race conditions detected
- Performance benchmarks within limits

## 🆘 **Troubleshooting Tests**

### Common Issues

#### Test Timeouts

```bash
# Increase timeout
make test

# Debug slow tests
make test
```

#### Resource Conflicts

```bash
# Clean up before tests
make clean

# Use unique namespaces
export TEST_NAMESPACE="test-$(date +%s)"
```

#### Flaky Tests

```bash
# Run tests multiple times
for i in {1..10}; do
    make test
done

# Run with race detection (included in make test)
make test
```

---

**Ready to write tests?** Check out our [Development Setup Guide](02-development-setup.md) to get your environment ready! 🚀
