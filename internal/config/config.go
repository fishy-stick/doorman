package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultPublicURL = "http://your-server:8080"

// Config holds the runtime configuration loaded from YAML.
type Config struct {
	Server ServerConfig `yaml:"server"`
}

// ServerConfig contains the HTTP and persistence settings for the server.
type ServerConfig struct {
	Port       string `yaml:"port"`
	TrustProxy *bool  `yaml:"trust_proxy"`
	DB         string `yaml:"db"`
	PublicURL  string `yaml:"public_url"`
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

	if strings.TrimSpace(c.Server.PublicURL) == "" {
		c.Server.PublicURL = defaultPublicURL
	}
}

func (c *Config) validate() error {
	publicURL, err := normalizePublicURL(c.Server.PublicURL)
	if err != nil {
		return err
	}
	c.Server.PublicURL = publicURL

	return nil
}

func boolPtr(v bool) *bool {
	return &v
}

func normalizePublicURL(rawURL string) (string, error) {
	value := strings.TrimSpace(rawURL)
	if value == "" {
		value = defaultPublicURL
	}

	if !strings.Contains(value, "://") {
		value = "http://" + value
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return "", fmt.Errorf("invalid server.public_url: %w", err)
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	parsed.Path = strings.TrimRight(parsed.Path, "/")

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("invalid server.public_url: scheme must be http or https")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("invalid server.public_url: host is required")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("invalid server.public_url: query and fragment are not supported")
	}

	return parsed.String(), nil
}
