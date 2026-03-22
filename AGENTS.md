# tracer — Claude Code Session Manager

A Go CLI TUI for browsing, inspecting, resuming, and deleting Claude Code sessions.

## Architecture

```
tracer/
├── main.go                         # Entry point — scans sessions, launches Bubbletea
├── internal/
│   ├── claude/                     # Data layer — reads/writes ~/.claude/
│   │   ├── parser.go              # JSONL line parser (Entry, RawMsg, Usage types)
│   │   ├── sessions.go            # ScanSessions, parseSessionFile, LoadConversation, loadRenames
│   │   └── delete.go              # DeleteSession — removes all session artifacts
│   ├── model/
│   │   └── session.go             # Session and Message structs, context window math
│   └── ui/
│       ├── app.go                 # Top-level Bubbletea model — view routing, key dispatch
│       ├── list.go                # List view — table, filtering, session selection
│       ├── detail.go              # Detail view — metadata, context progress bar, conversation
│       └── styles.go              # Lipgloss color and style definitions
```

## Tech Stack

- **Go** with modules (`go mod`)
- **Bubbletea v2** (`charm.land/bubbletea/v2`) — TUI framework, Model-View-Update pattern
- **Bubbles v2** (`charm.land/bubbles/v2`) — table, viewport, textinput, progress components
- **Lipgloss v2** (`charm.land/lipgloss/v2`) — terminal styling

## Key Concepts

### Data Sources

All data comes from `~/.claude/`:

- **`projects/{path}/{sessionId}.jsonl`** — one file per session, contains full conversation as JSONL. Each line is an `Entry` with type `user`, `assistant`, `system`, or `file-history-snapshot`. The path component encodes the working directory (e.g., `-Users-or-projects-myapp`).
- **`history.jsonl`** — global log of user prompts. Used to detect `/rename` commands that override the default session name (first user message).
- **`sessions/{pid}.json`** — maps PIDs to session IDs (not currently used by tracer).
- **`file-history/{sessionId}/`**, **`tasks/{sessionId}/`** — cleaned up on session deletion.

### Session Name Resolution

Default name = first user message (truncated to 80 chars). If the user ran `/rename <name>` in Claude Code, that name takes precedence. The rename is detected by scanning `history.jsonl` for entries where `display` starts with `/rename `.

### Bubbletea v2 Specifics

- `View()` returns `tea.View` (not `string`) — use `tea.NewView(content)`.
- Alt screen is set via `v.AltScreen = true` on the View struct, not as a program option.
- Key events are `tea.KeyPressMsg` (not `tea.KeyMsg`).
- Bubbles v2 components use option constructors: `viewport.New(viewport.WithWidth(w), viewport.WithHeight(h))`, `progress.New(progress.WithWidth(40))`.
- `tea.ExecProcess(cmd, callback)` to hand off the terminal to another process (used for `claude --resume`).

### UI Views

**List view** (`list.go`): Table with columns Name, Date, Directory, Branch. Supports `/` for filtering (substring match on name+dir+branch). The `listView` struct does not handle key events — `app.go` dispatches them.

**Detail view** (`detail.go`): Shows session metadata, a context usage progress bar (tokens used / max), and a scrollable conversation viewport. Same key dispatch pattern via `app.go`.

### Key Bindings

| Key | List | Detail |
|-----|------|--------|
| `Enter`/`v` | Open detail | Resume session |
| `c` | Copy session ID | Copy session ID |
| `d` | Delete (with confirm) | Delete (with confirm) |
| `/` | Filter mode | — |
| `Esc` | Clear filter | Back to list |
| `q` | Quit | Back to list |
| `Ctrl+C` | Quit | Quit |

## Build & Run

```bash
go build -o tracer .
./tracer
```

## Testing

```bash
go test ./... -v
```

Tests are in `internal/claude/` covering JSONL parsing, session scanning, and deletion. Tests use `t.TempDir()` to create isolated fixtures.

## Common Tasks

### Adding a new data field to sessions
1. Add field to `Session` struct in `model/session.go`
2. Populate it in `parseSessionFile()` in `claude/sessions.go`
3. Display it in `detail.go` (headerView) and/or `list.go` (table columns)

### Adding a new key binding
1. Add the case in `updateList()` or `updateDetail()` in `app.go`
2. Update the help text in `list.go:view()` or `detail.go:view()`

### Changing styles
All colors and styles are in `ui/styles.go`. Views reference these package-level vars.

### Releasing a new version

Releases are automatic. Pushing to `master` triggers `.github/workflows/release.yml`, which:
1. Analyzes commit messages since the last tag
2. Determines the version bump from conventional commit prefixes
3. Builds binaries for macOS (Intel + Apple Silicon) and Linux (amd64 + arm64)
4. Creates a git tag and GitHub release

**Version bump rules (from commit messages):**
- `fix:` or `fix(scope):` → **patch** (v0.1.0 → v0.1.1)
- `feat:` or `feat(scope):` → **minor** (v0.1.1 → v0.2.0)
- `feat!:`, `fix!:`, or `BREAKING CHANGE` in body → **major** (v0.2.0 → v1.0.0)
- No conventional prefix → defaults to **patch**

If there are no new commits since the last tag, the workflow skips the release.

To check the latest tag: `git describe --tags --abbrev=0`
