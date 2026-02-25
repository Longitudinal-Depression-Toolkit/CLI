package theme

import (
	"strings"

	"github.com/charmbracelet/glamour"
	glamouransi "github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/lipgloss"
)

const (
	colorPink       = "#DFA3C3"
	colorDeepPurple = "#6B7280"
	colorLightText  = "#ECECF3"
	colorMutedText  = "#A1A1AA"
)

type UITheme struct {
	titleColor              string
	subtitleColor           string
	subSubtitleColor        string
	textColor               string
	mutedTextColor          string
	accentColor             string
	borderColor             string
	errorColor              string
	selectedBackgroundColor string
	tabActiveBackground     string
	markdownHeadingColor    string
	markdownSubHeadingColor string
	markdownLinkColor       string
}

var App = UITheme{
	titleColor:              colorPink,
	subtitleColor:           colorLightText,
	subSubtitleColor:        colorPink,
	textColor:               colorLightText,
	mutedTextColor:          colorMutedText,
	accentColor:             colorPink,
	borderColor:             colorDeepPurple,
	errorColor:              "#FCA5A5",
	selectedBackgroundColor: "#2F2530",
	tabActiveBackground:     "#3B2B36",
	markdownHeadingColor:    "213",
	markdownSubHeadingColor: "177",
	markdownLinkColor:       "213",
}

func (t UITheme) Color(value string) lipgloss.Color {
	return lipgloss.Color(strings.TrimSpace(value))
}

func (t UITheme) TitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Color(t.titleColor)).
		Bold(true)
}

func (t UITheme) SubtitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Color(t.subtitleColor)).
		Bold(true)
}

func (t UITheme) SubSubtitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Color(t.subSubtitleColor)).
		Bold(true)
}

func (t UITheme) TextStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Color(t.textColor))
}

func (t UITheme) MutedTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Color(t.mutedTextColor))
}

func (t UITheme) AccentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.Color(t.accentColor)).
		Bold(true)
}

func (t UITheme) ErrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.Color(t.errorColor))
}

func (t UITheme) SelectedBackgroundColor() lipgloss.Color {
	return t.Color(t.selectedBackgroundColor)
}

func (t UITheme) PanelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Color(t.borderColor)).
		Padding(1, 2)
}

func (t UITheme) CompactPanelStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Color(t.borderColor)).
		Padding(0, 1)
}

func (t UITheme) TabActiveStyle(baseBorder lipgloss.Border) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(baseBorder, true).
		BorderForeground(t.Color("#B47B9A")).
		Foreground(t.Color(t.textColor)).
		Background(t.Color(t.tabActiveBackground)).
		Bold(true).
		Padding(0, 2).
		MarginRight(1)
}

func (t UITheme) TabInactiveStyle(baseBorder lipgloss.Border) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(baseBorder, true).
		BorderForeground(t.Color("#4B5563")).
		Foreground(t.Color(t.mutedTextColor)).
		Padding(0, 2).
		MarginRight(1)
}

func (t UITheme) MarkdownStyle() glamouransi.StyleConfig {
	style := glamour.DarkStyleConfig
	style.Heading.Color = stringRef(t.markdownHeadingColor)
	style.H2.Color = stringRef(t.markdownHeadingColor)
	style.H3.Color = stringRef(t.markdownSubHeadingColor)
	style.H4.Color = stringRef(t.markdownSubHeadingColor)
	style.H5.Color = stringRef(t.markdownSubHeadingColor)
	style.H6.Color = stringRef(t.markdownSubHeadingColor)
	style.Link.Color = stringRef(t.markdownLinkColor)
	style.LinkText.Color = stringRef(t.markdownLinkColor)
	style.LinkText.Bold = boolRef(true)
	return style
}

func stringRef(value string) *string {
	v := strings.TrimSpace(value)
	return &v
}

func boolRef(value bool) *bool {
	v := value
	return &v
}
