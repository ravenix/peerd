package handler

import (
	"context"

	"github.com/ravenix/peerd/internal/peer"
	"gopkg.in/yaml.v3"
)

type Initializer func(*yaml.Node) (Handler, error)
type Handler interface {
	PreExploration(context.Context, []*peer.Peer) error
	NewPeer(context.Context, *peer.Peer) error
	LostPeer(context.Context, *peer.Peer) error
	PostExploration(context.Context, []*peer.Peer, []*peer.Peer, []*peer.Peer) error
}
