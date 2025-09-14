package template

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/ravenix/peerd/internal/peer"
	"github.com/ravenix/peerd/pkg/handler"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type fileHandler struct {
	tpl            *template.Template
	outputFilename string
	outputFilemode os.FileMode
}

type fileHandlerConfig struct {
	Filename         string      `yaml:"filename"`
	Mode             os.FileMode `yaml:"mode"`
	TemplateFilename string      `yaml:"template_filename"`
	TemplateString   string      `yaml:"template_string"`
}

func fileHandlerInitializer(yamlConfig *yaml.Node) (handler.Handler, error) {
	var config fileHandlerConfig
	if err := yamlConfig.Decode(&config); err != nil {
		return nil, err
	}

	return newFileHandler(&config)
}

func newFileHandler(config *fileHandlerConfig) (*fileHandler, error) {
	if config.Filename == "" {
		return nil, fmt.Errorf("output filename must not be empty")
	}

	if config.TemplateFilename == "" && config.TemplateString == "" {
		return nil, fmt.Errorf("template filename and template string cannot both be empty")
	}

	if config.TemplateFilename != "" && config.TemplateString != "" {
		return nil, fmt.Errorf("template filename and template string cannot both be set")
	}

	r := &fileHandler{}

	var tplContents string

	if config.TemplateFilename != "" {
		tplContentsFile, err := os.ReadFile(config.TemplateFilename)
		if err != nil {
			return nil, err
		}

		tplContents = string(tplContentsFile)
	} else {
		tplContents = config.TemplateString
	}

	tpl, err := template.New(config.TemplateFilename).Parse(tplContents)
	if err != nil {
		return nil, err
	}

	r.outputFilename = config.Filename
	r.outputFilemode = config.Mode
	r.tpl = tpl

	return r, nil
}

func (r *fileHandler) PreExploration(context.Context, []*peer.Peer) error {
	return nil
}

func (r *fileHandler) NewPeer(context.Context, *peer.Peer) error {
	return nil
}

func (r *fileHandler) LostPeer(context.Context, *peer.Peer) error {
	return nil
}

func (r *fileHandler) PostExploration(ctx context.Context, peers []*peer.Peer, newPeers []*peer.Peer, lostPeers []*peer.Peer) error {
	var tplBuff bytes.Buffer
	tplContext := &TemplateContext{Peers: peers}

	if err := r.tpl.Execute(&tplBuff, tplContext); err != nil {
		log.Debugf("Failed to render template with context %v: %v", tplContext, err)
		return err
	}

	if err := os.WriteFile(r.outputFilename, tplBuff.Bytes(), r.outputFilemode); err != nil {
		log.Debugf("Failed to write file '%s' with mode %o: %v", r.outputFilename, r.outputFilemode, err)
		return err
	}

	return nil
}
