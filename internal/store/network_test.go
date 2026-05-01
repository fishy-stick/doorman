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
		DDNSConfig:  `{"domain":"example.com"}`,
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
