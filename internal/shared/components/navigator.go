package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"ldt-toolkit-cli/internal/shared/theme"
)

type StagePanelInput struct {
	Heading         string
	HeadingRendered string

	PanelWidth       int
	BodyWidth        int
	TruncateOverflow bool

	Loading          bool
	LoadingIndicator string
	LoadingMessage   string

	LoadError string
	LoadHint  string

	NodeMissingHint string

	Summary string
	Usage   string

	IsDataPreparationLeaf bool
	RunTargetPath         string
	RunHint               string

	ListView    string
	NextPreview string
}

func RenderNavigatorFooter(activeTab int) string {
	base := "Tabs: left/right | Navigate: up/down + enter | Back: b | Quit: q"
	if activeTab == 0 {
		base += " | Home: n/p pages"
	} else {
		base += " | Exit tool: x"
	}
	base += " | Palette: : or Ctrl+P"
	return theme.App.MutedTextStyle().Padding(0, 1).Render(base)
}

func RenderStagePanel(input StagePanelInput) string {
	panelStyle := theme.App.PanelStyle()
	if input.PanelWidth > 0 {
		panelStyle = panelStyle.Width(input.PanelWidth)
	}

	bodyWidth := input.BodyWidth
	if bodyWidth <= 0 {
		bodyWidth = input.PanelWidth - 6
	}

	headingSource := strings.TrimSpace(input.HeadingRendered)
	if headingSource == "" {
		headingSource = theme.App.SubSubtitleStyle().Render(strings.TrimSpace(input.Heading))
	}
	heading := fitPanelLine(headingSource, bodyWidth, input.TruncateOverflow)

	if input.Loading {
		loadingMessage := strings.TrimSpace(input.LoadingMessage)
		if loadingMessage == "" {
			loadingMessage = "Loading..."
		}
		loadingLine := strings.TrimSpace(input.LoadingIndicator)
		if loadingLine != "" {
			loadingLine = lipgloss.JoinHorizontal(
				lipgloss.Left,
				loadingLine,
				" ",
				theme.App.TextStyle().Render(loadingMessage),
			)
		} else {
			loadingLine = theme.App.TextStyle().Render(loadingMessage)
		}
		loadingLine = fitPanelLine(loadingLine, bodyWidth, input.TruncateOverflow)
		return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, heading, "", loadingLine))
	}

	if strings.TrimSpace(input.LoadError) != "" {
		hint := strings.TrimSpace(input.LoadHint)
		if hint == "" {
			hint = "Could not load this node. Press l to retry or b to go back."
		}
		errorBlock := fitPanelBlock(
			theme.App.ErrorStyle().Render(input.LoadError),
			bodyWidth,
			input.TruncateOverflow,
		)
		return panelStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				heading,
				"",
				fitPanelLine(theme.App.MutedTextStyle().Render(hint), bodyWidth, input.TruncateOverflow),
				"",
				errorBlock,
			),
		)
	}

	if strings.TrimSpace(input.Summary) == "" && strings.TrimSpace(input.ListView) == "" {
		missingHint := strings.TrimSpace(input.NodeMissingHint)
		if missingHint == "" {
			missingHint = "Press l to load this stage node."
		}
		return panelStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				heading,
				"",
				fitPanelLine(missingHint, bodyWidth, input.TruncateOverflow),
			),
		)
	}

	parts := []string{heading}
	if strings.TrimSpace(input.Usage) != "" {
		parts = append(parts, fitPanelLine(
			theme.App.SubtitleStyle().Render(fmt.Sprintf("Usage: %s", strings.TrimSpace(input.Usage))),
			bodyWidth,
			input.TruncateOverflow,
		))
	}
	if strings.TrimSpace(input.Summary) != "" {
		parts = append(parts, fitPanelBlock(
			theme.App.TextStyle().Render(input.Summary),
			bodyWidth,
			input.TruncateOverflow,
		))
	}

	if strings.TrimSpace(input.ListView) != "" {
		parts = append(parts, "", fitPanelBlock(input.ListView, bodyWidth, input.TruncateOverflow))
		if strings.TrimSpace(input.NextPreview) != "" {
			parts = append(parts, fitPanelLine(
				theme.App.MutedTextStyle().Render(input.NextPreview),
				bodyWidth,
				input.TruncateOverflow,
			))
		}
		return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
	}

	if input.IsDataPreparationLeaf {
		return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
	}

	runTarget := strings.TrimSpace(input.RunTargetPath)
	if runTarget != "" {
		parts = append(parts, "", fitPanelLine(
			theme.App.TextStyle().Render(fmt.Sprintf("Ready to run: %s", runTarget)),
			bodyWidth,
			input.TruncateOverflow,
		))
	}
	runHint := strings.TrimSpace(input.RunHint)
	if runHint == "" {
		runHint = "Press Enter to run this action."
	}
	parts = append(parts, fitPanelLine(
		theme.App.AccentStyle().Render(runHint),
		bodyWidth,
		input.TruncateOverflow,
	))
	return panelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func fitPanelBlock(content string, width int, truncate bool) string {
	if !truncate || width <= 0 || strings.TrimSpace(content) == "" {
		return content
	}
	lines := strings.Split(content, "\n")
	for index, line := range lines {
		lines[index] = fitPanelLine(line, width, truncate)
	}
	return strings.Join(lines, "\n")
}

func fitPanelLine(content string, width int, truncate bool) string {
	if !truncate || width <= 0 {
		return content
	}
	return lipgloss.NewStyle().
		Inline(true).
		MaxWidth(width).
		Render(content)
}
