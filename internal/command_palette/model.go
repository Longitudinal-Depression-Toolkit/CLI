package commandpalette

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/model"
	"ldt-toolkit-cli/internal/shared/theme"
)

const (
	inputMinWidth = 52
	inputMaxWidth = 96
)

type mode int

const (
	modeInput mode = iota
	modeHistory
)

type historyItem struct {
	value string
}

func (i historyItem) Title() string       { return i.value }
func (i historyItem) Description() string { return "" }
func (i historyItem) FilterValue() string { return i.value }

type Model struct {
	open      bool
	mode      mode
	input     textinput.Model
	history   list.Model
	errorText string
	width     int

	catalog  []Entry
	resolver resolver
	store    historyStore

	historyCommands []string
	exampleHint     components.RotatingHint
	titleShine      components.ShinyTitle
}

func New(projectRootFinder func() (string, error)) (Model, error) {
	catalog, err := buildCatalog()
	if err != nil {
		return Model{}, err
	}
	resolver := newResolver(catalog)
	examples := examplePool(catalog)
	exampleHint := components.NewRotatingHint(examples)
	titleShine := components.NewShinyTitle("Command Palette")

	input := textinput.New()
	input.Prompt = "⌁ "
	input.Placeholder = placeholderExample(exampleHint.Display())
	input.CharLimit = 220
	input.Width = 62
	input.TextStyle = theme.App.TextStyle().Bold(true)
	input.PromptStyle = theme.App.AccentStyle().Bold(true)
	input.PlaceholderStyle = theme.App.MutedTextStyle().Bold(true)
	input.Cursor.Style = theme.App.AccentStyle().Bold(true)
	input.ShowSuggestions = true
	input.SetSuggestions(suggestionsFromCatalog(catalog))
	input.Blur()

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(theme.App.Color("#DFA3C3")).
		BorderForeground(theme.App.Color("#6B7280")).
		Bold(true)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Foreground(theme.App.Color("#ECECF3"))

	historyList := list.New(nil, delegate, 62, 3)
	historyList.Title = ""
	historyList.SetShowTitle(false)
	historyList.SetShowFilter(false)
	historyList.SetFilteringEnabled(false)
	historyList.SetShowStatusBar(false)
	historyList.SetShowHelp(false)
	historyList.SetShowPagination(true)
	historyList.SetStatusBarItemName("command", "commands")
	historyList.Styles.PaginationStyle = historyList.Styles.PaginationStyle.Foreground(theme.App.Color("#A1A1AA"))
	historyList.Paginator.PerPage = 3
	historyList.InfiniteScrolling = true

	store := newHistoryStore(projectRootFinder)
	historyCommands := store.Load()

	palette := Model{
		open:            false,
		mode:            modeInput,
		input:           input,
		history:         historyList,
		catalog:         catalog,
		resolver:        resolver,
		store:           store,
		historyCommands: historyCommands,
		exampleHint:     exampleHint,
		titleShine:      titleShine,
	}
	palette.refreshHistoryList()
	return palette, nil
}

func (m Model) IsOpen() bool {
	return m.open
}

func (m *Model) Open() tea.Cmd {
	m.open = true
	m.mode = modeInput
	m.errorText = ""
	rotateCmd := m.exampleHint.Reset()
	titleCmd := m.titleShine.Reset()
	m.input.Placeholder = placeholderExample(m.exampleHint.Display())
	m.input.SetValue("")
	m.input.Focus()
	return tea.Batch(textinput.Blink, rotateCmd, titleCmd)
}

func (m *Model) Close() {
	m.open = false
	m.mode = modeInput
	m.errorText = ""
	m.input.Blur()
}

func (m *Model) Resize(width int) {
	m.width = width
	inputWidth := model.IntMin(inputMaxWidth, model.IntMax(inputMinWidth, width-34))
	m.input.Width = inputWidth
	m.history.SetSize(inputWidth, 3)
	m.syncHistoryPaginator()
}

func (m *Model) Update(msg tea.Msg) (tea.Cmd, *ResolvedCommand) {
	if !m.open {
		return nil, nil
	}

	hintCmd := m.updateHint(msg)
	titleCmd := m.updateTitle(msg)

	if key, ok := msg.(tea.KeyMsg); ok {
		if m.mode == modeHistory {
			cmd, resolved := m.updateHistory(key)
			return tea.Batch(cmd, hintCmd, titleCmd), resolved
		}
		cmd, resolved := m.updateInput(key)
		return tea.Batch(cmd, hintCmd, titleCmd), resolved
	}

	if m.mode == modeInput {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return tea.Batch(cmd, hintCmd, titleCmd), nil
	}

	if m.mode == modeHistory {
		var cmd tea.Cmd
		m.history, cmd = m.history.Update(msg)
		return tea.Batch(cmd, hintCmd, titleCmd), nil
	}

	return tea.Batch(hintCmd, titleCmd), nil
}

func (m *Model) updateHint(msg tea.Msg) tea.Cmd {
	changed, cmd := m.exampleHint.Update(msg)
	if changed {
		m.input.Placeholder = placeholderExample(m.exampleHint.Display())
	}
	return cmd
}

