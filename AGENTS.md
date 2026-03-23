# tracer — Claude Code Session, Skill & Settings Manager

A Go CLI TUI for managing Claude Code sessions, skills, and permission settings.

## Architecture

```
tracer/
├── main.go                         # Entry point, subcommands (update, theme, settings, man)
├── tracer.1                        # Man page (embedded into binary via go:embed)
├── install.sh                      # Quick install script
├── internal/
│   ├── claude/                     # Session data layer — reads/writes ~/.claude/
│   │   ├── parser.go              # JSONL line parser (Entry, RawMsg, Usage, IsRealUserMessage)
│   │   ├── sessions.go            # ScanSessions (parallel), scanSessionHead, LoadSessionDetails, LoadConversation, loadRenames
│   │   ├── delete.go              # DeleteSession — removes all session artifacts
│   │   └── export.go              # ExportMarkdown — exports session conversation as Markdown
│   ├── skills/                    # Skill data layer — scans and manages skills
│   │   ├── model.go               # Skill struct, Source type (user, command, project, plugin)
│   │   ├── scanner.go             # ScanSkills — scans 4 locations, parses YAML frontmatter
│   │   ├── scanner_test.go        # Tests for scanning and parsing
│   │   └── crud.go                # CreateSkill, DeleteSkill
│   ├── ccsettings/                # Claude Code settings.json management
│   │   ├── model.go               # SettingsFile, Permissions, PermRule structs
│   │   └── scanner.go             # ScanSettings, SavePermissions, AddRule, RemoveRule
│   ├── config/
│   │   ├── config.go              # Config struct (theme, sort, columns, confirm delete, auto update)
│   │   ├── pins.go                # Pinned session IDs
│   │   └── renames.go             # Custom session renames from tracer
│   ├── model/
│   │   └── session.go             # Session and Message structs, context window math
│   ├── ui/
│   │   ├── app.go                 # Top-level Bubbletea model — tab routing, view routing, key dispatch
│   │   ├── tabs.go                # Tab bar (Sessions / Skills / Permissions)
│   │   ├── list.go                # Sessions list — table, filtering, sorting, column visibility
│   │   ├── detail.go              # Session detail — metadata, context progress bar, conversation
│   │   ├── skillslist.go          # Skills list — table, filtering
│   │   ├── skilldetail.go         # Skill detail — metadata, full content viewport
│   │   ├── permslist.go            # Permissions list — settings files table
│   │   ├── permsdetail.go         # Permissions detail — rules table with add/toggle/delete
│   │   ├── permsadd.go            # Add rule flow — multi-step inline prompt
│   │   ├── settings.go            # Settings view + standalone SettingsApp
│   │   ├── theme.go               # 11 theme definitions and ApplyTheme
│   │   ├── themepicker.go         # Interactive theme picker (tracer theme command)
│   │   └── styles.go              # Default lipgloss styles (overwritten by ApplyTheme)
│   └── updater/
│       └── updater.go             # Self-updater with semver comparison, Homebrew detection
├── .github/workflows/
│   └── release.yml                # Auto-release on push to master (conventional commits)
```

## Tech Stack

- **Go** with modules (`go mod`)
- **Bubbletea v2** (`charm.land/bubbletea/v2`) — TUI framework, Model-View-Update pattern
- **Bubbles v2** (`charm.land/bubbles/v2`) — table, viewport, textinput, progress components
- **Lipgloss v2** (`charm.land/lipgloss/v2`) — terminal styling

## Key Concepts

### Tab System

The app has three tabs: **Sessions**, **Skills**, and **Permissions**, cycled with `Tab`/`Shift+Tab`. Each tab has its own list and detail views. The tab bar renders at the top of list views.

### Permissions Management

The Permissions tab scans for `settings.json` files across three scopes:
- **Global** — `~/.claude/settings.json`
- **Project** — `{project}/.claude/settings.json` for each project found in `~/.claude/projects/`
- **Local** — `{project}/.claude/settings.local.json`

The list view shows all discovered files with scope, rule count, and path. The detail view shows a table of allow/deny rules for the selected file. Users can:
- **Add rules** (`a`) — multi-step flow: pick allow/deny, type the rule pattern
- **Toggle rules** (`t`) — switch a rule between allow and deny
- **Delete rules** (`d`) — remove a rule

Changes are saved immediately to the settings file, preserving all other JSON fields.

### Data Sources

### Data Sources

#### Sessions

All session data comes from `~/.claude/`:

