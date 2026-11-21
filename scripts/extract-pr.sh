#!/usr/bin/env bash
# scripts/extract-pr.sh - BDD Chem Zero-Pollution Ephemeral PR Extractor
# Usage: ./scripts/extract-pr.sh <pr-name> [source-branch]
#
# Extracts a clean, upstream-ready PR branch directly from upstream/main,
# excluding docs/decisions/, features/, and personal agent configs.

set -euo pipefail

cd "$(dirname "$0")/.."

if [ $# -lt 1 ]; then
    echo "Usage: $0 <pr-name> [source-branch]"
    echo "Example: $0 pane-commander feat/pane-commander"
    exit 1
fi

PR_NAME="$1"
SOURCE_BRANCH="${2:-$(git branch --show-current)}"
TARGET_BRANCH="pr/${PR_NAME}"
UPSTREAM_BASE="upstream/main"

# Check for uncommitted working tree changes
if ! git diff-index --quiet HEAD --; then
    echo "Error: Working directory has uncommitted changes. Commit or stash them first."
    exit 1
fi

# Ensure upstream is available
if ! git rev-parse --verify "$UPSTREAM_BASE" >/dev/null 2>&1; then
    echo "Fetching upstream..."
    git fetch upstream main
fi

echo "==> Extracting clean PR branch '${TARGET_BRANCH}' from ${UPSTREAM_BASE} using ${SOURCE_BRANCH}"

# Create or reset target branch from upstream/main
git checkout -B "${TARGET_BRANCH}" "${UPSTREAM_BASE}"

# List changed files between upstream/main and source branch
CHANGED_FILES=$(git diff --name-only "${UPSTREAM_BASE}"..."${SOURCE_BRANCH}")

DISALLOWED_PATTERNS=("^docs/decisions/" "^features/" "^docs/adr/" "^\.gemini/" "^scratch/")
ALLOWED_FILES=()
REJECTED_FILES=()

for f in ${CHANGED_FILES}; do
    disallowed=0
    for pat in "${DISALLOWED_PATTERNS[@]}"; do
        if [[ "$f" =~ $pat ]]; then
            disallowed=1
            break
        fi
    done
    if [ "$disallowed" -eq 1 ]; then
        REJECTED_FILES+=("$f")
    else
        ALLOWED_FILES+=("$f")
    fi
done

if [ ${#ALLOWED_FILES[@]} -eq 0 ]; then
    echo "Error: No production files to extract from ${SOURCE_BRANCH}."
    git checkout "${SOURCE_BRANCH}"
    exit 1
fi

echo "==> Checked out production files (${#ALLOWED_FILES[@]} files):"
for f in "${ALLOWED_FILES[@]}"; do
    git checkout "${SOURCE_BRANCH}" -- "$f"
    echo "    + $f"
done

if [ ${#REJECTED_FILES[@]} -gt 0 ]; then
    echo "==> Filtered internal BDD / decision files (${#REJECTED_FILES[@]} files excluded):"
    for f in "${REJECTED_FILES[@]}"; do
        echo "    - $f"
    done
fi

# Validation check: Ensure NO disallowed files are staged or untracked
for pat in "${DISALLOWED_PATTERNS[@]}"; do
    if git status --porcelain | grep -qE "${pat}"; then
        echo "Error: Found internal files in working tree for ${TARGET_BRANCH}!"
        git status --porcelain
        git checkout "${SOURCE_BRANCH}"
        exit 1
    fi
done

echo ""
echo "==> Verification: Branch '${TARGET_BRANCH}' is clean and ready."
echo "    Review changes with: git diff --cached"
echo "    Commit with:         git commit -m 'feat(...): ...'"
echo "    Return to feature:   git checkout ${SOURCE_BRANCH}"
