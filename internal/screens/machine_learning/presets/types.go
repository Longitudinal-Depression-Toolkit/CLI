package presets

type BridgeExecutor func(operation string, params map[string]any) (map[string]any, error)

type Runtime struct {
	Execute     BridgeExecutor
	InNavigator func() bool
}
