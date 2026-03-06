package main

import (
	"context"
	"flag"
	"reflect"
	"time"

	"github.com/ravenix/peerd/internal/config"
	"github.com/ravenix/peerd/internal/group"
	"github.com/ravenix/peerd/pkg/explorer"
	"github.com/ravenix/peerd/pkg/plugin"
	_ "github.com/ravenix/peerd/plugin/exec"
	_ "github.com/ravenix/peerd/plugin/keepalived"
	_ "github.com/ravenix/peerd/plugin/kubernetes"
	_ "github.com/ravenix/peerd/plugin/multicast"
	_ "github.com/ravenix/peerd/plugin/netlink"
	_ "github.com/ravenix/peerd/plugin/template"
	log "github.com/sirupsen/logrus"
)

var conf string

func init() {
	flag.StringVar(&conf, "config", "/etc/peerd/peerd.yaml", "Configuration file")
}

func runGroupCycle(g *group.Group, cadence explorer.Cadence) {
	for _, h := range g.Handlers {
		if err := h.PreExploration(context.Background(), g.GetPeers()); err != nil {
			log.Warnf("Failed running pre-exploration hook for group '%s' of handler '%s': %v", g.Name, reflect.TypeOf(h).String(), err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), cadence.ExploreTimeout)
	defer cancel()

	for _, e := range g.Explorers {
		go func(currentExplorer explorer.Explorer) {
			if err := currentExplorer.Explore(ctx, g); err != nil {
				log.Warnf("Failed exploring peers for group '%s' with explorer '%s': %v", g.Name, reflect.TypeOf(currentExplorer).String(), err)
			}
		}(e)
	}

	<-ctx.Done()

	peers, newPeers, lostPeers := g.Reconcile(context.Background(), cadence.PeerTTL)

	for _, h := range g.Handlers {
		if err := h.PostExploration(context.Background(), peers, newPeers, lostPeers); err != nil {
			log.Warnf("Failed running post-exploration hook for group '%s' of handler '%s': %v", g.Name, reflect.TypeOf(h).String(), err)
		}
	}
}

func runGroup(g *group.Group) {
	cadence := explorer.ResolveCadence(g.Explorers)
	log.Infof(
		"Group '%s' cadence interval=%s timeout=%s peer_ttl=%s",
		g.Name,
		cadence.ExploreInterval,
		cadence.ExploreTimeout,
		cadence.PeerTTL,
	)

	ticker := time.NewTicker(cadence.ExploreInterval)
	defer ticker.Stop()

	for {
		runGroupCycle(g, cadence)
		<-ticker.C
	}
}

func main() {
	flag.Parse()
	log.Infof("Configuration %v", conf)
	cfg, err := config.NewConfig(conf)

	if err != nil {
		log.Fatalf("Error while loading configuration file: %v", err)
	}

	log.SetLevel(cfg.LogLevel)

	groups := make([]*group.Group, 0)

	for cgn, cg := range cfg.Groups {
		currentGroup := &group.Group{
			Name: cgn,
		}

		for _, ce := range cg.Explorers {
			if e, err := plugin.InitializeExplorer(ce.Name, &ce.Configuration); err == nil {
				currentGroup.Explorers = append(currentGroup.Explorers, e)
			} else {
				log.Fatalf("Could not initialize explorer '%s' for group '%s': %v", ce.Name, cgn, err)
			}
		}

		for _, ch := range cg.Handlers {
			if h, err := plugin.InitializeHandler(ch.Name, &ch.Configuration); err == nil {
				currentGroup.Handlers = append(currentGroup.Handlers, h)
			} else {
				log.Fatalf("Could not initialize handler '%s' for group '%s': %v", ch.Name, cgn, err)
			}
		}

		groups = append(groups, currentGroup)
	}

	for _, g := range groups {
		for _, e := range g.Explorers {
			go func(currentExplorer explorer.Explorer, currentGroup *group.Group) {
				if err := currentExplorer.Run(context.Background()); err != nil {
					log.Fatalf("Explorer '%s' for group '%s' could not be run: %v", reflect.TypeOf(currentExplorer).String(), currentGroup.Name, err)
				}
			}(e, g)
		}

		go runGroup(g)
	}

	select {}
}
