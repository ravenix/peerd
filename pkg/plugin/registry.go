package plugin

import (
	"github.com/ravenix/peerd/pkg/explorer"
	"github.com/ravenix/peerd/pkg/handler"
	log "github.com/sirupsen/logrus"
)

type registry struct {
	explorers map[string]explorer.Initializer
	handlers  map[string]handler.Initializer
}

func NewRegistry() *registry {
	r = new(registry)
	r.explorers = make(map[string]explorer.Initializer)
	r.handlers = make(map[string]handler.Initializer)
	return r
}

func (r *registry) RegisterExplorer(name string, handler explorer.Initializer) {
	r.explorers[name] = handler
	log.Infof("Registering explorer '%s'", name)
}

func (r *registry) RegisterHandler(name string, handler handler.Initializer) {
	r.handlers[name] = handler
	log.Infof("Registering handler '%s'", name)
}
