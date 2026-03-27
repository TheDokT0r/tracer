package ui

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	"tracer/internal/config"
	"tracer/internal/model"
)

type listView struct {
	table      table.Model
	filter     textinput.Model
	filtering  bool
	sessions   []model.Session
	filtered   []model.Session
	pins       map[string]bool
	cfg        config.Config
	columns    []config.UserColumn
	columnData map[string]map[string]string // column name -> session ID -> value
	width      int
	height     int
}

func newListView(sessions []model.Session, pins map[string]bool, cfg config.Config, width, height int) listView {
	ti := textinput.New()
	ti.Prompt = "Filter: "
	ti.Placeholder = "type to filter..."

	lv := listView{
		filter:     ti,
		sessions:   sessions,
		filtered:   sessions,
		pins:       pins,
		cfg:        cfg,
		columns:    config.ScanUserColumns(),
		columnData: make(map[string]map[string]string),
		width:      width,
		height:     height,
	}
	lv.sortSessions()
	lv.rebuildTable()
	return lv
}

func (lv *listView) sortSessions() {
	sorter := func(s []model.Session) func(i, j int) bool {
		return func(i, j int) bool {
			pi := lv.pins[s[i].ID]
			pj := lv.pins[s[j].ID]
			if pi != pj {
				return pi
			}
			switch lv.cfg.SortBy {
			case "name":
				return s[i].Name < s[j].Name
			case "directory":
				return s[i].Directory < s[j].Directory
			default:
				return s[i].StartedAt.After(s[j].StartedAt)
			}
		}
	}
	sort.SliceStable(lv.sessions, sorter(lv.sessions))
	if &lv.filtered != &lv.sessions {
		sort.SliceStable(lv.filtered, sorter(lv.filtered))
	}
}

func (lv *listView) rebuildTable() {
	dateWidth := 18

	// Collect visible custom columns
	var visibleCustom []config.UserColumn
	for _, col := range lv.columns {
		if !lv.cfg.IsColumnHidden(col.Name) {
			visibleCustom = append(visibleCustom, col)
		}
	}

	// Count visible columns to account for Cell padding (Padding(0,1) = 2 chars per column)
	numCols := 1 // Name is always shown
	if lv.cfg.ShowDate {
		numCols++
	}
	if lv.cfg.ShowDirectory {
		numCols++
	}
	if lv.cfg.ShowBranch {
		numCols++
	}
	if lv.cfg.ShowModel {
		numCols++
	}
	if lv.cfg.ShowAgent {
		numCols++
	}
	numCols += len(visibleCustom)
	cellPadding := 2 * numCols

	// Reserve space for custom columns
	customWidth := 0
	for _, col := range visibleCustom {
		customWidth += col.Width
	}

	remaining := lv.width - cellPadding - customWidth
	if lv.cfg.ShowDate {
		remaining -= dateWidth
	}

	agentWidth := 10
	if lv.cfg.ShowAgent {
		remaining -= agentWidth
	}
	modelWidth := 20
	if lv.cfg.ShowModel {
		remaining -= modelWidth
	}

	hasOptional := lv.cfg.ShowDirectory || lv.cfg.ShowBranch

	// Distribute remaining space
	var nameWidth, dirWidth, branchWidth int
	if !hasOptional {
		nameWidth = remaining
	} else {
		nameWidth = remaining * 40 / 100
		optionalSpace := remaining - nameWidth
		visibleOptional := 0
		if lv.cfg.ShowDirectory {
			visibleOptional++
		}
		if lv.cfg.ShowBranch {
			visibleOptional++
		}
		if visibleOptional > 0 {
			each := optionalSpace / visibleOptional
			if lv.cfg.ShowDirectory {
				dirWidth = each
			}
			if lv.cfg.ShowBranch {
				branchWidth = optionalSpace - dirWidth
			}
		} else {
			nameWidth = remaining
		}
	}

	// Build columns
	cols := []table.Column{{Title: "Name", Width: nameWidth}}
	if lv.cfg.ShowAgent {
		cols = append(cols, table.Column{Title: "Agent", Width: agentWidth})
	}
	if lv.cfg.ShowDate {
		cols = append(cols, table.Column{Title: "Date", Width: dateWidth})
	}
	if lv.cfg.ShowDirectory {
		cols = append(cols, table.Column{Title: "Directory", Width: dirWidth})
	}
	if lv.cfg.ShowBranch {
		cols = append(cols, table.Column{Title: "Branch", Width: branchWidth})
	}
	if lv.cfg.ShowModel {
		cols = append(cols, table.Column{Title: "Model", Width: modelWidth})
	}
	for _, col := range visibleCustom {
		cols = append(cols, table.Column{Title: col.Header, Width: col.Width})
	}

	home := os.Getenv("HOME")

	rows := make([]table.Row, 0, len(lv.filtered))
	for _, s := range lv.filtered {
		dir := s.Directory
		if home != "" && strings.HasPrefix(dir, home) {
			dir = "~" + dir[len(home):]
		}
		name := s.Name
		if !lv.cfg.ShowAgent {
			name = agentPrefix(s.Agent) + name
		}
		if lv.pins[s.ID] {
			name = "* " + name
		}
		row := table.Row{truncate(name, nameWidth)}
		if lv.cfg.ShowAgent {
			row = append(row, string(s.Agent))
		}
		if lv.cfg.ShowDate {
			row = append(row, s.StartedAt.Format("2006-01-02 15:04"))
		}
		if lv.cfg.ShowDirectory {
			row = append(row, truncate(dir, dirWidth))
		}
		if lv.cfg.ShowBranch {
			row = append(row, truncate(s.Branch, branchWidth))
		}
		if lv.cfg.ShowModel {
			row = append(row, truncate(shortModel(s.ModelID), modelWidth))
		}
		for _, col := range visibleCustom {
			val := "..."
			if data, ok := lv.columnData[col.Name]; ok {
				if v, ok := data[s.ID]; ok {
					val = v
				}
			}
			row = append(row, truncate(val, col.Width))
		}
		rows = append(rows, row)
	}

	tableHeight := lv.height - 6
	if tableHeight < 1 {
		tableHeight = 1
	}

	lv.table = table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(tableHeight),
		table.WithWidth(lv.width),
		table.WithStyles(themedTableStyles()),
		table.WithFocused(true),
	)
}

