package kubernetes

import "github.com/ravenix/peerd/pkg/plugin"

func init() {
	plugin.Register("kubernetes", setup)
}

func setup(api plugin.PluginApi) {
	api.RegisterExplorer("node", nodeExplorerInitializer)
	api.RegisterExplorer("pod", podExplorerInitializer)
}
