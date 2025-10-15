#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${RED}╔════════════════════════════════════════════════╗${NC}"
echo -e "${RED}║   Helios Operator - Infrastructure Teardown   ║${NC}"
echo -e "${RED}╚════════════════════════════════════════════════╝${NC}\n"

read -p "$(echo -e ${YELLOW}Are you sure you want to remove all infrastructure? [y/N]: ${NC})" -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${GREEN}Teardown cancelled.${NC}"
    exit 0
fi

# Remove Tekton Triggers
echo -e "\n${BLUE}[1/5]${NC} Removing Tekton Triggers..."
kubectl delete -f https://storage.googleapis.com/tekton-releases/triggers/latest/release.yaml --ignore-not-found=true
echo -e "${GREEN}✓ Tekton Triggers removed${NC}"

# Remove Tekton Pipelines
echo -e "\n${BLUE}[2/5]${NC} Removing Tekton Pipelines..."
kubectl delete -f https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml --ignore-not-found=true
echo -e "${GREEN}✓ Tekton Pipelines removed${NC}"

# Remove ArgoCD
echo -e "\n${BLUE}[3/5]${NC} Removing ArgoCD..."
helm uninstall argocd -n argocd --ignore-not-found 2>/dev/null || true
kubectl delete namespace argocd --ignore-not-found=true
echo -e "${GREEN}✓ ArgoCD removed${NC}"

# Clean up namespaces
echo -e "\n${BLUE}[4/5]${NC} Cleaning up namespaces..."
kubectl delete namespace tekton-pipelines --ignore-not-found=true 2>/dev/null || true
echo -e "${GREEN}✓ Namespaces cleaned${NC}"

# Remove /etc/hosts entry
echo -e "\n${BLUE}[5/5]${NC} Cleaning up /etc/hosts..."
if grep -q "argocd.local" /etc/hosts 2>/dev/null; then
    echo -e "${YELLOW}Removing argocd.local from /etc/hosts...${NC}"
    sudo sed -i '/argocd.local/d' /etc/hosts
    echo -e "${GREEN}✓ /etc/hosts cleaned${NC}"
else
    echo -e "${YELLOW}No entry for argocd.local in /etc/hosts${NC}"
fi

echo -e "\n${GREEN}╔════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║         Infrastructure Teardown Complete      ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════════════╝${NC}\n"

echo -e "${YELLOW}Note: Minikube cluster is still running.${NC}"
echo -e "${YELLOW}To stop Minikube, run: ${BLUE}minikube stop${NC}"
echo -e "${YELLOW}To delete Minikube, run: ${BLUE}minikube delete${NC}\n"

