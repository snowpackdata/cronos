# PR Guidelines

## Required Sections

| Section | Purpose |
|---------|---------|
| **Summary** | What this does (1-2 sentences) |
| **Why This Matters** | Impact on users/team/system |
| **Testing** | How verified |

## Conditional Sections

| Section | When |
|---------|------|
| **Changes** | Multi-file PRs |
| **Root Cause / Fix** | Bug fixes |
| **Alternatives Considered** | Non-trivial decisions |
| **Screenshots** | UI changes |
| **Migration** | Schema/config changes |
| **Breaking Changes** | API/behavior changes |

## Type Detection

- Commit messages contain "fix", "bug", "patch" → bugfix
- Commit messages contain "add", "feature", "implement" → feature
- Commit messages contain "refactor", "extract", "rename" → refactor
- Files are `.md`, config, deps only → docs

## Templates

- `.claude/templates/pr-bugfix.md`
- `.claude/templates/pr-feature.md`
- `.claude/templates/pr-refactor.md`
- `.claude/templates/pr-docs.md`

## Anti-Patterns

1. Vague titles
2. No "why"
3. Monster PRs
4. Missing tests
5. Secrets in diff
