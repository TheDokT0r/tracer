package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"tracer/internal/ccsettings"
)

type permsDetailView struct {
	file     *ccsettings.SettingsFile
	table    table.Model
	rules    []ruleEntry
	width    int
	height   int
}

type ruleEntry struct {
	rule string
	list string // "allow" or "deny"
}

func newPermsDetailView(file *ccsettings.SettingsFile, width, height int) permsDetailView {
	dv := permsDetailView{
		file:   file,
		width:  width,
		height: height,
	}
	dv.buildRules()
	dv.rebuildTable()
	return dv
}

func (dv *permsDetailView) buildRules() {
	dv.rules = nil
	for _, r := range dv.file.Permissions.Allow {
		dv.rules = append(dv.rules, ruleEntry{rule: r, list: "allow"})
	}
	for _, r := range dv.file.Permissions.Deny {
		dv.rules = append(dv.rules, ruleEntry{rule: r, list: "deny"})
	}
}

func (dv *permsDetailView) rebuildTable() {
	listWidth := dv.width * 10 / 100
	ruleWidth := dv.width - listWidth

	cols := []table.Column{
		{Title: "List", Width: listWidth},
		{Title: "Rule", Width: ruleWidth},
	}

	rows := make([]table.Row, 0, len(dv.rules))
	for _, r := range dv.rules {
		rows = append(rows, table.Row{
			r.list,
			truncate(r.rule, ruleWidth),
		})
	}

	tableHeight := dv.height - 10
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

	dv.table = table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(tableHeight),
		table.WithWidth(dv.width),
		table.WithStyles(s),
		table.WithFocused(true),
	)
}

func (dv *permsDetailView) selectedRule() *ruleEntry {
	if len(dv.rules) == 0 {
		return nil
	}
	idx := dv.table.Cursor()
	if idx < 0 || idx >= len(dv.rules) {
		return nil
	}
	return &dv.rules[idx]
}

func (dv *permsDetailView) deleteSelected() {
	r := dv.selectedRule()
	if r == nil {
		return
	}
	ccsettings.RemoveRule(dv.file, r.list, r.rule)
	ccsettings.SavePermissions(*dv.file)
	dv.buildRules()
	dv.rebuildTable()
}

func (dv *permsDetailView) toggleSelected() {
	r := dv.selectedRule()
	if r == nil {
		return
	}
	oldList := r.list
	newList := "deny"
	if oldList == "deny" {
		newList = "allow"
	}
	ccsettings.RemoveRule(dv.file, oldList, r.rule)
	ccsettings.AddRule(dv.file, newList, r.rule)
	ccsettings.SavePermissions(*dv.file)
	dv.buildRules()
	dv.rebuildTable()
}

func (dv *permsDetailView) addRule(list, rule string) {
	ccsettings.AddRule(dv.file, list, rule)
	ccsettings.SavePermissions(*dv.file)
	dv.buildRules()
	dv.rebuildTable()
}

func (dv permsDetailView) view() string {
	var b strings.Builder

	scope := string(dv.file.Scope)
	path := shortenHome(dv.file.Path)

	b.WriteString(titleStyle.Render(scope + " settings"))
	b.WriteString("\n\n")
	b.WriteString(labelStyle.Render("Path") + valueStyle.Render(path) + "\n")
	b.WriteString(labelStyle.Render("Allow rules") + valueStyle.Render(fmt.Sprintf("%d", len(dv.file.Permissions.Allow))) + "\n")
	b.WriteString(labelStyle.Render("Deny rules") + valueStyle.Render(fmt.Sprintf("%d", len(dv.file.Permissions.Deny))) + "\n")
	b.WriteString("\n")

	b.WriteString(dv.table.View())
	b.WriteString("\n")

	sep := helpSepStyle.Render(" • ")
	b.WriteString(
		helpKeyStyle.Render("↑/↓") + helpDescStyle.Render(" navigate") + sep +
			helpKeyStyle.Render("a") + helpDescStyle.Render(" add rule") + sep +
			helpKeyStyle.Render("t") + helpDescStyle.Render(" toggle allow/deny") + sep +
			helpKeyStyle.Render("d") + helpDescStyle.Render(" delete rule") + sep +
			helpKeyStyle.Render("esc") + helpDescStyle.Render(" back"),
	)

	return b.String()
}
