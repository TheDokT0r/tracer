# Custom Columns

Add your own columns to the sessions table, populated by shell scripts.

## Location

`~/.config/tracer/columns/<name>/`

Each column directory contains:
- `column.json` - metadata
- `run.sh` (or specified script) - executable that outputs one line

## column.json

```json
{
  "description": "What this column shows",
  "header": "Column Header",
  "shell": "run.sh",
  "width": 10,
  "timeout": 5
}
```

| Field       | Description                          | Default |
|-------------|--------------------------------------|---------|
| description | Shown in help                        | -       |
| header      | Column header text                   | -       |
| shell       | Script filename to run               | -       |
| width       | Column width in characters           | -       |
| timeout     | Max seconds before showing "---"     | 5       |

## Script Interface

The script runs once per session, in parallel across all sessions.

**Input:**
- `$1` - session's working directory

**Environment:**
- `SESSION_DIR` - session's working directory
- `SESSION_ID` - session UUID
- `TRACER_COL_DIR` - path to the column's directory

**Output:** Print one line. That line appears in the cell.

**Loading state:** Cells show `...` until the script completes.

## Example

A column showing the git remote URL:

```bash
#!/bin/bash
cd "$1" 2>/dev/null && git remote get-url origin 2>/dev/null || echo "-"
```

## Management

| Command                    | Description              |
|----------------------------|--------------------------|
| `:columns list`            | List custom columns      |
| `:columns new <name>`      | Create column            |
| `:columns new-ai <name>`   | Create with AI assistance |
| `:columns edit <name>`     | Edit metadata            |
| `:columns delete <name>`   | Delete column            |
| `:columns toggle <name>`   | Show/hide column         |

Hidden columns are stored in `~/.config/tracer/config.json` under `hidden_columns`.
