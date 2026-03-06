package presets

import (
	"encoding/json"
	"errors"
	"fmt"
)

const catalogOperation = "data_preprocessing.presets.catalog"

func Catalog(execute BridgeExecutor) ([]Preset, error) {
	if execute == nil {
		return nil, errors.New("data preprocessing presets runtime has no bridge executor")
	}
	result, err := execute(catalogOperation, map[string]any{})
	if err != nil {
		return nil, err
	}
	return decodePresetPayload(result)
}

func decodePresetPayload(result map[string]any) ([]Preset, error) {
	rawPresets, ok := result["presets"]
	if !ok {
		return nil, errors.New("bridge payload missing `presets`")
	}
	encoded, err := json.Marshal(rawPresets)
	if err != nil {
		return nil, fmt.Errorf("failed to encode presets payload: %w", err)
	}

	var presets []Preset
	if err := json.Unmarshal(encoded, &presets); err != nil {
		return nil, fmt.Errorf("failed to decode presets payload: %w", err)
	}
	return presets, nil
}
