package presets

import (
	"strings"

	preparemcsbyleap "ldt-toolkit-cli/internal/screens/data_preparation/presets/prepare_mcs_by_leap"
	"ldt-toolkit-cli/internal/screens/data_preparation/presets/spec"
)

type Preset = spec.Preset
type ModuleSpec = spec.ModuleSpec

type Builder func(preset Preset) ModuleSpec
type Runner func(preset Preset, runtime Runtime) error

type Module struct {
	Build Builder
	Run   Runner
}

var modules = map[string]Module{
	"prepare_mcs_by_leap": {
		Build: preparemcsbyleap.Build,
		Run: func(preset Preset, runtime Runtime) error {
			return preparemcsbyleap.Run(
				preset,
				preparemcsbyleap.Runtime{
					Execute: func(operation string, params map[string]any) (map[string]any, error) {
						return runtime.Execute(operation, params)
					},
					InNavigator: runtime.InNavigator,
				},
			)
		},
	},
}

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
