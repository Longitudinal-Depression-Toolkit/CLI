package schema

import (
	"encoding/json"
	"errors"
	"fmt"
)

type Option struct {
	Label       string `json:"label"`
	Value       any    `json:"value"`
	Description string `json:"description"`
}

type Condition struct {
	Field  string `json:"field"`
	Equals any    `json:"equals"`
}

type Parameter struct {
	Key         string     `json:"key"`
	Label       string     `json:"label"`
	Type        string     `json:"type"`
	Required    bool       `json:"required"`
	Default     any        `json:"default"`
	Hint        string     `json:"hint"`
	Placeholder string     `json:"placeholder"`
	Options     []Option   `json:"options"`
	Min         *float64   `json:"min"`
	Max         *float64   `json:"max"`
	When        *Condition `json:"when"`
}

type Technique struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  []Parameter `json:"parameters"`
}

func DecodeTechniquePayload(result map[string]any) ([]Technique, error) {
	rawTechniques, ok := result["techniques"]
	if !ok {
		return nil, errors.New("bridge payload missing `techniques`")
	}
	encoded, err := json.Marshal(rawTechniques)
	if err != nil {
		return nil, fmt.Errorf("failed to encode techniques payload: %w", err)
	}

	var techniques []Technique
	if err := json.Unmarshal(encoded, &techniques); err != nil {
		return nil, fmt.Errorf("failed to decode techniques payload: %w", err)
	}
	return techniques, nil
}
