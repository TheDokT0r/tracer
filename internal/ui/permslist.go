package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"tracer/internal/ccsettings"
)

type permsListView struct {
	table     table.Model
	filter    textinput.Model
	filtering bool
	files     []ccsettings.SettingsFile
	filtered  []ccsettings.SettingsFile
	width     int
	height    int
}

func newPermsListView(files []ccsettings.SettingsFile, width, height int) permsListView {
	ti := textinput.New()
	ti.Prompt = "Filter: "
	ti.Placeholder = "type to filter..."

	pv := permsListView{
		filter:   ti,
		files:    files,
		filtered: files,
		width:    width,
		height:   height,
	}
	pv.rebuildTable()
	return pv
}

func (pv *permsListView) rebuildTable() {
	numCols := 3
	cellPadding := 2 * numCols
	scopeWidth := (pv.width - cellPadding) * 10 / 100
	rulesWidth := (pv.width - cellPadding) * 15 / 100
	pathWidth := pv.width - cellPadding - scopeWidth - rulesWidth

	cols := []table.Column{
		{Title: "Scope", Width: scopeWidth},
		{Title: "Rules", Width: rulesWidth},
		{Title: "Path", Width: pathWidth},
	}

	rows := make([]table.Row, 0, len(pv.filtered))
	for _, f := range pv.filtered {
		count := len(f.Permissions.Allow) + len(f.Permissions.Deny)
		rows = append(rows, table.Row{
			string(f.Scope),
			fmt.Sprintf("%d", count),
			truncate(shortenHome(f.Path), pathWidth),
		})
	}

	tableHeight := pv.height - 6
	if tableHeight < 1 {
		tableHeight = 1
	}

	t := CurrentTheme()
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Bold(true).
		Foreground(t.Primary)
	s.Selected = s.Selected.
		Foreground(t.SelectFg).
		Background(t.SelectBg).
		Bold(true)

	pv.table = table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(tableHeight),
		table.WithWidth(pv.width),
		table.WithStyles(s),
		table.WithFocused(true),
	)
}

func (pv *permsListView) applyFilter() {
	query := strings.ToLower(pv.filter.Value())
	if query == "" {
		pv.filtered = pv.files
	} else {
		pv.filtered = nil
		for _, f := range pv.files {
			hay := strings.ToLower(string(f.Scope) + f.Path)
			if strings.Contains(hay, query) {
				pv.filtered = append(pv.filtered, f)
			}
		}
	}
	pv.rebuildTable()
}

func (pv *permsListView) selectedFile() *ccsettings.SettingsFile {
	if len(pv.filtered) == 0 {
		return nil
	}
	idx := pv.table.Cursor()
	if idx < 0 || idx >= len(pv.filtered) {
		return nil
	}
	return &pv.filtered[idx]
}

func (pv *permsListView) view() string {
	var b strings.Builder

	count := countStyle.Render(fmt.Sprintf("%d settings files", len(pv.filtered)))
	b.WriteString(count + "\n\n")

	b.WriteString(pv.table.View())
	b.WriteString("\n")

	if pv.filtering {
		b.WriteString(filterStyle.Render(pv.filter.View()))
	} else {
		sep := helpSepStyle.Render(" • ")
		b.WriteString(
			helpKeyStyle.Render("↑/↓") + helpDescStyle.Render(" navigate") + sep +
				helpKeyStyle.Render("enter/v") + helpDescStyle.Render(" view rules") + sep +
				helpKeyStyle.Render("/") + helpDescStyle.Render(" filter") + sep +
				helpKeyStyle.Render("tab") + helpDescStyle.Render(" switch tab") + sep +
				helpKeyStyle.Render("q") + helpDescStyle.Render(" quit"),
		)
	}

	return b.String()
}
