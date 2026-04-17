<p align="center">
  <h1 align="center">tracer</h1>
  <p align="center">
    A terminal UI for managing your AI coding sessions across <a href="https://docs.anthropic.com/en/docs/claude-code">Claude Code</a>, <a href="https://github.com/openai/codex">Codex CLI</a>, and <a href="https://github.com/google-gemini/gemini-cli">Gemini CLI</a>.
    <br />
    <br />
    <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white" />
    <img src="https://img.shields.io/badge/License-MIT-blue" />
    <img src="https://img.shields.io/badge/Platform-macOS%20%7C%20Linux-lightgrey" />
  </p>
</p>

<br />

<p align="center">
  <img src="assets/demo.png" alt="tracer demo" width="800" />
</p>

<br />

## Why tracer?

AI coding tools store sessions as raw files scattered across your home directory. tracer gives you a fast, searchable interface to **browse**, **resume**, **fork**, and **manage** sessions from all your agents in one place — plus skills and permission rules — without leaving the terminal.

### Supported Agents

| Agent | Sessions | Resume | Fork | Skills | Permissions |
|-------|----------|--------|------|--------|-------------|
| **Claude Code** | `~/.claude/` | yes | yes | yes | yes |
| **Codex CLI** | `~/.codex/` | yes | yes | — | — |
| **Gemini CLI** | `~/.gemini/` | view only | — | — | — |

All agents are scanned in parallel at startup. Enable or disable each in settings.

## Quick Start

```bash
curl -fsSL https://raw.githubusercontent.com/orkwitzel/tracer/master/install.sh | sh
```

Then just run `tracer`.

<details>
<summary>Other install methods</summary>

**Go install**
```bash
go install github.com/orkwitzel/tracer@latest
```

**From source**
```bash
git clone https://github.com/orkwitzel/tracer.git
cd tracer && go build -o tracer .
```

**Manual download** — grab the latest `.tar.gz` from [Releases](https://github.com/orkwitzel/tracer/releases).

</details>

## Features

### Sessions

| Action | How |
|--------|-----|
| Browse all sessions (all agents) | Just launch `tracer` |
| Filter by name, directory, or branch | `/` then type |
| Resume a session | `Enter` (routes to correct agent) |
| Fork a session | `f` (Claude and Codex) |
| Start a new session | `n` (pick agent if multiple enabled) |
| View details (context usage, conversation) | `v` |
| Rename a session | `r` in detail view |
| Export as Markdown or HTML | `x` in detail view |
| Pin to top | `p` |
| Copy session ID | `c` |
| Delete | `d` |

### Skills

Press `Tab` to switch to the Skills tab. Manages Claude Code skills.

| Action | How |
|--------|-----|
| Browse all skills (user, project, plugin) | Skills tab |
| View skill content | `Enter` or `v` |
| Edit a skill | `e` |
| Create a new skill | `n` |
| Delete a skill | `d` |

Plugin skills are read-only.

### Permissions

Press `Tab` again to reach the Permissions tab. Manages Claude Code permission rules.

| Action | How |
|--------|-----|
| Browse all settings files (global + project) | Permissions tab |
| View allow/deny rules | `Enter` or `v` |
| Add a new rule | `a` |
| Toggle allow/deny | `t` |
| Delete a rule | `d` |

Changes save immediately to `settings.json`.

### Themes

12 built-in themes. Preview them interactively:

```bash
tracer theme
```

Available: `default` `minimal` `mono` `ocean` `rose` `forest` `sunset` `nord` `dracula` `solarized` `monokai` `catppuccin`

### Command Palette

Press `:` in any view to open the command palette. Type commands with autocomplete:

```
:sort name          Sort sessions by name
:set theme dracula  Switch theme
:model opus         Set model for Claude sessions
:export html        Export session as HTML
:filter react       Filter by "react"
:help               List all commands
```

### User Commands

Create custom commands that extend the palette:

```bash
:commands new deploy       # Scaffold and open in $EDITOR
:commands new-ai deploy    # AI-assisted creation (launches Claude)
:commands edit deploy       # Edit existing command
:commands delete deploy     # Delete command
```

### Custom Columns

Add custom data columns to the session list:

```bash
:columns new cost          # Scaffold a new column
:columns new-ai cost       # AI-assisted creation
:columns toggle cost       # Show/hide column
```

### Settings

Press `s` or run `tracer settings`. Save with `⌘+s` / `ctrl+s`.

| Section | Settings |
|---------|----------|
| General | Theme, Sort by, Model, Confirm delete, Auto-update |
| Columns | Date, Directory, Branch, Model, Agent, custom columns |
| Command Palette | Dropdown, Ghost suggest, Max suggestions |
| Agents | Claude (on/off), Codex (on/off), Gemini (on/off) |

## Commands

```
tracer                Launch the TUI
tracer update         Update to the latest release
tracer theme          Interactive theme picker
tracer theme <name>   Set theme directly
tracer settings       Open settings
tracer man            View the manual page
tracer -v             Print version
```

## Documentation

Detailed docs for each feature are in the [`docs/`](docs/) directory:

- [Sessions](docs/sessions.md) — multi-agent browsing, resuming, forking, pinning, renaming
- [Skills](docs/skills.md) — browsing, creating, editing skills
- [Permissions](docs/permissions.md) — managing allow/deny rules
- [Commands](docs/commands.md) — command palette and user-defined commands
- [Custom Columns](docs/custom-columns.md) — script-powered table columns
- [Themes](docs/themes.md) — 12 built-in color themes
- [Settings](docs/settings.md) — all configuration options
- [Export](docs/export.md) — Markdown and HTML export
- [CLI](docs/cli.md) — subcommands, flags, auto-update
- [Keybindings](docs/keybindings.md) — complete keyboard reference

## How It Works

tracer scans sessions from three sources in parallel:

- **Claude Code** (`~/.claude/`) — JSONL files in `projects/`, only first message read for speed
- **Codex CLI** (`~/.codex/`) — JSONL files in `sessions/`, with thread names from `session_index.jsonl`
- **Gemini CLI** (`~/.gemini/`) — JSON files in `tmp/*/chats/`

Additionally for Claude Code:
- **Skills** — `skills/`, `commands/`, and `plugins/cache/`
- **Permissions** — `settings.json` at global, project, and local scopes

Full session details (token counts, conversation history) load on demand. Auto-update checks run in the background and apply after you exit.

## License

MIT
