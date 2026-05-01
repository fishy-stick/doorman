package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
}

type ServerConfig struct {
	Port          string `yaml:"port"`
	TrustProxy    bool   `yaml:"trust_proxy"`
	DB            string `yaml:"db"`
	AdminPassword string `yaml:"admin_password"`
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

func (c *Config) setDefaults() {
	if c.Server.Port == "" {
		c.Server.Port = ":8080"
	} else if c.Server.Port[0] != ':' {
		c.Server.Port = ":" + c.Server.Port
	}

	if !c.Server.TrustProxy {
		c.Server.TrustProxy = true
	}

	if c.Server.DB == "" {
		c.Server.DB = "doorman.db"
	}
}

func (c *Config) validate() error {
	if c.Server.AdminPassword == "" {
		return fmt.Errorf("server.admin_password is required")
	}
	return nil
}
