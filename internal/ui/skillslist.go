package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	skillspkg "tracer/internal/skills"
)

type skillsListView struct {
	table     table.Model
	filter    textinput.Model
	filtering bool
	skills    []skillspkg.Skill
	filtered  []skillspkg.Skill
	width     int
	height    int
}

func newSkillsListView(skills []skillspkg.Skill, width, height int) skillsListView {
	ti := textinput.New()
	ti.Prompt = "Filter: "
	ti.Placeholder = "type to filter..."

	sv := skillsListView{
		filter:   ti,
		skills:   skills,
		filtered: skills,
		width:    width,
		height:   height,
	}
	sv.rebuildTable()
	return sv
}

func (sv *skillsListView) rebuildTable() {
	numCols := 3
	cellPadding := 2 * numCols
	nameWidth := (sv.width - cellPadding) * 40 / 100
	sourceWidth := (sv.width - cellPadding) * 10 / 100
	descWidth := sv.width - cellPadding - nameWidth - sourceWidth

	cols := []table.Column{
		{Title: "Name", Width: nameWidth},
		{Title: "Source", Width: sourceWidth},
		{Title: "Description", Width: descWidth},
	}

	rows := make([]table.Row, 0, len(sv.filtered))
	for _, sk := range sv.filtered {
		rows = append(rows, table.Row{
			truncate(sk.Name, nameWidth),
			truncate(string(sk.Source), sourceWidth),
			truncate(sk.Description, descWidth),
		})
	}

	tableHeight := sv.height - 6
	if tableHeight < 1 {
		tableHeight = 1
	}

	sv.table = table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithHeight(tableHeight),
		table.WithWidth(sv.width),
		table.WithStyles(themedTableStyles()),
		table.WithFocused(true),
	)
}

func (sv *skillsListView) applyFilter() {
	query := strings.ToLower(sv.filter.Value())
	if query == "" {
		sv.filtered = sv.skills
	} else {
		sv.filtered = nil
		for _, sk := range sv.skills {
			hay := strings.ToLower(sk.Name + sk.Description)
			if strings.Contains(hay, query) {
				sv.filtered = append(sv.filtered, sk)
			}
		}
	}
	sv.rebuildTable()
}

func (sv *skillsListView) selectedSkill() *skillspkg.Skill {
	if len(sv.filtered) == 0 {
		return nil
	}
	idx := sv.table.Cursor()
	if idx < 0 || idx >= len(sv.filtered) {
		return nil
	}
	return &sv.filtered[idx]
}

func (sv *skillsListView) removeSkill(name string) {
	sv.skills = removeSkillByName(sv.skills, name)
	sv.filtered = removeSkillByName(sv.filtered, name)
	sv.rebuildTable()
}

func (sv *skillsListView) view() string {
	var b strings.Builder

	count := countStyle.Render(fmt.Sprintf("%d skills", len(sv.filtered)))
	b.WriteString(count + "\n\n")

	b.WriteString(sv.table.View())
	b.WriteString("\n")

	if sv.filtering {
		b.WriteString(filterStyle.Render(sv.filter.View()))
	} else {
		sep := helpSepStyle.Render(" • ")
		b.WriteString(
			helpItem("↑/↓", "navigate") + sep +
				helpItem("enter/v", "view") + sep +
				helpItem("e", "edit") + sep +
				helpItem("n", "new") + sep +
				helpItem("d", "delete") + sep +
				helpItem("/", "filter") + sep +
				helpItem("tab", "sessions") + sep +
				helpItem("q", "quit"),
		)
	}

	return b.String()
}

func removeSkillByName(skills []skillspkg.Skill, name string) []skillspkg.Skill {
	result := make([]skillspkg.Skill, 0, len(skills))
	for _, sk := range skills {
		if sk.Name != name {
			result = append(result, sk)
		}
	}
	return result
}
