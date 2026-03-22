# tracer — Claude Code Session Manager

A Go CLI TUI for browsing, inspecting, resuming, and deleting Claude Code sessions.

## Architecture

```
tracer/
├── main.go                         # Entry point, subcommands (update, theme, man, --version, --help)
├── tracer.1                        # Man page (embedded into binary via go:embed)
├── install.sh                      # Quick install script
├── internal/
│   ├── claude/                     # Data layer — reads/writes ~/.claude/
│   │   ├── parser.go              # JSONL line parser (Entry, RawMsg, Usage types, IsRealUserMessage)
│   │   ├── sessions.go            # ScanSessions (parallel), scanSessionHead, LoadSessionDetails, LoadConversation, loadRenames
│   │   └── delete.go              # DeleteSession — removes all session artifacts
│   ├── config/
│   │   ├── config.go              # Config struct (theme, sort, columns, confirm delete), load/save
│   │   └── pins.go                # Pinned session IDs, load/save/toggle
│   ├── model/
│   │   └── session.go             # Session and Message structs, context window math
│   ├── ui/
│   │   ├── app.go                 # Top-level Bubbletea model — view routing, key dispatch
│   │   ├── list.go                # List view — table, filtering, sorting, column visibility
│   │   ├── detail.go              # Detail view — metadata, context progress bar, conversation
│   │   ├── settings.go            # Settings view — navigate/cycle settings inline
│   │   ├── theme.go               # Theme definitions and ApplyTheme
│   │   ├── themepicker.go         # Interactive theme picker (tracer theme command)
│   │   └── styles.go              # Default lipgloss styles (overwritten by ApplyTheme)
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

1. **Fast scan** (`scanSessionHead`) — reads only until the first real user message. Skips meta messages (`isMeta: true`), XML-tagged system messages (`<local-command-caveat>`, `<command-message>`), and slash commands. Runs in parallel across 16 goroutines.
2. **Lazy detail loading** (`LoadSessionDetails`) — reads the full JSONL file only when the user opens the detail view. Populates token counts, message stats, and model ID.

### Session Name Resolution

Default name = first real user message (truncated to 80 chars). Messages with `isMeta: true`, XML tags, or `/` prefix are skipped. If the user ran `/rename <name>` in Claude Code, that name takes precedence (detected from `history.jsonl`).

### Context Token Calculation

Total context = `input_tokens + cache_creation_input_tokens + cache_read_input_tokens` from the last assistant message's usage block. The `input_tokens` field alone is just the uncached portion.

### Configuration

All settings are in `~/.config/tracer/config.json`:

| Setting | Field | Options | Default |
|---------|-------|---------|---------|
| Theme | `theme` | default, minimal, ocean, rose | default |
| Sort by | `sort_by` | date, name, directory | date |
| Show date | `show_date` | true/false | true |
| Show directory | `show_directory` | true/false | true |
| Show branch | `show_branch` | true/false | true |
| Confirm delete | `confirm_delete` | true/false | true |

Pinned sessions stored separately in `~/.config/tracer/pins.json`.

### Themes

Defined in `ui/theme.go`. Each theme is a `Theme` struct with `image/color.Color` fields (Primary, Accent, Text, Bright, Muted, Dim, Red, Green, SelectBg). `ApplyTheme()` overwrites the package-level style vars in `styles.go`. The table styles in `list.go` read from `CurrentTheme()` at rebuild time.

### Clipboard

Uses `pbcopy` on macOS. On Linux, tries `xclip -selection clipboard`, falls back to `xsel --clipboard --input`.

### Bubbletea v2 Specifics

- `View()` returns `tea.View` (not `string`) — use `tea.NewView(content)`.
- Alt screen is set via `v.AltScreen = true` on the View struct, not as a program option.
- Key events are `tea.KeyPressMsg` (not `tea.KeyMsg`).
- Bubbles v2 components use option constructors: `viewport.New(viewport.WithWidth(w), viewport.WithHeight(h))`, `progress.New(progress.WithWidth(40))`.
- `tea.ExecProcess(cmd, callback)` to hand off the terminal to another process (used for `claude --resume`).

### UI Views

**List view** (`list.go`): Table with configurable columns (Name always shown; Date, Directory, Branch toggleable). Column widths distributed dynamically. Supports `/` for filtering and configurable sort order. The `listView` does not handle key events — `app.go` dispatches them.

**Detail view** (`detail.go`): Shows session metadata, a context usage progress bar (tokens used / max), and a scrollable conversation viewport. Token counts loaded on demand via `LoadSessionDetails`.

**Settings view** (`settings.go`): Inline settings editor. Up/down to navigate, left/right to cycle values. Esc saves and returns to list. Changes to theme apply immediately via `ApplyTheme`. Changes to sort/columns apply on return to list.

**Theme picker** (`themepicker.go`): Standalone TUI for `tracer theme` command. Shows live preview with sample table, detail fields, and conversation. Left/right to switch themes, Enter to apply.

### Key Bindings

| Key | List | Detail | Settings |
|-----|------|--------|----------|
| `Enter` | Resume session | Resume session | Change value |
| `v` | Open detail | — | — |
| `c` | Copy session ID | Copy session ID | — |
| `p` | Pin/unpin | — | — |
| `d` | Delete | Delete | — |
| `s` | Open settings | — | — |
| `/` | Filter mode | — | — |
| `←/→` | — | — | Change value |
| `Esc` | Clear filter | Back to list | Save & back |
| `q` | Quit | Back to list | Save & back |
| `Ctrl+C` | Quit | Quit | Quit |

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
1. Add the case in `updateList()`, `updateDetail()`, or `updateSettings()` in `app.go`
2. Update the help text in the relevant view's `view()` method

### Adding a new subcommand
Add a case to the `switch os.Args[1]` block in `main.go`.

### Adding a new setting
1. Add field to `Config` struct in `config/config.go`, with default in `DefaultConfig()`
2. Add a `settingType` const in `settings.go`
3. Add `cycleRight`/`cycleLeft` handling for the new setting
4. Add it to the `items` slice in `settingsView.view()`
5. Use the setting where needed (e.g., `list.go`, `app.go`)

### Adding a new theme
Add a new entry to the `Themes` map and `ThemeNames()` slice in `ui/theme.go`.

### Changing styles
Default styles in `ui/styles.go` are overwritten by `ApplyTheme()`. To change the base styles, edit `styles.go`. To change theme-specific styles, edit `theme.go`.

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
