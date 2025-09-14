package template

import (
	"github.com/ravenix/peerd/internal/peer"
	"github.com/ravenix/peerd/pkg/plugin"
)

func init() {
	plugin.Register("template", setup)
}

func setup(api plugin.PluginApi) {
	api.RegisterHandler("file", fileHandlerInitializer)
}

type TemplateContext struct {
	Peers []*peer.Peer
}
