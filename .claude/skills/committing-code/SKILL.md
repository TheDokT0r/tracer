---
name: committing-code
description: Creates git commits following tracer's conventions and release workflow. Use when creating git commits, preparing commit messages, or when the user asks to commit changes. Triggers on "commit", "git commit", "save changes", or any request to record changes to version control. Ensures correct commit types so the CI release workflow bumps versions properly.
---

# Committing Code in Tracer

## Overview

Tracer uses **conventional commits with gitmoji** and has a CI release workflow that auto-bumps versions based on commit type. Choosing the wrong type means no release gets created.

## Format

```
<emoji> <type>: <short description>

<body — what changed and why>

Co-Authored-By: Claude <agent> <noreply@anthropic.com>
```

Short description: imperative mood, lowercase, no period. Body uses bullet points for multiple changes. Always use HEREDOC for multi-line messages.

## Release-Triggering Types

The release workflow (`.github/workflows/release.yml`) only creates a new version for these commit types:

| Type | Version Bump | When to use |
|------|-------------|-------------|
| `feat:` | **minor** (0.X.0) | New feature or capability |
| `fix:` | **patch** (0.0.X) | Bug fix |
| `perf:` | **patch** (0.0.X) | Performance improvement |
| `refactor:` | **patch** (0.0.X) | Code restructuring |
| `style:` | **patch** (0.0.X) | UI/cosmetic change |
| `security:` | **patch** (0.0.X) | Security fix |
| `<type>!:` or `BREAKING CHANGE` | **major** (X.0.0) | Breaking change to CLI or behavior |

## Non-Releasing Types

These types are valid but **will not trigger a release**. Use them only when no release is needed:

| Emoji | Type | When to use |
|-------|------|-------------|
| 📝 | `docs` | Documentation only |
| 🔧 | `chore` | Config, tooling, CI, non-code changes |
| ✅ | `test` | Adding or updating tests |

## Choosing the Right Type

**If the change improves user experience, it should trigger a release.** Ask:

- Does the app behave differently for the user? → `feat:` or `fix:`
- Did something get faster, more reliable, or less broken? → `fix:`
- Is it purely internal with zero user-visible effect? → `refactor:`, `chore:`, etc.

**Common mistakes:**
- `chore:` for a config change that affects behavior — use `fix:` or `feat:`.
- `docs:` for a commit that also includes code changes — use the code type instead.

## Merge Commits

The release workflow inspects non-merge commit messages (`git log --no-merges`). GitHub merge commits (`Merge pull request #N from ...`) are excluded. This means:

- **Squash merges:** The squashed commit message must start with `feat:` or `fix:` to trigger a release.
- **Merge commits:** At least one non-merge commit on the branch must start with `feat:` or `fix:`.

## Gitmoji Reference

| Emoji | Type | When to use |
|-------|------|-------------|
| ✨ | `feat` | New feature or capability |
| 🐛 | `fix` | Bug fix |
| 🚀 | `perf` | Performance improvement |
| ♻️ | `refactor` | Code restructuring without behavior change |
| 📝 | `docs` | Documentation only |
| 🔧 | `chore` | Config, tooling, non-code changes |
| ✅ | `test` | Adding or updating tests |
| 🔥 | `chore` | Removing code or files |
| 🏗️ | `refactor` | Architectural change |
| 💄 | `style` | UI/cosmetic change |
| 🔒 | `security` | Security fix |
| ⬆️ | `chore` | Dependency upgrade |
| 🚚 | `refactor` | Moving or renaming files |
| 🎉 | `feat` | Initial commit |

## Pre-Commit Checklist

Before committing, check if the change requires doc updates (see `updating-docs` skill):
- **AGENTS.md** — project structure, conventions, architecture
- **README.md** — user-facing features, settings, usage
- **tracer.1** — man page, CLI flags, feature descriptions

Documentation updates go in the **same commit** as the code, not a separate docs commit.

## Rules

1. **One type per commit.** Split unrelated changes into separate commits.
2. **Body explains WHY, not just WHAT.** The diff shows what; the message explains reasoning.
3. **Always include Co-Authored-By** when AI-assisted.
4. **Use HEREDOC** for multi-line messages:
   ```bash
   git commit -m "$(cat <<'EOF'
   ✨ feat: add session search

   - Add fuzzy search with `/` keybinding
   - Filter matches name, directory, and branch

   Co-Authored-By: Claude <agent> <noreply@anthropic.com>
   EOF
   )"
   ```
5. **Stage only relevant files** — never `git add -A` or `git add .`.

## Red Flags — STOP

| Temptation | Correct Action |
|-----------|----------------|
| Batching docs + code in a `docs:` commit | Use the code type (`feat:` / `fix:`) and include docs |
| Using `chore:` for something that affects users | Use `fix:` or `feat:` so a release is triggered |
| Amending a published commit | Create a new commit instead |
| Skipping hooks with `--no-verify` | Fix the hook issue, don't bypass it |
