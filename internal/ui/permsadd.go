package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
	"tracer/internal/ccsettings"
)

type addRuleStep int

const (
	addRuleStepList addRuleStep = iota
	addRuleStepRule
)

type addRuleState struct {
	active bool
	step   addRuleStep
	files  []ccsettings.SettingsFile
	list   string // "allow" or "deny"
	input  textinput.Model
}

func newAddRuleState(files []ccsettings.SettingsFile) addRuleState {
	ti := textinput.New()
	ti.Placeholder = "Bash(npm run *)"
	ti.CharLimit = 200
	return addRuleState{
		active: true,
		step:   addRuleStepList,
		files:  files,
		list:   "allow",
		input:  ti,
	}
}

func (ar *addRuleState) update(msg tea.KeyPressMsg) (done bool, result *addRuleResult) {
	switch ar.step {
	case addRuleStepList:
		switch msg.String() {
		case "esc":
			ar.active = false
			return true, nil
		case "left", "right", "h", "l":
			if ar.list == "allow" {
				ar.list = "deny"
			} else {
				ar.list = "allow"
			}
		case "enter":
			ar.step = addRuleStepRule
			ar.input.Focus()
		}
	case addRuleStepRule:
		switch msg.String() {
		case "esc":
			ar.step = addRuleStepList
		case "enter":
			rule := strings.TrimSpace(ar.input.Value())
			if rule != "" {
				ar.active = false
				fileIdx := 0
				if ar.list == "deny" {
					fileIdx = 1
				}
				return true, &addRuleResult{
					fileIdx: fileIdx,
					list:    ar.list,
					rule:    rule,
				}
			}
		default:
			ar.input, _ = ar.input.Update(msg)
		}
	}
	return false, nil
}

func (ar addRuleState) view() string {
	var b strings.Builder

	b.WriteString(helpKeyStyle.Render("Add permission rule") + "\n\n")

	b.WriteString(labelStyle.Render("List:") + " ")
	if ar.step == addRuleStepList && ar.list == "allow" {
		b.WriteString(titleStyle.Render(" allow "))
	} else if ar.list == "allow" {
		b.WriteString(valueStyle.Render("allow"))
	} else {
		b.WriteString(helpDescStyle.Render("allow"))
	}
	b.WriteString("  ")
	if ar.step == addRuleStepList && ar.list == "deny" {
		b.WriteString(titleStyle.Render(" deny "))
	} else if ar.list == "deny" {
		b.WriteString(valueStyle.Render("deny"))
	} else {
		b.WriteString(helpDescStyle.Render("deny"))
	}
	b.WriteString("\n")

	if ar.step >= addRuleStepRule {
		b.WriteString(labelStyle.Render("Rule:") + " " + ar.input.View() + "\n")
	}

	b.WriteString("\n")
	switch ar.step {
	case addRuleStepList:
		b.WriteString(helpDescStyle.Render("←/→ toggle allow/deny · enter confirm · esc cancel"))
	case addRuleStepRule:
		b.WriteString(helpDescStyle.Render("type rule (e.g. Bash(npm run *)) · enter add · esc back"))
	}

	return b.String()
}

type addRuleResult struct {
	fileIdx int
	list    string
	rule    string
}
