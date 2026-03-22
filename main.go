package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	tea "charm.land/bubbletea/v2"
	"tracer/internal/claude"
	"tracer/internal/config"
	"tracer/internal/skills"
	"tracer/internal/ui"
	"tracer/internal/updater"
)

//go:embed tracer.1
var manPage []byte

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "jungle":
			fmt.Println("Wondrous is our great blue ship")
			fmt.Println("That sails around the mighty sun")
			fmt.Println("And joy to everyone that rides along!")
			os.Exit(0)
		case "--version", "-v":
			fmt.Println("tracer", version)
			os.Exit(0)
		case "--help", "-h":
			fmt.Println("Usage: tracer [command]")
			fmt.Println()
			fmt.Println("A TUI for managing Claude Code sessions.")
			fmt.Println()
			fmt.Println("Commands:")
			fmt.Println("  update      Update tracer to the latest version")
			fmt.Println("  theme       View or set the color theme")
			fmt.Println("  settings    Open settings")
			fmt.Println("  man         View the manual page")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  -v, --version  Print version")
			fmt.Println("  -h, --help     Show this help")
			fmt.Println()
			fmt.Println("Run tracer with no arguments to launch the TUI.")
			os.Exit(0)
		case "update":
			if err := updater.Update(version); err != nil {
				fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		case "settings":
			cfg := config.LoadConfig()
			if t, ok := ui.Themes[cfg.Theme]; ok {
				ui.ApplyTheme(t)
			}
			sa := ui.NewSettingsApp(cfg)
			p := tea.NewProgram(sa)
			result, err := p.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if app := result.(ui.SettingsApp); app.Saved() {
				config.SaveConfig(app.Config())
				fmt.Println("Settings saved.")
			}
			os.Exit(0)
		case "theme":
			cfg := config.LoadConfig()
			if len(os.Args) > 2 {
				name := os.Args[2]
				if _, ok := ui.Themes[name]; !ok {
					fmt.Fprintf(os.Stderr, "Unknown theme: %s\n", name)
					fmt.Fprintf(os.Stderr, "Available: %s\n", strings.Join(ui.ThemeNames(), ", "))
					os.Exit(1)
				}
				cfg.Theme = name
				config.SaveConfig(cfg)
				fmt.Printf("Theme set to %s\n", name)
			} else {
				// Apply current theme before launching picker
				if t, ok := ui.Themes[cfg.Theme]; ok {
					ui.ApplyTheme(t)
				}
				picker := ui.NewThemePicker(cfg.Theme)
				p := tea.NewProgram(picker)
				result, err := p.Run()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				if chosen := result.(ui.ThemePicker).Chosen(); chosen != "" {
					cfg.Theme = chosen
					config.SaveConfig(cfg)
					fmt.Printf("Theme set to %s\n", chosen)
				}
			}
			os.Exit(0)
		case "man":
			tmp, err := os.CreateTemp("", "tracer-*.1")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			defer os.Remove(tmp.Name())
			tmp.Write(manPage)
			tmp.Close()
			cmd := exec.Command("man", tmp.Name())
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
			os.Exit(0)
		}
	}

	// Apply theme
	cfg := config.LoadConfig()
	if t, ok := ui.Themes[cfg.Theme]; ok {
		ui.ApplyTheme(t)
	}

	// Auto-update check (skip for Homebrew installs)
	if cfg.AutoUpdate && version != "dev" && !updater.IsHomebrew() {
		latest, err := updater.Check()
		if err == nil && updater.NeedsUpdate(version, latest) {
			fmt.Printf("Updating tracer %s -> %s...\n", version, latest)
			if err := updater.Update(version); err != nil {
				fmt.Fprintf(os.Stderr, "Auto-update failed: %v\n", err)
			} else {
				exe, err := os.Executable()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Updated successfully. Please restart tracer manually.\n")
				} else if err := syscall.Exec(exe, os.Args, os.Environ()); err != nil {
					fmt.Fprintf(os.Stderr, "Updated successfully. Please restart tracer manually.\n")
				}
			}
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	claudeDir := filepath.Join(home, ".claude")

	sessions, err := claude.ScanSessions(claudeDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning sessions: %v\n", err)
		os.Exit(1)
	}

	if len(sessions) == 0 {
		fmt.Println("No Claude Code sessions found.")
		os.Exit(0)
	}

	pins := config.LoadPins()
	renames := config.LoadRenames()

	// Apply tracer renames on top of Claude /rename
	for i := range sessions {
		if name, ok := renames[sessions[i].ID]; ok {
			sessions[i].Name = name
		}
	}

	allSkills, _ := skills.ScanSkills(claudeDir)

	app := ui.NewApp(claudeDir, sessions, pins, cfg, renames, allSkills)
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
