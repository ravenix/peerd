package group

import (
	"context"
	"reflect"
	"time"

	"github.com/ravenix/peerd/internal/peer"
	"github.com/ravenix/peerd/pkg/explorer"
	"github.com/ravenix/peerd/pkg/handler"

	log "github.com/sirupsen/logrus"
)

type Group struct {
	Name      string
	Explorers []explorer.Explorer
	Handlers  []handler.Handler

	peers []*peer.Peer
}

func (g *Group) Discovered(d *explorer.Discovery) {
	for _, p := range g.peers {
		if p.IPv4Addr.Equal(d.IPv4Addr) && p.IPv6Addr.Equal(d.IPv6Addr) && p.Port == d.Port {
			p.LastSeen = time.Now()
			return
		}
	}

	g.peers = append(g.peers, &peer.Peer{
		IPv4Addr:  d.IPv4Addr,
		IPv6Addr:  d.IPv6Addr,
		Port:      d.Port,
		FirstSeen: time.Now(),
	})
}

func (g *Group) GetPeers() []*peer.Peer {
	return copyPeers(g.peers)
}

func (g *Group) Reconcile(ctx context.Context) ([]*peer.Peer, []*peer.Peer, []*peer.Peer) {
	tmp := g.peers[:0]
	var newPeers []*peer.Peer
	var lostPeers []*peer.Peer

	for _, p := range g.peers {
		if p.LastSeen.IsZero() {
			log.Debugf("new peer %v", p)
			newPeers = append(newPeers, p)

			for _, h := range g.Handlers {
				if err := h.NewPeer(ctx, p); err != nil {
					log.Warnf("Failed running new-peer hook for group '%s' of handler '%s': %v", g.Name, reflect.TypeOf(h).String(), err)
				}
			}

			p.LastSeen = time.Now()
		}

		if p.LastSeen.Add(time.Second * 30).Before(time.Now()) {
			log.Debugf("lost peer %v", p)
			lostPeers = append(lostPeers, p)

			for _, h := range g.Handlers {
				if err := h.LostPeer(ctx, p); err != nil {
					log.Warnf("Failed running lost-peer hook for group '%s' of handler '%s': %v", g.Name, reflect.TypeOf(h).String(), err)
				}
			}
		} else {
			tmp = append(tmp, p)
		}
	}

	g.peers = tmp
	return copyPeers(g.peers), copyPeers(newPeers), copyPeers(lostPeers)
}

func copyPeers(peers []*peer.Peer) []*peer.Peer {
	var tmp []*peer.Peer
	for _, p := range peers {
		tmpP := *p
		tmp = append(tmp, &tmpP)
	}
	return tmp
}
