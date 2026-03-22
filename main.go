package main

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"tracer/internal/claude"
	"tracer/internal/ui"
	"tracer/internal/updater"
)

//go:embed tracer.1
var manPage []byte

var version = "dev"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
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

	app := ui.NewApp(claudeDir, sessions)
	p := tea.NewProgram(app)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
