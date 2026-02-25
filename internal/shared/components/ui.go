package components

import (
	"fmt"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	liptable "github.com/charmbracelet/lipgloss/table"

	anim "ldt-toolkit-cli/internal/shared/asciimotion"
	"ldt-toolkit-cli/internal/shared/model"
	"ldt-toolkit-cli/internal/shared/theme"
)

const (
	LeftLayoutMargin   = 3
	TopLayoutMarginRow = 1
)

type TableRow struct {
	Title       string
	Description string
}

var (
	defaultHeaderBannerOnce sync.Once
	defaultHeaderBanner     string
)

func ApplyLeftLayoutMargin(content string) string {
	return lipgloss.NewStyle().MarginLeft(LeftLayoutMargin).Render(content)
}

func RenderScreenHeader(width int) string {
	header := trimHeaderPadding(RenderInteractiveHeader(width))
	if header == "" {
		return ""
	}

	rendered := ApplyLeftLayoutMargin(header)
	if TopLayoutMarginRow <= 0 {
		return rendered
	}
	return strings.Repeat("\n", TopLayoutMarginRow) + rendered
}

func trimHeaderPadding(content string) string {
	if strings.TrimSpace(content) == "" {
		return ""
	}

	lines := strings.Split(content, "\n")
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}

	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}

	return strings.Join(lines[start:end], "\n")
}

func RenderGeneralActionList(actions []model.CommandDef, selectedIndex int, width int) string {
	items := make([]string, 0, len(actions))
	for index, action := range actions {
		items = append(items, renderGeneralActionItem(action, index == selectedIndex, width))
	}
	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

func renderGeneralActionItem(action model.CommandDef, selected bool, width int) string {
	width = model.IntMax(10, width)
	prefix := "•"
	if selected {
		prefix = "›"
	}

	selectedBackground := theme.App.SelectedBackgroundColor()
	titleStyle := theme.App.SubtitleStyle().
		Width(width).
		Padding(0, 1)
	descStyle := theme.App.MutedTextStyle().
		Width(width).
		Padding(0, 1).
		PaddingLeft(3)

	if selected {
		titleStyle = titleStyle.Background(selectedBackground)
		descStyle = descStyle.Background(selectedBackground)
	}

	title := titleStyle.Render(fmt.Sprintf("%s %s", prefix, model.CommandLabel(action)))
	desc := descStyle.Render(action.Description)
	return lipgloss.JoinVertical(lipgloss.Left, title, desc)
}

func StyledSubActionDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(theme.App.Color("#DFA3C3")).
		BorderForeground(theme.App.Color("#6B7280")).
		Bold(true)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(theme.App.Color("#A1A1AA")).
		BorderForeground(theme.App.Color("#6B7280"))
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Foreground(theme.App.Color("#ECECF3"))
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(theme.App.Color("#A1A1AA"))
	return delegate
}

func NewSubActionList(title string, items []list.Item, delegate list.DefaultDelegate, width int, height int) list.Model {
	mdl := list.New(items, delegate, width, height)
	mdl.Title = title
	mdl.SetShowTitle(false)
	mdl.SetShowStatusBar(false)
	mdl.SetFilteringEnabled(false)
	mdl.Styles.Title = mdl.Styles.Title.
		Foreground(theme.App.Color("#DFA3C3")).
		Bold(true)
	mdl.Styles.PaginationStyle = mdl.Styles.PaginationStyle.Foreground(theme.App.Color("#A1A1AA"))
	mdl.Styles.HelpStyle = mdl.Styles.HelpStyle.Foreground(theme.App.Color("#A1A1AA"))
	return mdl
}

func RenderNavigationTabs(tabs []string, active int) string {
	baseBorder := lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
	}
	activeStyle := theme.App.TabActiveStyle(baseBorder)
	inactiveStyle := theme.App.TabInactiveStyle(baseBorder)

	rendered := make([]string, 0, len(tabs))
	for index, tab := range tabs {
		label := strings.TrimSpace(tab)
		if label == "" {
			label = "Unnamed"
		}
		if index == active {
			rendered = append(rendered, activeStyle.Render(label))
			continue
		}
		rendered = append(rendered, inactiveStyle.Render(label))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func RenderNameDescriptionTable(rows []TableRow) string {
	headerStyle := theme.App.SubSubtitleStyle().
		Foreground(theme.App.Color("#DFA3C3")).
		Bold(true).
		Padding(0, 1)
	cellStyle := theme.App.TextStyle().Padding(0, 1)
	mutedCellStyle := theme.App.MutedTextStyle().Padding(0, 1)

	tbl := liptable.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(theme.App.Color("#6B7280"))).
		StyleFunc(func(row int, col int) lipgloss.Style {
			switch {
			case row == liptable.HeaderRow:
				return headerStyle
			case col == 0:
				return cellStyle
			default:
				return mutedCellStyle
			}
		}).
		Headers("Technique", "Description")

	for _, row := range rows {
		title := strings.TrimSpace(row.Title)
		if title == "" {
			title = "Unnamed"
		}
		description := strings.TrimSpace(row.Description)
		if description == "" {
			description = "No description configured."
		}
		tbl = tbl.Row(title, description)
	}

	return tbl.String()
}

func RenderInteractiveHeader(width int) string {
	return RenderInteractiveHeaderWithBanner(width, "")
}

func RenderInteractiveHeaderWithBanner(width int, banner string) string {
	_ = width

	bannerArt := banner
	hasCustomBanner := strings.TrimSpace(bannerArt) != ""
	if !hasCustomBanner {
		bannerArt = defaultInteractiveHeaderBanner()
	}
	return bannerArt
}

func defaultInteractiveHeaderBanner() string {
	defaultHeaderBannerOnce.Do(func() {
		model := anim.New(anim.Config{
			AutoPlay:          false,
			Loop:              false,
			HasDarkBackground: true,
			Width:             80,
			Height:            24,
		})
		total := model.TotalFrames()
		if total > 0 {
			model.SetFrame(total - 1)
		}
		defaultHeaderBanner = model.View()
	})
	return defaultHeaderBanner
}
