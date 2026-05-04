package ddns

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestGetProviderReturnsDNSPodProvider(t *testing.T) {
	provider, err := GetProvider("dnspod")
	if err != nil {
		t.Fatalf("GetProvider(dnspod) error = %v", err)
	}
	if _, ok := provider.(*DNSPodProvider); !ok {
		t.Fatalf("GetProvider(dnspod) = %T, want *DNSPodProvider", provider)
	}
}

func TestGetProviderRejectsUnsupportedProvider(t *testing.T) {
	provider, err := GetProvider("cloudflare")
	if err == nil {
		t.Fatal("GetProvider(cloudflare) error = nil, want error")
	}
	if provider != nil {
		t.Fatalf("GetProvider(cloudflare) provider = %T, want nil", provider)
	}
}

func TestDNSPodUpdateRejectsInvalidJSON(t *testing.T) {
	provider := &DNSPodProvider{}

	result, err := provider.Update("{", "198.51.100.10")
	if err == nil {
		t.Fatal("Update() error = nil, want error")
	}
	if result != nil {
		t.Fatalf("Update() result = %#v, want nil", result)
	}
}

func TestDNSPodUpdateRejectsMissingFields(t *testing.T) {
	provider := &DNSPodProvider{}

	result, err := provider.Update(`{"domain":"example.com"}`, "198.51.100.10")
	if err == nil {
		t.Fatal("Update() error = nil, want error")
	}
	if result != nil {
		t.Fatalf("Update() result = %#v, want nil", result)
	}
}

func TestDNSPodUpdateSkipsWhenIPUnchanged(t *testing.T) {
	restore := stubDefaultTransport(t,
		expectedRequest{
			path:     "/Record.List",
			contains: []string{"domain=example.com", "sub_domain=home"},
			response: `{"status":{"code":"1","message":"ok"},"records":[{"id":"42","name":"home","value":"198.51.100.10","type":"A"}]}`,
		},
	)
	defer restore()

	provider := &DNSPodProvider{}
	result, err := provider.Update(validDNSPodConfig(), "198.51.100.10")
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if !result.Skipped || result.SkipReason != "ip unchanged" {
		t.Fatalf("Update() skipped = %v / %q, want true / %q", result.Skipped, result.SkipReason, "ip unchanged")
	}
	if result.Updated {
		t.Fatal("Update() Updated = true, want false")
	}
}

func TestDNSPodUpdateUpdatesExistingRecord(t *testing.T) {
	restore := stubDefaultTransport(t,
		expectedRequest{
			path:     "/Record.List",
			contains: []string{"domain=example.com", "sub_domain=home"},
			response: `{"status":{"code":"1","message":"ok"},"records":[{"id":"42","name":"home","value":"198.51.100.1","type":"A"}]}`,
		},
		expectedRequest{
			path:     "/Record.Ddns",
			contains: []string{"record_id=42", "value=198.51.100.10"},
			response: `{"status":{"code":"1","message":"ok"}}`,
		},
	)
	defer restore()

	provider := &DNSPodProvider{}
	result, err := provider.Update(validDNSPodConfig(), "198.51.100.10")
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if !result.Updated {
		t.Fatal("Update() Updated = false, want true")
	}
	if result.OldIP != "198.51.100.1" || result.NewIP != "198.51.100.10" {
		t.Fatalf("Update() old/new = %q/%q, want %q/%q", result.OldIP, result.NewIP, "198.51.100.1", "198.51.100.10")
	}
}

func TestDNSPodUpdateCreatesRecordWhenMissing(t *testing.T) {
	restore := stubDefaultTransport(t,
		expectedRequest{
			path:     "/Record.List",
			contains: []string{"domain=example.com", "sub_domain=home"},
			response: `{"status":{"code":"1","message":"ok"},"records":[]}`,
		},
		expectedRequest{
			path:     "/Record.Create",
			contains: []string{"value=198.51.100.10", "record_line=%E9%BB%98%E8%AE%A4"},
			response: `{"status":{"code":"1","message":"ok"}}`,
		},
	)
	defer restore()

	provider := &DNSPodProvider{}
	result, err := provider.Update(validDNSPodConfig(), "198.51.100.10")
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if !result.Updated {
		t.Fatal("Update() Updated = false, want true")
	}
	if result.OldIP != "" || result.NewIP != "198.51.100.10" {
		t.Fatalf("Update() old/new = %q/%q, want empty/%q", result.OldIP, result.NewIP, "198.51.100.10")
	}
}

func TestDNSPodUpdateReturnsProviderError(t *testing.T) {
	restore := stubDefaultTransport(t,
		expectedRequest{
			path:     "/Record.List",
			response: `{"status":{"code":"10","message":"invalid token"},"records":[]}`,
		},
	)
	defer restore()

	provider := &DNSPodProvider{}
	result, err := provider.Update(validDNSPodConfig(), "198.51.100.10")
	if err == nil {
		t.Fatal("Update() error = nil, want error")
	}
	if result != nil {
		t.Fatalf("Update() result = %#v, want nil", result)
	}
	if !strings.Contains(err.Error(), "invalid token") {
		t.Fatalf("Update() error = %q, want substring %q", err.Error(), "invalid token")
	}
}

type expectedRequest struct {
	path     string
	contains []string
	response string
}

func stubDefaultTransport(t *testing.T, expected ...expectedRequest) func() {
	t.Helper()

	original := http.DefaultTransport
	requests := append([]expectedRequest(nil), expected...)

	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if len(requests) == 0 {
			t.Fatalf("unexpected request: %s", req.URL.Path)
		}

		current := requests[0]
		requests = requests[1:]

		if req.URL.Path != current.path {
			t.Fatalf("request path = %q, want %q", req.URL.Path, current.path)
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("ReadAll(req.Body) error = %v", err)
		}
		for _, fragment := range current.contains {
			if !strings.Contains(string(body), fragment) {
				t.Fatalf("request body = %q, want substring %q", string(body), fragment)
			}
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(current.response)),
		}, nil
	})

	return func() {
		http.DefaultTransport = original
		if len(requests) != 0 {
			t.Fatalf("unconsumed requests: %d", len(requests))
		}
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func validDNSPodConfig() string {
	return `{"domain":"example.com","record":"home","id":"abc","token":"def"}`
}
