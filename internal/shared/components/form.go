package components

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

func NewLDTForm(groups ...*huh.Group) *huh.Form {
	return huh.NewForm(groups...).
		WithTheme(ldtFormTheme()).
		WithKeyMap(ldtFormKeyMap())
}

func ldtFormTheme() *huh.Theme {
	formTheme := huh.ThemeCharm()
	accent := lipgloss.Color("#DFA3C3")
	text := lipgloss.Color("#ECECF3")
	muted := lipgloss.Color("#A1A1AA")
	border := lipgloss.Color("#6B7280")

	formTheme.Form.Base = formTheme.Form.Base.
		BorderForeground(border).
		MarginLeft(LeftLayoutMargin)
	formTheme.Group.Title = formTheme.Group.Title.Foreground(accent).Bold(true)
	formTheme.Group.Description = formTheme.Group.Description.Foreground(muted)
	formTheme.FieldSeparator = formTheme.FieldSeparator.Foreground(border)

	formTheme.Focused.Title = formTheme.Focused.Title.Foreground(accent).Bold(true)
	formTheme.Focused.Description = formTheme.Focused.Description.Foreground(muted)
	formTheme.Focused.SelectSelector = formTheme.Focused.SelectSelector.Foreground(accent).Bold(true)
	formTheme.Focused.Option = formTheme.Focused.Option.Foreground(text)
	formTheme.Focused.NextIndicator = formTheme.Focused.NextIndicator.Foreground(accent)
	formTheme.Focused.PrevIndicator = formTheme.Focused.PrevIndicator.Foreground(accent)
	formTheme.Focused.MultiSelectSelector = formTheme.Focused.MultiSelectSelector.Foreground(accent).Bold(true)
	formTheme.Focused.SelectedOption = formTheme.Focused.SelectedOption.Foreground(accent).Bold(true)
	formTheme.Focused.SelectedPrefix = formTheme.Focused.SelectedPrefix.Foreground(accent).Bold(true)
	formTheme.Focused.UnselectedOption = formTheme.Focused.UnselectedOption.Foreground(text)
	formTheme.Focused.UnselectedPrefix = formTheme.Focused.UnselectedPrefix.Foreground(muted)
	formTheme.Focused.TextInput.Cursor = formTheme.Focused.TextInput.Cursor.Foreground(accent)
	formTheme.Focused.TextInput.Prompt = formTheme.Focused.TextInput.Prompt.Foreground(accent)
	formTheme.Focused.TextInput.Placeholder = formTheme.Focused.TextInput.Placeholder.Foreground(muted)
	formTheme.Focused.TextInput.Text = formTheme.Focused.TextInput.Text.Foreground(text)

	return formTheme
}

func ldtFormKeyMap() *huh.KeyMap {
	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "exit"),
	)
	return keymap
}
