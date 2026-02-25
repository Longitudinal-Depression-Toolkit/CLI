package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/theme"
)

type returnCountdownModel struct {
	timer   timer.Model
	message string
}

func RunExitCountdown(message string, duration time.Duration, inNavigator func() bool) error {
	if inNavigator == nil || !inNavigator() {
		return nil
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) || !term.IsTerminal(int(os.Stdin.Fd())) {
		time.Sleep(duration)
		ClearActionScreen()
		return nil
	}

	model := returnCountdownModel{
		timer:   timer.NewWithInterval(duration, time.Second),
		message: strings.TrimSpace(message),
	}
	program := tea.NewProgram(model)
	if _, err := program.Run(); err != nil {
		return err
	}
	ClearActionScreen()
	return nil
}

func (m returnCountdownModel) Init() tea.Cmd {
	return m.timer.Init()
}

func (m returnCountdownModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(typed)
		return m, cmd
	case timer.TimeoutMsg:
		return m, tea.Quit
	case tea.KeyMsg:
		if typed.String() == "q" || typed.String() == "esc" || typed.String() == "ctrl+c" || typed.String() == "enter" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m returnCountdownModel) View() string {
	label := m.message
	if label == "" {
		label = "Returning to toolkit"
	}
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		theme.App.SubtitleStyle().Render(fmt.Sprintf("%s in %s", label, m.timer.View())),
		theme.App.MutedTextStyle().Render("Press q to return now."),
		"",
	)
	return components.ApplyLeftLayoutMargin(content)
}
