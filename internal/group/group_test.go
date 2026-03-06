package group

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ravenix/peerd/pkg/explorer"
)

func TestReconcileRespectsPeerTTL(t *testing.T) {
	g := &Group{Name: "test"}
	g.Discovered(&explorer.Discovery{
		IPv6Addr: net.ParseIP("fd00::1"),
		Port:     179,
	})

	_, newPeers, lostPeers := g.Reconcile(context.Background(), 100*time.Millisecond)
	if len(newPeers) != 1 {
		t.Fatalf("expected one new peer, got %d", len(newPeers))
	}

	if len(lostPeers) != 0 {
		t.Fatalf("expected no lost peers, got %d", len(lostPeers))
	}

	time.Sleep(20 * time.Millisecond)
	_, _, lostPeers = g.Reconcile(context.Background(), 10*time.Millisecond)
	if len(lostPeers) != 1 {
		t.Fatalf("expected one lost peer, got %d", len(lostPeers))
	}
}

func TestReconcileUsesDefaultTTLWhenUnset(t *testing.T) {
	g := &Group{Name: "test"}
	g.Discovered(&explorer.Discovery{
		IPv6Addr: net.ParseIP("fd00::2"),
		Port:     179,
	})

	_, _, _ = g.Reconcile(context.Background(), 0)

	time.Sleep(10 * time.Millisecond)
	_, _, lostPeers := g.Reconcile(context.Background(), 0)
	if len(lostPeers) != 0 {
		t.Fatalf("expected no lost peers with default ttl, got %d", len(lostPeers))
	}
}
