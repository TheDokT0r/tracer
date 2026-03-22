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

## Install

```bash
go install github.com/TheDokT0r/tracer@latest
```

Or build from source:

```bash
git clone https://github.com/TheDokT0r/tracer.git
cd tracer
go build -o tracer .
```

## Usage

```bash
tracer
```

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

The detail view shows a context usage progress bar (tokens used vs. model context window) and a scrollable conversation preview.

## License

MIT
