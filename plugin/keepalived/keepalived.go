package keepalived

import "github.com/ravenix/peerd/pkg/plugin"

func init() {
	plugin.Register("keepalived", setup)
}

func setup(api plugin.PluginApi) {
	api.RegisterExplorer("instance", instanceExplorerInitializer)
}
