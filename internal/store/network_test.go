package store

import (
	"errors"
	"testing"
)

func TestCreateNetworkRejectsEmptyToken(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	err := s.CreateNetwork(&Network{
		Name:       "home",
		DDNSConfig: "{}",
	})
	if !errors.Is(err, ErrInvalidNetwork) {
		t.Fatalf("CreateNetwork() error = %v, want ErrInvalidNetwork", err)
	}
}

func TestCreateNetworkRejectsDuplicateToken(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	first := &Network{
		Name:       "home",
		Token:      "shared-token",
		DDNSConfig: "{}",
	}
	if err := s.CreateNetwork(first); err != nil {
		t.Fatalf("CreateNetwork(first) error = %v", err)
	}

	second := &Network{
		Name:       "office",
		Token:      "shared-token",
		DDNSConfig: "{}",
	}
	err := s.CreateNetwork(second)
	if !errors.Is(err, ErrNetworkTokenConflict) {
		t.Fatalf("CreateNetwork(second) error = %v, want ErrNetworkTokenConflict", err)
	}
}

func TestCreateNetworkRejectsUnsupportedDDNSProvider(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	err := s.CreateNetwork(&Network{
		Name:        "home",
		Token:       "token-1",
		DDNSType:    "cloudflare",
		DDNSConfig:  "{}",
		DDNSEnabled: true,
	})
	if !errors.Is(err, ErrInvalidNetwork) {
		t.Fatalf("CreateNetwork() error = %v, want ErrInvalidNetwork", err)
	}
}

func TestCreateNetworkClearsDisabledDDNSSettings(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	network := &Network{
		Name:        "home",
		Token:       "token-1",
		DDNSEnabled: false,
		DDNSType:    "dnspod",
		DDNSConfig:  `{"domain":"","record":"","id":"","token":""}`,
	}
	if err := s.CreateNetwork(network); err != nil {
		t.Fatalf("CreateNetwork() error = %v", err)
	}

	got, err := s.GetNetwork(network.ID)
	if err != nil {
		t.Fatalf("GetNetwork() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetNetwork() = nil, want network")
	}
	if got.DDNSType != "" {
		t.Fatalf("DDNSType = %q, want empty", got.DDNSType)
	}
	if got.DDNSConfig != "{}" {
		t.Fatalf("DDNSConfig = %q, want {}", got.DDNSConfig)
	}
}

func TestCreateNetworkRejectsNonObjectDDNSConfig(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	err := s.CreateNetwork(&Network{
		Name:        "home",
		Token:       "token-1",
		DDNSEnabled: true,
		DDNSType:    "dnspod",
		DDNSConfig:  `[]`,
	})
	if !errors.Is(err, ErrInvalidNetwork) {
		t.Fatalf("CreateNetwork() error = %v, want ErrInvalidNetwork", err)
	}
}

func TestCreateNetworkRejectsInvalidDNSPodConfig(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	err := s.CreateNetwork(&Network{
		Name:        "home",
		Token:       "token-1",
		DDNSEnabled: true,
		DDNSType:    "dnspod",
		DDNSConfig:  `{"domain":"example.com","record":"","id":"abc","token":"def"}`,
	})
	if !errors.Is(err, ErrInvalidNetwork) {
		t.Fatalf("CreateNetwork() error = %v, want ErrInvalidNetwork", err)
	}
}

func TestUpdateNetworkPreservesCreatedAt(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	network := &Network{
		Name:       "home",
		Token:      "token-1",
		DDNSConfig: "{}",
	}
	if err := s.CreateNetwork(network); err != nil {
		t.Fatalf("CreateNetwork() error = %v", err)
	}

	createdAt := network.CreatedAt
	update := &Network{
		ID:          network.ID,
		Name:        "home-updated",
		Token:       "token-2",
		DDNSEnabled: true,
		DDNSType:    "dnspod",
		DDNSConfig:  `{"domain":"example.com","record":"home","id":"abc","token":"def"}`,
	}
	if err := s.UpdateNetwork(update); err != nil {
		t.Fatalf("UpdateNetwork() error = %v", err)
	}

	got, err := s.GetNetwork(network.ID)
	if err != nil {
		t.Fatalf("GetNetwork() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetNetwork() = nil, want network")
	}
	if !got.CreatedAt.Equal(createdAt) {
		t.Fatalf("CreatedAt changed: got %v, want %v", got.CreatedAt, createdAt)
	}
	if got.Name != "home-updated" {
		t.Fatalf("Name = %q, want %q", got.Name, "home-updated")
	}
	if got.Token != "token-2" {
		t.Fatalf("Token = %q, want %q", got.Token, "token-2")
	}
}

func TestUpdateNetworkRejectsUnsupportedDDNSProvider(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	network := &Network{
		Name:       "home",
		Token:      "token-1",
		DDNSConfig: "{}",
	}
	if err := s.CreateNetwork(network); err != nil {
		t.Fatalf("CreateNetwork() error = %v", err)
	}

	err := s.UpdateNetwork(&Network{
		ID:          network.ID,
		Name:        network.Name,
		Token:       network.Token,
		DDNSEnabled: true,
		DDNSType:    "cloudflare",
		DDNSConfig:  "{}",
	})
	if !errors.Is(err, ErrInvalidNetwork) {
		t.Fatalf("UpdateNetwork() error = %v, want ErrInvalidNetwork", err)
	}
}

func TestUpdateNetworkClearsDisabledDDNSSettings(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)

	network := &Network{
		Name:        "home",
		Token:       "token-1",
		DDNSEnabled: true,
		DDNSType:    "dnspod",
		DDNSConfig:  `{"domain":"example.com","record":"home","id":"abc","token":"def"}`,
	}
	if err := s.CreateNetwork(network); err != nil {
		t.Fatalf("CreateNetwork() error = %v", err)
	}

	if err := s.UpdateNetwork(&Network{
		ID:          network.ID,
		Name:        network.Name,
		Token:       network.Token,
		DDNSEnabled: false,
		DDNSType:    "dnspod",
		DDNSConfig:  `{"domain":"","record":"","id":"","token":""}`,
	}); err != nil {
		t.Fatalf("UpdateNetwork() error = %v", err)
	}

	got, err := s.GetNetwork(network.ID)
	if err != nil {
		t.Fatalf("GetNetwork() error = %v", err)
	}
	if got == nil {
		t.Fatal("GetNetwork() = nil, want network")
	}
	if got.DDNSEnabled {
		t.Fatal("DDNSEnabled = true, want false")
	}
	if got.DDNSType != "" {
		t.Fatalf("DDNSType = %q, want empty", got.DDNSType)
	}
	if got.DDNSConfig != "{}" {
		t.Fatalf("DDNSConfig = %q, want {}", got.DDNSConfig)
	}
}
