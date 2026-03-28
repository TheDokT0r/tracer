package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"tracer/internal/config"
)

var allExceptSettings = []viewState{
	viewList, viewDetail, viewSkillsList, viewSkillDetail, viewPermsList, viewPermsDetail,
}

func defaultRegistry() *registry {
	return newRegistry([]Command{
		{
			Name: "quit", Description: "Quit tracer",
			Contexts: allExceptSettings,
			Run:      func(a *App, args []string) tea.Cmd { return tea.Quit },
		},
		{
			Name: "delete", Description: "Delete selected item",
			Contexts: []viewState{viewList, viewDetail, viewSkillsList, viewSkillDetail, viewPermsDetail},
			Run: func(a *App, args []string) tea.Cmd {
				if a.view == viewPermsDetail {
					a.permsDetail.deleteSelected()
					return nil
				}
				if a.cfg.ConfirmDelete {
					a.confirmDelete = true
					return nil
				}
				a.executeDelete()
				return nil
			},
		},
		{
			Name: "pin", Description: "Toggle pin on session",
			Contexts: []viewState{viewList},
			Run: func(a *App, args []string) tea.Cmd {
				if s := a.list.selectedSession(); s != nil {
					config.TogglePin(a.list.pins, s.ID)
					config.SavePins(a.list.pins)
					a.list.sortSessions()
					a.list.rebuildTable()
					a.statusMsg = "Toggled pin"
					return statusClearCmd()
				}
				return nil
			},
		},
		{
			Name: "resume", Description: "Resume selected session",
			Contexts: []viewState{viewList, viewDetail},
			Run: func(a *App, args []string) tea.Cmd {
				_, cmd := a.resumeSession()
				return cmd
			},
		},
		{
			Name: "fork", Description: "Fork selected session",
			Contexts: []viewState{viewList, viewDetail},
			Run: func(a *App, args []string) tea.Cmd {
				_, cmd := a.forkSession()
				return cmd
			},
		},
		{
			Name: "view", Description: "View selected item detail",
			Contexts: []viewState{viewList, viewSkillsList, viewPermsList},
			Run: func(a *App, args []string) tea.Cmd {
				switch a.view {
				case viewList:
					a.openSessionDetail()
				case viewSkillsList:
					a.openSkillDetail()
				case viewPermsList:
					if f := a.permsList.selectedFile(); f != nil {
						a.permsDetail = newPermsDetailView(f, a.width, a.height)
						a.view = viewPermsDetail
					}
				}
				return nil
			},
		},
		{
			Name: "copy", Description: "Copy session ID",
			Contexts: []viewState{viewList, viewDetail},
			Run: func(a *App, args []string) tea.Cmd {
				a.copySessionID()
				a.statusMsg = "Copied session ID"
				return statusClearCmd()
			},
		},
		{
			Name: "new", Description: "New session",
			Args:     []CommandArg{{Name: "path", Required: false}},
			Contexts: []viewState{viewList},
			Run: func(a *App, args []string) tea.Cmd {
				if len(args) > 0 {
					// Pre-fill the dir and start the flow
					agents := a.getEnabledAgents()
					if len(agents) == 0 {
						a.statusMsg = "No agents enabled"
						return statusClearCmd()
					}
					if len(agents) == 1 {
						_, cmd := a.launchNewSession(agents[0], args[0])
						return cmd
					}
					a.enabledAgents = agents
					ti := textinput.New()
					ti.SetValue(args[0])
					a.newSessionDir = ti
					a.newSession = true
					a.agentPicker = true
					a.newSessionAgent = 0
					return nil
				}
				_, cmd := a.startNewSession()
				return cmd
			},
		},
		{
			Name: "new skill", Description: "Create new skill",
			Args:     []CommandArg{{Name: "name", Required: false}},
			Contexts: []viewState{viewSkillsList},
			Run: func(a *App, args []string) tea.Cmd {
				if len(args) > 0 {
					return a.createSkillDirect(args[0])
				}
				a.startNewSkill()
				return nil
			},
		},
		{
			Name: "rename", Description: "Rename selected session",
			Args:     []CommandArg{{Name: "name", Required: true}},
			Contexts: []viewState{viewDetail},
			Run: func(a *App, args []string) tea.Cmd {
				if len(args) == 0 {
					a.statusMsg = "Usage: rename <name>"
					return statusClearCmd()
				}
				name := strings.Join(args, " ")
				if s := a.list.selectedSession(); s != nil {
					a.renames[s.ID] = name
					config.SaveRenames(a.renames)
					a.writeRename(*s, name)
					for i := range a.list.sessions {
						if a.list.sessions[i].ID == s.ID {
							a.list.sessions[i].Name = name
							break
						}
					}
					s.Name = name
					a.detail.session.Name = name
					a.list.rebuildTable()
					a.statusMsg = "Renamed to " + name
					return statusClearCmd()
				}
				return nil
			},
		},
		{
			Name: "edit", Description: "Edit in editor",
			Contexts: []viewState{viewDetail, viewSkillsList, viewSkillDetail},
			Run: func(a *App, args []string) tea.Cmd {
				switch a.view {
				case viewDetail:
					_, cmd := a.editSessionFile()
					return cmd
				case viewSkillsList, viewSkillDetail:
					_, cmd := a.editSkillFile()
					return cmd
				}
				return nil
			},
		},
		{
			Name: "export", Description: "Export session",
			Args: []CommandArg{{
				Name: "format", Required: true,
				Options: func(a *App) []Completion {
					return []Completion{
						{Value: "html", Description: "Export as HTML with chat UI"},
						{Value: "md", Description: "Export as Markdown"},
					}
				},
			}},
			Contexts: []viewState{viewDetail},
			Run: func(a *App, args []string) tea.Cmd {
				if len(args) == 0 {
					a.statusMsg = "Usage: export <html|md>"
					return statusClearCmd()
				}
				switch args[0] {
				case "html":
					_, cmd := a.exportHTML()
					return cmd
				case "md":
					_, cmd := a.exportMarkdown()
					return cmd
				default:
					a.statusMsg = "Unknown format: " + args[0] + " (use html or md)"
					return statusClearCmd()
				}
			},
		},
		{
			Name: "filter", Description: "Filter current list",
			Args:     []CommandArg{{Name: "query", Required: false}},
			Contexts: []viewState{viewList, viewSkillsList, viewPermsList},
			Run: func(a *App, args []string) tea.Cmd {
				query := strings.Join(args, " ")
				switch a.view {
				case viewList:
					a.list.filter.SetValue(query)
					a.list.applyFilter()
				case viewSkillsList:
					a.skillsList.filter.SetValue(query)
					a.skillsList.applyFilter()
				case viewPermsList:
					a.permsList.filter.SetValue(query)
					a.permsList.applyFilter()
				}
				if query == "" {
					a.statusMsg = "Filter cleared"
				} else {
					a.statusMsg = "Filtered: " + query
				}
				return statusClearCmd()
			},
		},
		{
			Name: "sort", Description: "Sort sessions",
			Args: []CommandArg{{
				Name: "field", Required: true,
				Options: func(a *App) []Completion {
					return []Completion{
						{Value: "date", Description: "Sort by date (newest first)"},
						{Value: "name", Description: "Sort alphabetically"},
						{Value: "directory", Description: "Sort by working directory"},
					}
				},
			}},
			Contexts: []viewState{viewList},
			Run: func(a *App, args []string) tea.Cmd {
				if len(args) == 0 {
					a.statusMsg = "Usage: sort <date|name|directory>"
					return statusClearCmd()
				}
				switch args[0] {
				case "date", "name", "directory":
					a.cfg.SortBy = args[0]
					config.SaveConfig(a.cfg)
					a.list.cfg = a.cfg
					a.list.sortSessions()
					a.list.rebuildTable()
					a.statusMsg = "Sorted by " + args[0]
					return statusClearCmd()
				default:
					a.statusMsg = "Unknown sort field: " + args[0]
					return statusClearCmd()
				}
			},
		},
		{
			Name: "set", Description: "Change a setting",
			Args: []CommandArg{
				{
					Name: "key", Required: true,
					Options: func(a *App) []Completion {
						return []Completion{
							{Value: "theme", Description: "Color theme"},
							{Value: "sort_by", Description: "Sort order"},
							{Value: "show_date", Description: "Show date column"},
							{Value: "show_directory", Description: "Show directory column"},
							{Value: "show_branch", Description: "Show branch column"},
							{Value: "confirm_delete", Description: "Confirm before delete"},
							{Value: "auto_update", Description: "Auto-update on exit"},
							{Value: "cmd_dropdown", Description: "Command dropdown"},
							{Value: "cmd_ghost", Description: "Ghost text suggestions"},
							{Value: "cmd_max_suggestions", Description: "Max dropdown items"},
						}
					},
				},
				{
					Name: "value", Required: true,
					Options: func(a *App) []Completion {
						if a == nil || !a.commandMode {
							return nil
						}
						input := a.cmdInput.input.Value()
						_, args := a.cmdRegistry.resolve(input)
						if len(args) == 0 {
							return nil
						}
						switch args[0] {
						case "theme":
							names := ThemeNames()
							var comps []Completion
							for _, n := range names {
								comps = append(comps, Completion{Value: n})
							}
							return comps
						case "sort_by":
							return []Completion{{Value: "date", Description: "Newest first"}, {Value: "name", Description: "Alphabetical"}, {Value: "directory", Description: "By directory"}}
						case "show_date", "show_directory", "show_branch",
							"confirm_delete", "auto_update", "cmd_dropdown", "cmd_ghost":
							return []Completion{{Value: "on"}, {Value: "off"}}
						}
						return nil
					},
				},
			},
			Contexts: allExceptSettings,
			Run: func(a *App, args []string) tea.Cmd {
				if len(args) < 2 {
					a.statusMsg = "Usage: set <key> <value>"
					return statusClearCmd()
				}
				return a.runSet(args[0], args[1])
			},
		},
		{
			Name: "theme", Description: "Switch theme",
			Args: []CommandArg{{
				Name: "name", Required: true,
				Options: func(a *App) []Completion {
					names := ThemeNames()
					var comps []Completion
					for _, n := range names {
						comps = append(comps, Completion{Value: n})
					}
					return comps
				},
			}},
			Contexts: allExceptSettings,
			Run: func(a *App, args []string) tea.Cmd {
				if len(args) == 0 {
					a.statusMsg = "Usage: theme <name>"
					return statusClearCmd()
				}
				return a.runSet("theme", args[0])
			},
		},
		{
			Name: "model", Description: "Set model for Claude sessions",
			Args: []CommandArg{{
				Name: "model", Required: true,
				Options: func(a *App) []Completion {
					return []Completion{
						{Value: "opus", Description: "Claude Opus (latest)"},
						{Value: "sonnet", Description: "Claude Sonnet (latest)"},
						{Value: "haiku", Description: "Claude Haiku (latest)"},
						{Value: "claude-opus-4-6", Description: "Opus 4.6"},
						{Value: "claude-opus-4-6[1m]", Description: "Opus 4.6 (1M context)"},
						{Value: "claude-sonnet-4-6", Description: "Sonnet 4.6"},
						{Value: "claude-sonnet-4-6[1m]", Description: "Sonnet 4.6 (1M context)"},
						{Value: "", Description: "Clear (use Claude default)"},
					}
				},
			}},
			Contexts: allExceptSettings,
			Run: func(a *App, args []string) tea.Cmd {
				if len(args) == 0 {
					if a.cfg.Model == "" {
						a.statusMsg = "Model: (default)"
					} else {
						a.statusMsg = "Model: " + a.cfg.Model
					}
					return statusClearCmd()
				}
				return a.runSet("model", args[0])
			},
		},
		{
			Name: "settings", Description: "Open settings",
			Contexts: []viewState{viewList},
			Run: func(a *App, args []string) tea.Cmd {
				a.settings = newSettingsView(a.cfg, a.width, a.height)
				a.view = viewSettings
				return nil
			},
		},
		{
			Name: "tab", Description: "Switch tab",
			Args: []CommandArg{{
				Name: "name", Required: true,
				Options: func(a *App) []Completion {
					return []Completion{
						{Value: "sessions", Description: "Sessions tab"},
						{Value: "skills", Description: "Skills tab"},
						{Value: "permissions", Description: "Permissions tab"},
					}
				},
			}},
			Contexts: []viewState{viewList, viewSkillsList, viewPermsList},
			Run: func(a *App, args []string) tea.Cmd {
				if len(args) == 0 {
					a.statusMsg = "Usage: tab <sessions|skills|permissions>"
					return statusClearCmd()
				}
				switch args[0] {
				case "sessions":
					a.tab.active = TabSessions
					a.view = viewList
				case "skills":
					a.tab.active = TabSkills
					a.view = viewSkillsList
				case "permissions":
					a.tab.active = TabPermissions
					a.view = viewPermsList
				default:
					a.statusMsg = "Unknown tab: " + args[0]
					return statusClearCmd()
				}
				return nil
			},
		},
		{
			Name: "help", Description: "Show available commands",
			Args:     []CommandArg{{Name: "command", Required: false}},
			Contexts: allExceptSettings,
			Run: func(a *App, args []string) tea.Cmd {
				if len(args) > 0 {
					cmd, _ := a.cmdRegistry.resolve(args[0])
					if cmd == nil {
						a.statusMsg = "Unknown command: " + args[0]
					} else {
						desc := cmd.Name + " — " + cmd.Description
						if len(cmd.Args) > 0 {
							var argNames []string
							for _, arg := range cmd.Args {
								if arg.Required {
									argNames = append(argNames, "<"+arg.Name+">")
								} else {
									argNames = append(argNames, "["+arg.Name+"]")
								}
							}
							desc += " (" + strings.Join(argNames, " ") + ")"
						}
						a.statusMsg = desc
					}
				} else {
					matches := a.cmdRegistry.match("", a.view)
					var names []string
					for _, m := range matches {
						names = append(names, m.Name)
					}
					a.statusMsg = "Commands: " + strings.Join(names, ", ")
				}
				return statusClearCmd()
			},
		},
		{
			Name: "commands", Description: "Manage user commands",
			Args: []CommandArg{
				{
					Name: "subcommand", Required: false,
					Options: func(a *App) []Completion {
						return []Completion{
							{Value: "list", Description: "List user commands"},
							{Value: "new", Description: "Create new command"},
							{Value: "new-ai", Description: "Create with Claude AI"},
							{Value: "edit", Description: "Edit a command"},
							{Value: "delete", Description: "Delete a command"},
						}
					},
				},
				{
					Name: "name", Required: false,
					Options: func(a *App) []Completion {
						if a == nil || !a.commandMode {
							return nil
						}
						input := a.cmdInput.input.Value()
						_, args := a.cmdRegistry.resolve(input)
						if len(args) == 0 {
							return nil
						}
						switch args[0] {
						case "edit", "delete":
							names := config.UserCommandNames()
							var comps []Completion
							for _, n := range names {
								comps = append(comps, Completion{Value: n})
							}
							return comps
						}
						return nil
					},
				},
			},
			Contexts: allExceptSettings,
			Run: func(a *App, args []string) tea.Cmd {
				sub := "list"
				if len(args) > 0 {
					sub = args[0]
				}
				name := ""
				if len(args) > 1 {
					name = args[1]
				}

				switch sub {
				case "list":
					names := config.UserCommandNames()
					if len(names) == 0 {
						a.statusMsg = "No user commands. Use :commands new <name> to create one."
					} else {
						a.statusMsg = "User commands: " + strings.Join(names, ", ")
					}
					return statusClearCmd()

				case "new":
					if name == "" {
						a.statusMsg = "Usage: commands new <name>"
						return statusClearCmd()
					}
					if !config.ValidateCommandName(name) {
						a.statusMsg = "Invalid name (use lowercase letters, numbers, hyphens)"
						return statusClearCmd()
					}
					cmdDir := filepath.Join(os.Getenv("HOME"), ".config", "tracer", "commands", name)
					if _, err := os.Stat(cmdDir); err == nil {
						a.statusMsg = "Command already exists: " + name
						return statusClearCmd()
					}
					if err := config.ScaffoldCommand(name); err != nil {
						a.statusMsg = "Failed: " + err.Error()
						return statusClearCmd()
					}
					a.pendingRescan = true
					a.statusMsg = "Created command: " + name
					return openEditor(filepath.Join(cmdDir, "command.json"))

				case "new-ai":
					if name == "" {
						a.statusMsg = "Usage: commands new-ai <name>"
						return statusClearCmd()
					}
					if !config.ValidateCommandName(name) {
						a.statusMsg = "Invalid name (use lowercase letters, numbers, hyphens)"
						return statusClearCmd()
					}
					cmdDir := filepath.Join(os.Getenv("HOME"), ".config", "tracer", "commands", name)
					if _, err := os.Stat(cmdDir); err == nil {
						a.statusMsg = "Command already exists: " + name
						return statusClearCmd()
					}
					if err := config.ScaffoldCommand(name); err != nil {
						a.statusMsg = "Failed: " + err.Error()
						return statusClearCmd()
					}
					claudeBin, err := exec.LookPath("claude")
					if err != nil {
						a.statusMsg = "Claude not found in PATH"
						return statusClearCmd()
					}
					a.statusMsg = "Created command: " + name
					prompt := newAICommandPrompt(name, cmdDir)
					c := exec.Command(claudeBin, prompt)
					c.Dir = cmdDir
					return tea.ExecProcess(c, func(err error) tea.Msg {
						return userCommandFinishedMsg{}
					})

				case "edit":
					if name == "" {
						a.statusMsg = "Usage: commands edit <name>"
						return statusClearCmd()
					}
					cmdDir := filepath.Join(os.Getenv("HOME"), ".config", "tracer", "commands", name)
					if _, err := os.Stat(filepath.Join(cmdDir, "command.json")); err != nil {
						a.statusMsg = "Command not found: " + name
						return statusClearCmd()
					}
					a.pendingRescan = true
					return openEditor(filepath.Join(cmdDir, "command.json"))

				case "delete":
					if name == "" {
						a.statusMsg = "Usage: commands delete <name>"
						return statusClearCmd()
					}
					cmdDir := filepath.Join(os.Getenv("HOME"), ".config", "tracer", "commands", name)
					if _, err := os.Stat(cmdDir); err != nil {
						a.statusMsg = "Command not found: " + name
						return statusClearCmd()
					}
					if a.cfg.ConfirmDelete {
						a.pendingCommandDelete = name
						a.confirmDelete = true
						return nil
					}
					config.DeleteCommand(name)
					a.rescanCommands()
					a.statusMsg = "Deleted command: " + name
					return statusClearCmd()

				default:
					a.statusMsg = "Unknown subcommand: " + sub
					return statusClearCmd()
				}
			},
		},
		{
			Name: "columns", Description: "Manage custom columns",
			Args: []CommandArg{
				{
					Name: "subcommand", Required: false,
					Options: func(a *App) []Completion {
						return []Completion{
							{Value: "list", Description: "List custom columns"},
							{Value: "new", Description: "Create new column"},
							{Value: "new-ai", Description: "Create with Claude AI"},
							{Value: "edit", Description: "Edit a column"},
							{Value: "delete", Description: "Delete a column"},
							{Value: "toggle", Description: "Show/hide a column"},
						}
					},
				},
				{
					Name: "name", Required: false,
					Options: func(a *App) []Completion {
						if a == nil || !a.commandMode {
							return nil
						}
						input := a.cmdInput.input.Value()
						_, args := a.cmdRegistry.resolve(input)
						if len(args) == 0 {
							return nil
						}
						switch args[0] {
						case "edit", "delete", "toggle":
							names := config.UserColumnNames()
							var comps []Completion
							for _, n := range names {
								comps = append(comps, Completion{Value: n})
							}
							return comps
						}
						return nil
					},
				},
			},
			Contexts: allExceptSettings,
			Run: func(a *App, args []string) tea.Cmd {
				sub := "list"
				if len(args) > 0 {
					sub = args[0]
				}
				name := ""
				if len(args) > 1 {
					name = args[1]
				}

				switch sub {
				case "list":
					names := config.UserColumnNames()
					if len(names) == 0 {
						a.statusMsg = "No custom columns. Use :columns new <name> to create one."
					} else {
						a.statusMsg = "Custom columns: " + strings.Join(names, ", ")
					}
					return statusClearCmd()

				case "new":
					if name == "" {
						a.statusMsg = "Usage: columns new <name>"
						return statusClearCmd()
					}
					if !config.ValidateCommandName(name) {
						a.statusMsg = "Invalid name (use lowercase letters, numbers, hyphens)"
						return statusClearCmd()
					}
					colDir := filepath.Join(os.Getenv("HOME"), ".config", "tracer", "columns", name)
					if _, err := os.Stat(colDir); err == nil {
						a.statusMsg = "Column already exists: " + name
						return statusClearCmd()
					}
					if err := config.ScaffoldColumn(name); err != nil {
						a.statusMsg = "Failed: " + err.Error()
						return statusClearCmd()
					}
					a.pendingRescan = true
					a.statusMsg = "Created column: " + name
					return openEditor(filepath.Join(colDir, "column.json"))

				case "new-ai":
					if name == "" {
						a.statusMsg = "Usage: columns new-ai <name>"
						return statusClearCmd()
					}
					if !config.ValidateCommandName(name) {
						a.statusMsg = "Invalid name (use lowercase letters, numbers, hyphens)"
						return statusClearCmd()
					}
					colDir := filepath.Join(os.Getenv("HOME"), ".config", "tracer", "columns", name)
					if _, err := os.Stat(colDir); err == nil {
						a.statusMsg = "Column already exists: " + name
						return statusClearCmd()
					}
					if err := config.ScaffoldColumn(name); err != nil {
						a.statusMsg = "Failed: " + err.Error()
						return statusClearCmd()
					}
					claudeBin, err := exec.LookPath("claude")
					if err != nil {
						a.statusMsg = "Claude not found in PATH"
						return statusClearCmd()
					}
					a.statusMsg = "Created column: " + name
					prompt := newAIColumnPrompt(name, colDir)
					c := exec.Command(claudeBin, prompt)
					c.Dir = colDir
					return tea.ExecProcess(c, func(err error) tea.Msg {
						return userCommandFinishedMsg{}
					})

				case "edit":
					if name == "" {
						a.statusMsg = "Usage: columns edit <name>"
						return statusClearCmd()
					}
					colDir := filepath.Join(os.Getenv("HOME"), ".config", "tracer", "columns", name)
					if _, err := os.Stat(filepath.Join(colDir, "column.json")); err != nil {
						a.statusMsg = "Column not found: " + name
						return statusClearCmd()
					}
					a.pendingRescan = true
					return openEditor(filepath.Join(colDir, "column.json"))

				case "delete":
					if name == "" {
						a.statusMsg = "Usage: columns delete <name>"
						return statusClearCmd()
					}
					colDir := filepath.Join(os.Getenv("HOME"), ".config", "tracer", "columns", name)
					if _, err := os.Stat(colDir); err != nil {
						a.statusMsg = "Column not found: " + name
						return statusClearCmd()
					}
					if a.cfg.ConfirmDelete {
						a.pendingCommandDelete = "col:" + name
						a.confirmDelete = true
						return nil
					}
					config.DeleteColumn(name)
					a.rescanColumns()
					a.statusMsg = "Deleted column: " + name
					return statusClearCmd()

				case "toggle":
					if name == "" {
						a.statusMsg = "Usage: columns toggle <name>"
						return statusClearCmd()
					}
					// Check column exists
					found := false
					for _, col := range a.list.columns {
						if col.Name == name {
							found = true
							break
						}
					}
					if !found {
						a.statusMsg = "Column not found: " + name
						return statusClearCmd()
					}
					a.cfg.ToggleColumn(name)
					config.SaveConfig(a.cfg)
					a.list.cfg = a.cfg
					a.list.rebuildTable()
					if a.cfg.IsColumnHidden(name) {
						a.statusMsg = "Hidden column: " + name
					} else {
						a.statusMsg = "Showing column: " + name
					}
					return statusClearCmd()

				default:
					a.statusMsg = "Unknown subcommand: " + sub
					return statusClearCmd()
				}
			},
		},
	})
}

