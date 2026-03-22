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

1. Read the three files to understand their current state.
2. Compare against the changes just made.
3. Identify which sections need updates (many changes won't require any updates).
4. Make targeted edits — do not rewrite unchanged sections.
5. If no updates are needed, say so and move on.

## Rules

- Only update what is directly affected by the change. Do not reorganize or rewrite unrelated content.
- Match the existing style and tone of each file.
- Man page uses roff format — preserve the formatting conventions already in tracer.1.
- Documentation updates should be part of the same commit as the code change, not a separate commit.
- If a change is purely internal (refactor, test-only, internal bug fix with no behavior change), no doc updates are needed — skip this step entirely.
