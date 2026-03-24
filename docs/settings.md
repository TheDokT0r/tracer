# Settings

Open settings with `s` from the list view, or `:settings` from the command palette.

## Available Settings

| Setting              | Values           | Default | Description                      |
|----------------------|------------------|---------|----------------------------------|
| theme                | see [themes](themes.md) | default | Color theme               |
| sort_by              | date, name, directory | date | Session sort order              |
| show_date            | on, off          | on      | Show Date column                 |
| show_directory       | on, off          | on      | Show Directory column            |
| show_branch          | on, off          | on      | Show Branch column               |
| confirm_delete       | on, off          | on      | Require y/N before deleting      |
| auto_update          | on, off          | off     | Auto-install updates on exit     |
| cmd_dropdown         | on, off          | on      | Show autocomplete dropdown       |
| cmd_ghost            | on, off          | off     | Show inline ghost text           |
| cmd_max_suggestions  | 3-12             | 8       | Max dropdown items               |

## Navigation (Settings View)

| Key             | Action        |
|-----------------|---------------|
| `up`/`down` or `k`/`j` | Navigate      |
| `left`/`right` or `h`/`l` or `enter` | Change value |
| `esc` or `q`   | Save and exit |

## Changing from Command Palette

```
:set <key> <value>
```

Examples:
```
:set theme dracula
:set sort_by name
:set show_branch off
:set cmd_max_suggestions 5
```

## Storage

`~/.config/tracer/config.json`
