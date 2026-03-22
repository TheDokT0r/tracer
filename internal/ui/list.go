package ui

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"tracer/internal/model"
)

type listView struct {
	table    table.Model
	filter   textinput.Model
	filtering bool
	sessions []model.Session
	filtered []model.Session
	width    int
	height   int
}

func newListView(sessions []model.Session, width, height int) listView {
	ti := textinput.New()
	ti.Prompt = "Filter: "
	ti.Placeholder = "type to filter..."

	lv := listView{
		filter:   ti,
		sessions: sessions,
		filtered: sessions,
		width:    width,
		height:   height,
	}
	lv.rebuildTable()
	return lv
}

func (lv *listView) rebuildTable() {
	dateWidth := 18
	padding := 8
	remaining := lv.width - dateWidth - padding
	if remaining < 30 {
		remaining = 30
	}
	nameWidth := remaining * 40 / 100
	dirWidth := remaining * 30 / 100
	branchWidth := remaining - nameWidth - dirWidth

	cols := []table.Column{
		{Title: "Name", Width: nameWidth},
		{Title: "Date", Width: dateWidth},
		{Title: "Directory", Width: dirWidth},
		{Title: "Branch", Width: branchWidth},
	}

	home := os.Getenv("HOME")

	rows := make([]table.Row, 0, len(lv.filtered))
	for _, s := range lv.filtered {
		dir := s.Directory
		if home != "" && strings.HasPrefix(dir, home) {
			dir = "~" + dir[len(home):]
		}
		rows = append(rows, table.Row{
			truncate(s.Name, nameWidth),
			s.StartedAt.Format("2006-01-02 15:04"),
			truncate(dir, dirWidth),
			truncate(s.Branch, branchWidth),
		})
	}

	tableHeight := lv.height - 6
	if tableHeight < 1 {
		tableHeight = 1
	}

	s := table.DefaultStyles()
	s.Header = s.Header.
		Bold(true).
		Foreground(purple).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder())
	s.Selected = s.Selected.
		Foreground(white).
		Background(purple).
		Bold(true)

	lv.table = table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(tableHeight),
		table.WithWidth(lv.width),
		table.WithStyles(s),
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

	title := titleStyle.Render("tracer")
	count := countStyle.Render(fmt.Sprintf(" %d sessions", len(lv.filtered)))
	b.WriteString(title + count + "\n\n")

	b.WriteString(lv.table.View())
	b.WriteString("\n")

	if lv.filtering {
		b.WriteString(filterStyle.Render(lv.filter.View()))
	} else {
		sep := helpSepStyle.Render(" • ")
		b.WriteString(
			helpKeyStyle.Render("↑/↓") + helpDescStyle.Render(" navigate") + sep +
				helpKeyStyle.Render("enter") + helpDescStyle.Render(" resume") + sep +
				helpKeyStyle.Render("v") + helpDescStyle.Render(" view") + sep +
				helpKeyStyle.Render("c") + helpDescStyle.Render(" copy") + sep +
				helpKeyStyle.Render("/") + helpDescStyle.Render(" filter") + sep +
				helpKeyStyle.Render("d") + helpDescStyle.Render(" delete") + sep +
				helpKeyStyle.Render("q") + helpDescStyle.Render(" quit"),
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

func truncate(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if len(s) <= maxWidth {
		return s
	}
	if maxWidth <= 3 {
		return s[:maxWidth]
	}
	return s[:maxWidth-3] + "..."
}
