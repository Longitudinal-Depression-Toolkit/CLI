package benchstandardandlongitudinalml

import (
	"ldt-toolkit-cli/internal/screens/machine_learning/tools/spec"
)

func Build(tool spec.Tool) spec.FlowSpec {
	flow := spec.Default(tool)
	flow.PromptTechniqueSelection = PromptTechniqueSelection
	flow.PromptTechniqueParams = PromptTechniqueParameters
	return flow
}
