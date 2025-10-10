#!/bin/bash
set -euo pipefail

echo "🚀 Setting up Helios Operator Development Environment..."

readonly WORKSPACE="/workspaces/helios-operator"
readonly DAEMON_CONFIG_SOURCE="$WORKSPACE/.devcontainer/daemon.json"
readonly DAEMON_CONFIG_TARGET="/etc/docker/daemon.json"

# Tool versions (keep in sync with Makefile when possible)
readonly KUBEBUILDER_VERSION="4.9.0"
readonly KIND_VERSION="v0.30.0"
readonly KUSTOMIZE_VERSION="v5.7.1"
readonly CONTROLLER_TOOLS_VERSION="v0.19.0"
readonly GOLANGCI_LINT_VERSION="v2.5.0"
readonly GOIMPORTS_VERSION="v0.38.0"
readonly MOCKGEN_VERSION="v0.6.0"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
	echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
	echo -e "${GREEN}✅ $1${NC}"
}

log_warn() {
	echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
	echo -e "${RED}❌ $1${NC}" >&2
}

ensure_directory() {
	local dir="$1"
	sudo mkdir -p "$dir"
}

wait_for_docker() {
	log_info "Waiting for Docker daemon to be ready..."
	local attempts=0
	until docker info >/dev/null 2>&1 || [ "$attempts" -ge 30 ]; do
		attempts=$((attempts + 1))
		sleep 1
	done

	if docker info >/dev/null 2>&1; then
		log_success "Docker daemon is ready"
	else
		log_warn "Docker daemon not ready after 30 seconds, continuing anyway"
	fi
}

configure_docker_daemon() {
	log_info "Configuring Docker daemon..."
	ensure_directory "$(dirname "$DAEMON_CONFIG_TARGET")"

	if [ -f "$DAEMON_CONFIG_SOURCE" ]; then
		if [ -f "$DAEMON_CONFIG_TARGET" ] && ! sudo cmp -s "$DAEMON_CONFIG_SOURCE" "$DAEMON_CONFIG_TARGET"; then
			sudo cp "$DAEMON_CONFIG_TARGET" "${DAEMON_CONFIG_TARGET}.bak"
		fi
		sudo cp "$DAEMON_CONFIG_SOURCE" "$DAEMON_CONFIG_TARGET"
		log_success "Docker daemon configuration updated"
	else
		log_warn "Daemon configuration source not found (skipping)"
	fi
}

install_apt_packages() {
	log_info "Updating package list..."
	sudo apt-get update -qq

	local packages=(
		make
		wget
		curl
		jq
		vim
		less
		ca-certificates
		gnupg
		build-essential
		tree
		htop
		bash-completion
	)

	log_info "Installing essential development tools..."
	sudo apt-get install -y --no-install-recommends "${packages[@]}"
	log_success "Essential tools installed"
}

install_go_tool() {
	local binary="$1"
	local module="$2"
	local version="$3"
	local temp_dir
	temp_dir=$(mktemp -d)

	if command -v "$binary" >/dev/null 2>&1; then
		local current_version
		current_version=$("$binary" --version 2>/dev/null | head -n1 || true)
		if [[ -n "$current_version" && "$current_version" == *"$version"* ]]; then
			log_success "$binary $version already installed"
			rm -rf "$temp_dir"
			return
		fi
	fi

	log_info "Installing $binary ($version)..."
	if ! output=$(GOBIN="$temp_dir" go install "${module}@${version}" 2>&1); then
		rm -rf "$temp_dir"
		log_error "Failed to install $binary ${version}"
		if [[ -n "$output" ]]; then
			echo "$output" >&2
		fi
		exit 1
	fi
	sudo install -m 0755 "$temp_dir/$binary" "/usr/local/bin/$binary"
	log_success "$binary $version installed"
	rm -rf "$temp_dir"
}

install_kubebuilder() {
	if command -v kubebuilder >/dev/null 2>&1 && kubebuilder version 2>&1 | grep -q "$KUBEBUILDER_VERSION"; then
		log_success "Kubebuilder ${KUBEBUILDER_VERSION} already installed"
		return
	fi

	log_info "Installing Kubebuilder ${KUBEBUILDER_VERSION}..."
	local temp_dir
	temp_dir=$(mktemp -d)
	curl -fsSL -o "$temp_dir/kubebuilder" "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_VERSION}/kubebuilder_linux_amd64"
	sudo install -m 0755 "$temp_dir/kubebuilder" /usr/local/bin/kubebuilder
	rm -rf "$temp_dir"
	log_success "Kubebuilder ${KUBEBUILDER_VERSION} installed"
}

