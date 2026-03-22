package ui

import "charm.land/lipgloss/v2"

var (
	purple    = lipgloss.Color("#7D56F4")
	gray      = lipgloss.Color("#626262")
	lightGray = lipgloss.Color("#ADADAD")
	white     = lipgloss.Color("#FAFAFA")
	red       = lipgloss.Color("#FF4444")
	green     = lipgloss.Color("#44FF44")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(white).
			Background(purple).
			Padding(0, 1)

	countStyle = lipgloss.NewStyle().
			Foreground(lightGray)

	helpStyle = lipgloss.NewStyle().
			Foreground(gray)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(purple).
			Width(14)

	valueStyle = lipgloss.NewStyle().
			Foreground(white)

	userStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(green)

	assistantStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(purple)

	dimmedStyle = lipgloss.NewStyle().
			Foreground(gray)

	deletePromptStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(red)

	filterStyle = lipgloss.NewStyle().
			Foreground(purple)

	helpKeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(purple)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(gray)

	helpSepStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444"))

)
