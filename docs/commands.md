# Command Palette

Press `:` to open the command palette from any view. Type a command and press `enter` to execute.

## Features

- **Autocomplete dropdown** with descriptions (toggle: `:set cmd_dropdown on|off`)
- **Ghost text** inline suggestions (toggle: `:set cmd_ghost on|off`)
- **Persistent history** of last 100 commands (`up`/`down` to cycle)
- **Context-aware** - commands only appear when available in the current view

### Key Bindings

| Key         | Action             |
|-------------|--------------------|
| `enter`     | Execute command    |
| `tab`       | Accept suggestion  |
| `up`/`down` | Navigate dropdown or history |
| `esc`       | Cancel             |
| `backspace` (empty) | Cancel     |

## Built-In Commands

### Session Commands
| Command              | Description                   |
|----------------------|-------------------------------|
| `:resume`            | Resume selected session (agent-aware) |
| `:fork`              | Fork selected session (agent-aware)   |
| `:copy`              | Copy session ID to clipboard  |
| `:pin`               | Toggle pin                    |
| `:rename <name>`     | Rename session                |
| `:new [path]`        | New session (pick agent if multiple enabled) |
| `:model <name>`      | Set model for Claude sessions |
| `:export <html\|md>` | Export conversation           |

### Navigation
| Command                          | Description         |
|----------------------------------|---------------------|
| `:view`                          | View selected item  |
| `:edit`                          | Edit in $EDITOR     |
| `:delete`                        | Delete selected     |
| `:filter [query]`               | Filter current list |
| `:sort <date\|name\|directory>` | Change sort order   |
| `:tab <sessions\|skills\|permissions>` | Switch tab  |
| `:settings`                      | Open settings       |
| `:quit`                          | Exit tracer         |

### Configuration
| Command                | Description        |
|------------------------|--------------------|
| `:set <key> <value>`  | Change a setting   |
| `:theme <name>`       | Switch theme       |
| `:help [command]`     | Show help          |

## User-Defined Commands

Create custom commands that appear in the command palette.

### Location

`~/.config/tracer/commands/<name>/`

Each command directory contains:
- `command.json` - metadata
- `run.sh` (or specified script) - executable

### command.json

```json
{
  "description": "What this command does",
  "shell": "run.sh",
  "mode": "status",
  "args": [
    {
      "name": "target",
      "required": false,
      "completions": [
        {"value": "option1", "description": "First option"}
      ]
    }
  ],
  "autostart": false
}
```

### Modes

| Mode     | Behavior                                        |
|----------|------------------------------------------------|
| `status` | First line of output shown in status bar (3s)  |
| `exec`   | Full terminal handoff for interactive programs |

### Aliases

Instead of `shell`, use `alias` to point to another command:

```json
{
  "description": "Shortcut for sort by name",
  "alias": "sort name"
}
```

### Environment

Scripts receive:
- `$1`, `$2`, ... - positional arguments
- `TRACER_CMD_DIR` - path to the command's directory

### Autostart

Set `"autostart": true` to run the command automatically when tracer launches (status mode only).

### Management

| Command                     | Description              |
|-----------------------------|--------------------------|
| `:commands list`            | List user commands       |
| `:commands new <name>`      | Create command           |
| `:commands new-ai <name>`   | Create with AI assistance |
| `:commands edit <name>`     | Edit command             |
| `:commands delete <name>`   | Delete command           |
