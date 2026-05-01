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
}

func TestLoadPreservesExplicitTrustProxyFalse(t *testing.T) {
	t.Parallel()

	path := writeConfigFile(t, "server:\n  port: 9090\n  trust_proxy: false\n  db: custom.db\n")

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
}

func writeConfigFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	return path
}
