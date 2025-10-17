#!/bin/bash

# Teardown script for Helios Operator development environment
# This script removes all resources created during development

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="helios-operator-system"
HELIOS_NAMESPACE="helios"

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

confirm_teardown() {
    echo ""
    log_warning "This will remove all Helios Operator resources and dependencies."
    echo "The following will be removed:"
    echo "- Helios Operator deployment"
    echo "- All HeliosApp resources"
    echo "- cert-manager (if installed by this script)"
    echo "- Tekton Pipelines (if installed by this script)"
    echo "- ArgoCD (if installed by this script)"
    echo ""
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Teardown cancelled"
        exit 0
    fi
}

remove_helios_apps() {
    log_info "Removing all HeliosApp resources..."
    
    # Get all namespaces that might have HeliosApp resources
    namespaces=$(kubectl get namespaces -o jsonpath='{.items[*].metadata.name}')
    
    for ns in $namespaces; do
        if kubectl get heliosapps.platform.helios.io -n "$ns" &> /dev/null; then
            log_info "Removing HeliosApp resources from namespace: $ns"
            kubectl delete heliosapps.platform.helios.io --all -n "$ns" --ignore-not-found=true
        fi
    done
    
    log_success "HeliosApp resources removed"
}

remove_operator() {
    log_info "Removing Helios Operator..."
    
    # Undeploy the operator
    make undeploy || true
    
    # Remove any remaining resources
    kubectl delete namespace ${NAMESPACE} --ignore-not-found=true
    
    log_success "Helios Operator removed"
}

remove_test_namespace() {
    log_info "Removing test namespace..."
    
    kubectl delete namespace ${HELIOS_NAMESPACE} --ignore-not-found=true
    
    log_success "Test namespace removed"
}

remove_cert_manager() {
    log_info "Removing cert-manager..."
    
    if kubectl get namespace cert-manager &> /dev/null; then
        kubectl delete -f "https://github.com/cert-manager/cert-manager/releases/download/v1.19.1/cert-manager.yaml" --ignore-not-found=true
        log_success "cert-manager removed"
    else
        log_info "cert-manager not found, skipping"
    fi
}

remove_tekton() {
    log_info "Removing Tekton Pipelines..."
    
    if kubectl get namespace tekton-pipelines &> /dev/null; then
        kubectl delete -f "https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml" --ignore-not-found=true
        log_success "Tekton Pipelines removed"
    else
        log_info "Tekton Pipelines not found, skipping"
    fi
}

remove_argocd() {
    log_info "Removing ArgoCD..."
    
    if kubectl get namespace argocd &> /dev/null; then
        kubectl delete -f "https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml" --ignore-not-found=true
        kubectl delete namespace argocd --ignore-not-found=true
        log_success "ArgoCD removed"
    else
        log_info "ArgoCD not found, skipping"
    fi
}

cleanup_images() {
    log_info "Cleaning up Docker images..."
    
    # Remove operator image from minikube
    minikube image rm helios.dev/helios-operator:latest || true
    
    # Remove local Docker images
    docker rmi helios.dev/helios-operator:latest || true
    
    log_success "Docker images cleaned up"
}

cleanup_files() {
    log_info "Cleaning up temporary files..."
    
    # Remove test files
    rm -f examples/test-webhooks/test-*.yaml
    rm -f examples/test-webhooks/webhook.*
    
    # Remove build artifacts
    make clean || true
    
    log_success "Temporary files cleaned up"
}

show_remaining_resources() {
    log_info "Checking for remaining resources..."
    
    echo ""
    echo "Remaining namespaces:"
    kubectl get namespaces | grep -E "(helios|cert-manager|tekton|argocd)" || echo "No related namespaces found"
    
    echo ""
    echo "Remaining CRDs:"
    kubectl get crd | grep helios || echo "No HeliosApp CRDs found"
    
    echo ""
    echo "Remaining webhook configurations:"
    kubectl get validatingwebhookconfiguration | grep helios || echo "No Helios webhooks found"
    kubectl get mutatingwebhookconfiguration | grep helios || echo "No Helios webhooks found"
}

# Main execution
main() {
    log_info "Starting Helios Operator development environment teardown..."
    echo ""
    
    confirm_teardown
    
    remove_helios_apps
    remove_operator
    remove_test_namespace
    remove_cert_manager
    remove_tekton
    remove_argocd
    cleanup_images
    cleanup_files
    show_remaining_resources
    
    echo ""
    log_success "Development environment teardown completed successfully!"
    echo ""
    echo "Note: Some resources might still exist if they were created outside of this script."
    echo "Use 'kubectl get all --all-namespaces' to check for any remaining resources."
}

# Run main function
main "$@"
