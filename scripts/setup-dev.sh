#!/bin/bash

# Setup script for Helios Operator development environment
# This script sets up a complete development environment with all dependencies

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="helios-operator-system"
CERT_MANAGER_VERSION="v1.19.1"
TEKTON_VERSION="latest"
ARGOCD_VERSION="stable"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    # Check if minikube is running
    if ! minikube status &> /dev/null; then
        log_error "Minikube is not running. Please start minikube first."
        exit 1
    fi
    
    # Check if Docker is running
    if ! docker info &> /dev/null; then
        log_error "Docker is not running. Please start Docker first."
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

install_cert_manager() {
    log_info "Installing cert-manager ${CERT_MANAGER_VERSION}..."
    
    if kubectl get namespace cert-manager &> /dev/null; then
        log_warning "cert-manager is already installed"
        return
    fi
    
    kubectl apply -f "https://github.com/cert-manager/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml"
    
    # Wait for cert-manager to be ready
    log_info "Waiting for cert-manager to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=cert-manager -n cert-manager --timeout=300s
    
    log_success "cert-manager installed successfully"
}

install_tekton() {
    log_info "Installing Tekton Pipelines..."
    
    if kubectl get namespace tekton-pipelines &> /dev/null; then
        log_warning "Tekton Pipelines is already installed"
        return
    fi
    
    kubectl apply -f "https://storage.googleapis.com/tekton-releases/pipeline/${TEKTON_VERSION}/release.yaml"
    
    # Wait for Tekton to be ready
    log_info "Waiting for Tekton Pipelines to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=tekton-pipelines-controller -n tekton-pipelines --timeout=300s
    
    log_success "Tekton Pipelines installed successfully"
}

install_argocd() {
    log_info "Installing ArgoCD..."
    
    if kubectl get namespace argocd &> /dev/null; then
        log_warning "ArgoCD is already installed"
        return
    fi
    
    kubectl create namespace argocd
    kubectl apply -n argocd -f "https://raw.githubusercontent.com/argoproj/argo-cd/${ARGOCD_VERSION}/manifests/install.yaml"
    
    # Wait for ArgoCD to be ready
    log_info "Waiting for ArgoCD to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=argocd-server -n argocd --timeout=300s
    
    log_success "ArgoCD installed successfully"
}

build_operator() {
    log_info "Building Helios Operator..."
    
    # Build the operator
    make docker-build
    
    # Load image into minikube
    minikube image load helios.dev/helios-operator:latest
    
    log_success "Helios Operator built successfully"
}

deploy_operator() {
    log_info "Deploying Helios Operator..."
    
    # Deploy the operator
    make deploy
    
    # Wait for operator to be ready
    log_info "Waiting for Helios Operator to be ready..."
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=helios-operator -n ${NAMESPACE} --timeout=300s
    
    log_success "Helios Operator deployed successfully"
}

create_test_namespace() {
    log_info "Creating test namespace..."
    
    kubectl create namespace helios --dry-run=client -o yaml | kubectl apply -f -
    
    log_success "Test namespace created"
}

verify_installation() {
    log_info "Verifying installation..."
    
    # Check if operator is running
    if kubectl get pods -n ${NAMESPACE} | grep -q "Running"; then
        log_success "Helios Operator is running"
    else
        log_error "Helios Operator is not running"
        exit 1
    fi
    
    # Check if CRDs are installed
    if kubectl get crd heliosapps.platform.helios.io &> /dev/null; then
        log_success "HeliosApp CRD is installed"
    else
        log_error "HeliosApp CRD is not installed"
        exit 1
    fi
    
    # Check if webhook is working
    if kubectl get validatingwebhookconfiguration helios-operator-validating-webhook-configuration &> /dev/null; then
        log_success "Webhook is configured"
    else
        log_warning "Webhook is not configured"
    fi
    
    log_success "Installation verification completed"
}

show_status() {
    log_info "Current status:"
    echo ""
    echo "Namespaces:"
    kubectl get namespaces | grep -E "(helios|cert-manager|tekton|argocd)"
    echo ""
    echo "Helios Operator pods:"
    kubectl get pods -n ${NAMESPACE}
    echo ""
    echo "HeliosApp CRD:"
    kubectl get crd heliosapps.platform.helios.io
    echo ""
    echo "Webhook configurations:"
    kubectl get validatingwebhookconfiguration | grep helios
    kubectl get mutatingwebhookconfiguration | grep helios
}

# Main execution
main() {
    log_info "Setting up Helios Operator development environment..."
    echo ""
    
    check_prerequisites
    install_cert_manager
    install_tekton
    install_argocd
    build_operator
    deploy_operator
    create_test_namespace
    verify_installation
    show_status
    
    echo ""
    log_success "Development environment setup completed successfully!"
    echo ""
    echo "Next steps:"
    echo "1. Test the operator: kubectl apply -f examples/test-webhooks/test-complete-webhook.yaml"
    echo "2. Check logs: kubectl logs -n ${NAMESPACE} deployment/helios-operator-controller-manager"
    echo "3. Access ArgoCD UI: kubectl port-forward -n argocd svc/argocd-server 8080:443"
    echo "4. Access ArgoCD admin password: kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' | base64 -d"
}

# Run main function
main "$@"
