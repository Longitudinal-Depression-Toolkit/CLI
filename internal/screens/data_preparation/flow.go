package dataprep

import toolsmodule "ldt-toolkit-cli/internal/screens/data_preparation/tools"

func runDataPreparationToolFlow(tool ToolConfig, listOnly bool) error {
	err := toolsmodule.Run(
		toolsmodule.Definition{
			ID:               tool.ID,
			Name:             tool.Name,
			Subtitle:         tool.Subtitle,
			TableTitle:       tool.TableTitle,
			SelectionTitle:   tool.SelectionTitle,
			CatalogOperation: tool.CatalogOperation,
			RunOperation:     tool.RunOperation,
		},
		listOnly,
		toolsmodule.Runtime{
			Execute:     executeBridge,
			InNavigator: currentNavigatorState,
		},
	)
	if toolsmodule.IsFlowCancelled(err) {
		return nil
	}
	return err
}
