---
name: updating-docs
description: Checks and updates AGENTS.md, README.md, and the man page (tracer.1) when features are added or changed. Use after adding new features, modifying existing features, changing CLI flags, adding commands, or altering project structure. Triggers on any code change that affects user-facing behavior, CLI interface, project architecture, or tab/view functionality.
---

# Updating Documentation

## When This Applies

After any code change that:
- Adds a new feature, tab, view, command, or flag
- Modifies existing feature behavior or CLI interface
- Changes project structure (new packages, moved files)
- Adds or removes dependencies
- Changes keybindings, settings, or configuration options
- Alters the TUI layout or navigation

## Files to Check

Review each file and update **only the sections affected** by the change:

### 1. AGENTS.md
- Architecture overview, package descriptions, data flow
- Conventions and patterns
- Available features and their locations

### 2. README.md
- Feature list and descriptions
- Screenshots or demo references (flag if outdated)
- Installation or usage instructions
- CLI flags and options

### 3. tracer.1 (man page)
- Command synopsis and description
- Flag/option documentation
- Feature descriptions in the DESCRIPTION section
- Any new sections needed for new functionality

## Process

### Step 1: Determine what changed

Run `git diff` (staged + unstaged) to get the full picture of code changes. Understand what was added, modified, or removed.

### Step 2: Triage — do docs need updating?

Read AGENTS.md, README.md, and tracer.1. For each file, ask:
- Is there anything in the current changes that is **not already documented** or that **contradicts** what's documented?
- Would a user or contributor be misled by the current docs after this change lands?

If the answer is **no** for all three files — **skip updates entirely** and say "No doc updates needed" with a one-line reason why.

Most changes do NOT need doc updates. Examples that don't:
- Bug fixes that don't change documented behavior
- Internal refactors (renaming private functions, restructuring internals)
- Test additions or changes
- Performance improvements with no user-facing difference
- Code style or linting changes

### Step 3: Update only what's stale

If updates are needed, make targeted edits to only the sections that are now inaccurate or incomplete. Do not rewrite unchanged sections.

## Rules

- Only update what is directly affected by the change. Do not reorganize or rewrite unrelated content.
- Match the existing style and tone of each file.
- Man page uses roff format — preserve the formatting conventions already in tracer.1.
- Documentation updates should be part of the same commit as the code change, not a separate commit.
- If a change is purely internal (refactor, test-only, internal bug fix with no behavior change), no doc updates are needed — skip this step entirely.
