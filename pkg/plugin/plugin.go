package plugin

import (
	"fmt"

	"github.com/ravenix/peerd/pkg/explorer"
	"github.com/ravenix/peerd/pkg/handler"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type PluginApi interface {
	RegisterExplorer(name string, initializer explorer.Initializer)
	RegisterHandler(name string, initializer handler.Initializer)
}

type proxyPluginApi struct {
	pluginName             string
	registerExplorerMethod func(string, explorer.Initializer)
	registerHandlerMethod  func(string, handler.Initializer)
}

type SetupFunc func(PluginApi)

var r *registry

func init() {
	r = NewRegistry()
}

func Register(name string, setupFunc SetupFunc) {
	log.Infof("Loading plugin '%s'", name)

	setupFunc(&proxyPluginApi{
		pluginName:             name,
		registerExplorerMethod: r.RegisterExplorer,
		registerHandlerMethod:  r.RegisterHandler,
	})
}

func InitializeExplorer(name string, config *yaml.Node) (explorer.Explorer, error) {
	if r.explorers[name] == nil {
		return nil, fmt.Errorf("explorer with name '%s' does not exist", name)
	}

	return r.explorers[name](config)
}

func InitializeHandler(name string, config *yaml.Node) (handler.Handler, error) {
	if r.handlers[name] == nil {
		return nil, fmt.Errorf("handler with name '%s' does not exist", name)
	}

	return r.handlers[name](config)
}

func (p *proxyPluginApi) RegisterExplorer(name string, initializer explorer.Initializer) {
	p.registerExplorerMethod(p.pluginName+":"+name, initializer)
}

func (p *proxyPluginApi) RegisterHandler(name string, initializer handler.Initializer) {
	p.registerHandlerMethod(p.pluginName+":"+name, initializer)
}
