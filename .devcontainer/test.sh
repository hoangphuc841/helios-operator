#!/bin/bash
# Quick DevContainer Test Script
# This script helps verify the devcontainer configuration before full rebuild

set -e

echo "🧪 DevContainer Pre-flight Check"
echo "================================="
echo ""

# Check if required files exist
echo "📁 Checking configuration files..."
FILES=(
    ".devcontainer/devcontainer.json"
    ".devcontainer/setup.sh"
    ".devcontainer/health-check.sh"
    ".devcontainer/daemon.json"
)

for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "  ✅ $file exists"
    else
        echo "  ❌ $file is missing"
        exit 1
    fi
done
echo ""

# Validate JSON files
echo "🔍 Validating JSON syntax..."
if command -v jq &> /dev/null; then
    if jq . ".devcontainer/devcontainer.json" > /dev/null 2>&1; then
        echo "  ✅ devcontainer.json is valid"
    else
        echo "  ❌ devcontainer.json has syntax errors"
        exit 1
    fi

    if jq . ".devcontainer/daemon.json" > /dev/null 2>&1; then
        echo "  ✅ daemon.json is valid"
    else
        echo "  ❌ daemon.json has syntax errors"
        exit 1
    fi
else
    echo "  ⚠️  jq not found, skipping JSON validation"
fi
echo ""

# Check shell script syntax
echo "🐚 Checking shell scripts..."
for script in .devcontainer/*.sh; do
    if bash -n "$script" 2>/dev/null; then
        echo "  ✅ $(basename "$script") syntax is valid"
    else
        echo "  ❌ $(basename "$script") has syntax errors"
        exit 1
    fi
done
echo ""

# Check Docker
echo "🐳 Checking Docker..."
if command -v docker &> /dev/null; then
    echo "  ✅ Docker is installed"
    if docker info > /dev/null 2>&1; then
        echo "  ✅ Docker daemon is running"
    else
        echo "  ⚠️  Docker daemon is not running (this is OK, it will run in the container)"
    fi
else
    echo "  ❌ Docker is not installed"
    exit 1
fi
echo ""

# Check for old containers
echo "🗑️  Checking for old containers..."
OLD_CONTAINERS=$(docker ps -a --filter "label=devcontainer.local_folder" -q 2>/dev/null | wc -l)
if [ "$OLD_CONTAINERS" -gt 0 ]; then
    echo "  ⚠️  Found $OLD_CONTAINERS old devcontainer(s)"
    echo "  💡 Run: bash .devcontainer/cleanup.sh"
else
    echo "  ✅ No old containers found"
fi
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✨ Pre-flight check complete!"
echo ""
echo "📝 Next steps:"
echo "  1. Open VS Code"
echo "  2. Press F1 → 'Dev Containers: Reopen in Container'"
echo "  3. Wait for setup to complete (~2-5 minutes)"
echo "  4. Run: bash .devcontainer/health-check.sh"
echo ""
