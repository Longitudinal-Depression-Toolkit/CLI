package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/theme"
)

const DefaultHeaderPanelWidth = 132

func PrepareActionScreen(title string, subtitle string, inNavigator func() bool) {
	if inNavigator != nil && inNavigator() {
		ClearActionScreen()
	}
	header := components.RenderScreenHeader(DefaultHeaderPanelWidth)

	lines := []string{theme.App.SubSubtitleStyle().Render(strings.TrimSpace(title))}
	if strings.TrimSpace(subtitle) != "" {
		lines = append(lines, theme.App.MutedTextStyle().Render(strings.TrimSpace(subtitle)))
	}
	lines = append(lines, theme.App.AccentStyle().Render("Exit tool: Esc"))
	panel := components.ApplyLeftLayoutMargin(
		theme.App.CompactPanelStyle().Render(
			lipgloss.JoinVertical(lipgloss.Left, lines...),
		),
	)

	fmt.Println(header)
	fmt.Println(panel)
	fmt.Println()
}

func ClearActionScreen() {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return
	}
	fmt.Print("\033[2J\033[H")
}
