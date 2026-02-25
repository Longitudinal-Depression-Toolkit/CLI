package components

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
)

type NumericParameterPromptConfig struct {
	Key         string
	Name        string
	Type        string
	Hint        string
	Description string
	Required    bool
	Default     any
	Min         *float64
	Max         *float64
}

func IsNumericParameterType(typeName string) bool {
	switch strings.ToLower(strings.TrimSpace(typeName)) {
	case "int", "integer", "float", "number":
		return true
	default:
		return false
	}
}

func PromptNumericParameter(config NumericParameterPromptConfig) (any, error) {
	normalizedType := strings.ToLower(strings.TrimSpace(config.Type))
	if !IsNumericParameterType(normalizedType) {
		return nil, fmt.Errorf("parameter type `%s` is not numeric", config.Type)
	}

	name := strings.TrimSpace(config.Name)
	if name == "" {
		name = strings.TrimSpace(config.Key)
	}
	if name == "" {
		name = "numeric value"
	}

	defaultValue, hasDefault := parseNumericDefault(config.Default)
	if !config.Required && !hasDefault {
		shouldSet := false
		form := NewLDTForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Set %s?", name)).
					Description("Choose Yes to provide a value, or No to leave it unset.").
					Value(&shouldSet),
			),
		)
		if err := form.Run(); err != nil {
			return nil, err
		}
		if !shouldSet {
			return nil, nil
		}
	}

	min, max := cloneNumericBounds(config.Min, config.Max)
	description := numericPromptDescription(config)

	if normalizedType == "int" || normalizedType == "integer" {
		initial := 0
		if hasDefault {
			initial = int(math.Round(defaultValue))
		} else if min != nil {
			initial = int(math.Ceil(*min))
		}

		value, err := PromptIntStepper(NumberStepperConfig{
			Title:       name,
			Description: description,
			Initial:     float64(initial),
			Step:        1,
			Min:         min,
			Max:         max,
			Precision:   0,
			InputWidth:  18,
		})
		if err != nil {
			return nil, err
		}
		return value, nil
	}

	initial := 0.0
	if hasDefault {
		initial = defaultValue
	} else if min != nil {
		initial = *min
	}

	precision := inferNumericPromptPrecision(defaultValue, hasDefault, min, max)
	step := math.Pow(10, -float64(precision))
	if step <= 0 || math.IsNaN(step) || math.IsInf(step, 0) {
		step = 1
	}

	value, err := PromptNumberStepper(NumberStepperConfig{
		Title:       name,
		Description: description,
		Initial:     initial,
		Step:        step,
		Min:         min,
		Max:         max,
		Precision:   precision,
		InputWidth:  18,
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

func parseNumericDefault(value any) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case float32:
		return float64(typed), true
	case float64:
		return typed, true
	case string:
		trimmed := strings.TrimSpace(typed)
		if trimmed == "" {
			return 0, false
		}
		parsed, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func numericPromptDescription(config NumericParameterPromptConfig) string {
	description := strings.TrimSpace(config.Description)
	if description != "" {
		return description
	}

	hint := strings.TrimSpace(config.Hint)
	if hint != "" {
		return hint
	}

	return "Use up/down arrows to increase or decrease the value."
}

func cloneNumericBounds(min *float64, max *float64) (*float64, *float64) {
	var clonedMin *float64
	var clonedMax *float64
	if min != nil {
		value := *min
		clonedMin = &value
	}
	if max != nil {
		value := *max
		clonedMax = &value
	}
	return clonedMin, clonedMax
}

func inferNumericPromptPrecision(defaultValue float64, hasDefault bool, min *float64, max *float64) int {
	precision := 0
	if hasDefault {
		precision = maxNumericInt(precision, decimalPlaces(defaultValue))
	}
	if min != nil {
		precision = maxNumericInt(precision, decimalPlaces(*min))
	}
	if max != nil {
		precision = maxNumericInt(precision, decimalPlaces(*max))
	}

	if precision == 0 && isUnitInterval(min, max) {
		precision = 2
	}
	if precision > 6 {
		precision = 6
	}
	return precision
}

func decimalPlaces(value float64) int {
	if value == math.Trunc(value) {
		return 0
	}

	formatted := strconv.FormatFloat(value, 'f', -1, 64)
	parts := strings.SplitN(formatted, ".", 2)
	if len(parts) < 2 {
		return 0
	}
	trimmed := strings.TrimRight(parts[1], "0")
	if trimmed == "" {
		return 0
	}
	return len(trimmed)
}

func isUnitInterval(min *float64, max *float64) bool {
	if min == nil || max == nil {
		return false
	}
	return *min >= 0 && *max <= 1
}

func maxNumericInt(left int, right int) int {
	if left >= right {
		return left
	}
	return right
}
