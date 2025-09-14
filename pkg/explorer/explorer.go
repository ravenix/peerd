package explorer

import (
	"context"
	"net"

	"gopkg.in/yaml.v3"
)

type Initializer func(*yaml.Node) (Explorer, error)
type Explorer interface {
	Run(context.Context) error
	Explore(context.Context, DiscoveryHandler) error
}

type Discovery struct {
	IPv4Addr net.IP
	IPv6Addr net.IP
	Port     uint16
}

type DiscoveryHandler interface {
	Discovered(*Discovery)
}
