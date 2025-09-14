package multicast

import "github.com/ravenix/peerd/pkg/plugin"

func init() {
	plugin.Register("multicast", setup)
}

func setup(api plugin.PluginApi) {
	api.RegisterExplorer("dns", dnsExplorerInitializer)
}
