#!/bin/bash
# DevContainer Health Check Script
set -e

PASS="✅"
FAIL="❌"
WARN="⚠️ "

echo "🔍 Helios Operator DevContainer Health Check"
echo "=============================================="
echo ""

# Function to check command
check_command() {
    if command -v "$1" &> /dev/null; then
        echo "$PASS $1 is installed"
        if [ -n "$2" ]; then
            echo "   Version: $($1 $2 2>&1 | head -n1)"
        fi
        return 0
    else
        echo "$FAIL $1 is NOT installed"
        return 1
    fi
}

# Check user
echo "👤 User Configuration"
CURRENT_USER=$(whoami)
if [ "$CURRENT_USER" = "vscode" ]; then
    echo "$PASS Running as vscode user (UID: $(id -u), GID: $(id -g))"
else
    echo "$WARN Running as $CURRENT_USER (expected: vscode)"
fi
echo ""

# Check essential tools
echo "🛠️  Essential Tools"
check_command "go" "version"
check_command "docker" "--version"
check_command "kubectl" "version --client"
check_command "make" "--version | head -n1"
check_command "git" "--version"
echo ""

# Check Kubernetes tools
echo "☸️  Kubernetes Tools"
check_command "kind" "version"
check_command "kustomize" "version --short"
check_command "helm" "version --short"
echo ""

# Check Go tools
echo "🔧 Go Development Tools"
check_command "kubebuilder" "version"
check_command "controller-gen" "--version"
check_command "golangci-lint" "version"
command -v goimports &> /dev/null && echo "$PASS goimports is installed" || echo "$FAIL goimports is NOT installed"
command -v mockgen &> /dev/null && echo "$PASS mockgen is installed" || echo "$FAIL mockgen is NOT installed"
command -v setup-envtest &> /dev/null && echo "$PASS setup-envtest is installed" || echo "$FAIL setup-envtest is NOT installed"
echo ""

# Check Docker daemon
echo "🐳 Docker Status"
if docker info &> /dev/null; then
    echo "$PASS Docker daemon is running"
    echo "   Images: $(docker images -q 2>/dev/null | wc -l)"
    echo "   Containers: $(docker ps -a -q 2>/dev/null | wc -l)"
else
    echo "$FAIL Docker daemon is NOT running"
fi
echo ""

# Check volumes
echo "💾 Volume Mounts"
if [ -d "/go/pkg/mod" ]; then
    MOD_SIZE=$(du -sh /go/pkg/mod 2>/dev/null | cut -f1)
    echo "$PASS Go mod cache: /go/pkg/mod ($MOD_SIZE)"
else
    echo "$FAIL Go mod cache not found"
fi

if [ -d "/home/vscode/.cache/go-build" ]; then
    BUILD_SIZE=$(du -sh /home/vscode/.cache/go-build 2>/dev/null | cut -f1)
    echo "$PASS Go build cache: /home/vscode/.cache/go-build ($BUILD_SIZE)"
else
    echo "$FAIL Go build cache not found"
fi
echo ""

# Check environment variables
echo "🌍 Environment Variables"
ENV_VARS=("CGO_ENABLED" "GOPROXY" "GOSUMDB" "GOPATH")
for var in "${ENV_VARS[@]}"; do
    if [ -n "${!var}" ]; then
        echo "$PASS $var=${!var}"
    else
        echo "$FAIL $var is not set"
    fi
done
echo ""

echo "✨ Health check complete!"
echo ""
echo "💡 Run 'make help' to see available commands"
echo ""
echo ""

# Check workspace
echo "📁 Workspace"
if [ -f "/workspaces/helios-operator/go.mod" ]; then
    echo "$PASS go.mod found"
else
    echo "$FAIL go.mod not found"
fi

if [ -f "/workspaces/helios-operator/Makefile" ]; then
    echo "$PASS Makefile found"
else
    echo "$FAIL Makefile not found"
fi

if [ -d "/workspaces/helios-operator/.git" ]; then
    echo "$PASS Git repository found"
    BRANCH=$(git branch --show-current 2>/dev/null || echo "unknown")
    echo "   Branch: $BRANCH"
else
    echo "$FAIL Git repository not found"
fi
echo ""

# Check Git config
echo "🔐 Git Configuration"
SAFE_DIR=$(git config --global --get safe.directory 2>/dev/null || echo "")
if [ "$SAFE_DIR" = "/workspaces/helios-operator" ]; then
    echo "$PASS Git safe directory configured"
else
    echo "$WARN Git safe directory: $SAFE_DIR"
    echo "   Fixing..."
    git config --global --add safe.directory /workspaces/helios-operator
    echo "$PASS Fixed"
fi
echo ""

# Test Go modules
echo "📦 Go Modules"
cd /workspaces/helios-operator
if go list ./... &> /dev/null; then
    echo "$PASS Go modules are valid"
    PKG_COUNT=$(go list ./... | wc -l)
    echo "   Packages: $PKG_COUNT"
else
    echo "$FAIL Go modules have errors"
fi
echo ""

# Test make targets
echo "🎯 Makefile Targets"
if make help &> /dev/null; then
    echo "$PASS make help works"
else
    echo "$WARN make help failed"
fi
echo ""

# Performance check
echo "⚡ Performance Metrics"
echo "   CPU cores: $(nproc)"
echo "   Memory: $(free -h | awk '/^Mem:/ {print $2}')"
echo "   Disk free: $(df -h /workspaces | awk 'NR==2 {print $4}')"
echo ""

# Summary
echo "=============================================="
echo "🎉 Health check complete!"
echo ""
echo "Quick commands to try:"
echo "  make help       - Show available make targets"
echo "  make test       - Run tests"
echo "  go mod tidy     - Clean up dependencies"
echo "  kind-create     - Create local cluster"
echo ""
