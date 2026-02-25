package tools

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"ldt-toolkit-cli/internal/screens/machine_learning/support/schema"
	"ldt-toolkit-cli/internal/screens/machine_learning/support/ui"
	"ldt-toolkit-cli/internal/screens/machine_learning/tools/prompting"
	"ldt-toolkit-cli/internal/shared/components"
	"ldt-toolkit-cli/internal/shared/theme"
)

type toolCatalog struct {
	Techniques       []schema.Technique
	EstimatorPreview []schema.Technique
}

type discoveredEstimator struct {
	Key         string
	Name        string
	Description string
}

var curatedMetricOptions = []schema.Option{
	{Label: "f1 macro", Value: "f1_macro", Description: "Macro-averaged F1 score."},
	{Label: "f1 micro", Value: "f1_micro", Description: "Micro-averaged F1 score."},
	{Label: "precision macro", Value: "precision_macro", Description: "Macro-averaged precision."},
	{Label: "precision micro", Value: "precision_micro", Description: "Micro-averaged precision."},
	{Label: "recall macro", Value: "recall_macro", Description: "Macro-averaged recall."},
	{Label: "recall micro", Value: "recall_micro", Description: "Micro-averaged recall."},
	{Label: "f1 weighted", Value: "f1_weighted", Description: "Support-weighted F1 score."},
	{Label: "precision weighted", Value: "precision_weighted", Description: "Support-weighted precision."},
	{Label: "recall weighted", Value: "recall_weighted", Description: "Support-weighted recall."},
	{Label: "accuracy", Value: "accuracy", Description: "Overall classification accuracy."},
	{Label: "balanced accuracy", Value: "balanced_accuracy", Description: "Mean class recall."},
	{Label: "AUROC", Value: "auroc", Description: "Area under the ROC curve."},
	{Label: "AUPRC", Value: "auprc", Description: "Area under the precision-recall curve."},
	{Label: "geometric mean", Value: "geometric_mean", Description: "Geometric mean of class recall."},
}

var validationSplitOptions = []schema.Option{
	{Label: "auto", Value: "auto", Description: "Cross-validation only."},
	{Label: "90/10", Value: "90/10", Description: "90% train, 10% validation."},
	{Label: "80/20", Value: "80/20", Description: "80% train, 20% validation."},
	{Label: "70/30", Value: "70/30", Description: "70% train, 30% validation."},
	{Label: "60/40", Value: "60/40", Description: "60% train, 40% validation."},
	{Label: "50/50", Value: "50/50", Description: "50% train, 50% validation."},
	{Label: "40/60", Value: "40/60", Description: "40% train, 60% validation."},
}

func Run(def Definition, listOnly bool, runtime Runtime) error {
	if runtime.Execute == nil {
		return errors.New("machine learning tools runtime has no bridge executor")
	}

	screenTool := Tool{
		ID:             def.ID,
		Name:           def.Name,
		Subtitle:       def.Subtitle,
		TableTitle:     def.TableTitle,
		SelectionTitle: def.SelectionTitle,
	}
	spec, ok := Resolve(screenTool)
	if !ok {
		spec = Default(screenTool)
	}

	ui.PrepareActionScreen(spec.Title, spec.Subtitle, runtime.InNavigator)

	catalog, err := fetchToolCatalog(def, runtime.Execute)
	if err != nil {
		return err
	}
	techniques := catalog.Techniques
	if len(techniques) == 0 {
		return fmt.Errorf("no techniques were found for %s", strings.ToLower(spec.Title))
	}

	var selected schema.Technique
	if len(techniques) == 1 {
		selected = techniques[0]
		if len(catalog.EstimatorPreview) > 0 {
			ui.PrintTechniqueTable(catalog.EstimatorPreview)
		}
		if listOnly {
			return nil
		}
	} else {
		ui.PrintTechniqueTable(techniques)
		if listOnly {
			return nil
		}

		selectionTitle := strings.TrimSpace(spec.SelectionTitle)
		if selectionTitle == "" {
			selectionTitle = fmt.Sprintf("%s technique", strings.TrimSpace(spec.Title))
		}

		if spec.PromptTechniqueSelection == nil {
			spec.PromptTechniqueSelection = prompting.PromptTechniqueSelection
		}
		selected, err = spec.PromptTechniqueSelection(selectionTitle, techniques)
		if err != nil {
			return err
		}
	}

	if spec.PromptTechniqueParams == nil {
		spec.PromptTechniqueParams = prompting.PromptTechniqueParameters
	}
	params, err := spec.PromptTechniqueParams(selected)
	if err != nil {
		return err
	}

	operationTechnique := selected.ID
	if base, estimatorKey, ok := splitSyntheticTechniqueID(selected.ID); ok {
		operationTechnique = base
		if strings.TrimSpace(estimatorKey) != "" {
			if _, exists := params["estimator_key"]; !exists {
				params["estimator_key"] = estimatorKey
			}
		}
	}

	result, err := runtime.Execute(
		def.RunOperation,
		map[string]any{
			"technique": operationTechnique,
			"params":    params,
		},
	)
	if err != nil {
		return err
	}

	if spec.SummaryPrinter != nil {
		spec.SummaryPrinter(selected, result)
	} else {
		printGenericSummary(selected, result)
	}

	return ui.RunExitCountdown("Returning to toolkit", 10*time.Second, runtime.InNavigator)
}

