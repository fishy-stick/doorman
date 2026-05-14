package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSetsDefaultsWhenOptionalFieldsAreOmitted(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "server:\n  port: 8080\n")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != ":8080" {
		t.Fatalf("Port = %q, want %q", cfg.Server.Port, ":8080")
	}
	if !cfg.Server.TrustProxyEnabled() {
		t.Fatal("TrustProxyEnabled() = false, want true")
	}
	if cfg.Server.DB != "doorman.db" {
		t.Fatalf("DB = %q, want %q", cfg.Server.DB, "doorman.db")
	}
	if cfg.Server.PublicURL != "http://your-server:8080" {
		t.Fatalf("PublicURL = %q, want %q", cfg.Server.PublicURL, "http://your-server:8080")
	}
}

func TestLoadPreservesExplicitTrustProxyFalse(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "server:\n  port: 9090\n  trust_proxy: false\n  db: custom.db\n  public_url: https://Example.com/prefix/\n")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != ":9090" {
		t.Fatalf("Port = %q, want %q", cfg.Server.Port, ":9090")
	}
	if cfg.Server.TrustProxyEnabled() {
		t.Fatal("TrustProxyEnabled() = true, want false")
	}
	if cfg.Server.DB != "custom.db" {
		t.Fatalf("DB = %q, want %q", cfg.Server.DB, "custom.db")
	}
	if cfg.Server.PublicURL != "https://example.com/prefix" {
		t.Fatalf("PublicURL = %q, want %q", cfg.Server.PublicURL, "https://example.com/prefix")
	}
}

func TestLoadNormalizesPublicURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "full url", raw: "https://EXAMPLE.com:8443/prefix/", want: "https://example.com:8443/prefix"},
		{name: "bare host", raw: "example.com", want: "http://example.com"},
		{name: "bare host with path", raw: "www.abc.com/prefix", want: "http://www.abc.com/prefix"},
		{name: "root path", raw: "https://example.com/", want: "https://example.com"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := writeConfigFile(t, "server:\n  public_url: \""+tc.raw+"\"\n")

			cfg, err := Load(path)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.Server.PublicURL != tc.want {
				t.Fatalf("PublicURL = %q, want %q", cfg.Server.PublicURL, tc.want)
			}
		})
	}
}

func TestLoadRejectsInvalidPublicURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "unsupported scheme", raw: "ftp://example.com"},
		{name: "missing host", raw: "https:///prefix"},
		{name: "query", raw: "https://example.com/prefix?x=1"},
		{name: "fragment", raw: "https://example.com/prefix#section"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := writeConfigFile(t, "server:\n  public_url: \""+tc.raw+"\"\n")

			if _, err := Load(path); err == nil {
				t.Fatal("Load() error = nil, want error")
			}
		})
	}
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	return path
}