install_kind() {
	if command -v kind >/dev/null 2>&1 && kind version 2>&1 | grep -q "$KIND_VERSION"; then
		log_success "kind ${KIND_VERSION} already installed"
		return
	fi

	log_info "Installing kind ${KIND_VERSION}..."
	local temp_dir
	temp_dir=$(mktemp -d)
	curl -fsSL -o "$temp_dir/kind" "https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-linux-amd64"
	sudo install -m 0755 "$temp_dir/kind" /usr/local/bin/kind
	rm -rf "$temp_dir"
	log_success "kind ${KIND_VERSION} installed"
}

configure_path() {
	local profile_d="/etc/profile.d/helios-operator-path.sh"
	if ! sudo test -f "$profile_d"; then
		echo "export PATH=\"$WORKSPACE/bin:\$PATH\"" | sudo tee "$profile_d" >/dev/null
		sudo chmod 0644 "$profile_d"
	fi
	export PATH="$WORKSPACE/bin:$PATH"
}

install_project_tools() {
	log_info "Installing Go-based CLI tools..."
	install_go_tool "kustomize" "sigs.k8s.io/kustomize/kustomize/v5" "$KUSTOMIZE_VERSION"
	install_go_tool "controller-gen" "sigs.k8s.io/controller-tools/cmd/controller-gen" "$CONTROLLER_TOOLS_VERSION"
	install_go_tool "golangci-lint" "github.com/golangci/golangci-lint/v2/cmd/golangci-lint" "$GOLANGCI_LINT_VERSION"
	install_go_tool "goimports" "golang.org/x/tools/cmd/goimports" "$GOIMPORTS_VERSION"
	install_go_tool "mockgen" "go.uber.org/mock/mockgen" "$MOCKGEN_VERSION"
}

install_envtest() {
	if [ ! -f "$WORKSPACE/go.mod" ]; then
		log_warn "go.mod not found; skipping envtest installation"
		return
	fi

	local controller_runtime_version
	controller_runtime_version=$(cd "$WORKSPACE" && go list -m -f "{{ .Version }}" sigs.k8s.io/controller-runtime 2>/dev/null || echo "")
	local envtest_version="release-0.19"
	if [[ -n "$controller_runtime_version" ]]; then
		envtest_version=$(echo "$controller_runtime_version" | awk -F'[v.]' '{printf "release-%d.%d", $2, $3}')
	fi

	local temp_dir
	temp_dir=$(mktemp -d)
	log_info "Installing setup-envtest (${envtest_version})..."
	if ! GOBIN="$temp_dir" go install "sigs.k8s.io/controller-runtime/tools/setup-envtest@${envtest_version}" >/dev/null 2>&1; then
		rm -rf "$temp_dir"
		log_error "Failed to install setup-envtest ${envtest_version}"
		exit 1
	fi
	sudo install -m 0755 "$temp_dir/setup-envtest" /usr/local/bin/setup-envtest
	log_success "setup-envtest ${envtest_version} installed"
	rm -rf "$temp_dir"
}

download_go_modules() {
	if [ -f "$WORKSPACE/go.mod" ]; then
		log_info "Downloading Go modules..."
		(cd "$WORKSPACE" && go mod download)
		log_success "Go modules downloaded"
	else
		log_warn "go.mod not found; skipping Go module download"
	fi
}

configure_git() {
	log_info "Configuring git..."
	git config --global --add safe.directory "$WORKSPACE"
	git config --global init.defaultBranch main
	log_success "Git configured"
}

fix_permissions() {
	log_info "Fixing permissions..."
	sudo chown -R vscode:vscode /go "$WORKSPACE" 2>/dev/null || true
	log_success "Permissions fixed"
}

cleanup() {
	log_info "Cleaning up package cache..."
	sudo apt-get clean
	sudo rm -rf /var/lib/apt/lists/*
	log_success "Cleanup complete"
}

main() {
	wait_for_docker
	configure_docker_daemon
	install_apt_packages
	configure_path
	download_go_modules
	install_kubebuilder
	install_kind
	install_project_tools
	install_envtest
	configure_git
	fix_permissions
	cleanup

	echo ""
	echo "🎉 Setup complete! Development environment is ready."
	echo ""
	echo "📝 Quick Start:"
	echo "  - Run 'make help' to see available commands"
	echo "  - Run 'bash .devcontainer/health-check.sh' to verify the setup"
	echo ""
}

main "$@"