func fetchToolCatalog(def Definition, execute BridgeExecutor) (toolCatalog, error) {
	result, err := execute(def.CatalogOperation, map[string]any{})
	if err != nil {
		return toolCatalog{}, err
	}
	techniques, err := schema.DecodeTechniquePayload(result)
	if err != nil {
		return toolCatalog{}, err
	}

	interactiveTechniques := filterNonInteractiveTechniques(techniques)
	if len(interactiveTechniques) == 0 {
		interactiveTechniques = techniques
	}

	estimators, err := fetchDiscoveredEstimators(def, execute)
	if err != nil {
		return toolCatalog{}, err
	}
	estimatorPreview := estimatorsToTechniques(estimators)

	enriched := enrichParameters(interactiveTechniques, estimators)
	if shouldUseEstimatorTechniqueChoices(def.ID) {
		if expanded := expandEstimatorTechniques(enriched, estimators); len(expanded) > 0 {
			enriched = expanded
		}
	}

	return toolCatalog{
		Techniques:       enriched,
		EstimatorPreview: estimatorPreview,
	}, nil
}

func printGenericSummary(selected schema.Technique, result map[string]any) {
	components.PrintBlankLine()
	components.PrintLine(theme.App.SubtitleStyle().Render("Run complete"))
	components.PrintfLine("- Technique: %s", nonEmptyText(selected.Name, selected.ID))
	for key, value := range result {
		components.PrintfLine("- %s: %v", strings.TrimSpace(key), value)
	}
}

func nonEmptyText(primary string, fallback string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	if strings.TrimSpace(fallback) != "" {
		return fallback
	}
	return "Unnamed technique"
}

func splitSyntheticTechniqueID(raw string) (base string, estimatorKey string, ok bool) {
	if !strings.Contains(raw, "::") {
		return "", "", false
	}
	parts := strings.SplitN(raw, "::", 2)
	base = strings.TrimSpace(parts[0])
	estimatorKey = strings.TrimSpace(parts[1])
	if base == "" {
		return "", "", false
	}
	return base, estimatorKey, true
}

func filterNonInteractiveTechniques(techniques []schema.Technique) []schema.Technique {
	filtered := make([]schema.Technique, 0, len(techniques))
	for _, technique := range techniques {
		identifier := strings.ToLower(strings.TrimSpace(technique.ID))
		if identifier == "list_estimators" || identifier == "list_metrics" {
			continue
		}
		filtered = append(filtered, technique)
	}
	return filtered
}

func shouldUseEstimatorTechniqueChoices(toolID string) bool {
	switch strings.TrimSpace(toolID) {
	case "standard_machine_learning", "longitudinal_machine_learning":
		return true
	default:
		return false
	}
}

func expandEstimatorTechniques(
	techniques []schema.Technique,
	estimators []discoveredEstimator,
) []schema.Technique {
	if len(techniques) == 0 || len(estimators) == 0 {
		return nil
	}

	base := techniques[0]
	expanded := make([]schema.Technique, 0, len(estimators))
	for _, estimator := range estimators {
		parameters := removeParameter(base.Parameters, "estimator_key")
		expanded = append(expanded, schema.Technique{
			ID:          fmt.Sprintf("%s::%s", strings.TrimSpace(base.ID), estimator.Key),
			Name:        nonEmptyText(estimator.Name, estimator.Key),
			Description: strings.TrimSpace(estimator.Description),
			Parameters:  parameters,
		})
	}
	return expanded
}

