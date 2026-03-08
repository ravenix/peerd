package group

import (
	"context"
	"reflect"
	"sync"
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

	mu    sync.RWMutex
	peers []*peer.Peer
}

func (g *Group) Discovered(d *explorer.Discovery) {
	g.mu.Lock()
	defer g.mu.Unlock()

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
	g.mu.RLock()
	defer g.mu.RUnlock()

	return copyPeers(g.peers)
}

func (g *Group) Reconcile(ctx context.Context, peerTTL time.Duration) ([]*peer.Peer, []*peer.Peer, []*peer.Peer) {
	if peerTTL <= 0 {
		peerTTL = 5 * time.Second
	}

	now := time.Now()

	g.mu.Lock()
	tmp := g.peers[:0]
	var newPeers []*peer.Peer
	var lostPeers []*peer.Peer

	for _, p := range g.peers {
		if p.LastSeen.IsZero() {
			log.Debugf("new peer %v", p)
			newPeers = append(newPeers, p)
			p.LastSeen = now
		}

		if p.LastSeen.Add(peerTTL).Before(now) {
			log.Debugf("lost peer %v", p)
			lostPeers = append(lostPeers, p)
		} else {
			tmp = append(tmp, p)
		}
	}

	g.peers = tmp
	peers := copyPeers(g.peers)
	g.mu.Unlock()

	for _, p := range newPeers {
		for _, h := range g.Handlers {
			if err := h.NewPeer(ctx, p); err != nil {
				log.Warnf("Failed running new-peer hook for group '%s' of handler '%s': %v", g.Name, reflect.TypeOf(h).String(), err)
			}
		}
	}

	for _, p := range lostPeers {
		for _, h := range g.Handlers {
			if err := h.LostPeer(ctx, p); err != nil {
				log.Warnf("Failed running lost-peer hook for group '%s' of handler '%s': %v", g.Name, reflect.TypeOf(h).String(), err)
			}
		}
	}

	return peers, copyPeers(newPeers), copyPeers(lostPeers)
}

func copyPeers(peers []*peer.Peer) []*peer.Peer {
	var tmp []*peer.Peer
	for _, p := range peers {
		tmpP := *p
		tmp = append(tmp, &tmpP)
	}
	return tmp
}
