package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	green = lipgloss.Color("#34C759")
	cyan  = lipgloss.Color("#5AC8FA")
	red   = lipgloss.Color("#FF3B30")
	white = lipgloss.Color("#F5F5F7")
	dim   = lipgloss.Color("#6E6E73")
	bgBar = lipgloss.Color("#2C2C2E")

	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(green)
	dimStyle     = lipgloss.NewStyle().Foreground(dim)
	cmdStyle     = lipgloss.NewStyle().Foreground(white).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(red)
	successStyle = lipgloss.NewStyle().Foreground(cyan)
	promptStyle  = lipgloss.NewStyle().Foreground(green).Bold(true)

	statusBarStyle = lipgloss.NewStyle().Foreground(dim).Background(bgBar).Padding(0, 1)
)

func (m model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(),
		m.viewport.View(),
		m.renderBottom(),
		m.renderStatusBar(),
	)
}

func (m model) renderHeader() string {
	return titleStyle.Render("◈ INTI") +
		dimStyle.Render(fmt.Sprintf("  voice: %s", m.currentVoice))
}

func (m model) renderBottom() string {
	switch m.state {
	case stateProcessing:
		return promptStyle.Render("CMD: ") + m.spinner.View() + dimStyle.Render(" processing...")
	case stateMenu:
		return m.cmdList.View()
	default:
		return promptStyle.Render("CMD: ") + m.input.View()
	}
}

func (m model) renderStatusBar() string {
	hint := "↑↓ scroll  |  enter for commands  |  ctrl+c quit"
	if m.state == stateMenu {
		hint = "↑↓ navigate  |  enter select  |  esc dismiss"
	}
	return statusBarStyle.Width(m.width).Render(
		fmt.Sprintf("%s  |  24kHz PCM-16  |  %s", m.currentModel, hint),
	)
}
