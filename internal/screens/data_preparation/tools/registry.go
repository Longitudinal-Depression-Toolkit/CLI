package tools

import (
	"strings"

	dataconversion "ldt-toolkit-cli/internal/screens/data_preparation/tools/data_conversion"
	"ldt-toolkit-cli/internal/screens/data_preparation/tools/spec"
	syntheticdatageneration "ldt-toolkit-cli/internal/screens/data_preparation/tools/synthetic_data_generation"
)

type Tool = spec.Tool
type FlowSpec = spec.FlowSpec

type Builder func(tool Tool) FlowSpec

var builders = map[string]Builder{
	"synthetic_data_generation": syntheticdatageneration.Build,
	"data_conversion":           dataconversion.Build,
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
