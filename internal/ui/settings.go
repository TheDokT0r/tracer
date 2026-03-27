package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"tracer/internal/config"
)

// settingItem is a single row in the settings view.
// Each item carries its own read/write functions — no index math needed.
type settingItem struct {
	label   string
	section string         // non-empty = render section header above this item
	value   func() string  // read current display value
	cycle   func(dir int)  // change value (+1 = right, -1 = left)
}

type settingsView struct {
	cfg          config.Config
	items        []settingItem
	cursor       int
	width        int
	height       int
	dirty        bool // changed since last save
	confirmExit  bool // asking "save before exit?"
}

func newSettingsView(cfg config.Config, width, height int) *settingsView {
	sv := &settingsView{
		cfg:    cfg,
		width:  width,
		height: height,
	}
	sv.items = sv.buildItems()
	return sv
}

// cycleEnum cycles through a list of options by dir (+1/-1).
func cycleEnum(options []string, current *string, dir int) {
	for i, o := range options {
		if o == *current {
			*current = options[(i+dir+len(options))%len(options)]
			return
		}
	}
}

func toggleBool(b *bool, _ int) { *b = !*b }

func (sv *settingsView) buildItems() []settingItem {
	items := []settingItem{
		{
			label: "Theme",
			value: func() string { return sv.cfg.Theme },
			cycle: func(dir int) {
				cycleEnum(ThemeNames(), &sv.cfg.Theme, dir)
				ApplyTheme(Themes[sv.cfg.Theme])
			},
		},
		{
			label: "Sort by",
			value: func() string { return sv.cfg.SortBy },
			cycle: func(dir int) {
				cycleEnum([]string{"date", "name", "directory"}, &sv.cfg.SortBy, dir)
			},
		},
		{
			label: "Model",
			value: func() string {
				if sv.cfg.Model == "" {
					return "(default)"
				}
				return sv.cfg.Model
			},
			cycle: func(dir int) {
				models := []string{"", "opus", "sonnet", "haiku",
					"claude-opus-4-6[1m]", "claude-sonnet-4-6[1m]",
					"claude-opus-4-6", "claude-sonnet-4-6"}
				cycleEnum(models, &sv.cfg.Model, dir)
			},
		},
		{
			label: "Confirm delete",
			value: func() string { return boolDisplay(sv.cfg.ConfirmDelete) },
			cycle: func(_ int) { sv.cfg.ConfirmDelete = !sv.cfg.ConfirmDelete },
		},
		{
			label: "Auto update",
			value: func() string { return boolDisplay(sv.cfg.AutoUpdate) },
			cycle: func(_ int) { sv.cfg.AutoUpdate = !sv.cfg.AutoUpdate },
		},

		// -- Columns --
		{
			label:   "Date",
			section: "Columns",
			value:   func() string { return boolDisplay(sv.cfg.ShowDate) },
			cycle:   func(_ int) { sv.cfg.ShowDate = !sv.cfg.ShowDate },
		},
		{
			label: "Directory",
			value: func() string { return boolDisplay(sv.cfg.ShowDirectory) },
			cycle: func(_ int) { sv.cfg.ShowDirectory = !sv.cfg.ShowDirectory },
		},
		{
			label: "Branch",
			value: func() string { return boolDisplay(sv.cfg.ShowBranch) },
			cycle: func(_ int) { sv.cfg.ShowBranch = !sv.cfg.ShowBranch },
		},
		{
			label: "Model",
			value: func() string { return boolDisplay(sv.cfg.ShowModel) },
			cycle: func(_ int) { sv.cfg.ShowModel = !sv.cfg.ShowModel },
		},
		{
			label: "Agent",
			value: func() string { return boolDisplay(sv.cfg.ShowAgent) },
			cycle: func(_ int) { sv.cfg.ShowAgent = !sv.cfg.ShowAgent },
		},
	}

	// Custom column toggles
	userColumns := config.ScanUserColumns()
	for _, col := range userColumns {
		col := col
		items = append(items, settingItem{
			label: col.Header,
			value: func() string { return boolDisplay(!sv.cfg.IsColumnHidden(col.Name)) },
			cycle: func(_ int) { sv.cfg.ToggleColumn(col.Name) },
		})
	}

	// -- Command Palette --
	items = append(items,
		settingItem{
			label:   "Dropdown",
			section: "Command Palette",
			value:   func() string { return boolDisplay(sv.cfg.CmdDropdown) },
			cycle:   func(_ int) { sv.cfg.CmdDropdown = !sv.cfg.CmdDropdown },
		},
		settingItem{
			label: "Ghost suggest",
			value: func() string { return boolDisplay(sv.cfg.CmdGhost) },
			cycle: func(_ int) { sv.cfg.CmdGhost = !sv.cfg.CmdGhost },
		},
		settingItem{
			label: "Max suggestions",
			value: func() string { return fmt.Sprintf("%d", sv.cfg.CmdMaxSuggestions) },
			cycle: func(dir int) {
				sv.cfg.CmdMaxSuggestions += dir
				if sv.cfg.CmdMaxSuggestions > 12 {
					sv.cfg.CmdMaxSuggestions = 3
				} else if sv.cfg.CmdMaxSuggestions < 3 {
					sv.cfg.CmdMaxSuggestions = 12
				}
			},
		},
	)

	// -- Agents --
	items = append(items,
		settingItem{
			label:   "Claude",
			section: "Agents",
			value:   func() string { return boolDisplay(sv.cfg.AgentClaude) },
			cycle:   func(_ int) { sv.cfg.AgentClaude = !sv.cfg.AgentClaude },
		},
		settingItem{
			label: "Codex",
			value: func() string { return boolDisplay(sv.cfg.AgentCodex) },
			cycle: func(_ int) { sv.cfg.AgentCodex = !sv.cfg.AgentCodex },
		},
		settingItem{
			label: "Gemini",
			value: func() string { return boolDisplay(sv.cfg.AgentGemini) },
			cycle: func(_ int) { sv.cfg.AgentGemini = !sv.cfg.AgentGemini },
		},
	)

	return items
}

