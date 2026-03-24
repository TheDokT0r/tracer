# Permissions

The Permissions tab manages Claude Code's `settings.json` allow/deny rules.

## List View

Displays all settings files found across three scopes:

| Scope   | Path                                  |
|---------|---------------------------------------|
| Global  | `~/.claude/settings.json`             |
| Project | `<project>/.claude/settings.json`     |
| Local   | `<project>/.claude/settings.local.json` |

Columns: Scope, Rules (count), Path.

### Key Bindings

| Key       | Action              |
|-----------|---------------------|
| `enter`/`v` | View rules        |
| `/`       | Filter by path      |
| `tab`     | Switch to next tab  |

## Detail View

Shows all allow and deny rules in a table.

### Key Bindings

| Key   | Action                     |
|-------|----------------------------|
| `a`   | Add new rule               |
| `t`   | Toggle rule (allow/deny)   |
| `d`   | Delete selected rule       |
| `esc` | Back to list               |

## Adding Rules

Press `a` to add a rule. The flow has two steps:

1. **Choose list** - Use `left`/`right` to toggle between `allow` and `deny`, then `enter` to confirm.
2. **Enter rule** - Type the rule pattern (e.g. `Bash(npm run *)`) and press `enter`.

Rules follow Claude Code's permission format (e.g. `Bash(git *)`, `Read(*.md)`).
