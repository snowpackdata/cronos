#!/bin/bash
# Hook script for PostToolUse on gh issue create
# Captures the issue number and outputs a reminder to run @issue-reviewer

set -e

# Parse stdin JSON
HOOK_DATA=$(cat)
TOOL_NAME=$(echo "$HOOK_DATA" | jq -r '.tool_name // empty')
COMMAND=$(echo "$HOOK_DATA" | jq -r '.tool_input.command // empty')
STDOUT=$(echo "$HOOK_DATA" | jq -r '.tool_response.stdout // empty')
EXIT_CODE=$(echo "$HOOK_DATA" | jq -r '.tool_response.exitCode // empty')

# Only proceed if this was a successful gh issue create command
if [[ "$TOOL_NAME" != "Bash" ]]; then
  exit 0
fi

if [[ ! "$COMMAND" =~ "gh issue create" ]]; then
  exit 0
fi

if [[ "$EXIT_CODE" != "0" ]]; then
  exit 0
fi

# Extract issue URL from stdout (gh issue create outputs the URL)
ISSUE_URL=$(echo "$STDOUT" | head -1 | tr -d '\n')
ISSUE_NUMBER=$(echo "$ISSUE_URL" | grep -oE '[0-9]+$' || true)

if [ -z "$ISSUE_NUMBER" ]; then
  exit 0
fi

# Output reminder (this will be visible in hook output)
echo "Issue #$ISSUE_NUMBER created: $ISSUE_URL"
echo "Run '@issue-reviewer #$ISSUE_NUMBER' to review and enrich the issue."

exit 0
