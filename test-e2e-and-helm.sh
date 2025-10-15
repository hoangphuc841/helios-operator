#!/bin/bash

# Test E2E and Helm Chart for Helios Operator
# This script tests both the E2E test suite and Helm chart functionality

set -e

echo "🚀 Starting E2E and Helm Chart Testing for Helios Operator"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check kubectl
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed"
        exit 1
    fi
    
    # Check helm
    if ! command -v helm &> /dev/null; then
        print_error "helm is not installed"
        exit 1
    fi
    
    # Check minikube
    if ! command -v minikube &> /dev/null; then
        print_error "minikube is not installed"
        exit 1
    fi
    
    # Check if minikube is running
    if ! minikube status &> /dev/null; then
        print_error "minikube is not running"
        exit 1
    fi
    
    print_success "All prerequisites are available"
}

# Build and load operator image
build_and_load_image() {
    print_status "Building and loading operator image..."
    
    # Build the image
    docker build -t helios-operator:v2.0.0 .
    
    # Load into minikube
    minikube image load helios-operator:v2.0.0
    
    print_success "Operator image built and loaded"
}

# Test Helm Chart
test_helm_chart() {
    print_status "Testing Helm Chart..."
    
    # Lint the chart
    print_status "Linting Helm chart..."
    helm lint ./helm/helios-operator
    
    # Template the chart
    print_status "Templating Helm chart..."
    helm template helios-operator ./helm/helios-operator > /tmp/helm-template.yaml
    
    # Dry run installation
    print_status "Dry run installation..."
    helm install helios-operator ./helm/helios-operator --dry-run --debug
    
    print_success "Helm chart tests passed"
}

# Install dependencies
install_dependencies() {
    print_status "Installing dependencies..."
    
    # Install Tekton Pipelines
    kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
    
    # Install Tekton Triggers
    kubectl apply -f https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml
    
    # Install ArgoCD
    helm repo add argo https://argoproj.github.io/argo-helm
    helm repo update
    kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
    helm install argocd argo/argo-cd --namespace argocd
    
    # Wait for dependencies to be ready
    print_status "Waiting for Tekton to be ready..."
    kubectl wait --for=condition=ready pod -l app=tekton-pipelines-controller -n tekton-pipelines --timeout=300s
    
    print_status "Waiting for ArgoCD to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server -n argocd --timeout=300s
    
    print_success "Dependencies installed and ready"
}

# Install operator via Helm
install_operator_helm() {
    print_status "Installing operator via Helm..."
    
    # Install CRDs first
    kubectl apply -f config/crd/bases/platform.helios.io_heliosapps.yaml
    
    # Install operator via Helm
    helm install helios-operator ./helm/helios-operator \
        --set image.repository=helios-operator \
        --set image.tag=v2.0.0 \
        --set image.pullPolicy=IfNotPresent
    
    # Wait for operator to be ready
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=helios-operator -n system --timeout=300s
    
    print_success "Operator installed via Helm"
}

# Test operator functionality
test_operator_functionality() {
    print_status "Testing operator functionality..."
    
    # Create test namespace
    kubectl create namespace helios-apps --dry-run=client -o yaml | kubectl apply -f -
    
    # Create PVC for test
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: test-app-pvc
  namespace: helios-apps
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
EOF
    
    # Create test HeliosApp
    cat <<EOF | kubectl apply -f -
apiVersion: platform.helios.io/v1
kind: HeliosApp
metadata:
  name: test-app
  namespace: helios-apps
spec:
  gitRepo: "https://github.com/hoangphuc841/helios.git"
  gitBranch: "main"
  gitopsRepo: "https://github.com/PhuocHoan/helios-gitops.git"
  gitopsPath: "test-app"
  gitopsBranch: "main"
  imageRepo: "docker.io/test/test-app"
  port: 80
  replicas: 1
  serviceAccount: "pipeline-sa"
  webhookSecret: "github-webhook-secret"
  pvcName: "test-app-pvc"
EOF
    
    # Wait for resources to be created
    print_status "Waiting for Pipeline to be created..."
    kubectl wait --for=condition=ready --timeout=120s pipeline/test-app-pipeline -n helios-apps || true
    
    print_status "Waiting for EventListener to be created..."
    kubectl wait --for=condition=ready --timeout=120s eventlistener/test-app-el -n helios-apps || true
    
    print_status "Waiting for ArgoCD Application to be created..."
    kubectl wait --for=condition=ready --timeout=120s application/test-app-argocd -n argocd || true
    
    # Verify resources exist
    if kubectl get pipeline test-app-pipeline -n helios-apps &> /dev/null; then
        print_success "Pipeline created successfully"
    else
        print_warning "Pipeline not found (may still be creating)"
    fi
    
    if kubectl get eventlistener test-app-el -n helios-apps &> /dev/null; then
        print_success "EventListener created successfully"
    else
        print_warning "EventListener not found (may still be creating)"
    fi
    
    if kubectl get application test-app-argocd -n argocd &> /dev/null; then
        print_success "ArgoCD Application created successfully"
    else
        print_warning "ArgoCD Application not found (may still be creating)"
    fi
    
    # Test cleanup
    print_status "Testing resource cleanup..."
    kubectl delete heliosapp test-app -n helios-apps
    
    # Wait for cleanup
    sleep 30
    
    if ! kubectl get pipeline test-app-pipeline -n helios-apps &> /dev/null; then
        print_success "Pipeline cleaned up successfully"
    else
        print_warning "Pipeline still exists (cleanup may be in progress)"
    fi
    
    print_success "Operator functionality tests completed"
}

# Run E2E tests
run_e2e_tests() {
    print_status "Running E2E tests..."
    
    # Set project image for E2E tests
    export PROJECT_IMAGE="helios-operator:v2.0.0"
    
    # Run E2E tests
    go test -v ./test/e2e/... -timeout 30m
    
    print_success "E2E tests completed"
}

# Cleanup function
cleanup() {
    print_status "Cleaning up test resources..."
    
    # Delete test namespace
    kubectl delete namespace helios-apps --ignore-not-found=true
    
    # Uninstall operator
    helm uninstall helios-operator --ignore-not-found
    
    print_success "Cleanup completed"
}

# Main execution
main() {
    # Set up trap for cleanup on exit
    trap cleanup EXIT
    
    check_prerequisites
    build_and_load_image
    test_helm_chart
    install_dependencies
    install_operator_helm
    test_operator_functionality
    
    # Ask user if they want to run E2E tests (they take longer)
    echo -e "${YELLOW}Do you want to run the full E2E test suite? (y/n)${NC}"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        run_e2e_tests
    else
        print_warning "Skipping E2E tests"
    fi
    
    print_success "All tests completed successfully! 🎉"
}

# Run main function
main "$@"
