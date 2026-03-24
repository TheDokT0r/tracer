# Sessions

The Sessions tab is the main view. It lists all Claude Code sessions found in `~/.claude/`.

## List View

Displays sessions in a table with configurable columns:

| Column    | Default | Toggle                     |
|-----------|---------|----------------------------|
| Name      | always  | -                          |
| Date      | on      | `:set show_date on\|off`   |
| Directory | on      | `:set show_directory on\|off` |
| Branch    | on      | `:set show_branch on\|off` |

Pinned sessions appear at the top (marked with `*`). Directories show `~` instead of the full home path.

### Key Bindings

| Key     | Action                              |
|---------|-------------------------------------|
| `enter` | Resume session (launches Claude)    |
| `v`     | View session detail                 |
| `n`     | New session (prompts for directory)  |
| `f`     | Fork session                        |
| `c`     | Copy session ID to clipboard        |
| `p`     | Pin/unpin session                   |
| `d`     | Delete session                      |
| `/`     | Filter by name, directory, or branch |
| `s`     | Open settings                       |
| `tab`   | Switch to Skills tab                |

## Detail View

Shows session metadata and a scrollable conversation preview.

**Metadata displayed:**
- Session ID, date, directory, git branch
- Message count (user + assistant)
- Context usage (tokens / max, with progress bar)
- Output token count

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

Override session names with `r` in detail view or `:rename <name>`. Renames are stored in `~/.config/tracer/renames.json` and are independent of Claude's `/rename` command.

## Forking

Fork creates a new Claude session that continues from the same conversation. Useful for branching a conversation in a different direction.
