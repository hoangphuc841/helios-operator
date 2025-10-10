#!/bin/bash
set -euo pipefail

WORKSPACE="/workspaces/helios-operator"
CACHE_DIR="/home/vscode/.cache/go-build"
MODULE_CACHE="/go/pkg/mod"
GOCACHE_PARENT="/go/pkg"

ensure_dir() {
  local dir="$1"
  if [ ! -d "$dir" ]; then
    sudo mkdir -p "$dir"
  fi
}

ensure_dir "$CACHE_DIR"
ensure_dir "$GOCACHE_PARENT"
ensure_dir "$MODULE_CACHE"

sudo chown -R vscode:vscode "$CACHE_DIR"
sudo chown -R vscode:vscode "$GOCACHE_PARENT"
sudo chown -R vscode:vscode /go

if [ -f "$WORKSPACE/go.mod" ]; then
  (cd "$WORKSPACE" && go mod download)
fi
