package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the runtime configuration loaded from YAML.
type Config struct {
	Server ServerConfig `yaml:"server"`
}

// ServerConfig contains the HTTP and persistence settings for the server.
type ServerConfig struct {
	Port       string `yaml:"port"`
	TrustProxy *bool  `yaml:"trust_proxy"`
	DB         string `yaml:"db"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	cfg.setDefaults()

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// TrustProxyEnabled returns the effective proxy-trust setting after defaults.
func (c ServerConfig) TrustProxyEnabled() bool {
	return c.TrustProxy == nil || *c.TrustProxy
}

func (c *Config) setDefaults() {
	if c.Server.Port == "" {
		c.Server.Port = ":8080"
	} else if c.Server.Port[0] != ':' {
		c.Server.Port = ":" + c.Server.Port
	}

	if c.Server.TrustProxy == nil {
		c.Server.TrustProxy = boolPtr(true)
	}

	if c.Server.DB == "" {
		c.Server.DB = "doorman.db"
	}
}

func (c *Config) validate() error {
	return nil
}

func boolPtr(v bool) *bool {
	return &v
}