- **`projects/{path}/{sessionId}.jsonl`** — one file per session, contains full conversation as JSONL. Each line is an `Entry` with type `user`, `assistant`, `system`, or `file-history-snapshot`. The path component encodes the working directory (e.g., `-Users-or-projects-myapp`).
- **`history.jsonl`** — global log of user prompts. Used to detect `/rename` commands that override the default session name.
- **`file-history/{sessionId}/`**, **`tasks/{sessionId}/`** — cleaned up on session deletion.

#### Skills

Skills are scanned from 4 locations:

- **`~/.claude/skills/*/SKILL.md`** — user-created skills (source: `user`). Each skill lives in a subdirectory with a `SKILL.md` file containing YAML frontmatter (`name`, `description`) and markdown content.
- **`~/.claude/commands/*.md`** — user-created commands (source: `command`). Flat markdown files.
- **Project `.claude/commands/*.md`** — project-specific commands (source: `project`). Discovered by decoding project paths from `~/.claude/projects/`.
- **`~/.claude/plugins/cache/claude-plugins-official/*/[version]/skills/*/SKILL.md`** — plugin-bundled skills (source: `plugin`, read-only). Scans the latest non-orphaned version directory per plugin.

YAML frontmatter is parsed with simple string matching (no YAML library). If no frontmatter exists, the filename is used as the skill name.

### Session Scanning (Performance)

Startup uses a two-phase approach for speed:

1. **Fast scan** (`scanSessionHead`) — reads only until the first real user message. Skips meta messages (`isMeta: true`), XML-tagged system messages (`<local-command-caveat>`, `<command-message>`), and slash commands. Runs in parallel across 16 goroutines.
2. **Lazy detail loading** (`LoadSessionDetails`) — reads the full JSONL file only when the user opens the detail view. Populates token counts, message stats, and model ID.

### Session Name Resolution

Default name = first real user message (truncated to 80 chars). Messages with `isMeta: true`, XML tags, or `/` prefix are skipped. Priority order:
1. Tracer rename (`~/.config/tracer/renames.json`) — highest
2. Claude `/rename` command (from `history.jsonl`)
3. First real user message — lowest

### Context Token Calculation

Total context = `input_tokens + cache_creation_input_tokens + cache_read_input_tokens` from the last assistant message's usage block. Stored as `ContextTokens` on the Session struct.

### Configuration

All settings in `~/.config/tracer/config.json`:

| Setting | Field | Options | Default |
|---------|-------|---------|---------|
| Theme | `theme` | 11 themes | default |
| Sort by | `sort_by` | date, name, directory | date |
| Show date | `show_date` | true/false | true |
| Show directory | `show_directory` | true/false | true |
| Show branch | `show_branch` | true/false | true |
| Confirm delete | `confirm_delete` | true/false | true |
| Auto update | `auto_update` | true/false | false |

Other config files:
- `~/.config/tracer/pins.json` — pinned session IDs
- `~/.config/tracer/renames.json` — custom session names

### Themes

11 themes defined in `ui/theme.go`. Each theme is a `Theme` struct with `image/color.Color` fields (Primary, Accent, Text, Bright, Muted, Dim, Red, Green, SelectBg, SelectFg). `ApplyTheme()` overwrites package-level style vars. Table styles in `list.go` and `skillslist.go` read from `CurrentTheme()` at rebuild time.

### Clipboard

Uses `pbcopy` on macOS. On Linux, tries `xclip -selection clipboard`, falls back to `xsel --clipboard --input`.

### Self-Updater

Uses proper semver comparison (prevents downgrades). Detects Homebrew installs via binary path and redirects to `brew upgrade`. Auto-update is disabled for Homebrew installs and `dev` builds. The update check runs in a background goroutine during TUI use and applies after the user exits. Downloads are capped at 200MB via `io.LimitReader`.

### Bubbletea v2 Specifics

- `View()` returns `tea.View` (not `string`) — use `tea.NewView(content)`.
- Alt screen is set via `v.AltScreen = true` on the View struct, not as a program option.
- Key events are `tea.KeyPressMsg` (not `tea.KeyMsg`).
- Bubbles v2 components use option constructors: `viewport.New(viewport.WithWidth(w), viewport.WithHeight(h))`, `progress.New(progress.WithWidth(40))`.
- `tea.ExecProcess(cmd, callback)` to hand off the terminal to another process.

### UI Views

**Sessions list** (`list.go`): Table with configurable columns (Name always shown; Date, Directory, Branch toggleable). Column widths distributed dynamically. Supports filtering and configurable sort order.

**Session detail** (`detail.go`): Session metadata, context usage progress bar, scrollable conversation viewport. Supports rename (`r`), edit (`e`), resume (`Enter`).

