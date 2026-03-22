# tracer

See [AGENTS.md](./AGENTS.md) for full architecture, data sources, and conventions.

## Quick Reference

- **Build:** `go build -o tracer .`
- **Test:** `go test ./... -v`
- **Deps:** Bubbletea v2, Bubbles v2, Lipgloss v2 (all from `charm.land`)
- Data layer: `internal/claude/` — reads `~/.claude/`
- Types: `internal/model/` — `Session`, `Message`
- UI: `internal/ui/` — Bubbletea app with list and detail views
