package exec

import (
	"bytes"
	"context"
	"fmt"
	osexec "os/exec"

	"github.com/ravenix/peerd/internal/peer"
	"github.com/ravenix/peerd/pkg/handler"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type commandHandler struct {
	c *commandHandlerConfig
}

type commandHandlerConfig struct {
	Command           string   `yaml:"command"`
	Args              []string `yaml:"args"`
	OnPreExploration  bool     `yaml:"on_pre_exploration"`
	OnNewPeer         bool     `yaml:"on_new_peer"`
	OnLostPeer        bool     `yaml:"on_lost_peer"`
	OnPostExploration struct {
		Always    bool `yaml:"always"`
		NewPeers  bool `yaml:"new_peers"`
		LostPeers bool `yaml:"lost_peers"`
	} `yaml:"on_post_exploration"`
}

func commandHandlerInitializer(yamlConfig *yaml.Node) (handler.Handler, error) {
	var config commandHandlerConfig
	if err := yamlConfig.Decode(&config); err != nil {
		return nil, err
	}

	return newcommandHandler(&config)
}

func newcommandHandler(config *commandHandlerConfig) (*commandHandler, error) {
	if config.Command == "" {
		return nil, fmt.Errorf("command must not be empty")
	}

	r := &commandHandler{
		c: config,
	}

	return r, nil
}

func (r *commandHandler) PreExploration(context.Context, []*peer.Peer) error {
	if r.c.OnPreExploration {
		return r.run()
	}

	return nil
}

func (r *commandHandler) NewPeer(context.Context, *peer.Peer) error {
	if r.c.OnNewPeer {
		return r.run()
	}

	return nil
}

func (r *commandHandler) LostPeer(context.Context, *peer.Peer) error {
	if r.c.OnLostPeer {
		return r.run()
	}

	return nil
}

func (r *commandHandler) PostExploration(ctx context.Context, peers []*peer.Peer, newPeers []*peer.Peer, lostPeers []*peer.Peer) error {
	if r.c.OnPostExploration.Always || (r.c.OnPostExploration.NewPeers && len(newPeers) > 0) || (r.c.OnPostExploration.LostPeers && len(lostPeers) > 0) {
		return r.run()
	}

	return nil
}

func (r *commandHandler) run() error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := osexec.Command(r.c.Command, r.c.Args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err == nil {
		log.Debugf("Command '%s' with args %v ran successfully, stdout=%v, stderr=%v", r.c.Command, r.c.Args, stdout.String(), stderr.String())
	} else {
		log.Warnf("Command '%s' with args %v failed, stdout=%v, stderr=%v", r.c.Command, r.c.Args, stdout.String(), stderr.String())
	}

	return err
}
