package home

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/model"
	"ldt-toolkit-cli/internal/shared/theme"
)

const (
	homeActionsPerPage = 3
	homeListMinWidth   = 44
	homeListMaxWidth   = 88
)

type Screen struct {
	actions        []model.CommandDef
	cursor         int
	page           int
	actionsPerPage int
}

func NewScreen(actions []model.CommandDef) Screen {
	screen := Screen{
		actions:        cloneCommandDefs(actions),
		cursor:         0,
		page:           0,
		actionsPerPage: homeActionsPerPage,
	}
	screen.ensureValidState()
	return screen
}

func (s *Screen) Update(msg tea.Msg) []string {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return nil
	}

	switch key.String() {
	case "enter":
		actions := s.visibleActions()
		if len(actions) == 0 || s.cursor < 0 || s.cursor >= len(actions) {
			return nil
		}
		return []string{actions[s.cursor].Name}

	case "j", "down":
		actions := s.visibleActions()
		if len(actions) > 0 {
			s.cursor = (s.cursor + 1) % len(actions)
		}

	case "k", "up":
		actions := s.visibleActions()
		if len(actions) > 0 {
			s.cursor = (s.cursor - 1 + len(actions)) % len(actions)
		}

	case "n":
		if s.page+1 < s.pageCount() {
			s.page++
			s.cursor = 0
		}

	case "p":
		if s.page > 0 {
			s.page--
			s.cursor = 0
		}
	}

	s.ensureValidState()
	return nil
}

func (s Screen) Render(width int, heading string) string {
	headingLine := strings.TrimSpace(heading)
	if headingLine == "" {
		headingLine = theme.App.SubSubtitleStyle().Render("Home")
	}

	intro := theme.App.TextStyle().Render(
		"The LDT-Toolkit supports you in exploring\n" +
			"(1) a new longitudinal data study and\n" +
			"(2) reproducing other studies via reproducibility presets.",
	)
	intro2 := theme.App.MutedTextStyle().Render("Use tabs for stage navigation, or run a command directly from the command palette (:).")

	actions := s.visibleActions()
	itemWidth := model.IntMin(homeListMaxWidth, model.IntMax(homeListMinWidth, width-38))
	actionList := components.RenderGeneralActionList(actions, s.cursor, itemWidth)

	pageInfo := theme.App.MutedTextStyle().Render(
		fmt.Sprintf("Actions page %d/%d • n/p paginate", s.page+1, s.pageCount()),
	)

	preview := ""
	if len(actions) > 0 && s.cursor >= 0 && s.cursor < len(actions) {
		preview = theme.App.MutedTextStyle().Render(
			fmt.Sprintf("Selected action: %s", model.CommandLabel(actions[s.cursor])),
		)
	}

	if len(actions) == 0 {
		actionList = theme.App.MutedTextStyle().Render("No actions configured.")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headingLine,
		"",
		intro,
		intro2,
		"",
		"",
		actionList,
		"",
		"",
		pageInfo,
		preview,
	)
}

func (s Screen) HasAction(name string) bool {
	needle := strings.TrimSpace(name)
	for _, action := range s.actions {
		if action.Name == needle {
			return true
		}
	}
	return false
}

func (s Screen) Actions() []model.CommandDef {
	return cloneCommandDefs(s.actions)
}

func (s *Screen) ensureValidState() {
	totalPages := s.pageCount()
	if s.page < 0 {
		s.page = 0
	}
	if s.page >= totalPages {
		s.page = totalPages - 1
	}
	if s.page < 0 {
		s.page = 0
	}

	visible := s.visibleActions()
	if len(visible) == 0 {
		s.cursor = 0
		return
	}
	if s.cursor < 0 {
		s.cursor = 0
	}
	if s.cursor >= len(visible) {
		s.cursor = len(visible) - 1
	}
}

func (s Screen) pageCount() int {
	if len(s.actions) == 0 || s.actionsPerPage <= 0 {
		return 1
	}
	return (len(s.actions) + s.actionsPerPage - 1) / s.actionsPerPage
}

func (s Screen) visibleActions() []model.CommandDef {
	if len(s.actions) == 0 {
		return nil
	}
	start := s.page * s.actionsPerPage
	if start >= len(s.actions) {
		return nil
	}
	end := model.IntMin(start+s.actionsPerPage, len(s.actions))
	return s.actions[start:end]
}
