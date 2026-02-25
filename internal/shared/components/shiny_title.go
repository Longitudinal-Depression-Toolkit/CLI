package components

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ldt-toolkit-cli/internal/shared/theme"
)

const (
	defaultShinyTickEvery  = 45 * time.Millisecond
	defaultShinyTrailWidth = 3
	defaultShinyLeadPad    = 4
	defaultShinyTailPad    = 4
)

type shinyTitleTickMsg struct {
	cycleID uint64
}

// ShinyTitle renders a one-time left-to-right shine pass over a title.
type ShinyTitle struct {
	text string

	position int
	active   bool

	tickEvery  time.Duration
	trailWidth int
	leadPad    int
	tailPad    int

	cycleID uint64
}

func NewShinyTitle(text string) ShinyTitle {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		trimmed = "Title"
	}
	return ShinyTitle{
		text:       trimmed,
		position:   0,
		active:     false,
		tickEvery:  defaultShinyTickEvery,
		trailWidth: defaultShinyTrailWidth,
		leadPad:    defaultShinyLeadPad,
		tailPad:    defaultShinyTailPad,
		cycleID:    0,
	}
}

func (s *ShinyTitle) Reset() tea.Cmd {
	s.cycleID++
	s.position = -s.leadPad
	s.active = true
	return s.scheduleTick()
}

func (s *ShinyTitle) SetText(text string) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		trimmed = "Title"
	}
	s.text = trimmed
}

// Update returns whether the visual title state changed and a follow-up command.
func (s *ShinyTitle) Update(msg tea.Msg) (bool, tea.Cmd) {
	typed, ok := msg.(shinyTitleTickMsg)
	if !ok || typed.cycleID != s.cycleID || !s.active {
		return false, nil
	}

	s.position++
	if s.position > len([]rune(s.text))+s.tailPad {
		s.active = false
		return true, nil
	}
	return true, s.scheduleTick()
}

func (s ShinyTitle) View(baseStyle lipgloss.Style) string {
	return s.ViewWithText(baseStyle, s.text)
}

func (s ShinyTitle) ViewWithText(baseStyle lipgloss.Style, text string) string {
	title := strings.TrimSpace(text)
	if title == "" {
		title = strings.TrimSpace(s.text)
	}
	if title == "" {
		title = "Title"
	}

	if !s.active {
		return baseStyle.Render(title)
	}

	runes := []rune(title)
	rendered := make([]string, 0, len(runes))
	for index, char := range runes {
		dist := abs(index - s.position)
		style := baseStyle
		switch {
		case dist == 0:
			style = baseStyle.Copy().Foreground(theme.App.Color("#FFFFFF")).Bold(true)
		case dist == 1:
			style = baseStyle.Copy().Foreground(theme.App.Color("#FDE8F3")).Bold(true)
		case dist == 2:
			style = baseStyle.Copy().Foreground(theme.App.Color("#F3CADF")).Bold(true)
		case dist <= s.trailWidth:
			style = baseStyle.Copy().Foreground(theme.App.Color("#E9B8D4")).Bold(true)
		}
		rendered = append(rendered, style.Render(string(char)))
	}
	return strings.Join(rendered, "")
}

func (s ShinyTitle) scheduleTick() tea.Cmd {
	cycleID := s.cycleID
	return tea.Tick(s.tickEvery, func(time.Time) tea.Msg {
		return shinyTitleTickMsg{cycleID: cycleID}
	})
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
