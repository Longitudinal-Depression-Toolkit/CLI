package components

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"ldt-toolkit-cli/internal/shared/theme"
)

var ErrPathPickerCancelled = errors.New("path picker cancelled")

type PathPickerMode int

const (
	PathPickerFile PathPickerMode = iota
	PathPickerDirectory
)

type confirmPathOutcome struct {
	path      string
	repick    bool
	cancelled bool
}

type confirmPathModel struct {
	title       string
	description string

	input textinput.Model
	width int
	mode  PathPickerMode

	errorText string
	outcome   confirmPathOutcome
}

func PickPathWithFilePicker(
	title string,
	subtitle string,
	mode PathPickerMode,
	initialPath string,
) (string, error) {
	currentPath := strings.TrimSpace(initialPath)
	currentDirectory := resolvePickerStartDirectory(currentPath)
	for {
		selectedPath := currentPath
		picker := huh.NewFilePicker().
			Title(defaultPathPickerTitle(title, mode)).
			Description(defaultPathPickerSubtitle(subtitle)).
			CurrentDirectory(currentDirectory).
			Value(&selectedPath).
			ShowHidden(false).
			ShowPermissions(false).
			ShowSize(false).
			Height(14).
			Picking(true).
			Validate(func(value string) error {
				return validatePickerSelection(value, mode)
			})

		if mode == PathPickerDirectory {
			picker = picker.
				DirAllowed(true).
				FileAllowed(true)
		} else {
			picker = picker.
				DirAllowed(false).
				FileAllowed(true)
		}

		form := NewLDTForm(
			huh.NewGroup(picker),
		)
		if err := form.Run(); err != nil {
			clearTerminalScreen()
			if errors.Is(err, huh.ErrUserAborted) {
				return "", ErrPathPickerCancelled
			}
			return "", err
		}
		clearTerminalScreen()

		finalPath, err := resolvePickedPath(selectedPath, mode)
		if err != nil {
			// Keep the last attempted path as the next picker starting location.
			currentPath = strings.TrimSpace(selectedPath)
			currentDirectory = resolvePickerStartDirectory(currentPath)
			continue
		}

		confirmedPath, repick, err := ConfirmPickedPath(
			defaultPathPickerTitle(title, mode),
			"Selected path. Press Enter to continue, edit if needed. Ctrl+R to reopen picker. Esc to cancel.",
			finalPath,
			mode,
		)
		if err != nil {
			return "", err
		}
		if repick {
			currentDirectory = resolvePickerStartDirectory(finalPath)
			currentPath = ""
			continue
		}
		return strings.TrimSpace(confirmedPath), nil
	}
}

func ConfirmPickedPath(
	title string,
	description string,
	initialPath string,
	mode PathPickerMode,
) (string, bool, error) {
	input := textinput.New()
	input.Prompt = "⌁ "
	input.SetValue(strings.TrimSpace(initialPath))
	input.CharLimit = 4096
	input.Width = 76
	input.TextStyle = theme.App.TextStyle().Bold(true)
	input.PromptStyle = theme.App.AccentStyle().Bold(true)
	input.PlaceholderStyle = theme.App.MutedTextStyle().Bold(true)
	input.Cursor.Style = theme.App.AccentStyle().Bold(true)
	input.Focus()

	model := confirmPathModel{
		title:       strings.TrimSpace(title),
		description: strings.TrimSpace(description),
		input:       input,
		mode:        mode,
	}
	finalModel, err := tea.NewProgram(model).Run()
	clearTerminalScreen()
	if err != nil {
		return "", false, err
	}

	resolved, ok := finalModel.(confirmPathModel)
	if !ok {
		return "", false, errors.New("unexpected confirm path model type")
	}
	if resolved.outcome.cancelled {
		return "", false, ErrPathPickerCancelled
	}
	if resolved.outcome.repick {
		return "", true, nil
	}
	return strings.TrimSpace(resolved.outcome.path), false, nil
}

