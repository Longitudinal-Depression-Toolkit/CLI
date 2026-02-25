package datapp

import "errors"

const usageCommandPrefix = "ldt-toolkit"

type bridgeExecutor func(operation string, params map[string]any) (map[string]any, error)

var (
	currentBridgeExecutor bridgeExecutor = func(_ string, _ map[string]any) (map[string]any, error) {
		return nil, errors.New("data preprocessing bridge runtime is not configured")
	}
	currentNavigatorState = func() bool { return false }
)

func ConfigureRuntime(executor bridgeExecutor, inNavigator func() bool) {
	if executor != nil {
		currentBridgeExecutor = executor
	}
	if inNavigator != nil {
		currentNavigatorState = inNavigator
	}
}

func executeBridge(operation string, params map[string]any) (map[string]any, error) {
	return currentBridgeExecutor(operation, params)
}
