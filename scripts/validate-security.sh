#!/bin/bash

# Security Validation Script for Helios Operator
# This script validates that the operator deployment follows Pod Security Standards

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
NAMESPACE="${NAMESPACE:-helios-operator-system}"
OPERATOR_NAME="${OPERATOR_NAME:-helios-operator}"

echo -e "${GREEN}🔒 Starting Helios Operator Security Validation${NC}"
echo "Namespace: $NAMESPACE"
echo "Operator: $OPERATOR_NAME"
echo

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
echo -e "${YELLOW}📋 Checking prerequisites...${NC}"
if ! command_exists kubectl; then
    echo -e "${RED}❌ kubectl not found. Please install kubectl.${NC}"
    exit 1
fi

if ! kubectl cluster-info >/dev/null 2>&1; then
    echo -e "${RED}❌ Cannot connect to Kubernetes cluster.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Prerequisites check passed${NC}"
echo

# Function to validate pod security context
validate_pod_security() {
    local pod_name="$1"
    local namespace="$2"
    
    echo -e "${YELLOW}🔍 Validating pod security context for $pod_name...${NC}"
    
    # Get pod security context
    local pod_sec_ctx=$(kubectl get pod "$pod_name" -n "$namespace" -o jsonpath='{.spec.securityContext}' 2>/dev/null || echo "{}")
    local container_sec_ctx=$(kubectl get pod "$pod_name" -n "$namespace" -o jsonpath='{.spec.containers[0].securityContext}' 2>/dev/null || echo "{}")
    
    local errors=0
    
    # Check pod-level security context
    if echo "$pod_sec_ctx" | jq -e '.runAsNonRoot == true' >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ Pod runs as non-root${NC}"
    else
        echo -e "${RED}  ❌ Pod does not run as non-root${NC}"
        ((errors++))
    fi
    
    if echo "$pod_sec_ctx" | jq -e '.seccompProfile.type == "RuntimeDefault"' >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ Pod uses RuntimeDefault seccomp profile${NC}"
    else
        echo -e "${RED}  ❌ Pod does not use RuntimeDefault seccomp profile${NC}"
        ((errors++))
    fi
    
    # Check container-level security context
    if echo "$container_sec_ctx" | jq -e '.allowPrivilegeEscalation == false' >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ Container disallows privilege escalation${NC}"
    else
        echo -e "${RED}  ❌ Container allows privilege escalation${NC}"
        ((errors++))
    fi
    
    if echo "$container_sec_ctx" | jq -e '.readOnlyRootFilesystem == true' >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ Container uses read-only root filesystem${NC}"
    else
        echo -e "${RED}  ❌ Container does not use read-only root filesystem${NC}"
        ((errors++))
    fi
    
    if echo "$container_sec_ctx" | jq -e '.capabilities.drop | index("ALL")' >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ Container drops ALL capabilities${NC}"
    else
        echo -e "${RED}  ❌ Container does not drop ALL capabilities${NC}"
        ((errors++))
    fi
    
    return $errors
}

# Function to validate namespace security labels
validate_namespace_security() {
    local namespace="$1"
    
    echo -e "${YELLOW}🔍 Validating namespace security labels for $namespace...${NC}"
    
    local labels=$(kubectl get namespace "$namespace" -o jsonpath='{.metadata.labels}' 2>/dev/null || echo "{}")
    local errors=0
    
    if echo "$labels" | jq -e '.["pod-security.kubernetes.io/enforce"] == "restricted"' >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ Namespace enforces restricted pod security${NC}"
    else
        echo -e "${RED}  ❌ Namespace does not enforce restricted pod security${NC}"
        ((errors++))
    fi
    
    return $errors
}

# Function to validate RBAC permissions
validate_rbac() {
    local namespace="$1"
    local service_account="$2"
    
    echo -e "${YELLOW}🔍 Validating RBAC permissions for $service_account in $namespace...${NC}"
    
    local errors=0
    
    # Check if service account exists
    if kubectl get serviceaccount "$service_account" -n "$namespace" >/dev/null 2>&1; then
        echo -e "${GREEN}  ✅ Service account exists${NC}"
    else
        echo -e "${RED}  ❌ Service account does not exist${NC}"
        ((errors++))
        return $errors
    fi
    
    # Check for role binding
    local role_binding=$(kubectl get rolebinding -n "$namespace" -o name 2>/dev/null | grep "$service_account" || echo "")
    if [ -n "$role_binding" ]; then
        echo -e "${GREEN}  ✅ Role binding exists${NC}"
    else
        echo -e "${RED}  ❌ Role binding not found${NC}"
        ((errors++))
    fi
    
    # Check for cluster role binding
    local cluster_role_binding=$(kubectl get clusterrolebinding -o name 2>/dev/null | grep "$service_account" || echo "")
    if [ -n "$cluster_role_binding" ]; then
        echo -e "${GREEN}  ✅ Cluster role binding exists${NC}"
    else
        echo -e "${YELLOW}  ⚠️  No cluster role binding found (may be intentional)${NC}"
    fi
    
    return $errors
}

# Main validation
main() {
    local total_errors=0
    
    echo -e "${YELLOW}🚀 Starting security validation...${NC}"
    echo
    
    # Check if namespace exists
    if ! kubectl get namespace "$NAMESPACE" >/dev/null 2>&1; then
        echo -e "${RED}❌ Namespace $NAMESPACE does not exist${NC}"
        exit 1
    fi
    
    # Validate namespace security labels
    validate_namespace_security "$NAMESPACE"
    ((total_errors += $?))
    echo
    
    # Get operator pods
    local pods=$(kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/name=$OPERATOR_NAME" -o name 2>/dev/null || echo "")
    
    if [ -z "$pods" ]; then
        echo -e "${RED}❌ No operator pods found in namespace $NAMESPACE${NC}"
        exit 1
    fi
    
    # Validate each pod
    while IFS= read -r pod_line; do
        if [ -n "$pod_line" ]; then
            local pod_name=$(echo "$pod_line" | sed 's|pod/||')
            validate_pod_security "$pod_name" "$NAMESPACE"
            ((total_errors += $?))
            echo
        fi
    done <<< "$pods"
    
    # Validate RBAC
    validate_rbac "$NAMESPACE" "helios-operator-controller-manager"
    ((total_errors += $?))
    echo
    
    # Summary
    echo -e "${YELLOW}📊 Security Validation Summary${NC}"
    if [ $total_errors -eq 0 ]; then
        echo -e "${GREEN}✅ All security validations passed!${NC}"
        echo -e "${GREEN}🎉 Helios Operator is compliant with Pod Security Standards${NC}"
        exit 0
    else
        echo -e "${RED}❌ Found $total_errors security issues${NC}"
        echo -e "${RED}🔧 Please review and fix the issues above${NC}"
        exit 1
    fi
}

# Run main function
main "$@"
