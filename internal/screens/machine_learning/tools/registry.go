package tools

import (
	"strings"

	benchstandardandlongitudinalml "ldt-toolkit-cli/internal/screens/machine_learning/tools/bench_standard_and_longitudinal_ml"
	benchmarklongitudinalml "ldt-toolkit-cli/internal/screens/machine_learning/tools/benchmark_longitudinal_ml"
	benchmarkstandardml "ldt-toolkit-cli/internal/screens/machine_learning/tools/benchmark_standard_ml"
	longitudinalmachinelearning "ldt-toolkit-cli/internal/screens/machine_learning/tools/longitudinal_machine_learning"
	"ldt-toolkit-cli/internal/screens/machine_learning/tools/shap_analysis"
	"ldt-toolkit-cli/internal/screens/machine_learning/tools/spec"
	standardmachinelearning "ldt-toolkit-cli/internal/screens/machine_learning/tools/standard_machine_learning"
)

type Tool = spec.Tool
type FlowSpec = spec.FlowSpec

type Builder func(tool Tool) FlowSpec

var builders = map[string]Builder{
	"bench_standard_and_longitudinal_ml": benchstandardandlongitudinalml.Build,
	"benchmark_longitudinal_ml":          benchmarklongitudinalml.Build,
	"benchmark_standard_ml":              benchmarkstandardml.Build,
	"longitudinal_machine_learning":      longitudinalmachinelearning.Build,
	"shap_analysis":                      shapanalysis.Build,
	"standard_machine_learning":          standardmachinelearning.Build,
}

func Resolve(tool Tool) (FlowSpec, bool) {
	builder, ok := builders[canonicalToolKey(tool.ID)]
	if !ok {
		return FlowSpec{}, false
	}
	return builder(tool), true
}

func Default(tool Tool) FlowSpec {
	return spec.Default(tool)
}

func canonicalToolKey(raw string) string {
	return strings.TrimSpace(strings.ToLower(raw))
}
