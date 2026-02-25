package presets

import (
	"strings"

	"ldt-toolkit-cli/internal/screens/machine_learning/presets/spec"
)

type Preset = spec.Preset
type ModuleSpec = spec.ModuleSpec

type Builder func(preset Preset) ModuleSpec
type Runner func(preset Preset, runtime Runtime) error

type Module struct {
	Build Builder
	Run   Runner
}

// Preset-specific modules can be registered here as they are implemented.
var modules = map[string]Module{}

func ResolveModule(preset Preset) (Module, bool) {
	key := canonicalPresetKey(preset.ID)
	module, ok := modules[key]
	return module, ok
}

func Default(preset Preset) ModuleSpec {
	return spec.Default(preset)
}

func canonicalPresetKey(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}