func (m *Model) updateTitle(msg tea.Msg) tea.Cmd {
	_, cmd := m.titleShine.Update(msg)
	return cmd
}

func (m *Model) updateInput(key tea.KeyMsg) (tea.Cmd, *ResolvedCommand) {
	switch key.String() {
	case "esc":
		m.Close()
		return nil, nil
	case "ctrl+l":
		m.input.SetValue("")
		m.errorText = ""
		return nil, nil
	case "ctrl+h":
		m.mode = modeHistory
		m.errorText = ""
		m.refreshHistoryList()
		return nil, nil
	case "enter":
		resolved, err := m.resolver.Resolve(m.input.Value())
		if err != nil {
			m.errorText = err.Error()
			return nil, nil
		}
		m.remember(resolved.Raw)
		m.Close()
		return nil, &resolved
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(key)
	if m.errorText != "" {
		m.errorText = ""
	}
	return cmd, nil
}

func (m *Model) updateHistory(key tea.KeyMsg) (tea.Cmd, *ResolvedCommand) {
	switch key.String() {
	case "b", "q", "esc", "ctrl+h":
		m.mode = modeInput
		m.errorText = ""
		return nil, nil
	case "enter":
		selected, ok := m.history.SelectedItem().(historyItem)
		if !ok {
			m.errorText = "select a command first"
			return nil, nil
		}
		resolved, err := m.resolver.Resolve(selected.value)
		if err != nil {
			m.errorText = err.Error()
			return nil, nil
		}
		m.remember(resolved.Raw)
		m.Close()
		return nil, &resolved
	}

	var cmd tea.Cmd
	m.history, cmd = m.history.Update(key)
	return cmd, nil
}

func (m *Model) remember(raw string) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return
	}

	next := make([]string, 0, len(m.historyCommands)+1)
	next = append(next, value)
	for _, existing := range m.historyCommands {
		if normalize(existing) == normalize(value) {
			continue
		}
		next = append(next, existing)
		if len(next) >= historyMaxItems {
			break
		}
	}

	m.historyCommands = next
	m.refreshHistoryList()
	_ = m.store.Save(m.historyCommands)
}

func (m *Model) refreshHistoryList() {
	items := make([]list.Item, 0, len(m.historyCommands))
	for _, command := range m.historyCommands {
		items = append(items, historyItem{value: command})
	}
	m.history.SetItems(items)
	m.syncHistoryPaginator()
	if len(items) == 0 {
		return
	}
	m.history.Select(0)
}

func (m *Model) syncHistoryPaginator() {
	const perPage = 3
	m.history.Paginator.PerPage = perPage

	itemCount := len(m.historyCommands)
	if itemCount <= 0 {
		m.history.Paginator.Page = 0
		m.history.Paginator.TotalPages = 1
		return
	}

	totalPages := m.history.Paginator.SetTotalPages(itemCount)
	if totalPages <= 0 {
		totalPages = 1
		m.history.Paginator.TotalPages = totalPages
	}
	if m.history.Paginator.Page >= totalPages {
		m.history.Paginator.Page = totalPages - 1
	}
	if m.history.Paginator.Page < 0 {
		m.history.Paginator.Page = 0
	}
}

func (m Model) View() string {
	label := m.titleShine.View(theme.App.SubSubtitleStyle())
	hint := theme.App.MutedTextStyle().Render("Run your fav. command here.")

	if m.mode == modeHistory {
		label = theme.App.SubSubtitleStyle().Render("Command Palette — History")
		hint = ""
	}

	errorLine := ""
	if strings.TrimSpace(m.errorText) != "" {
		errorLine = lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8")).Render(m.errorText)
	}

	rows := []string{label}
	if m.mode == modeHistory {
		rows = append(rows, "")
		if len(m.historyCommands) == 0 {
			rows = append(rows, theme.App.MutedTextStyle().Render("No command history yet."))
		} else {
			rows = append(rows, m.history.View())
		}
		rows = append(rows, theme.App.MutedTextStyle().Render("Enter run | b/q back"))
	} else {
		inputLine := lipgloss.NewStyle().
			Padding(1, 1).
			Width(m.input.Width + 2).
			Bold(true).
			Render(m.input.View())
		rows = append(rows, hint)
		rows = append(rows, "")
		rows = append(rows, inputLine)
		rows = append(rows, "")
		rows = append(rows, theme.App.MutedTextStyle().Render("Tab autocomplete | Ctrl+H history | Ctrl+L clear | Esc close"))
	}
	if errorLine != "" {
		rows = append(rows, errorLine)
	}

	panelWidth := model.IntMax(58, m.input.Width+8)
	if m.width > 0 {
		maxAllowedWidth := model.IntMax(58, m.width-14)
		panelWidth = model.IntMin(panelWidth, maxAllowedWidth)
	}
	panel := theme.App.PanelStyle().Width(panelWidth)
	return panel.Render(lipgloss.JoinVertical(lipgloss.Left, rows...))
}

func placeholderExample(command string) string {
	candidate := strings.TrimSpace(command)
	if candidate == "" {
		candidate = "data conversion"
	}
	return "Type here, e.g. " + candidate
}
