#!/bin/bash
# Hook: Prevent PR creation from main branch
# Triggered by: gh pr create

CURRENT_BRANCH=$(git branch --show-current)

if [ "$CURRENT_BRANCH" = "main" ] || [ "$CURRENT_BRANCH" = "master" ]; then
    # Check for unpushed commits
    COMMITS_AHEAD=$(git rev-list --count origin/main..HEAD 2>/dev/null || echo "0")

    if [ "$COMMITS_AHEAD" -gt 0 ]; then
        echo "ERROR: Cannot create PR from main branch with $COMMITS_AHEAD unpushed commit(s)."
        echo ""
        echo "Auto-fixing: Creating new branch and moving commits..."

        # Generate branch name from first commit message
        BRANCH_NAME=$(git log -1 --format=%s | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | cut -c1-50)
        BRANCH_NAME="${USER:-dev}/${BRANCH_NAME}"

        # Create new branch at current HEAD
        git checkout -b "$BRANCH_NAME"

        # Reset main to origin/main
        git branch -f main origin/main

        echo "SUCCESS: Moved to branch '$BRANCH_NAME'"
        echo "Continuing with PR creation..."
        exit 0
    fi

    echo "ERROR: Cannot create PR from main branch."
    exit 1
fi

exit 0