func enrichParameters(
	techniques []schema.Technique,
	estimators []discoveredEstimator,
) []schema.Technique {
	enriched := make([]schema.Technique, 0, len(techniques))
	for _, technique := range techniques {
		parameters := make([]schema.Parameter, 0, len(technique.Parameters))
		for _, parameter := range technique.Parameters {
			key := strings.ToLower(strings.TrimSpace(parameter.Key))
			switch key {
			case "metric_keys":
				parameter.Options = append([]schema.Option(nil), curatedMetricOptions...)
				parameter.Hint = "Tick one or more metrics. Example: f1_macro + auroc + geometric_mean."
			case "validation_split":
				parameter.Options = append([]schema.Option(nil), validationSplitOptions...)
				parameter.Default = "auto"
				parameter.Hint = "Leave as auto for CV-only. Or use 70/30 where 70% trains and 30% validates (also accepts 0.3)."
			case "excluded_estimators":
				parameter.Label = "Include estimators"
				parameter.Options = estimatorOptions(estimators)
				parameter.Hint = "Tick estimators to include in this run. Unticked estimators are not considered."
			}
			parameters = append(parameters, parameter)
		}
		technique.Parameters = parameters
		enriched = append(enriched, technique)
	}
	return enriched
}

func removeParameter(parameters []schema.Parameter, key string) []schema.Parameter {
	needle := strings.ToLower(strings.TrimSpace(key))
	filtered := make([]schema.Parameter, 0, len(parameters))
	for _, parameter := range parameters {
		if strings.ToLower(strings.TrimSpace(parameter.Key)) == needle {
			continue
		}
		filtered = append(filtered, parameter)
	}
	return filtered
}

func estimatorOptions(estimators []discoveredEstimator) []schema.Option {
	options := make([]schema.Option, 0, len(estimators))
	for _, estimator := range estimators {
		options = append(options, schema.Option{
			Label:       nonEmptyText(estimator.Name, estimator.Key),
			Value:       estimator.Key,
			Description: strings.TrimSpace(estimator.Description),
		})
	}
	return options
}

func estimatorsToTechniques(estimators []discoveredEstimator) []schema.Technique {
	rows := make([]schema.Technique, 0, len(estimators))
	for _, estimator := range estimators {
		rows = append(rows, schema.Technique{
			ID:          estimator.Key,
			Name:        nonEmptyText(estimator.Name, estimator.Key),
			Description: strings.TrimSpace(estimator.Description),
		})
	}
	return rows
}

func fetchDiscoveredEstimators(
	def Definition,
	execute BridgeExecutor,
) ([]discoveredEstimator, error) {
	switch strings.TrimSpace(def.ID) {
	case "standard_machine_learning",
		"longitudinal_machine_learning",
		"benchmark_standard_ml",
		"benchmark_longitudinal_ml",
		"bench_standard_and_longitudinal_ml":
	default:
		return nil, nil
	}

	payload, err := execute(def.RunOperation, map[string]any{
		"technique": "list_estimators",
		"params":    map[string]any{},
	})
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(def.ID) == "bench_standard_and_longitudinal_ml" {
		standard := decodeEstimatorSlice(payload["standard_estimators"], "standard")
		longitudinal := decodeEstimatorSlice(payload["longitudinal_estimators"], "longitudinal")
		combined := append(standard, longitudinal...)
		return dedupeEstimators(combined), nil
	}
	return decodeEstimatorSlice(payload["estimators"], ""), nil
}

func decodeEstimatorSlice(raw any, sourceLabel string) []discoveredEstimator {
	rows, ok := raw.([]any)
	if !ok {
		return nil
	}

	discovered := make([]discoveredEstimator, 0, len(rows))
	for _, row := range rows {
		item, ok := row.(map[string]any)
		if !ok {
			continue
		}
		key := valueAsString(item["key"])
		if key == "" {
			continue
		}
		name := valueAsString(item["name"])
		description := valueAsString(item["description"])
		if source := valueAsString(item["source"]); source != "" {
			sourceLabel = source
		}
		if sourceLabel != "" {
			name = strings.TrimSpace(fmt.Sprintf("%s (%s)", nonEmptyText(name, key), sourceLabel))
		}
		discovered = append(discovered, discoveredEstimator{
			Key:         key,
			Name:        nonEmptyText(name, key),
			Description: description,
		})
	}

	sort.SliceStable(discovered, func(i, j int) bool {
		return strings.ToLower(discovered[i].Name) < strings.ToLower(discovered[j].Name)
	})
	return discovered
}

func dedupeEstimators(estimators []discoveredEstimator) []discoveredEstimator {
	seen := make(map[string]struct{}, len(estimators))
	unique := make([]discoveredEstimator, 0, len(estimators))
	for _, estimator := range estimators {
		key := strings.TrimSpace(estimator.Key)
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, estimator)
	}
	return unique
}

func valueAsString(raw any) string {
	if raw == nil {
		return ""
	}
	rendered := strings.TrimSpace(fmt.Sprintf("%v", raw))
	if strings.EqualFold(rendered, "<nil>") {
		return ""
	}
	return rendered
}
