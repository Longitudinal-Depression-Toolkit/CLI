package components

import (
	"strings"

	"github.com/charmbracelet/huh"
)

type MultiSelectOption struct {
	Label       string
	Value       string
	Description string
	Selected    bool
}

func PromptMultiSelect(
	title string,
	description string,
	options []MultiSelectOption,
) ([]string, error) {
	selected := make([]string, 0, len(options))
	huhOptions := make([]huh.Option[string], 0, len(options))

	for _, option := range options {
		label := strings.TrimSpace(option.Label)
		value := strings.TrimSpace(option.Value)
		if value == "" {
			continue
		}
		if label == "" {
			label = value
		}
		current := huh.NewOption(label, value)
		if option.Selected {
			current = current.Selected(true)
			selected = append(selected, value)
		}
		huhOptions = append(huhOptions, current)
	}

	field := huh.NewMultiSelect[string]().
		Title(strings.TrimSpace(title)).
		Options(huhOptions...).
		Value(&selected)
	if strings.TrimSpace(description) != "" {
		field = field.Description(strings.TrimSpace(description))
	}

	form := NewLDTForm(
		huh.NewGroup(field),
	)
	if err := form.Run(); err != nil {
		return nil, err
	}
	return selected, nil
}
