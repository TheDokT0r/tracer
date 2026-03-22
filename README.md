# tracer

A TUI for managing your [Claude Code](https://docs.anthropic.com/en/docs/claude-code) sessions and skills.

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-blue)

![tracer demo](assets/demo.png)

## Features

### Sessions
- Browse all your Claude Code sessions in an interactive table
- Filter sessions by name, directory, or branch
- View session details — metadata, context usage, and conversation history
- Resume or start new sessions directly from the TUI
- Rename sessions inline
- Edit session files in `$EDITOR`
- Copy session IDs to clipboard
- Pin sessions to the top of the list
- Delete sessions permanently

### Skills
- Browse all installed skills across user, command, project, and plugin sources
- View full skill content with metadata (source, path, size)
- Edit user and command skills in `$EDITOR`
- Create new skills from a template
- Delete user and command skills (plugin skills are read-only)
- Filter skills by name or description

### Permissions
- Browse all Claude Code settings files (global, project, local)
- View and manage allow/deny permission rules per file
- Add new rules interactively (pick allow/deny, type the rule)
- Toggle rules between allow and deny
- Delete rules — changes saved immediately

### General
- **Tab switching** — `Tab` to cycle between Sessions, Skills, and Permissions
- **12 color themes** — default, minimal, mono, ocean, rose, forest, sunset, nord, dracula, solarized, monokai, catppuccin
- **In-app settings** — theme, sort order, column visibility, confirm delete, auto-update
- **Interactive theme picker** — `tracer theme` with live preview
- **Self-updating** — `tracer update`
- **Built-in man page** — `tracer man`
- Respects Claude Code's `/rename` command

## Install

### Quick install (macOS / Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/TheDokT0r/tracer/master/install.sh | sh
```

Installs binary to `~/.local/bin` and man page to `~/.local/share/man/man1`.

### Go install

```bash
go install github.com/TheDokT0r/tracer@latest
```

### From source

```bash
git clone https://github.com/TheDokT0r/tracer.git
cd tracer
go build -o tracer .
```

### Manual download

Grab the latest `.tar.gz` for your platform from [Releases](https://github.com/TheDokT0r/tracer/releases) and extract it somewhere on your `$PATH`.

## Usage

```bash
tracer
```

### Commands

| Command | Description |
|---------|-------------|
| `tracer` | Launch the TUI |
| `tracer update` | Update to the latest release |
| `tracer theme` | Interactive theme picker with live preview |
| `tracer theme <name>` | Set theme directly |
| `tracer settings` | Open settings |
| `tracer man` | View the manual page |
| `tracer -v` | Print version |
| `tracer -h` | Show help |

### Key Bindings

#### Sessions Tab

| Key | List View | Detail View |
|-----|-----------|-------------|
| `Enter` | Resume session | Resume session |
| `n` | New session | — |
| `v` | View details | — |
| `r` | — | Rename session |
| `e` | — | Edit session file |
| `c` | Copy session ID | Copy session ID |
| `p` | Pin/unpin | — |
| `d` | Delete | Delete |
| `s` | Open settings | — |
| `/` | Filter | — |
| `f` | Fork session | Fork session |
| `Tab` | Next tab | — |
| `↑/↓` | Navigate | Scroll |
| `Esc` | Clear filter | Back to list |
| `q` | Quit | Back to list |

#### Skills Tab

| Key | List View | Detail View |
|-----|-----------|-------------|
| `Enter`/`v` | View details | — |
| `e` | Edit skill | Edit skill |
| `n` | Create new skill | — |
| `d` | Delete skill | Delete skill |
| `/` | Filter | — |
| `Tab` | Next tab | — |
| `↑/↓` | Navigate | Scroll |
| `Esc` | Clear filter | Back to list |
| `q` | Quit | Back to list |

#### Permissions Tab

| Key | List View | Detail View |
|-----|-----------|-------------|
| `Enter`/`v` | View rules | — |
| `a` | — | Add rule |
| `t` | — | Toggle allow/deny |
| `d` | — | Delete rule |
| `/` | Filter | — |
| `Tab` | Next tab | — |
| `↑/↓` | Navigate | Navigate |
| `Esc` | Clear filter | Back to list |
| `q` | Quit | Back to list |

### Settings

Press `s` from the sessions list or run `tracer settings`. All settings persist in `~/.config/tracer/config.json`.

| Setting | Options | Default |
|---------|---------|---------|
| Theme | 11 themes (run `tracer theme` to preview) | default |
| Sort by | date, name, directory | date |
| Show date | on/off | on |
| Show directory | on/off | on |
| Show branch | on/off | on |
| Confirm delete | on/off | on |
| Auto update | on/off | off |

## How It Works

tracer reads data from `~/.claude/`:

- **Session files** (`projects/{path}/{sessionId}.jsonl`) — full conversation history
- **History** (`history.jsonl`) — detects `/rename` commands for custom session names
- **Skills** (`skills/`, `commands/`, `plugins/cache/`) — skill definitions and commands
- **Settings** (`settings.json`, `.claude/settings.json`) — permission rules (allow/deny)

Startup is fast — sessions are scanned in parallel, reading only the first message per file. Auto-update checks run in the background and apply after the TUI exits, so they never block startup. Full details (token counts, conversation) are loaded on demand when opening the detail view.

## License

MIT
