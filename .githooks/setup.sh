#!/bin/bash
# Setup Git Hooks for Helios Operator

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GIT_DIR="$(git rev-parse --git-dir 2>/dev/null || echo ".git")"

echo "🔧 Setting up Git hooks for Helios Operator..."

# Check if we're in a git repository
if [ ! -d "$GIT_DIR" ]; then
    echo "❌ Error: Not in a git repository"
    exit 1
fi

# Configure git to use our hooks directory
git config core.hooksPath "$SCRIPT_DIR"

echo "✅ Git hooks configured successfully!"
echo ""
echo "The following hooks are now active:"
echo "  📝 pre-commit: Runs format, vet, lint, and tests before each commit"
echo ""
echo "To bypass hooks (not recommended), use:"
echo "  git commit --no-verify"
echo ""
echo "To disable hooks completely, run:"
echo "  git config --unset core.hooksPath"
