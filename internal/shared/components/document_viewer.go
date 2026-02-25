package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"ldt-toolkit-cli/internal/shared/model"
	"ldt-toolkit-cli/internal/shared/theme"
)

var (
	docViewerTitleStyle = theme.App.TitleStyle().
				Background(theme.App.SelectedBackgroundColor()).
				Bold(true).
				Padding(0, 1)
	docViewerInfoStyle = theme.App.AccentStyle()
	docViewerHintStyle = theme.App.MutedTextStyle().
				Italic(true)
	docViewerPanelStyle = theme.App.CompactPanelStyle()
)

type documentViewerModel struct {
	title    string
	markdown string

	ready    bool
	viewport viewport.Model
	width    int
	height   int
}

type markdownDocumentComponent struct {
	Title    string
	Markdown string
}

func newDocumentViewerModel(title string, markdown string) documentViewerModel {
	return documentViewerModel{
		title:    strings.TrimSpace(title),
		markdown: strings.TrimSpace(markdown),
	}
}

func (m documentViewerModel) Init() tea.Cmd {
	return nil
}

func (m documentViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch message := msg.(type) {
	case tea.KeyMsg:
		switch message.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = message.Width
		m.height = message.Height
		m.resize()
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m documentViewerModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	header := m.headerView()
	footer := m.footerView()
	body := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.viewport.View(),
		footer,
	)

	panel := docViewerPanelStyle.Render(body)

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		panel,
	)
}

func (m *documentViewerModel) resize() {
	if m.width <= 0 || m.height <= 0 {
		return
	}

	panelWidth := model.IntMin(112, model.IntMax(58, m.width-16))
	maxPanelHeight := model.IntMax(10, m.height-6)
	panelHeight := model.IntMin(34, maxPanelHeight)

	headerHeight := lipgloss.Height(m.headerViewForWidth(panelWidth))
	footerHeight := lipgloss.Height(m.footerViewForWidth(panelWidth, 0))
	viewportHeight := model.IntMax(5, panelHeight-headerHeight-footerHeight)

	if !m.ready {
		m.viewport = viewport.New(panelWidth, viewportHeight)
		m.viewport.MouseWheelEnabled = true
		m.ready = true
	} else {
		m.viewport.Width = panelWidth
		m.viewport.Height = viewportHeight
	}

	m.viewport.SetContent(renderMarkdownForViewport(m.markdown, panelWidth))
}

func (m documentViewerModel) headerView() string {
	return m.headerViewForWidth(m.viewport.Width)
}

func (m documentViewerModel) headerViewForWidth(width int) string {
	title := docViewerTitleStyle.Render(m.title)
	if lipgloss.Width(title) >= width {
		return title
	}
	return title + strings.Repeat(" ", width-lipgloss.Width(title))
}

func (m documentViewerModel) footerView() string {
	return m.footerViewForWidth(m.viewport.Width, m.viewport.ScrollPercent())
}

func (m documentViewerModel) footerViewForWidth(width int, scrollPercent float64) string {
	info := docViewerInfoStyle.Render(fmt.Sprintf("%3.f%%  q close", scrollPercent*100))
	hint := docViewerHintStyle.Render("↑/↓ scroll")
	infoWidth := lipgloss.Width(info)
	hintWidth := lipgloss.Width(hint)

	if infoWidth >= width {
		return info
	}
	if hintWidth+infoWidth+1 >= width {
		return strings.Repeat(" ", width-infoWidth) + info
	}

	gap := strings.Repeat(" ", width-hintWidth-infoWidth)
	return hint + gap + info
}

func renderMarkdownForViewport(markdown string, width int) string {
	style := theme.App.MarkdownStyle()

	renderer, err := glamour.NewTermRenderer(
		glamour.WithStyles(style),
		glamour.WithWordWrap(model.IntMax(32, width-2)),
	)
	if err != nil {
		return markdown
	}

	rendered, err := renderer.Render(markdown)
	if err != nil {
		return markdown
	}
	return strings.TrimSpace(rendered)
}

func runDocumentViewer(title string, markdown string) error {
	program := tea.NewProgram(
		newDocumentViewerModel(title, markdown),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := program.Run(); err != nil {
		return fmt.Errorf("document viewer failed: %w", err)
	}
	return nil
}

func (c markdownDocumentComponent) Show() error {
	return runDocumentViewer(c.Title, c.Markdown)
}

func ShowMarkdownDocument(title string, markdown string) error {
	component := markdownDocumentComponent{
		Title:    strings.TrimSpace(title),
		Markdown: strings.TrimSpace(markdown),
	}
	return component.Show()
}
