package ui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69")).MarginBottom(1)
	errStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	helpStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	tabsBarStyle    = lipgloss.NewStyle().Padding(0, 1).BorderStyle(lipgloss.NormalBorder()).BorderBottom(true).BorderForeground(lipgloss.Color("238"))
	activeTabStyle  = lipgloss.NewStyle().Padding(0, 1).Bold(true).Foreground(lipgloss.Color("81"))
	tabStyle        = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("252"))
	usernameStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("135"))
	sepStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	metaStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	messageStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81"))
	contextBarStyle = lipgloss.NewStyle().PaddingLeft(1).BorderStyle(lipgloss.NormalBorder()).BorderLeft(true).BorderForeground(lipgloss.Color("240")).MarginBottom(1)
	contextSepStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(0, 1)
	titleChipStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("111")).Padding(0, 1)
	metaChipStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Padding(0, 1)
	panelStyle      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("238")).Padding(0, 1)
	overlayStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Width(60)
)
