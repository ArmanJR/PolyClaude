package tui

import "charm.land/lipgloss/v2"

var (
	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("170"))
	subtitleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	successStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	mutedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	highlightStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	boldStyle      = lipgloss.NewStyle().Bold(true)
	codeStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("236")).Padding(0, 1)
	warnStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
)
