#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print status messages
print_status() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${CYAN}ℹ${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if resource exists
resource_exists() {
    kubectl get "$1" -n "$2" "$3" >/dev/null 2>&1
}

# Function to wait for resource to be ready
wait_for_resource() {
    local resource_type="$1"
    local namespace="$2"
    local selector="$3"
    local timeout="${4:-300}"
    
    print_info "Waiting for $resource_type to be ready in namespace $namespace..."
    if kubectl wait --for=condition=ready pod \
        -l "$selector" \
        -n "$namespace" \
        --timeout="${timeout}s" 2>/dev/null; then
        print_success "$resource_type is ready"
        return 0
    else
        print_warning "$resource_type may not be fully ready, but continuing..."
        return 1
    fi
}

echo -e "${GREEN}╔════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║   Helios Operator - Infrastructure Setup      ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════╝${NC}\n"

# Pre-flight checks
print_status "Performing pre-flight checks..."

# Check required tools
REQUIRED_TOOLS=("kubectl" "helm" "minikube")
for tool in "${REQUIRED_TOOLS[@]}"; do
    if command_exists "$tool"; then
        print_success "$tool is installed"
    else
        print_error "$tool is not installed. Please install it first."
        exit 1
    fi
done

# Check if kubectl can connect to cluster
if kubectl cluster-info >/dev/null 2>&1; then
    print_success "kubectl can connect to cluster"
else
    print_error "kubectl cannot connect to cluster. Please check your cluster configuration."
    exit 1
fi

print_success "Pre-flight checks completed\n"

# Check if Minikube is running
print_status "Checking Minikube status..."
if ! minikube status &> /dev/null; then
    print_warning "Minikube is not running. Starting Minikube..."
    minikube start --memory=4096 --cpus=2 --driver=docker
    print_success "Minikube started"
else
    print_success "Minikube is already running"
fi

# Enable Ingress addon
print_status "Enabling Ingress Controller..."
if minikube addons list | grep -q "ingress.*enabled"; then
    print_success "Ingress Controller is already enabled"
else
    minikube addons enable ingress
    print_success "Ingress Controller enabled"
fi

# Wait for Ingress Controller to be ready
wait_for_resource "Ingress Controller" "ingress-nginx" "app.kubernetes.io/name=ingress-nginx"

# Install ArgoCD via Helm
print_status "Installing ArgoCD via Helm..."
if resource_exists "deployment" "argocd" "argocd-server"; then
    print_success "ArgoCD is already installed"
else
    helm repo add argo https://argoproj.github.io/argo-helm 2>/dev/null || true
    helm repo update

    print_info "Installing ArgoCD (this may take 2-3 minutes)..."
    helm upgrade --install argocd argo/argo-cd \
        --namespace argocd \
        --create-namespace \
        --values infrastructure/helm-values/argocd-values.yaml \
        --wait \
        --timeout 5m

    print_success "ArgoCD installed"
fi

# Install Tekton Pipelines
print_status "Installing Tekton Pipelines..."
if resource_exists "deployment" "tekton-pipelines" "tekton-pipelines-controller"; then
    print_success "Tekton Pipelines is already installed"
else
    kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
    wait_for_resource "Tekton Pipelines" "tekton-pipelines" "app=tekton-pipelines-controller"
    print_success "Tekton Pipelines installed"
fi

# Install Tekton Triggers
print_status "Installing Tekton Triggers..."
if resource_exists "deployment" "tekton-pipelines" "tekton-triggers-controller"; then
    print_success "Tekton Triggers is already installed"
else
    kubectl apply -f https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml
    wait_for_resource "Tekton Triggers" "tekton-pipelines" "app.kubernetes.io/name=controller"
    print_success "Tekton Triggers installed"
fi

# Update /etc/hosts for ArgoCD
print_status "Configuring local DNS..."
MINIKUBE_IP=$(minikube ip)

# Check if entry already exists
if grep -q "argocd.local" /etc/hosts; then
    print_success "Entry for argocd.local already exists in /etc/hosts"
    print_info "If you need to update it, run:"
    print_info "  sudo sed -i '/argocd.local/d' /etc/hosts"
    print_info "  echo \"$MINIKUBE_IP argocd.local\" | sudo tee -a /etc/hosts"
else
    print_info "Adding argocd.local to /etc/hosts..."
    if echo "$MINIKUBE_IP argocd.local" | sudo tee -a /etc/hosts >/dev/null; then
        print_success "DNS configured"
    else
        print_warning "Failed to update /etc/hosts. You may need to add manually:"
        print_warning "  echo \"$MINIKUBE_IP argocd.local\" | sudo tee -a /etc/hosts"
    fi
fi

# Final verification
print_status "Performing final verification..."
sleep 5  # Give services time to stabilize

# Check ArgoCD
if kubectl get pods -n argocd -l app.kubernetes.io/name=argocd-server --no-headers | grep -q Running; then
    print_success "ArgoCD server is running"
else
    print_warning "ArgoCD server may not be fully ready yet"
fi

# Check Tekton
if kubectl get pods -n tekton-pipelines -l app=tekton-pipelines-controller --no-headers | grep -q Running; then
    print_success "Tekton Pipelines is running"
else
    print_warning "Tekton Pipelines may not be fully ready yet"
fi

print_success "Installation complete! 🎉\n"

echo -e "${GREEN}╔════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║            Access Information                  ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════╝${NC}\n"

echo -e "${YELLOW}ArgoCD UI:${NC}"
echo -e "  URL (Ingress): https://argocd.local"
echo -e "  URL (Port-Forward): https://localhost:8080"
echo -e "  Username: ${GREEN}admin${NC}"

ARGOCD_PASSWORD=$(kubectl get secret argocd-initial-admin-secret -n argocd -o jsonpath="{.data.password}" 2>/dev/null | base64 -d)
echo -e "  Password: ${GREEN}${ARGOCD_PASSWORD}${NC}\n"

echo -e "${YELLOW}Port-forward command (if Ingress doesn't work):${NC}"
echo -e "  ${BLUE}kubectl port-forward svc/argocd-server -n argocd 8080:443${NC}\n"

echo -e "${YELLOW}Verify installation:${NC}"
echo -e "  ${BLUE}kubectl get pods -n argocd${NC}"
echo -e "  ${BLUE}kubectl get pods -n tekton-pipelines${NC}\n"

echo -e "${GREEN}╔════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║              Next Steps                        ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════╝${NC}\n"

echo -e "1. Create your GitOps repository with structure:"
echo -e "   ${BLUE}your-gitops-repo/${NC}"
echo -e "     ${BLUE}└── helios-app/${NC}  (app name, NO 'apps/' prefix)"
echo -e "         ${BLUE}├── deployment.yaml${NC}"
echo -e "         ${BLUE}└── service.yaml${NC}\n"

echo -e "2. Build and deploy Helios Operator:"
echo -e "   ${BLUE}make docker-build IMG=helios-operator:latest${NC}"
echo -e "   ${BLUE}minikube image load helios-operator:latest${NC}"
echo -e "   ${BLUE}make install${NC}"
echo -e "   ${BLUE}make deploy IMG=helios-operator:latest${NC}\n"

echo -e "3. Apply your HeliosApp CR:"
echo -e "   ${BLUE}kubectl apply -f config/samples/heliosapp_v1_heliosapp.yaml${NC}\n"

echo -e "${GREEN}Happy coding! 🚀${NC}\n"

