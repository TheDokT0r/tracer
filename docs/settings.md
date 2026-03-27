# Settings

Open settings with `s` from the list view, or `:settings` from the command palette.

Settings are saved explicitly with `⌘+s` (macOS) or `ctrl+s`. If you exit with unsaved changes, a confirmation prompt appears.

## Available Settings

### General

| Setting              | Values           | Default | Description                      |
|----------------------|------------------|---------|----------------------------------|
| theme                | see [themes](themes.md) | default | Color theme               |
| sort_by              | date, name, directory | date | Session sort order              |
| model                | model name       | (empty) | Model to pass to Claude on resume |
| confirm_delete       | on, off          | on      | Require y/N before deleting      |
| auto_update          | on, off          | off     | Auto-install updates on exit     |

### Columns

| Setting              | Values           | Default | Description                      |
|----------------------|------------------|---------|----------------------------------|
| show_date            | on, off          | on      | Show Date column                 |
| show_directory       | on, off          | on      | Show Directory column            |
| show_branch          | on, off          | on      | Show Branch column               |
| show_model           | on, off          | off     | Show Model column                |
| show_agent           | on, off          | off     | Show Agent column (claude/codex/gemini) |

Custom columns also appear here when defined.

### Command Palette

| Setting              | Values           | Default | Description                      |
|----------------------|------------------|---------|----------------------------------|
| cmd_dropdown         | on, off          | on      | Show autocomplete dropdown       |
| cmd_ghost            | on, off          | off     | Show inline ghost text           |
| cmd_max_suggestions  | 3-12             | 8       | Max dropdown items               |

### Agents

| Setting              | Values           | Default | Description                      |
|----------------------|------------------|---------|----------------------------------|
| agent_claude         | on, off          | on      | Scan Claude Code sessions        |
| agent_codex          | on, off          | on      | Scan Codex CLI sessions          |
| agent_gemini         | on, off          | on      | Scan Gemini CLI sessions         |

Disabling an agent skips its session scan entirely at startup.

## Navigation (Settings View)

| Key             | Action             |
|-----------------|--------------------|
| `up`/`down` or `k`/`j` | Navigate   |
| `left`/`right` or `h`/`l` or `enter` | Change value |
| `⌘+s` / `ctrl+s` | Save             |
| `esc` or `q`   | Exit (prompts if unsaved) |

## Changing from Command Palette

```
:set <key> <value>
```

Examples:
```
:set theme dracula
:set sort_by name
:set show_agent on
:set agent_codex off
:set model claude-opus-4-6[1m]
```

## Storage

`~/.config/tracer/config.json`
