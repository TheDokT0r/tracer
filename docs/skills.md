# Skills

The Skills tab lets you browse, view, create, and manage Claude Code skills.

Skills are read from `~/.claude/` and include user skills, project skills, and built-in command skills.

## List View

Displays skills in a table with columns: Name, Source, Description.

### Key Bindings

| Key       | Action                   |
|-----------|--------------------------|
| `enter`/`v` | View skill content    |
| `e`       | Edit skill in $EDITOR    |
| `n`       | Create new skill         |
| `d`       | Delete skill             |
| `/`       | Filter by name or description |
| `tab`     | Switch to next tab       |

## Detail View

Shows the full content of a skill's SKILL.md file in a scrollable viewport.

| Key   | Action            |
|-------|-------------------|
| `e`   | Edit in $EDITOR   |
| `d`   | Delete            |
| `esc` | Back to list      |

## Creating Skills

Press `n` on the skills list to create a new skill. You'll be prompted for a name (kebab-case). Tracer creates `~/.claude/skills/<name>/SKILL.md` with a template and opens it in your editor.

You can also use `:new skill <name>` from the command palette.

## Read-Only Skills

Skills from built-in sources (like project `.claude/` directories) are read-only and cannot be edited or deleted through tracer.
