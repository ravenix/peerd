package keepalived

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/ravenix/peerd/pkg/explorer"
)

type fakeDiscoveryHandler struct {
	discoveries []*explorer.Discovery
}

func (h *fakeDiscoveryHandler) Discovered(d *explorer.Discovery) {
	h.discoveries = append(h.discoveries, d)
}

func TestExploreDiscoveredWhenStateIsMaster(t *testing.T) {
	e, err := newInstanceExplorer(&instanceExplorerConfig{
		Interface:       "eth0",
		VirtualRouterID: 42,
		PeerIPv4:        "10.0.0.2",
		PeerIPv6:        "fd00::2",
		Port:            179,
	})
	if err != nil {
		t.Fatalf("unexpected error creating explorer: %v", err)
	}

	e.readState = func(context.Context, string) (int64, error) {
		return keepalivedMasterState, nil
	}

	h := &fakeDiscoveryHandler{}
	if err := e.Explore(context.Background(), h); err != nil {
		t.Fatalf("unexpected explore error: %v", err)
	}

	if len(h.discoveries) != 1 {
		t.Fatalf("expected exactly one discovery, got %d", len(h.discoveries))
	}

	discovery := h.discoveries[0]
	if !discovery.IPv4Addr.Equal(net.ParseIP("10.0.0.2").To4()) {
		t.Fatalf("unexpected ipv4 address: %v", discovery.IPv4Addr)
	}

	if !discovery.IPv6Addr.Equal(net.ParseIP("fd00::2").To16()) {
		t.Fatalf("unexpected ipv6 address: %v", discovery.IPv6Addr)
	}

	if discovery.Port != 179 {
		t.Fatalf("unexpected port: %d", discovery.Port)
	}
}

func TestExploreNoDiscoveryWhenStateIsNotMaster(t *testing.T) {
	e, err := newInstanceExplorer(&instanceExplorerConfig{
		Interface:       "eth0",
		VirtualRouterID: 42,
	})
	if err != nil {
		t.Fatalf("unexpected error creating explorer: %v", err)
	}

	e.readState = func(context.Context, string) (int64, error) {
		return 1, nil
	}

	h := &fakeDiscoveryHandler{}
	if err := e.Explore(context.Background(), h); err != nil {
		t.Fatalf("unexpected explore error: %v", err)
	}

	if len(h.discoveries) != 0 {
		t.Fatalf("expected no discoveries, got %d", len(h.discoveries))
	}
}

func TestExploreReturnsReadStateError(t *testing.T) {
	e, err := newInstanceExplorer(&instanceExplorerConfig{
		Interface:       "eth0",
		VirtualRouterID: 42,
	})
	if err != nil {
		t.Fatalf("unexpected error creating explorer: %v", err)
	}

	expectedErr := errors.New("read state failed")
	e.readState = func(context.Context, string) (int64, error) {
		return 0, expectedErr
	}

	h := &fakeDiscoveryHandler{}
	if err := e.Explore(context.Background(), h); !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestNewInstanceExplorerAllowsMissingOptionalPeerFields(t *testing.T) {
	e, err := newInstanceExplorer(&instanceExplorerConfig{
		Interface:       "eth0",
		VirtualRouterID: 42,
	})
	if err != nil {
		t.Fatalf("unexpected error creating explorer: %v", err)
	}

	if e.peerIPv4 != nil {
		t.Fatalf("expected nil peer ipv4, got %v", e.peerIPv4)
	}

	if e.peerIPv6 != nil {
		t.Fatalf("expected nil peer ipv6, got %v", e.peerIPv6)
	}

	if e.port != 0 {
		t.Fatalf("expected port 0, got %d", e.port)
	}
}

func TestNewInstanceExplorerRejectsInvalidOptionalIPs(t *testing.T) {
	_, err := newInstanceExplorer(&instanceExplorerConfig{
		Interface:       "eth0",
		VirtualRouterID: 42,
		PeerIPv4:        "not-an-ip",
	})
	if err == nil {
		t.Fatalf("expected error for invalid peer_ipv4")
	}

	_, err = newInstanceExplorer(&instanceExplorerConfig{
		Interface:       "eth0",
		VirtualRouterID: 42,
		PeerIPv6:        "10.0.0.1",
	})
	if err == nil {
		t.Fatalf("expected error for invalid peer_ipv6")
	}
}
