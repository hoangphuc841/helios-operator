#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}╔════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║   Helios Operator - Infrastructure Setup      ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════╝${NC}\n"

# Check if Minikube is running
echo -e "${BLUE}[1/7]${NC} Checking Minikube status..."
if ! minikube status &> /dev/null; then
    echo -e "${YELLOW}Minikube is not running. Starting Minikube...${NC}"
    minikube start --memory=4096 --cpus=2 --driver=docker
else
    echo -e "${GREEN}✓ Minikube is running${NC}"
fi

# Enable Ingress addon
echo -e "\n${BLUE}[2/7]${NC} Enabling Ingress Controller..."
minikube addons enable ingress
echo -e "${GREEN}✓ Ingress Controller enabled${NC}"

# Wait for Ingress Controller to be ready
echo -e "${YELLOW}Waiting for Ingress Controller to be ready...${NC}"
kubectl wait --for=condition=ready pod \
    -l app.kubernetes.io/name=ingress-nginx \
    -n ingress-nginx \
    --timeout=300s 2>/dev/null || true

# Install ArgoCD via Helm
echo -e "\n${BLUE}[3/7]${NC} Installing ArgoCD via Helm..."
helm repo add argo https://argoproj.github.io/argo-helm 2>/dev/null || true
helm repo update

echo -e "${YELLOW}Installing ArgoCD (this may take 2-3 minutes)...${NC}"
helm upgrade --install argocd argo/argo-cd \
    --namespace argocd \
    --create-namespace \
    --values infrastructure/helm-values/argocd-values.yaml \
    --wait \
    --timeout 5m

echo -e "${GREEN}✓ ArgoCD installed${NC}"

# Install Tekton Pipelines
echo -e "\n${BLUE}[4/7]${NC} Installing Tekton Pipelines..."
kubectl apply -f https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml

echo -e "${YELLOW}Waiting for Tekton Pipelines to be ready...${NC}"
kubectl wait --for=condition=ready pod \
    -l app=tekton-pipelines-controller \
    -n tekton-pipelines \
    --timeout=300s

echo -e "${GREEN}✓ Tekton Pipelines installed${NC}"

# Install Tekton Triggers
echo -e "\n${BLUE}[5/7]${NC} Installing Tekton Triggers..."
kubectl apply -f https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml

echo -e "${YELLOW}Waiting for Tekton Triggers to be ready...${NC}"
kubectl wait --for=condition=ready pod \
    -l app.kubernetes.io/name=controller \
    -n tekton-pipelines \
    --timeout=300s 2>/dev/null || true

echo -e "${GREEN}✓ Tekton Triggers installed${NC}"

# Update /etc/hosts for ArgoCD
echo -e "\n${BLUE}[6/7]${NC} Configuring local DNS..."
MINIKUBE_IP=$(minikube ip)

# Check if entry already exists
if grep -q "argocd.local" /etc/hosts; then
    echo -e "${YELLOW}Entry for argocd.local already exists in /etc/hosts${NC}"
    echo -e "${YELLOW}If you need to update it, run:${NC}"
    echo -e "${YELLOW}  sudo sed -i '/argocd.local/d' /etc/hosts${NC}"
    echo -e "${YELLOW}  echo \"$MINIKUBE_IP argocd.local\" | sudo tee -a /etc/hosts${NC}"
else
    echo -e "${YELLOW}Adding argocd.local to /etc/hosts...${NC}"
    echo "$MINIKUBE_IP argocd.local" | sudo tee -a /etc/hosts
    echo -e "${GREEN}✓ DNS configured${NC}"
fi

# Display access information
echo -e "\n${BLUE}[7/7]${NC} Installation complete! 🎉\n"

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

