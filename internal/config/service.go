package config

import "gopkg.in/yaml.v3"

type ServiceConfig struct {
	Enabled  bool      `yaml:"enabled"`
	Host     string    `yaml:"host"`
	Port     int       `yaml:"port"`
	Settings yaml.Node `yaml:"settings,omitempty"`
}