func (a *App) runSet(key, value string) tea.Cmd {
	switch key {
	case "theme":
		if _, ok := Themes[value]; !ok {
			a.statusMsg = "Unknown theme: " + value
			return statusClearCmd()
		}
		a.cfg.Theme = value
		ApplyTheme(Themes[value])
	case "sort_by":
		switch value {
		case "date", "name", "directory":
			a.cfg.SortBy = value
			a.list.cfg = a.cfg
			a.list.sortSessions()
			a.list.rebuildTable()
		default:
			a.statusMsg = "Invalid sort: " + value
			return statusClearCmd()
		}
	case "show_date":
		a.cfg.ShowDate = parseBool(value)
		a.list.cfg = a.cfg
		a.list.rebuildTable()
	case "show_directory":
		a.cfg.ShowDirectory = parseBool(value)
		a.list.cfg = a.cfg
		a.list.rebuildTable()
	case "show_branch":
		a.cfg.ShowBranch = parseBool(value)
		a.list.cfg = a.cfg
		a.list.rebuildTable()
	case "show_model":
		a.cfg.ShowModel = parseBool(value)
		a.list.cfg = a.cfg
		a.list.rebuildTable()
	case "show_agent":
		a.cfg.ShowAgent = parseBool(value)
		a.list.cfg = a.cfg
		a.list.rebuildTable()
	case "model":
		a.cfg.Model = value
	case "confirm_delete":
		a.cfg.ConfirmDelete = parseBool(value)
	case "auto_update":
		a.cfg.AutoUpdate = parseBool(value)
	case "cmd_dropdown":
		a.cfg.CmdDropdown = parseBool(value)
	case "cmd_ghost":
		a.cfg.CmdGhost = parseBool(value)
	case "cmd_max_suggestions":
		n, err := strconv.Atoi(value)
		if err != nil || n < 3 || n > 12 {
			a.statusMsg = "Invalid value (3-12): " + value
			return statusClearCmd()
		}
		a.cfg.CmdMaxSuggestions = n
	default:
		a.statusMsg = "Unknown setting: " + key
		return statusClearCmd()
	}
	config.SaveConfig(a.cfg)
	a.statusMsg = fmt.Sprintf("Set %s = %s", key, value)
	return statusClearCmd()
}

