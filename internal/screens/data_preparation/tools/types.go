package tools

type BridgeExecutor func(operation string, params map[string]any) (map[string]any, error)

type Runtime struct {
	Execute     BridgeExecutor
	InNavigator func() bool
}

type Definition struct {
	ID               string
	Name             string
	Subtitle         string
	TableTitle       string
	SelectionTitle   string
	CatalogOperation string
	RunOperation     string
}
