package exec

import (
	"github.com/ravenix/peerd/pkg/plugin"
)

func init() {
	plugin.Register("exec", setup)
}

func setup(api plugin.PluginApi) {
	api.RegisterHandler("command", commandHandlerInitializer)
}
