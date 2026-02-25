package components

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ldt-toolkit-cli/internal/shared/model"
	"ldt-toolkit-cli/internal/shared/theme"
)

var ErrNumberStepperCancelled = errors.New("number stepper cancelled")

type NumberStepperConfig struct {
	Title       string
	Description string
	Initial     float64
	Step        float64
	Min         *float64
	Max         *float64
	Precision   int
	InputWidth  int
}

type numberStepperOutcome struct {
	value     float64
	cancelled bool
}

type numberStepperModel struct {
	config  NumberStepperConfig
	value   float64
	input   textinput.Model
	width   int
	error   string
	outcome numberStepperOutcome
}

func PromptNumberStepper(config NumberStepperConfig) (float64, error) {
	normalized := normalizeNumberStepperConfig(config)
	start := clampNumber(normalized.Initial, normalized.Min, normalized.Max)

	input := textinput.New()
	input.Prompt = "⌁ "
	input.SetValue(formatStepperValue(start, normalized.Precision))
	input.CharLimit = 64
	input.Width = normalized.InputWidth
	input.TextStyle = theme.App.TextStyle().Bold(true)
	input.PromptStyle = theme.App.AccentStyle().Bold(true)
	input.PlaceholderStyle = theme.App.MutedTextStyle().Bold(true)
	input.Cursor.Style = theme.App.AccentStyle().Bold(true)
	input.Focus()

	component := numberStepperModel{
		config: normalized,
		value:  start,
		input:  input,
	}

	finalModel, err := tea.NewProgram(component).Run()
	clearTerminalScreen()
	if err != nil {
		return 0, err
	}

	resolved, ok := finalModel.(numberStepperModel)
	if !ok {
		return 0, errors.New("unexpected number stepper model type")
	}
	if resolved.outcome.cancelled {
		return 0, ErrNumberStepperCancelled
	}

	return resolved.outcome.value, nil
}

func PromptIntStepper(config NumberStepperConfig) (int, error) {
	normalized := normalizeNumberStepperConfig(config)
	normalized.Precision = 0
	value, err := PromptNumberStepper(normalized)
	if err != nil {
		return 0, err
	}
	return int(math.Round(value)), nil
}

func (m numberStepperModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m numberStepperModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.input.Width = model.IntMax(12, model.IntMin(32, typed.Width-40))
		return m, nil

	case tea.KeyMsg:
		switch typed.String() {
		case "esc", "ctrl+c":
			m.outcome.cancelled = true
			return m, tea.Sequence(tea.ClearScreen, tea.Quit)
		case "up", "k":
			m.stepBy(m.config.Step)
			return m, nil
		case "down", "j":
			m.stepBy(-m.config.Step)
			return m, nil
		case "enter":
			parsed, err := m.parseCurrentValue()
			if err != nil {
				m.error = err.Error()
				return m, nil
			}
			m.outcome.value = parsed
			return m, tea.Sequence(tea.ClearScreen, tea.Quit)
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.error != "" {
		m.error = ""
	}
	return m, cmd
}

func (m numberStepperModel) View() string {
	title := strings.TrimSpace(m.config.Title)
	if title == "" {
		title = "Select number"
	}

	description := strings.TrimSpace(m.config.Description)
	if description == "" {
		description = "Use up/down arrows to increase or decrease the value."
	}

	info := []string{
		fmt.Sprintf("Step %s", formatStepperValue(m.config.Step, m.config.Precision)),
	}
	if m.config.Min != nil {
		info = append(info, fmt.Sprintf("Min %s", formatStepperValue(*m.config.Min, m.config.Precision)))
	}
	if m.config.Max != nil {
		info = append(info, fmt.Sprintf("Max %s", formatStepperValue(*m.config.Max, m.config.Precision)))
	}

	rows := []string{
		theme.App.SubSubtitleStyle().Render(title),
		theme.App.MutedTextStyle().Render(description),
		theme.App.MutedTextStyle().Render(strings.Join(info, " | ")),
		"",
		lipgloss.NewStyle().Padding(1, 1).Bold(true).Render(m.input.View()),
		"",
		theme.App.MutedTextStyle().Render("↑/↓ adjust | Enter confirm | Esc cancel"),
	}
	if strings.TrimSpace(m.error) != "" {
		rows = append(rows, theme.App.ErrorStyle().Render(m.error))
	}

	panelWidth := model.IntMax(58, m.input.Width+18)
	if m.width > 0 {
		panelWidth = model.IntMin(panelWidth, model.IntMax(58, m.width-12))
	}

	return ApplyLeftLayoutMargin(
		lipgloss.NewStyle().Width(panelWidth).Render(
			lipgloss.JoinVertical(lipgloss.Left, rows...),
		),
	)
}

func (m *numberStepperModel) stepBy(delta float64) {
	current, err := m.parseCurrentValue()
	if err == nil {
		m.value = current
	}

	next := m.value + delta
	next = clampNumber(next, m.config.Min, m.config.Max)
	m.value = next
	m.input.SetValue(formatStepperValue(next, m.config.Precision))
	m.error = ""
}

func (m numberStepperModel) parseCurrentValue() (float64, error) {
	raw := strings.TrimSpace(m.input.Value())
	if raw == "" {
		return m.value, nil
	}

	parsed, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, errors.New("enter a numeric value")
	}

	if m.config.Min != nil && parsed < *m.config.Min {
		return 0, fmt.Errorf("value must be >= %s", formatStepperValue(*m.config.Min, m.config.Precision))
	}
	if m.config.Max != nil && parsed > *m.config.Max {
		return 0, fmt.Errorf("value must be <= %s", formatStepperValue(*m.config.Max, m.config.Precision))
	}
	return parsed, nil
}

func normalizeNumberStepperConfig(config NumberStepperConfig) NumberStepperConfig {
	normalized := config
	if normalized.Step == 0 {
		normalized.Step = 1
	}
	if normalized.Step < 0 {
		normalized.Step = math.Abs(normalized.Step)
	}
	if normalized.Precision < 0 {
		normalized.Precision = 0
	}
	if normalized.InputWidth <= 0 {
		normalized.InputWidth = 16
	}
	if normalized.Precision == 0 {
		normalized.Precision = inferPrecision(normalized.Step)
	}
	return normalized
}

func inferPrecision(step float64) int {
	if step == math.Trunc(step) {
		return 0
	}
	formatted := strconv.FormatFloat(step, 'f', -1, 64)
	parts := strings.SplitN(formatted, ".", 2)
	if len(parts) < 2 {
		return 0
	}
	trimmed := strings.TrimRight(parts[1], "0")
	precision := len(trimmed)
	if precision < 0 {
		return 0
	}
	if precision > 6 {
		return 6
	}
	return precision
}

func clampNumber(value float64, min *float64, max *float64) float64 {
	clamped := value
	if min != nil && clamped < *min {
		clamped = *min
	}
	if max != nil && clamped > *max {
		clamped = *max
	}
	return clamped
}

func formatStepperValue(value float64, precision int) string {
	if precision <= 0 {
		if value == math.Trunc(value) {
			return strconv.FormatInt(int64(value), 10)
		}
		return strconv.FormatFloat(value, 'f', -1, 64)
	}

	formatted := strconv.FormatFloat(value, 'f', precision, 64)
	formatted = strings.TrimRight(formatted, "0")
	formatted = strings.TrimRight(formatted, ".")
	if formatted == "" || formatted == "-" {
		return "0"
	}
	return formatted
}
