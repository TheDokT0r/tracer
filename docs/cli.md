# CLI

Run `tracer` with no arguments to launch the TUI. Subcommands are available for specific tasks.

## Usage

```
tracer [command]
```

## Commands

| Command            | Description                    |
|--------------------|--------------------------------|
| (none)             | Launch the TUI                 |
| `update`           | Update to the latest version   |
| `theme`            | Interactive theme picker       |
| `theme <name>`     | Set theme directly             |
| `settings`         | Open settings editor           |
| `man`              | View the manual page           |

## Flags

| Flag             | Description   |
|------------------|---------------|
| `-v`, `--version` | Print version |
| `-h`, `--help`    | Show help     |

## Auto-Update

When `auto_update` is enabled in settings, tracer checks for updates in the background while the TUI is running. If a new version is available, it installs after the TUI exits and restarts automatically.

Manual update: `tracer update`
