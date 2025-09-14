package config

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	LogLevel log.Level        `yaml:"log_level"`
	Groups   map[string]Group `yaml:"groups"`
}

type Group struct {
	Explorers []Explorer `yaml:"explorers"`
	Handlers  []Handler  `yaml:"handlers"`
}

type Explorer struct {
	Name          string    `yaml:"name"`
	Configuration yaml.Node `yaml:"configuration"`
}

type Handler struct {
	Name          string    `yaml:"name"`
	Configuration yaml.Node `yaml:"configuration"`
}

func NewConfig(filename string) (*Config, error) {
	absFilename, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	yamlFile, err := os.ReadFile(absFilename)
	if err != nil {
		return nil, err
	}

	config := Config{
		LogLevel: log.InfoLevel,
	}

	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
