# Themes

Tracer includes 12 built-in color themes and supports custom user themes.

## Built-In Themes

| Theme      | Style                        |
|------------|------------------------------|
| default    | Purple accent on dark        |
| minimal    | Subtle purple, low contrast  |
| mono       | Monochrome white/gray        |
| ocean      | Cyan/blue                    |
| rose       | Pink/mauve                   |
| forest     | Green                        |
| sunset     | Orange/gold                  |
| nord       | Cool blue/gray               |
| dracula    | Vivid purple/pink            |
| solarized  | Classic blue/teal            |
| monokai    | Pink/yellow on dark          |
| catppuccin | Purple/pink pastel           |

## Switching Themes

- **Interactive picker:** `tracer theme` (live preview)
- **CLI:** `tracer theme <name>`
- **Command palette:** `:theme <name>`
- **Settings:** `:set theme <name>`

The active theme is saved in `~/.config/tracer/config.json`.

## Custom Themes

Custom themes are JSON files in `~/.config/tracer/themes/`. Each file defines 10 hex colors:

```json
{
  "primary": "#7D56F4",
  "accent": "#7D56F4",
  "text": "#FAFAFA",
  "bright": "#FFFFFF",
  "muted": "#626262",
  "dim": "#444444",
  "red": "#FF4444",
  "green": "#44FF44",
  "select_bg": "#7D56F4",
  "select_fg": "#FFFFFF"
}
```

### Color Reference

| Field     | Where it appears |
|-----------|-----------------|
| primary   | Tab highlights, key hints, labels, table headers, borders |
| accent    | Assistant name label in conversation view |
| text      | Session names, setting values, conversation content |
| bright    | Active tab text, selected setting value |
| muted     | Inactive tabs, help descriptions, session count |
| dim       | Separators, section dividers |
| red       | Error messages, delete prompts, "unsaved" indicator |
| green     | "You:" label in conversation |
| select_bg | Selected table row background |
| select_fg | Selected table row text |

### Creating a Custom Theme

**Manually:**
```
:theme new <name>
```
Scaffolds a theme JSON with default colors and opens it in your editor.

**With AI assistance:**
```
:theme new-ai <name> [claude|codex|gemini]
```
Launches an AI agent that asks what kind of theme you want, then creates the colors. Optionally specify which AI provider to use.

### Editing a Custom Theme

```
:theme edit <name>
```

Opens the theme JSON in your editor. Changes take effect when you return to tracer.

### Using a Custom Theme

Custom themes appear alongside built-in themes everywhere — the theme picker, settings, autocomplete, and `:theme <name>`.
