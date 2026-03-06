package explorer

import (
	"context"
	"net"
	"time"

	"gopkg.in/yaml.v3"
)

type Initializer func(*yaml.Node) (Explorer, error)
type Explorer interface {
	Run(context.Context) error
	Explore(context.Context, DiscoveryHandler) error
}

type Cadence struct {
	ExploreInterval time.Duration
	ExploreTimeout  time.Duration
	PeerTTL         time.Duration
}

type CadenceProvider interface {
	Cadence() Cadence
}

func DefaultCadence() Cadence {
	return Cadence{
		ExploreInterval: 2 * time.Second,
		ExploreTimeout:  2 * time.Second,
		PeerTTL:         5 * time.Second,
	}
}

func normalizeCadence(c Cadence) Cadence {
	defaults := DefaultCadence()

	if c.ExploreInterval <= 0 {
		c.ExploreInterval = defaults.ExploreInterval
	}

	if c.ExploreTimeout <= 0 {
		c.ExploreTimeout = defaults.ExploreTimeout
	}

	if c.PeerTTL <= 0 {
		c.PeerTTL = defaults.PeerTTL
	}

	return c
}

func ResolveCadence(explorers []Explorer) Cadence {
	if len(explorers) == 0 {
		return DefaultCadence()
	}

	var cadence Cadence
	for idx, e := range explorers {
		provider, ok := e.(CadenceProvider)
		current := DefaultCadence()
		if ok {
			current = provider.Cadence()
		}

		current = normalizeCadence(current)
		if idx == 0 {
			cadence = current
			continue
		}

		if current.ExploreInterval < cadence.ExploreInterval {
			cadence.ExploreInterval = current.ExploreInterval
		}

		if current.ExploreTimeout > cadence.ExploreTimeout {
			cadence.ExploreTimeout = current.ExploreTimeout
		}

		if current.PeerTTL > cadence.PeerTTL {
			cadence.PeerTTL = current.PeerTTL
		}
	}

	return cadence
}

type Discovery struct {
	IPv4Addr net.IP
	IPv6Addr net.IP
	Port     uint16
}

type DiscoveryHandler interface {
	Discovered(*Discovery)
}
