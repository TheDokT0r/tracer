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
- Delete sessions permanently
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
| `tracer man` | View the manual page |
| `tracer -v` | Print version |
| `tracer -h` | Show help |

### Key Bindings

| Key | List View | Detail View |
|-----|-----------|-------------|
| `Enter` | Resume session | Resume session |
| `v` | View details | — |
| `c` | Copy session ID | Copy session ID |
| `d` | Delete session | Delete session |
| `/` | Filter | — |
| `Esc` | Clear filter | Back to list |
| `↑/↓` | Navigate | Scroll |
| `q` | Quit | Back to list |

## How It Works

tracer reads session data from `~/.claude/`:

- **Session files** (`projects/{path}/{sessionId}.jsonl`) — full conversation history
- **History** (`history.jsonl`) — detects `/rename` commands for custom session names

Startup is fast — sessions are scanned in parallel, reading only the first message per file. Full details (token counts, conversation) are loaded on demand when opening the detail view.

## License

MIT