func (sv *settingsView) cycle(dir int) {
	if sv.cursor >= 0 && sv.cursor < len(sv.items) {
		sv.items[sv.cursor].cycle(dir)
		sv.dirty = true
	}
}

func (sv *settingsView) save() {
	config.SaveConfig(sv.cfg)
	sv.dirty = false
}

func (sv settingsView) view() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("settings"))
	b.WriteString("\n\n")

	for i, item := range sv.items {
		if item.section != "" {
			b.WriteString("\n")
			b.WriteString(dimmedStyle.Render("  ── " + item.section + " ──"))
			b.WriteString("\n")
		}

		cursor := "  "
		if i == sv.cursor {
			cursor = "> "
		}

		label := fmt.Sprintf("%-18s", item.label)
		val := item.value()

		if i == sv.cursor {
			b.WriteString(
				helpKeyStyle.Render(cursor) +
					valueStyle.Render(label) +
					helpKeyStyle.Render("< ") +
					titleStyle.Render(fmt.Sprintf(" %s ", val)) +
					helpKeyStyle.Render(" >"))
		} else {
			b.WriteString(
				dimmedStyle.Render(cursor) +
					dimmedStyle.Render(label) +
					dimmedStyle.Render("  ") +
					helpDescStyle.Render(val))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	sep := helpSepStyle.Render(" • ")
	saved := ""
	if sv.dirty {
		saved = sep + deletePromptStyle.Render("unsaved")
	}
	b.WriteString(
		helpItem("↑/↓", "navigate") + sep +
			helpItem("←/→", "change") + sep +
			helpItem(metaKey()+"+s", "save") + sep +
			helpItem("esc", "back") + saved,
	)

	content := b.String()

	if sv.confirmExit {
		content = overlayPopup(content, sv.width)
	}

	return content
}

func overlayPopup(content string, width int) string {
	lines := strings.Split(content, "\n")

	innerWidth := 42
	sep := helpSepStyle.Render(" • ")
	titleText := "Unsaved changes"
	titlePad := (innerWidth - len(titleText)) / 2
	title := strings.Repeat(" ", titlePad) + deletePromptStyle.Render(titleText)
	body := " " + helpItem("y", "save & exit") + sep +
		helpItem("n", "discard") + sep +
		helpItem("esc", "cancel")
	border := helpKeyStyle.Render

	topLine := border("╭" + strings.Repeat("─", innerWidth) + "╮")
	emptyLine := border("│") + strings.Repeat(" ", innerWidth) + border("│")
	botLine := border("╰" + strings.Repeat("─", innerWidth) + "╯")
	textLine := func(s string) string {
		pad := innerWidth - visibleLen(s)
		if pad < 0 {
			pad = 0
		}
		return border("│") + s + strings.Repeat(" ", pad) + border("│")
	}

	popupLines := []string{
		topLine,
		emptyLine,
		textLine(title),
		emptyLine,
		textLine(body),
		emptyLine,
		botLine,
	}

	startY := (len(lines) - len(popupLines)) / 2
	if startY < 1 {
		startY = 1
	}

	padLeft := (width - innerWidth - 2) / 2
	if padLeft < 0 {
		padLeft = 0
	}
	prefix := strings.Repeat(" ", padLeft)

	for i, pline := range popupLines {
		y := startY + i
		if y < len(lines) {
			lines[y] = prefix + pline
		}
	}

	return strings.Join(lines, "\n")
}

func visibleLen(s string) int {
	n := 0
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
		} else if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
		} else {
			n++
		}
	}
	return n
}

// SettingsApp wraps settingsView as a standalone tea.Model for the subcommand.
type SettingsApp struct {
	sv    *settingsView
	saved bool
}

func NewSettingsApp(cfg config.Config) SettingsApp {
	return SettingsApp{
		sv: newSettingsView(cfg, 80, 24),
	}
}

func (sa SettingsApp) Init() tea.Cmd { return nil }

func (sa SettingsApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		sa.sv.width = msg.Width
		sa.sv.height = msg.Height
		return sa, nil
	case tea.KeyPressMsg:
		if sa.sv.confirmExit {
			switch msg.String() {
			case "y", "Y":
				sa.sv.save()
				sa.saved = true
				return sa, tea.Quit
			case "n", "N":
				return sa, tea.Quit
			case "esc":
				sa.sv.confirmExit = false
			}
			return sa, nil
		}

		switch msg.String() {
		case "esc", "q":
			if sa.sv.dirty {
				sa.sv.confirmExit = true
				return sa, nil
			}
			return sa, tea.Quit
		case "ctrl+s", "super+s":
			sa.sv.save()
			sa.saved = true
			return sa, nil
		case "ctrl+c":
			return sa, tea.Quit
		case "up", "k":
			if sa.sv.cursor > 0 {
				sa.sv.cursor--
			}
		case "down", "j":
			if sa.sv.cursor < len(sa.sv.items)-1 {
				sa.sv.cursor++
			}
		case "right", "l", "enter":
			sa.sv.cycle(1)
		case "left", "h":
			sa.sv.cycle(-1)
		}
	}
	return sa, nil
}

func (sa SettingsApp) View() tea.View {
	return tea.NewView(sa.sv.view())
}

func (sa SettingsApp) Config() config.Config {
	return sa.sv.cfg
}

func (sa SettingsApp) Saved() bool {
	return sa.saved
}

func boolDisplay(v bool) string {
	if v {
		return "on"
	}
	return "off"
}