**Skills list** (`skillslist.go`): Table with Name, Source, Description columns. Supports filtering. Edit (`e`) and create (`n`) open `$EDITOR`.

**Skill detail** (`skilldetail.go`): Skill metadata (name, source, plugin, path, size) and full file content in scrollable viewport.

**Settings** (`settings.go`): Inline settings editor + standalone `SettingsApp` for `tracer settings` subcommand. Up/down to navigate, left/right to cycle values.

**Theme picker** (`themepicker.go`): Standalone TUI for `tracer theme`. Live preview with sample table and conversation.

**Permissions list** (`permslist.go`): Table of settings files with scope, rule count, path. Enter opens the detail view.

**Permissions detail** (`permsdetail.go`): Table of allow/deny rules for a single settings file. Add (`a`), toggle (`t`), delete (`d`) rules. Saves to disk immediately.

**Add rule** (`permsadd.go`): Multi-step inline prompt. Step 1: pick allow/deny. Step 2: type the rule pattern. Esc cancels at any step.

**Tab bar** (`tabs.go`): Renders Sessions/Skills/Permissions tabs. Active tab highlighted with theme primary color.

### Key Bindings

#### Sessions Tab

| Key | List | Detail |
|-----|------|--------|
| `Enter` | Resume session | Resume session |
| `n` | New session | — |
| `v` | Open detail | — |
| `r` | — | Rename |
| `e` | — | Edit in $EDITOR |
| `c` | Copy session ID | Copy session ID |
| `x` | — | Export as Markdown |
| `p` | Pin/unpin | — |
| `d` | Delete | Delete |
| `s` | Settings | — |
| `/` | Filter | — |
| `Tab` | Switch to Skills | — |
| `Esc` | Clear filter | Back to list |

#### Skills Tab

| Key | List | Detail |
|-----|------|--------|
| `Enter`/`v` | View detail | — |
| `e` | Edit | Edit |
| `n` | Create new | — |
| `d` | Delete | Delete |
| `/` | Filter | — |
| `Tab` | Switch to Sessions | — |
| `Esc` | Clear filter | Back to list |

#### Permissions Tab

| Key | List | Detail |
|-----|------|--------|
| `Enter`/`v` | View rules | — |
| `a` | — | Add rule |
| `t` | — | Toggle allow/deny |
| `d` | — | Delete rule |
| `/` | Filter | — |
| `Tab` | Next tab | — |
| `Esc` | Clear filter | Back to list |

#### Settings / General

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate |
| `←/→` / `Enter` | Change value |
| `Esc` / `q` | Save & back |
| `Ctrl+C` | Quit |

## Build & Run

```bash
go build -o tracer .
./tracer
```

## Testing

```bash
go test ./... -v
```

Tests cover: JSONL parsing, session scanning, session deletion, skill scanning, and frontmatter parsing. Tests use `t.TempDir()` for isolated fixtures.

## Common Tasks

### Adding a new data field to sessions
1. Add field to `Session` struct in `model/session.go`
2. Populate it in `scanSessionHead()` (list view) or `LoadSessionDetails()` (detail view) in `claude/sessions.go`
3. Display it in `detail.go` and/or `list.go`

### Adding a new key binding
1. Add the case in the relevant `update*()` method in `app.go`
2. Update the help text in the relevant view's `view()` method

### Adding a new tab
1. Add a `Tab` const in `tabs.go` and update `tabNames`
2. Add `viewState` consts for the new tab's list and detail views
3. Create list and detail view files (follow `skillslist.go` / `skilldetail.go` pattern)
4. Wire into `app.go`: add fields to `App`, handle tab switching, add update/view methods

### Adding a new subcommand
Add a case to the `switch os.Args[1]` block in `main.go`.

### Adding a new setting
1. Add field to `Config` in `config/config.go` with default in `DefaultConfig()`
2. Add `settingType` const in `settings.go`
3. Add `cycleRight`/`cycleLeft` handling
4. Add to `items` slice in `settingsView.view()`
5. Use the setting where needed

### Adding a new theme
Add a new entry to `Themes` map and `ThemeNames()` slice in `ui/theme.go`. Include all color fields including `SelectFg` and `SelectBg`.

### Adding a new skill source
1. Add a `Source` const in `skills/model.go`
2. Add scanning logic in `skills/scanner.go` within `ScanSkills()`
3. Set `ReadOnly` appropriately

### Changing styles
Default styles in `styles.go` are overwritten by `ApplyTheme()`. To change base styles, edit `styles.go`. To change theme-specific styles, edit `theme.go`.

### Updating the man page
Edit `tracer.1` (troff format). Embedded at build time — no extra steps needed.

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