func parseBool(s string) bool {
	return s == "on" || s == "true" || s == "1" || s == "yes"
}

func newAIColumnPrompt(name, colDir string) string {
	return fmt.Sprintf(`You are creating a custom column called "%s" for tracer, a TUI session manager for Claude Code.

## What you're building

A custom column adds a new data column to tracer's session list table. Each cell in the column is populated by running a script once per session, with the session's working directory passed as $1.

The column is a directory containing:
- column.json — metadata that tells tracer how to display the column
- run.sh (or any script) — runs per session row, stdout becomes the cell value

The template files are already created in the current directory with placeholders — edit them.

## column.json schema

{
  "description": "What this column shows",
  "header": "Header",
  "shell": "run.sh",
  "width": 10,
  "timeout": 5
}

### Fields explained:
- **description** (required): explains what this column shows (for documentation)
- **header** (required): the column header text displayed in the table
- **shell** (required): filename of script to run, relative to this directory
- **width**: column width in characters (default: 10). Choose based on expected output length.
- **timeout**: max seconds per script execution (default: 5). If exceeded, shows "—".

## How the script works:
- Receives the session's working directory as $1 (positional parameter)
- Must output exactly ONE line to stdout — this becomes the cell value
- Runs once per session row, in parallel across all sessions
- Runs asynchronously — the table shows "..." while loading, then fills in values

## Environment variables:
- $1: session working directory (e.g., /Users/or/projects/myapp)
- SESSION_DIR: same as $1
- SESSION_ID: the session's unique ID
- TRACER_COL_DIR: path to this column's directory

## Examples of useful columns:
- Git branch: git -C "$1" branch --show-current 2>/dev/null || echo "—"
- Dirty files count: git -C "$1" status --porcelain 2>/dev/null | wc -l | tr -d ' '
- Last commit age: git -C "$1" log -1 --format="%%cr" 2>/dev/null || echo "—"
- Language: ls "$1"/*.go 2>/dev/null && echo "Go" || (ls "$1"/*.py 2>/dev/null && echo "Python" || echo "—")
- Directory size: du -sh "$1" 2>/dev/null | cut -f1

## Guidelines:
- Keep output short — it must fit in the column width
- Handle missing directories gracefully (the session dir may not exist anymore)
- Handle missing tools (git may not be installed) — use "—" as fallback
- Fast scripts make a better experience (under 1 second ideal)
- The script runs in its own directory, not the session directory. Use $1 to reference the session dir.

## Your task:
1. Ask the user what data they want this column to show
2. Edit column.json with the right header, width, and timeout
3. Write run.sh implementing the column logic
4. chmod +x run.sh
5. Test by running: ./run.sh /some/project/directory

Current directory: %s`, name, colDir)
}

