package store

import "testing"

func TestGetLatestIPReturnsLatestObservedIP(t *testing.T) {
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

	firstIP := "198.51.100.10"
	if err := s.InsertKnock(&Knock{NetworkID: network.ID, IP: firstIP}); err != nil {
		t.Fatalf("InsertKnock(first) error = %v", err)
	}

	secondIP := "198.51.100.11"
	if err := s.InsertKnock(&Knock{NetworkID: network.ID, IP: secondIP, PreviousIP: &firstIP}); err != nil {
		t.Fatalf("InsertKnock(second) error = %v", err)
	}

	got, err := s.GetLatestIP(network.ID)
	if err != nil {
		t.Fatalf("GetLatestIP() error = %v", err)
	}
	if got == nil || *got != secondIP {
		t.Fatalf("GetLatestIP() = %v, want %q", got, secondIP)
	}
}

func TestListNetworksUsesLatestKnockSnapshot(t *testing.T) {
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

	firstIP := "198.51.100.10"
	if err := s.InsertKnock(&Knock{NetworkID: network.ID, IP: firstIP, DDNSStatus: "success"}); err != nil {
		t.Fatalf("InsertKnock(first) error = %v", err)
	}

	secondIP := "198.51.100.11"
	if err := s.InsertKnock(&Knock{NetworkID: network.ID, IP: secondIP, PreviousIP: &firstIP, DDNSStatus: "skipped"}); err != nil {
		t.Fatalf("InsertKnock(second) error = %v", err)
	}

	networks, err := s.ListNetworks()
	if err != nil {
		t.Fatalf("ListNetworks() error = %v", err)
	}
	if len(networks) != 1 {
		t.Fatalf("len(ListNetworks()) = %d, want 1", len(networks))
	}

	got := networks[0]
	if got.CurrentIP == nil || *got.CurrentIP != secondIP {
		t.Fatalf("CurrentIP = %v, want %q", got.CurrentIP, secondIP)
	}
	if got.PreviousIP == nil || *got.PreviousIP != firstIP {
		t.Fatalf("PreviousIP = %v, want %q", got.PreviousIP, firstIP)
	}
	if got.DDNSStatus == nil || *got.DDNSStatus != "skipped" {
		t.Fatalf("DDNSStatus = %v, want %q", got.DDNSStatus, "skipped")
	}
}
