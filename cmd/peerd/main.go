package main

import (
	"context"
	"flag"
	"reflect"
	"time"

	"github.com/ravenix/peerd/internal/config"
	"github.com/ravenix/peerd/internal/group"
	"github.com/ravenix/peerd/pkg/plugin"
	_ "github.com/ravenix/peerd/plugin/exec"
	_ "github.com/ravenix/peerd/plugin/kubernetes"
	_ "github.com/ravenix/peerd/plugin/multicast"
	_ "github.com/ravenix/peerd/plugin/template"
	log "github.com/sirupsen/logrus"
)

var conf string

func init() {
	flag.StringVar(&conf, "config", "/etc/peerd/peerd.yaml", "Configuration file")
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
		group := &group.Group{
			Name: cgn,
		}

		for _, ce := range cg.Explorers {
			if e, err := plugin.InitializeExplorer(ce.Name, &ce.Configuration); err == nil {
				group.Explorers = append(group.Explorers, e)
			} else {
				log.Fatalf("Could not initialize explorer '%s' for group '%s': %v", ce.Name, cgn, err)
			}
		}

		for _, ch := range cg.Handlers {
			if h, err := plugin.InitializeHandler(ch.Name, &ch.Configuration); err == nil {
				group.Handlers = append(group.Handlers, h)
			} else {
				log.Fatalf("Could not initialize handler '%s' for group '%s': %v", ch.Name, cgn, err)
			}
		}

		groups = append(groups, group)
	}

	for _, g := range groups {
		for _, e := range g.Explorers {
			go func() {
				if err := e.Run(context.Background()); err != nil {
					log.Fatalf("Explorer '%s' for group '%s' could not be run: %v", reflect.TypeOf(e).String(), g.Name, err)
				}
			}()
		}
	}

	for {
		for _, g := range groups {
			for _, h := range g.Handlers {
				if err := h.PreExploration(context.Background(), g.GetPeers()); err != nil {
					log.Warnf("Failed running pre-exploration hook for group '%s' of handler '%s': %v", g.Name, reflect.TypeOf(h).String(), err)
				}
			}
		}

		ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
		for _, g := range groups {
			for _, e := range g.Explorers {
				go func() {
					if err := e.Explore(ctx, g); err != nil {
						log.Warnf("Failed exploring peers for group '%s' with explorer '%s': %v", g.Name, reflect.TypeOf(e).String(), err)
					}
				}()
			}
		}
		<-ctx.Done()

		for _, g := range groups {
			peers, newPeers, lostPeers := g.Reconcile(context.Background())

			for _, h := range g.Handlers {
				if err := h.PostExploration(context.Background(), peers, newPeers, lostPeers); err != nil {
					log.Warnf("Failed running post-exploration hook for group '%s' of handler '%s': %v", g.Name, reflect.TypeOf(h).String(), err)
				}
			}
		}
	}
}
