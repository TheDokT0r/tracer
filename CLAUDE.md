# tracer

See [AGENTS.md](./AGENTS.md) for full architecture, data sources, and conventions.

## Quick Reference

- **Build:** `go build -o tracer .`
- **Test:** `go test ./... -v`
- **Deps:** Bubbletea v2, Bubbles v2, Lipgloss v2 (all from `charm.land`)
- Session data: `internal/claude/` — reads `~/.claude/`
- Skill data: `internal/skills/` — scans skills, commands, plugins
- Permissions: `internal/ccsettings/` — manages settings.json allow/deny rules
- Config: `internal/config/` — tracer settings, pins, renames (`~/.config/tracer/`)
- Types: `internal/model/` — `Session`, `Message`
- UI: `internal/ui/` — 3 tabs (Sessions/Skills/Permissions), list, detail, settings, themes
- Updater: `internal/updater/` — self-update via GitHub releases
- Man page: `tracer.1` — embedded via `go:embed`
- Release: automatic on push to master via conventional commits
