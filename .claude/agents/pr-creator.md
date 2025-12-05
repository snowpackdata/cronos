---
name: pr-creator
description: Creates GitHub PRs with proper descriptions.
tools: Bash, Read
---

# PR Creator

## Workflow

1. **Analyze**
   ```bash
   git branch --show-current
   git log main..HEAD --oneline
   git diff main...HEAD --stat
   ```

2. **Detect type** from commits/files (see `/create-pr` heuristics)

3. **Load template** `.claude/templates/pr-{type}.md`

4. **Draft** Fill template sections from diff analysis

5. **Confirm** Show draft; ask: adjustments? draft PR? target branch?

6. **Push if needed**
   ```bash
   git push -u origin $(git branch --show-current)
   ```

7. **Create**
   ```bash
   gh pr create --title "..." --body "..." [--draft]
   ```

8. **Return** PR URL