func newAICommandPrompt(name, cmdDir string) string {
	return fmt.Sprintf(`You are creating a user command called "%s" for tracer, a TUI session manager for Claude Code.

## What you're building

A user command is a directory containing:
- command.json — metadata that tells tracer how to register and run the command
- run.sh (or any script) — the script that executes when the user invokes the command

When the user types ":%s" in tracer's command palette, tracer runs your script.
The template files are already created in the current directory with placeholders — edit them.

## command.json schema

{
  "description": "One-line description shown in the command palette dropdown",
  "shell": "run.sh",
  "mode": "status",
  "args": [
    {
      "name": "arg-name",
      "required": true,
      "completions": [
        {"value": "option1", "description": "What option1 does"},
        {"value": "option2", "description": "What option2 does"}
      ]
    }
  ],
  "autostart": false
}

### Fields explained:
- **description** (required): one-line text shown in the intellisense dropdown when the user types ":"
- **shell** (required): filename of the script to execute, relative to this directory
- **mode** (required): either "status" or "exec"
  - "status": captures stdout and displays the first line in tracer's status bar (bottom of screen). Use for quick info commands (git status, file counts, etc.)
  - "exec": hands off the full terminal to the script (like opening vim or running an interactive program). Use for interactive commands.
- **args**: optional array of argument definitions. Each arg has:
  - "name": displayed in help text
  - "required": whether the command fails without this arg
  - "completions": array of {value, description} pairs shown in the intellisense dropdown when the user is typing this argument. Descriptions are optional but helpful.
- **autostart**: if true, this command runs automatically when tracer launches. Only works with mode "status". Useful for commands that populate the status bar on startup.

## Environment variables available to the script:
- TRACER_CMD_DIR: absolute path to this command's directory (useful for accessing bundled data files)
- Positional parameters ($1, $2, ...): the arguments the user typed after the command name

## Guidelines:
- For "status" mode: output exactly ONE line to stdout. It appears in the status bar and auto-clears after 3 seconds.
- For "exec" mode: you have the full terminal. When the script exits, tracer resumes.
- Make run.sh executable (chmod +x).
- If the command needs configuration or data files, put them in this directory alongside the script.
- Keep the description concise — it's shown in a dropdown with limited width.

## Your task:
1. First, ask the user what this command should do — what behavior they want, what output, what arguments.
2. Edit command.json with appropriate metadata, description, mode, args, and completions.
3. Write run.sh implementing the command.
4. chmod +x run.sh
5. Test the script by running it directly to make sure it works.

Current directory: %s`, name, name, cmdDir)
}
