package ui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69")).MarginBottom(1)
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	activeTabStyle = lipgloss.NewStyle().Padding(0, 1).Bold(true).Foreground(lipgloss.Color("69")).Underline(true)
	tabStyle      = lipgloss.NewStyle().Padding(0, 1)
	usernameStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("135"))
	sepStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	overlayStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Width(60)
)
