# Export

Export conversations from the session detail view.

## Formats

| Format   | Command          | Output                          |
|----------|------------------|---------------------------------|
| Markdown | `:export md`     | Plain `.md` with message blocks |
| HTML     | `:export html`   | Styled chat interface           |

## How to Export

1. Open a session detail view (`v` from the list)
2. Press `x` to open the format picker, then `m` or `h`
3. Or use `:export md` / `:export html` from the command palette

The export path is automatically copied to your clipboard.

## Output Location

Exports are saved to `~/.config/tracer/exports/`:
- `<sessionId>.md` for Markdown
- `<sessionId>.html` for HTML
