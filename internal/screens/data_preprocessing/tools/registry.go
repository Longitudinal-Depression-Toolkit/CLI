package tools

import (
	"strings"

	aggregatelongtocrosssectional "ldt-toolkit-cli/internal/screens/data_preprocessing/tools/aggregate_long_to_cross_sectional"
	buildtrajectories "ldt-toolkit-cli/internal/screens/data_preprocessing/tools/build_trajectories"
	cleandataset "ldt-toolkit-cli/internal/screens/data_preprocessing/tools/clean_dataset"
	combinedatasetwithtrajectories "ldt-toolkit-cli/internal/screens/data_preprocessing/tools/combine_dataset_with_trajectories"
	harmonisecategories "ldt-toolkit-cli/internal/screens/data_preprocessing/tools/harmonise_categories"
	missingimputation "ldt-toolkit-cli/internal/screens/data_preprocessing/tools/missing_imputation"
	pivotlongtowide "ldt-toolkit-cli/internal/screens/data_preprocessing/tools/pivot_long_to_wide"
	"ldt-toolkit-cli/internal/screens/data_preprocessing/tools/remove_columns"
	renamefeature "ldt-toolkit-cli/internal/screens/data_preprocessing/tools/rename_feature"
	showtable "ldt-toolkit-cli/internal/screens/data_preprocessing/tools/show_table"
	"ldt-toolkit-cli/internal/screens/data_preprocessing/tools/spec"
	trajectoriesviz "ldt-toolkit-cli/internal/screens/data_preprocessing/tools/trajectories_viz"
)

type Tool = spec.Tool
type FlowSpec = spec.FlowSpec

type Builder func(tool Tool) FlowSpec

var builders = map[string]Builder{
	"aggregate_long_to_cross_sectional": aggregatelongtocrosssectional.Build,
	"build_trajectories":                buildtrajectories.Build,
	"clean_dataset":                     cleandataset.Build,
	"combine_dataset_with_trajectories": combinedatasetwithtrajectories.Build,
	"harmonise_categories":              harmonisecategories.Build,
	"missing_imputation":                missingimputation.Build,
	"pivot_long_to_wide":                pivotlongtowide.Build,
	"remove_columns":                    removecolumns.Build,
	"rename_feature":                    renamefeature.Build,
	"show_table":                        showtable.Build,
	"trajectories_viz":                  trajectoriesviz.Build,
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