func resolvePickerStartDirectory(initialPath string) string {
	candidate := strings.TrimSpace(initialPath)
	if candidate != "" {
		expanded := filepath.Clean(os.ExpandEnv(candidate))
		if filepath.IsAbs(expanded) {
			if info, err := os.Stat(expanded); err == nil {
				if info.IsDir() {
					return expanded
				}
				return filepath.Dir(expanded)
			}
		}
	}

	if cwd, err := os.Getwd(); err == nil && strings.TrimSpace(cwd) != "" {
		return cwd
	}
	if home, err := os.UserHomeDir(); err == nil && strings.TrimSpace(home) != "" {
		return home
	}
	return "."
}

func defaultPathPickerTitle(title string, mode PathPickerMode) string {
	trimmed := strings.TrimSpace(title)
	if trimmed != "" {
		return trimmed
	}
	if mode == PathPickerDirectory {
		return "Select folder path"
	}
	return "Select file path"
}

func defaultPathPickerSubtitle(subtitle string) string {
	trimmed := strings.TrimSpace(subtitle)
	nav := "← to go backward, → to open folder, Enter to choose, Esc to cancel."
	if trimmed == "" {
		return nav
	}
	return fmt.Sprintf("%s\n%s", trimmed, nav)
}

func validatePickerSelection(value string, mode PathPickerMode) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return errors.New("select a path")
	}

	info, err := os.Stat(trimmed)
	if err != nil {
		return fmt.Errorf("path not found: %s", trimmed)
	}

	// In directory mode we accept both folders and files while browsing.
	// If a file is selected we resolve to its parent directory after submit.
	if mode == PathPickerDirectory {
		_ = info
		return nil
	}

	if info.IsDir() {
		return errors.New("select a file path")
	}
	return nil
}

func (m confirmPathModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m confirmPathModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.input.Width = maxInt(40, minInt(100, typed.Width-28))
		return m, nil

	case tea.KeyMsg:
		switch typed.String() {
		case "esc":
			m.outcome.cancelled = true
			return m, tea.Sequence(tea.ClearScreen, tea.Quit)
		case "ctrl+r":
			m.outcome.repick = true
			return m, tea.Sequence(tea.ClearScreen, tea.Quit)
		case "enter":
			value := strings.TrimSpace(m.input.Value())
			resolvedPath, err := resolvePickedPath(value, m.mode)
			if err != nil {
				m.errorText = err.Error()
				return m, nil
			}
			m.outcome.path = resolvedPath
			return m, tea.Sequence(tea.ClearScreen, tea.Quit)
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.errorText != "" {
		m.errorText = ""
	}
	return m, cmd
}

func (m confirmPathModel) View() string {
	title := strings.TrimSpace(m.title)
	if title == "" {
		title = "Confirm selected path"
	}

	description := strings.TrimSpace(m.description)
	if description == "" {
		description = "Review and edit the selected path if needed."
	}

	rows := []string{
		theme.App.SubSubtitleStyle().Render(title),
		theme.App.MutedTextStyle().Render(description),
		"",
		lipgloss.NewStyle().Padding(1, 1).Bold(true).Render(m.input.View()),
		"",
		theme.App.MutedTextStyle().Render("Enter continue | Ctrl+R reopen picker | Esc cancel"),
	}
	if strings.TrimSpace(m.errorText) != "" {
		rows = append(rows, theme.App.ErrorStyle().Render(m.errorText))
	}

	panelWidth := maxInt(58, m.input.Width+8)
	if m.width > 0 {
		panelWidth = minInt(panelWidth, maxInt(58, m.width-12))
	}

	return ApplyLeftLayoutMargin(
		lipgloss.NewStyle().Width(panelWidth).Render(
			lipgloss.JoinVertical(lipgloss.Left, rows...),
		),
	)
}

func resolvePickedPath(value string, mode PathPickerMode) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", errors.New("select a path")
	}

	info, err := os.Stat(trimmed)
	if err != nil {
		return "", fmt.Errorf("path not found: %s", trimmed)
	}

	if mode == PathPickerDirectory {
		if info.IsDir() {
			return trimmed, nil
		}
		parent := filepath.Dir(trimmed)
		parentInfo, parentErr := os.Stat(parent)
		if parentErr != nil || !parentInfo.IsDir() {
			return "", errors.New("select a folder path")
		}
		return parent, nil
	}

	if info.IsDir() {
		return "", errors.New("select a file path")
	}
	return trimmed, nil
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
