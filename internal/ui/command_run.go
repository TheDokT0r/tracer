package ui

import (
	"fmt"
	"strconv"
	"strings"

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
					_, cmd := a.startSessionInDir(args[0])
					return cmd
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
				Options: func(a *App) []string { return []string{"html", "md"} },
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
				Options: func(a *App) []string { return []string{"date", "name", "directory"} },
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
					Options: func(a *App) []string {
						return []string{
							"theme", "sort_by", "show_date", "show_directory",
							"show_branch", "confirm_delete", "auto_update",
							"cmd_dropdown", "cmd_ghost", "cmd_max_suggestions",
						}
					},
				},
				{
					Name: "value", Required: true,
					Options: func(a *App) []string {
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
							return ThemeNames()
						case "sort_by":
							return []string{"date", "name", "directory"}
						case "show_date", "show_directory", "show_branch",
							"confirm_delete", "auto_update", "cmd_dropdown", "cmd_ghost":
							return []string{"on", "off"}
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
				Options: func(a *App) []string { return ThemeNames() },
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
				Options: func(a *App) []string { return []string{"sessions", "skills", "permissions"} },
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
