#!/bin/bash
# Hook: Prevent PR creation from main branch
# Triggered by: gh pr create
#
# If user has unpushed commits on main, auto-fixes by:
# 1. Creating a new branch from the commit message
# 2. Moving the commits to that branch
# 3. Resetting main to origin/main

CURRENT_BRANCH=$(git branch --show-current)

if [ "$CURRENT_BRANCH" = "main" ]; then
    # Check for unpushed commits
    COMMITS_AHEAD=$(git rev-list --count origin/main..HEAD 2>/dev/null || echo "0")

    if [ "$COMMITS_AHEAD" -gt 0 ]; then
        echo "ERROR: Cannot create PR from main branch with $COMMITS_AHEAD unpushed commit(s)."
        echo ""
        echo "Auto-fixing: Creating new branch and moving commits..."

        # Generate branch name from first commit message
        BRANCH_NAME=$(git log -1 --format=%s | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | cut -c1-50)
        BRANCH_NAME="${USER:-dev}/${BRANCH_NAME}"

        # Handle branch name collision by appending timestamp
        if git show-ref --verify --quiet "refs/heads/$BRANCH_NAME"; then
            TIMESTAMP=$(date +%s)
            BRANCH_NAME="${BRANCH_NAME}-${TIMESTAMP}"
            echo "Branch already exists, using: $BRANCH_NAME"
        fi

        # Create new branch at current HEAD (with error handling)
        if ! git checkout -b "$BRANCH_NAME"; then
            echo "FAILED: Could not create branch '$BRANCH_NAME'"
            echo "Please manually create a branch and move your commits."
            exit 1
        fi

        # Reset main to origin/main (with error handling)
        if ! git branch -f main origin/main; then
            echo "WARNING: Could not reset main to origin/main"
            echo "You may need to manually run: git branch -f main origin/main"
        fi

        echo "SUCCESS: Moved to branch '$BRANCH_NAME'"
        echo "Continuing with PR creation..."
        exit 0
    fi

    echo "ERROR: Cannot create PR from main branch."
    exit 1
fi

exit 0
