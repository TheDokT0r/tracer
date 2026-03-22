package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/viewport"
	skillspkg "tracer/internal/skills"
)

type skillDetailView struct {
	skill    skillspkg.Skill
	viewport viewport.Model
	width    int
	height   int
}

func newSkillDetailView(skill skillspkg.Skill, content string, width, height int) skillDetailView {
	vpHeight := height - 10
	if vpHeight < 1 {
		vpHeight = 1
	}

	vp := viewport.New(
		viewport.WithWidth(width),
		viewport.WithHeight(vpHeight),
	)
	vp.SetContent(content)

	return skillDetailView{
		skill:    skill,
		viewport: vp,
		width:    width,
		height:   height,
	}
}

func (d skillDetailView) headerView() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(d.skill.Name))
	b.WriteString("\n\n")
	b.WriteString(labelStyle.Render("Name") + valueStyle.Render(d.skill.Name) + "\n")
	b.WriteString(labelStyle.Render("Source") + valueStyle.Render(string(d.skill.Source)) + "\n")
	if d.skill.PluginName != "" {
		b.WriteString(labelStyle.Render("Plugin") + valueStyle.Render(d.skill.PluginName) + "\n")
	}
	b.WriteString(labelStyle.Render("Path") + valueStyle.Render(shortenHome(d.skill.Path)) + "\n")
	b.WriteString(labelStyle.Render("Size") + valueStyle.Render(formatSize(d.skill.Size)) + "\n")
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", d.width) + "\n")

	return b.String()
}

func (d skillDetailView) view() string {
	header := d.headerView()
	body := d.viewport.View()
	sep := helpSepStyle.Render(" • ")
	help := helpKeyStyle.Render("↑/↓") + helpDescStyle.Render(" scroll") + sep +
		helpKeyStyle.Render("e") + helpDescStyle.Render(" edit") + sep +
		helpKeyStyle.Render("d") + helpDescStyle.Render(" delete") + sep +
		helpKeyStyle.Render("esc") + helpDescStyle.Render(" back")

	return header + body + "\n" + help
}

func formatSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/1024/1024)
	case bytes >= 1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
