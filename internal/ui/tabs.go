package ui

import "strings"

type Tab int

const (
	TabSessions Tab = iota
	TabSkills
	TabPermissions
)

var tabNames = []string{"Sessions", "Skills", "Permissions"}

type tabBar struct {
	active Tab
}

func (tb tabBar) view(width int) string {
	var b strings.Builder
	for i, name := range tabNames {
		if Tab(i) == tb.active {
			b.WriteString(titleStyle.Render(name))
		} else {
			b.WriteString(helpDescStyle.Render(" " + name + " "))
		}
		if i < len(tabNames)-1 {
			b.WriteString(" ")
		}
	}
	b.WriteString("\n\n")
	return b.String()
}
