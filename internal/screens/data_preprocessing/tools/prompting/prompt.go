package prompting

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"

	"ldt-toolkit-cli/internal/screens/data_preprocessing/support/schema"
	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/model"
)

func PromptTechniqueSelection(title string, techniques []schema.Technique) (schema.Technique, error) {
	options := make([]huh.Option[string], 0, len(techniques))
	for _, technique := range techniques {
		label := technique.Name
		options = append(options, huh.NewOption(label, technique.ID))
	}

	selectedID := ""
	if len(techniques) > 0 {
		selectedID = techniques[0].ID
	}

	form := components.NewLDTForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Description("Enter to choose. Esc to exit tool.").
				Options(options...).
				Value(&selectedID),
		),
	)
	if err := form.Run(); err != nil {
		return schema.Technique{}, err
	}

	for _, technique := range techniques {
		if technique.ID == selectedID {
			return technique, nil
		}
	}
	return schema.Technique{}, fmt.Errorf("selected technique not found: %s", selectedID)
}

func PromptTechniqueParameters(technique schema.Technique) (map[string]any, error) {
	collected := make(map[string]any)

	for _, parameter := range technique.Parameters {
		if !shouldPromptParameter(parameter, collected) {
			continue
		}
		if strings.TrimSpace(parameter.Key) == "" {
			continue
		}

		value, err := promptParameterValue(parameter)
		if err != nil {
			return nil, err
		}
		if value == nil {
			continue
		}
		collected[parameter.Key] = value
	}

	return collected, nil
}

func promptParameterValue(parameter schema.Parameter) (any, error) {
	typeName := strings.ToLower(strings.TrimSpace(parameter.Type))
	title := parameterPromptTitle(parameter)

	if components.IsNumericParameterType(typeName) {
		return components.PromptNumericParameter(components.NumericParameterPromptConfig{
			Key:      parameter.Key,
			Name:     parameterPromptName(parameter),
			Type:     typeName,
			Hint:     parameter.Hint,
			Required: parameter.Required,
			Default:  parameter.Default,
			Min:      parameter.Min,
			Max:      parameter.Max,
		})
	}

	if typeName == "path" && shouldUsePathPickerForParameter(parameter) {
		mode := components.PathPickerFile
		if parameterExpectsDirectory(parameter) {
			mode = components.PathPickerDirectory
		}
		selectedPath, err := components.PickPathWithFilePicker(
			parameterPromptName(parameter),
			strings.TrimSpace(parameter.Hint),
			mode,
			defaultInputValue(parameter),
		)
		if err != nil {
			if errors.Is(err, components.ErrPathPickerCancelled) {
				return nil, err
			}
			return nil, err
		}
		return parseInputParameter(parameter, selectedPath)
	}

	switch typeName {
	case "bool", "boolean":
		value := defaultBool(parameter.Default)
		form := components.NewLDTForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(title).
					Value(&value),
			),
		)
		if err := form.Run(); err != nil {
			return nil, err
		}
		return value, nil

	case "enum":
		if len(parameter.Options) == 0 {
			return nil, fmt.Errorf("enum parameter `%s` has no options", parameter.Key)
		}
		selected := defaultString(parameter.Default)
		options := make([]huh.Option[string], 0, len(parameter.Options))
		for _, option := range parameter.Options {
			label := strings.TrimSpace(option.Label)
			value := strings.TrimSpace(fmt.Sprintf("%v", option.Value))
			if label == "" {
				label = value
			}
			if strings.TrimSpace(option.Description) != "" {
				label = fmt.Sprintf("%s — %s", label, option.Description)
			}
			options = append(options, huh.NewOption(label, value))
		}
		if selected == "" {
			selected = strings.TrimSpace(fmt.Sprintf("%v", parameter.Options[0].Value))
		}

		form := components.NewLDTForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(title).
					Options(options...).
					Value(&selected),
			),
		)
		if err := form.Run(); err != nil {
			return nil, err
		}
		if strings.TrimSpace(selected) == "" && parameter.Required {
			return nil, fmt.Errorf("%s is required", parameter.Label)
		}
		if strings.TrimSpace(selected) == "" {
			return nil, nil
		}
		return selected, nil
	}

	rawValue := defaultInputValue(parameter)
	inputField := huh.NewInput().
		Title(title).
		Value(&rawValue).
		Validate(func(value string) error {
			_, err := parseInputParameter(parameter, value)
			return err
		})
	if strings.TrimSpace(parameter.Placeholder) != "" {
		inputField = inputField.Placeholder(strings.TrimSpace(parameter.Placeholder))
	}

	form := components.NewLDTForm(
		huh.NewGroup(inputField),
	)
	if err := form.Run(); err != nil {
		return nil, err
	}

	parsed, err := parseInputParameter(parameter, rawValue)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

