package netlink

import "github.com/ravenix/peerd/pkg/plugin"

func init() {
	plugin.Register("netlink", setup)
}

func setup(api plugin.PluginApi) {
	api.RegisterExplorer("hardwareaddr", hardwareAddrExplorerInitializer)
}
