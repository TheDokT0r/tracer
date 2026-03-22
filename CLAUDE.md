# tracer

See [AGENTS.md](./AGENTS.md) for full architecture, data sources, and conventions.

## Quick Reference

- **Build:** `go build -o tracer .`
- **Test:** `go test ./... -v`
- **Deps:** Bubbletea v2, Bubbles v2, Lipgloss v2 (all from `charm.land`)
- Data layer: `internal/claude/` — reads `~/.claude/`
- Config: `internal/config/` — settings and pins (`~/.config/tracer/`)
- Types: `internal/model/` — `Session`, `Message`
- UI: `internal/ui/` — list, detail, settings views + theme system
- Updater: `internal/updater/` — self-update via GitHub releases
- Man page: `tracer.1` — embedded via `go:embed`
- Release: automatic on push to master via conventional commits
