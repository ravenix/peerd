package explorer

import (
	"context"
	"testing"
	"time"
)

type testExplorer struct{}

func (testExplorer) Run(context.Context) error                       { return nil }
func (testExplorer) Explore(context.Context, DiscoveryHandler) error { return nil }

type testCadenceExplorer struct {
	c Cadence
}

func (e testCadenceExplorer) Run(context.Context) error                       { return nil }
func (e testCadenceExplorer) Explore(context.Context, DiscoveryHandler) error { return nil }
func (e testCadenceExplorer) Cadence() Cadence                                { return e.c }

func TestResolveCadenceDefaults(t *testing.T) {
	cadence := ResolveCadence(nil)
	expected := DefaultCadence()

	if cadence != expected {
		t.Fatalf("expected default cadence %v, got %v", expected, cadence)
	}
}

func TestResolveCadenceAggregatesExplorerProfiles(t *testing.T) {
	cadence := ResolveCadence([]Explorer{
		testExplorer{},
		testCadenceExplorer{
			c: Cadence{
				ExploreInterval: 500 * time.Millisecond,
				ExploreTimeout:  time.Second,
				PeerTTL:         10 * time.Second,
			},
		},
		testCadenceExplorer{
			c: Cadence{
				ExploreInterval: 250 * time.Millisecond,
				ExploreTimeout:  3 * time.Second,
				PeerTTL:         7 * time.Second,
			},
		},
	})

	if cadence.ExploreInterval != 250*time.Millisecond {
		t.Fatalf("expected min explore interval 250ms, got %s", cadence.ExploreInterval)
	}

	if cadence.ExploreTimeout != 3*time.Second {
		t.Fatalf("expected max explore timeout 3s, got %s", cadence.ExploreTimeout)
	}

	if cadence.PeerTTL != 10*time.Second {
		t.Fatalf("expected max peer ttl 10s, got %s", cadence.PeerTTL)
	}
}

func TestResolveCadenceSingleProviderUsesProviderValues(t *testing.T) {
	cadence := ResolveCadence([]Explorer{
		testCadenceExplorer{
			c: Cadence{
				ExploreInterval: 100 * time.Millisecond,
				ExploreTimeout:  80 * time.Millisecond,
				PeerTTL:         300 * time.Millisecond,
			},
		},
	})

	if cadence.ExploreInterval != 100*time.Millisecond {
		t.Fatalf("expected explore interval 100ms, got %s", cadence.ExploreInterval)
	}

	if cadence.ExploreTimeout != 80*time.Millisecond {
		t.Fatalf("expected explore timeout 80ms, got %s", cadence.ExploreTimeout)
	}

	if cadence.PeerTTL != 300*time.Millisecond {
		t.Fatalf("expected peer ttl 300ms, got %s", cadence.PeerTTL)
	}
}
