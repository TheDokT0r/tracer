# Sessions

The Sessions tab is the main view. It lists sessions from all enabled agents (Claude, Codex, Gemini) in a single unified list, sorted by date.

## Multi-Agent Support

Tracer scans sessions from three AI coding tools:

| Agent  | Source Directory | Resume | Fork |
|--------|-----------------|--------|------|
| Claude | `~/.claude/`    | `claude --resume` | `claude --resume --fork-session` |
| Codex  | `~/.codex/`     | `codex resume` | `codex fork` |
| Gemini | `~/.gemini/`    | not supported | not supported |

Enable/disable agents in Settings > Agents, or with `:set agent_claude on|off`, `:set agent_codex on|off`, `:set agent_gemini on|off`.

Non-Claude sessions show a `[codex]` or `[gemini]` prefix on their name, or in a separate Agent column if enabled.

## List View

Displays sessions in a table with configurable columns:

| Column    | Default | Toggle                        |
|-----------|---------|-------------------------------|
| Name      | always  | -                             |
| Agent     | off     | `:set show_agent on\|off`     |
| Date      | on      | `:set show_date on\|off`      |
| Directory | on      | `:set show_directory on\|off` |
| Branch    | on      | `:set show_branch on\|off`    |
| Model     | off     | `:set show_model on\|off`     |

Pinned sessions appear at the top (marked with `*`). Directories show `~` instead of the full home path.

### Key Bindings

| Key     | Action                              |
|---------|-------------------------------------|
| `enter` | Resume session (agent-aware)        |
| `v`     | View session detail                 |
| `n`     | New session (pick agent + directory) |
| `f`     | Fork session (agent-aware)          |
| `c`     | Copy session ID to clipboard        |
| `p`     | Pin/unpin session                   |
| `d`     | Delete session                      |
| `/`     | Filter by name, directory, or branch |
| `s`     | Open settings                       |
| `tab`   | Switch to Skills tab                |

## Creating a New Session

Press `n` to start a new session:

1. Enter the working directory (defaults to current directory)
2. If multiple agents are enabled, pick which one with `←/→` then `enter`
3. If only one agent is enabled, it launches directly

## Detail View

Shows session metadata and a scrollable conversation preview.

**Metadata displayed:**
- Session ID, date, directory, git branch (if available)
- Message count (user + assistant)
- Context usage (tokens / max, with progress bar)
- Output token count
- Agent name (for non-Claude sessions)

**Conversation preview** shows the first 500 characters of each message.

### Key Bindings (Detail)

| Key     | Action              |
|---------|---------------------|
| `enter` | Resume session      |
| `f`     | Fork session        |
| `r`     | Rename session      |
| `e`     | Edit JSONL in $EDITOR |
| `x`     | Export (pick format) |
| `c`     | Copy session ID     |
| `d`     | Delete              |
| `esc`   | Back to list        |

## Sorting

Change sort order with `:sort <field>`:
- `date` (default, newest first)
- `name` (alphabetical)
- `directory`

## Pinning

Pin important sessions to keep them at the top of the list. Pins persist across restarts in `~/.config/tracer/pins.json`.

## Renaming

Override session names with `r` in detail view or `:rename <name>`. For Claude sessions, renames are also written to Claude's `history.jsonl` so they persist across both tracer and Claude Code.

## Forking

Fork creates a new session that continues from the same conversation. Supported for Claude and Codex sessions.

## Model Override

Set a default model for Claude sessions with `:model <name>` or `:set model <name>`. This passes `--model` to Claude on resume, fork, and new session.
