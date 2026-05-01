package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"doorman/internal/auth"
	"doorman/internal/store"

	"github.com/gin-gonic/gin"
)

func TestKnockHandlerTracksLatestIPAcrossRequests(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	network := &store.Network{
		Name:       "home",
		Token:      "token-1",
		DDNSConfig: "{}",
	}
	if err := s.CreateNetwork(network); err != nil {
		t.Fatalf("CreateNetwork() error = %v", err)
	}

	h := NewKnockHandler(s, false)
	router := gin.New()
	router.GET("/knock", func(c *gin.Context) {
		c.Set(string(auth.NetworkContextKey), network)
		h.Handle(c)
	})

	first := performRequest(t, router, http.MethodGet, "/knock", "", "198.51.100.10:1000", nil)
	assertKnockResponse(t, first, "198.51.100.10", true)

	firstKnock, err := s.GetLatestKnock(network.ID)
	if err != nil {
		t.Fatalf("GetLatestKnock(first) error = %v", err)
	}
	if firstKnock.PreviousIP != nil {
		t.Fatalf("PreviousIP on first knock = %v, want nil", *firstKnock.PreviousIP)
	}

	second := performRequest(t, router, http.MethodGet, "/knock", "", "198.51.100.11:1001", nil)
	assertKnockResponse(t, second, "198.51.100.11", true)

	secondKnock, err := s.GetLatestKnock(network.ID)
	if err != nil {
		t.Fatalf("GetLatestKnock(second) error = %v", err)
	}
	if secondKnock.PreviousIP == nil || *secondKnock.PreviousIP != "198.51.100.10" {
		t.Fatalf("PreviousIP on second knock = %v, want %q", secondKnock.PreviousIP, "198.51.100.10")
	}

	third := performRequest(t, router, http.MethodGet, "/knock", "", "198.51.100.11:1002", nil)
	assertKnockResponse(t, third, "198.51.100.11", false)

	thirdKnock, err := s.GetLatestKnock(network.ID)
	if err != nil {
		t.Fatalf("GetLatestKnock(third) error = %v", err)
	}
	if thirdKnock.PreviousIP == nil || *thirdKnock.PreviousIP != "198.51.100.11" {
		t.Fatalf("PreviousIP on third knock = %v, want %q", thirdKnock.PreviousIP, "198.51.100.11")
	}
}

func TestKnockHandlerHonorsDisabledTrustProxySetting(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	network := &store.Network{
		Name:       "home",
		Token:      "token-1",
		DDNSConfig: "{}",
	}
	if err := s.CreateNetwork(network); err != nil {
		t.Fatalf("CreateNetwork() error = %v", err)
	}

	h := NewKnockHandler(s, false)
	router := gin.New()
	router.GET("/knock", func(c *gin.Context) {
		c.Set(string(auth.NetworkContextKey), network)
		h.Handle(c)
	})

	resp := performRequest(t, router, http.MethodGet, "/knock", "", "198.51.100.20:1000", map[string]string{
		"X-Forwarded-For": "203.0.113.10",
		"X-Real-IP":       "203.0.113.11",
	})
	assertKnockResponse(t, resp, "198.51.100.20", true)
}

func assertKnockResponse(t *testing.T, recorder *httptest.ResponseRecorder, wantIP string, wantChanged bool) {
	t.Helper()

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var payload struct {
		IP      string `json:"ip"`
		Changed bool   `json:"changed"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if payload.IP != wantIP {
		t.Fatalf("ip = %q, want %q", payload.IP, wantIP)
	}
	if payload.Changed != wantChanged {
		t.Fatalf("changed = %v, want %v", payload.Changed, wantChanged)
	}
}