func shouldUsePathPickerForParameter(parameter schema.Parameter) bool {
	key := strings.ToLower(strings.TrimSpace(parameter.Key))
	if key == "" {
		return false
	}
	return strings.HasPrefix(key, "input_") &&
		(strings.Contains(key, "path") || strings.Contains(key, "folder"))
}

func parameterExpectsDirectory(parameter schema.Parameter) bool {
	key := strings.ToLower(strings.TrimSpace(parameter.Key))
	if strings.Contains(key, "folder") || strings.Contains(key, "directory") {
		return true
	}
	return false
}

func parseInputParameter(parameter schema.Parameter, raw string) (any, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		if parameter.Required {
			return nil, fmt.Errorf("%s is required", parameterPromptName(parameter))
		}
		return nil, nil
	}

	typeName := strings.ToLower(strings.TrimSpace(parameter.Type))
	switch typeName {
	case "int", "integer":
		parsed, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, fmt.Errorf("%s must be an integer", parameterPromptName(parameter))
		}
		if parameter.Min != nil && float64(parsed) < *parameter.Min {
			return nil, fmt.Errorf("%s must be >= %v", parameterPromptName(parameter), *parameter.Min)
		}
		if parameter.Max != nil && float64(parsed) > *parameter.Max {
			return nil, fmt.Errorf("%s must be <= %v", parameterPromptName(parameter), *parameter.Max)
		}
		return parsed, nil

	case "float", "number":
		parsed, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return nil, fmt.Errorf("%s must be numeric", parameterPromptName(parameter))
		}
		if parameter.Min != nil && parsed < *parameter.Min {
			return nil, fmt.Errorf("%s must be >= %v", parameterPromptName(parameter), *parameter.Min)
		}
		if parameter.Max != nil && parsed > *parameter.Max {
			return nil, fmt.Errorf("%s must be <= %v", parameterPromptName(parameter), *parameter.Max)
		}
		return parsed, nil

	case "string_list":
		items := model.ParseCSVValues(trimmed)
		if len(items) == 0 {
			if parameter.Required {
				return nil, fmt.Errorf("%s must include at least one value", parameterPromptName(parameter))
			}
			return nil, nil
		}
		return items, nil

	case "json":
		var parsed any
		if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
			return nil, fmt.Errorf("%s must be valid JSON", parameterPromptName(parameter))
		}
		return parsed, nil

	case "path", "string", "text":
		return trimmed, nil

	default:
		return trimmed, nil
	}
}

func shouldPromptParameter(parameter schema.Parameter, collected map[string]any) bool {
	if parameter.When == nil || strings.TrimSpace(parameter.When.Field) == "" {
		return true
	}

	actualValue, ok := collected[strings.TrimSpace(parameter.When.Field)]
	if !ok {
		return false
	}

	expectedValue := parameter.When.Equals
	if expectedString, ok := expectedValue.(string); ok {
		return strings.EqualFold(
			strings.TrimSpace(fmt.Sprintf("%v", actualValue)),
			strings.TrimSpace(expectedString),
		)
	}
	return fmt.Sprintf("%v", actualValue) == fmt.Sprintf("%v", expectedValue)
}

func parameterPromptTitle(parameter schema.Parameter) string {
	name := parameterPromptName(parameter)
	hint := strings.TrimSpace(parameter.Hint)
	if hint == "" {
		return name
	}
	return fmt.Sprintf("%s (%s)", name, hint)
}

func parameterPromptName(parameter schema.Parameter) string {
	label := strings.TrimSpace(parameter.Label)
	if label != "" {
		return label
	}
	return strings.TrimSpace(parameter.Key)
}

func defaultInputValue(parameter schema.Parameter) string {
	if parameter.Default == nil {
		return ""
	}

	switch value := parameter.Default.(type) {
	case string:
		return value
	case float64:
		if strings.EqualFold(parameter.Type, "int") || strings.EqualFold(parameter.Type, "integer") {
			return strconv.Itoa(int(value))
		}
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value)
	case []any:
		parts := make([]string, 0, len(value))
		for _, entry := range value {
			parts = append(parts, strings.TrimSpace(fmt.Sprintf("%v", entry)))
		}
		return strings.Join(parts, ",")
	case []string:
		return strings.Join(value, ",")
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", value))
	}
}

func defaultString(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", value))
}

func defaultBool(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		normalized := strings.ToLower(strings.TrimSpace(typed))
		return normalized == "1" || normalized == "true" || normalized == "yes" || normalized == "y"
	case float64:
		return typed != 0
	default:
		return false
	}
}