func (lv *listView) applyFilter() {
	query := strings.ToLower(lv.filter.Value())
	if query == "" {
		lv.filtered = lv.sessions
	} else {
		lv.filtered = nil
		for _, s := range lv.sessions {
			hay := strings.ToLower(s.Name + s.Directory + s.Branch)
			if strings.Contains(hay, query) {
				lv.filtered = append(lv.filtered, s)
			}
		}
	}
	lv.sortSessions()
	lv.rebuildTable()
}

func (lv *listView) selectedSession() *model.Session {
	if len(lv.filtered) == 0 {
		return nil
	}
	idx := lv.table.Cursor()
	if idx < 0 || idx >= len(lv.filtered) {
		return nil
	}
	return &lv.filtered[idx]
}

func (lv *listView) removeSession(id string) {
	lv.sessions = removeByID(lv.sessions, id)
	lv.filtered = removeByID(lv.filtered, id)
	lv.rebuildTable()
}

func (lv *listView) view() string {
	var b strings.Builder

	count := countStyle.Render(fmt.Sprintf("%d sessions", len(lv.filtered)))
	b.WriteString(count + "\n\n")

	b.WriteString(lv.table.View())
	b.WriteString("\n")

	if lv.filtering {
		b.WriteString(filterStyle.Render(lv.filter.View()))
	} else {
		sep := helpSepStyle.Render(" • ")
		b.WriteString(
			helpItem("↑/↓", "navigate") + sep +
				helpItem("n", "new") + sep +
				helpItem("enter", "resume") + sep +
				helpItem("f", "fork") + sep +
				helpItem("v", "view") + sep +
				helpItem("c", "copy") + sep +
				helpItem("p", "pin") + sep +
				helpItem("/", "filter") + sep +
				helpItem("d", "delete") + sep +
				helpItem("s", "settings") + sep +
				helpItem("tab", "skills") + sep +
				helpItem("q", "quit"),
		)
	}

	return b.String()
}

// helpers

func removeByID(sessions []model.Session, id string) []model.Session {
	result := make([]model.Session, 0, len(sessions))
	for _, s := range sessions {
		if s.ID != id {
			result = append(result, s)
		}
	}
	return result
}

func agentPrefix(agent model.Agent) string {
	switch agent {
	case model.AgentCodex:
		return "[codex] "
	case model.AgentGemini:
		return "[gemini] "
	default:
		return ""
	}
}

func shortModel(id string) string {
	if id == "" {
		return "-"
	}
	return strings.TrimPrefix(id, "claude-")
}

func truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return string(runes[:maxWidth])
	}
	return string(runes[:maxWidth-3]) + "..."
}
