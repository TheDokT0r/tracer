# tracer — Claude Code Session Manager

A Go CLI TUI for browsing, inspecting, resuming, and deleting Claude Code sessions.

## Architecture

```
tracer/
├── main.go                         # Entry point, subcommands (update, man, --version, --help)
├── tracer.1                        # Man page (embedded into binary via go:embed)
├── install.sh                      # Quick install script
├── internal/
│   ├── claude/                     # Data layer — reads/writes ~/.claude/
│   │   ├── parser.go              # JSONL line parser (Entry, RawMsg, Usage types)
│   │   ├── sessions.go            # ScanSessions (parallel), scanSessionHead, LoadSessionDetails, LoadConversation, loadRenames
│   │   └── delete.go              # DeleteSession — removes all session artifacts
│   ├── model/
│   │   └── session.go             # Session and Message structs, context window math
│   ├── ui/
│   │   ├── app.go                 # Top-level Bubbletea model — view routing, key dispatch
│   │   ├── list.go                # List view — table, filtering, session selection
│   │   ├── detail.go              # Detail view — metadata, context progress bar, conversation
│   │   └── styles.go              # Lipgloss color and style definitions
│   └── updater/
│       └── updater.go             # Self-updater — checks GitHub releases, downloads, replaces binary
├── .github/workflows/
│   └── release.yml                # Auto-release on push to master (conventional commits)
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

### Session Scanning (Performance)

Startup uses a two-phase approach for speed:

1. **Fast scan** (`scanSessionHead`) — reads only until the first user message. Extracts name, directory, branch, timestamp. Runs in parallel across 16 goroutines.
2. **Lazy detail loading** (`LoadSessionDetails`) — reads the full JSONL file only when the user opens the detail view. Populates token counts, message stats, and model ID.

### Session Name Resolution

Default name = first user message (truncated to 80 chars). If the user ran `/rename <name>` in Claude Code, that name takes precedence. The rename is detected by scanning `history.jsonl` for entries where `display` starts with `/rename `.

### Context Token Calculation

Total context = `input_tokens + cache_creation_input_tokens + cache_read_input_tokens` from the last assistant message's usage block. The `input_tokens` field alone is just the uncached portion.

### Clipboard

Uses `pbcopy` on macOS. On Linux, tries `xclip -selection clipboard`, falls back to `xsel --clipboard --input`.

### Bubbletea v2 Specifics

- `View()` returns `tea.View` (not `string`) — use `tea.NewView(content)`.
- Alt screen is set via `v.AltScreen = true` on the View struct, not as a program option.
- Key events are `tea.KeyPressMsg` (not `tea.KeyMsg`).
- Bubbles v2 components use option constructors: `viewport.New(viewport.WithWidth(w), viewport.WithHeight(h))`, `progress.New(progress.WithWidth(40))`.
- `tea.ExecProcess(cmd, callback)` to hand off the terminal to another process (used for `claude --resume`).

### UI Views

**List view** (`list.go`): Table with columns Name, Date, Directory, Branch. Column widths: 40% name, 30% directory, 30% branch (of remaining space after date). Supports `/` for filtering (substring match on name+dir+branch). The `listView` struct does not handle key events — `app.go` dispatches them.

**Detail view** (`detail.go`): Shows session metadata, a context usage progress bar (tokens used / max), and a scrollable conversation viewport. Token counts and message stats are loaded on demand via `LoadSessionDetails`. Same key dispatch pattern via `app.go`.

### Key Bindings

| Key | List | Detail |
|-----|------|--------|
| `Enter` | Resume session | Resume session |
| `v` | Open detail | — |
| `c` | Copy session ID | Copy session ID |
| `d` | Delete (with confirm) | Delete (with confirm) |
| `/` | Filter mode | — |
| `Esc` | Clear filter | Back to list |
| `q` | Quit | Back to list |
| `Ctrl+C` | Quit | Quit |

### Self-Updater

`tracer update` checks the GitHub releases API for the latest version, downloads the correct `.tar.gz` for the current OS/arch, and replaces the binary. If the binary path is not writable, it falls back to `sudo mv`. Skips update if running a `dev` build.

### Man Page

The man page (`tracer.1`) is embedded into the binary via `go:embed`. `tracer man` writes it to a temp file and opens it with `man`. The install script also places it in `~/.local/share/man/man1/`.

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
2. Populate it in `scanSessionHead()` (if needed for list view) or `LoadSessionDetails()` (if only for detail view) in `claude/sessions.go`
3. Display it in `detail.go` (headerView) and/or `list.go` (table columns)

### Adding a new key binding
1. Add the case in `updateList()` or `updateDetail()` in `app.go`
2. Update the help text in `list.go:view()` or `detail.go:view()`

### Adding a new subcommand
Add a case to the `switch os.Args[1]` block in `main.go`.

### Changing styles
All colors and styles are in `ui/styles.go`. Views reference these package-level vars.

### Updating the man page
Edit `tracer.1` (troff format). It gets embedded at build time — no extra steps needed.

### Releasing a new version

Releases are automatic. Pushing to `master` triggers `.github/workflows/release.yml`, which:
1. Analyzes commit messages since the last tag
2. Determines the version bump from conventional commit prefixes
3. Builds binaries for macOS (Intel + Apple Silicon) and Linux (amd64 + arm64)
4. Bundles man page in each `.tar.gz` archive
5. Creates a git tag and GitHub release

**Version bump rules (from commit messages):**
- `fix:` or `fix(scope):` → **patch** (v0.1.0 → v0.1.1)
- `feat:` or `feat(scope):` → **minor** (v0.1.1 → v0.2.0)
- `feat!:`, `fix!:`, or `BREAKING CHANGE` in body → **major** (v0.2.0 → v1.0.0)
- Other prefixes (`docs:`, `ci:`, `chore:`, `refactor:`, `test:`) → no release

To check the latest tag: `git describe --tags --abbrev=0`
