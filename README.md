# tracer

A TUI for browsing, inspecting, resuming, and deleting your [Claude Code](https://docs.anthropic.com/en/docs/claude-code) sessions.

![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-blue)

## Features

- Browse all your Claude Code sessions in an interactive table
- Filter sessions by name, directory, or branch
- View session details — metadata, context usage, and conversation history
- Resume any session directly from the TUI
- Copy session IDs to clipboard
- Pin sessions to the top of the list
- Delete sessions permanently
- In-app settings — theme, sort order, column visibility, confirm delete
- 4 color themes — default, minimal, ocean, rose
- Respects `/rename` — shows custom session names
- Self-updating — `tracer update`
- Built-in man page — `tracer man`

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
| `tracer man` | View the manual page |
| `tracer -v` | Print version |
| `tracer -h` | Show help |

### Key Bindings

| Key | List View | Detail View | Settings |
|-----|-----------|-------------|----------|
| `Enter` | Resume session | Resume session | Change value |
| `v` | View details | — | — |
| `c` | Copy session ID | Copy session ID | — |
| `p` | Pin/unpin | — | — |
| `d` | Delete session | Delete session | — |
| `s` | Open settings | — | — |
| `/` | Filter | — | — |
| `←/→` | — | — | Change value |
| `↑/↓` | Navigate | Scroll | Navigate |
| `Esc` | Clear filter | Back to list | Save & back |
| `q` | Quit | Back to list | Save & back |

### Settings

Press `s` from the list view to open settings. All settings persist in `~/.config/tracer/config.json`.

| Setting | Options | Default |
|---------|---------|---------|
| Theme | default, minimal, ocean, rose | default |
| Sort by | date, name, directory | date |
| Show date | on/off | on |
| Show directory | on/off | on |
| Show branch | on/off | on |
| Confirm delete | on/off | on |

## How It Works

tracer reads session data from `~/.claude/`:

- **Session files** (`projects/{path}/{sessionId}.jsonl`) — full conversation history
- **History** (`history.jsonl`) — detects `/rename` commands for custom session names

Startup is fast — sessions are scanned in parallel, reading only the first message per file. Full details (token counts, conversation) are loaded on demand when opening the detail view.

## License

MIT
