package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"doorman/internal/auth"
	"doorman/internal/store"

	"github.com/gin-gonic/gin"
)

func TestListNetworksReturnsEmptyArray(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	h := NewAdminHandler(s, auth.NewSessionManager())
	router := gin.New()
	router.GET("/admin/api/networks", h.ListNetworks)

	resp := performRequest(t, router, http.MethodGet, "/admin/api/networks", "", "127.0.0.1:1234", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusOK, resp.Body.String())
	}

	var body struct {
		Networks []store.NetworkSummary `json:"networks"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body.Networks == nil {
		t.Fatal("networks = nil, want empty array")
	}
	if len(body.Networks) != 0 {
		t.Fatalf("len(networks) = %d, want 0", len(body.Networks))
	}
}

func TestListKnocksReturnsEmptyRecordsArray(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	h := NewAdminHandler(s, auth.NewSessionManager())
	router := gin.New()
	router.GET("/admin/api/networks/:id/knocks", h.ListKnocks)

	resp := performRequest(t, router, http.MethodGet, "/admin/api/networks/1/knocks", "", "127.0.0.1:1234", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusOK, resp.Body.String())
	}

	var body struct {
		Records []store.Knock `json:"records"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body.Records == nil {
		t.Fatal("records = nil, want empty array")
	}
	if len(body.Records) != 0 {
		t.Fatalf("len(records) = %d, want 0", len(body.Records))
	}
}

func TestSessionReturnsAuthenticatedForValidSession(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	sm := auth.NewSessionManager()
	token, err := sm.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	h := NewAdminHandler(s, sm)
	router := gin.New()
	protected := router.Group("/admin/api")
	protected.Use(auth.AdminAuth(sm))
	protected.GET("/session", h.Session)

	resp := performRequest(t, router, http.MethodGet, "/admin/api/session", "", "127.0.0.1:1234", map[string]string{
		"Cookie": auth.SessionCookieName + "=" + token,
	})
	if resp.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusOK, resp.Body.String())
	}
	if body := resp.Body.String(); body != "{\"authenticated\":true}" {
		t.Fatalf("body = %s, want %s", body, `{"authenticated":true}`)
	}
}

func TestSessionReturnsUnauthorizedWithoutValidSession(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	sm := auth.NewSessionManager()

	h := NewAdminHandler(s, sm)
	router := gin.New()
	protected := router.Group("/admin/api")
	protected.Use(auth.AdminAuth(sm))
	protected.GET("/session", h.Session)

	resp := performRequest(t, router, http.MethodGet, "/admin/api/session", "", "127.0.0.1:1234", nil)
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusUnauthorized, resp.Body.String())
	}
	if body := resp.Body.String(); body != "{\"error\":\"not authenticated\"}" {
		t.Fatalf("body = %s, want %s", body, `{"error":"not authenticated"}`)
	}
}

func TestChangePasswordInvalidatesExistingSessions(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	if _, err := s.InitAdmin(); err != nil {
		t.Fatalf("InitAdmin() error = %v", err)
	}
	if err := s.UpdatePassword("old-password"); err != nil {
		t.Fatalf("UpdatePassword(old) error = %v", err)
	}

	sm := auth.NewSessionManager()
	token, err := sm.Create()
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	h := NewAdminHandler(s, sm)
	router := gin.New()
	router.PUT("/admin/api/password", h.ChangePassword)

	resp := performRequest(t, router, http.MethodPut, "/admin/api/password", `{"old_password":"old-password","new_password":"new-password"}`, "127.0.0.1:1234", map[string]string{
		"Content-Type": "application/json",
	})
	if resp.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusOK, resp.Body.String())
	}

	if sm.Validate(token) {
		t.Fatal("session is still valid after ChangePassword()")
	}

	valid, err := s.VerifyPassword("new-password")
	if err != nil {
		t.Fatalf("VerifyPassword(new) error = %v", err)
	}
	if !valid {
		t.Fatal("VerifyPassword(new) = false, want true")
	}
}

func TestCreateNetworkReturnsBadRequestForInvalidPayload(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	h := NewAdminHandler(s, auth.NewSessionManager())
	router := gin.New()
	router.POST("/admin/api/networks", h.CreateNetwork)

	resp := performRequest(t, router, http.MethodPost, "/admin/api/networks", `{"name":"home","token":"","ddns_config":"{}"}`, "127.0.0.1:1234", map[string]string{
		"Content-Type": "application/json",
	})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusBadRequest, resp.Body.String())
	}
}

func TestCreateNetworkReturnsConflictForDuplicateToken(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	if err := s.CreateNetwork(&store.Network{
		Name:       "home",
		Token:      "shared-token",
		DDNSConfig: "{}",
	}); err != nil {
		t.Fatalf("CreateNetwork(seed) error = %v", err)
	}

	h := NewAdminHandler(s, auth.NewSessionManager())
	router := gin.New()
	router.POST("/admin/api/networks", h.CreateNetwork)

	resp := performRequest(t, router, http.MethodPost, "/admin/api/networks", `{"name":"office","token":"shared-token","ddns_config":"{}"}`, "127.0.0.1:1234", map[string]string{
		"Content-Type": "application/json",
	})
	if resp.Code != http.StatusConflict {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusConflict, resp.Body.String())
	}
}

func TestCreateNetworkReturnsBadRequestForUnsupportedProvider(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	h := NewAdminHandler(s, auth.NewSessionManager())
	router := gin.New()
	router.POST("/admin/api/networks", h.CreateNetwork)

	resp := performRequest(t, router, http.MethodPost, "/admin/api/networks", `{"name":"home","token":"token-1","ddns_type":"cloudflare","ddns_config":"{}"}`, "127.0.0.1:1234", map[string]string{
		"Content-Type": "application/json",
	})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusBadRequest, resp.Body.String())
	}
}

func TestCreateNetworkReturnsBadRequestForInvalidDDNSConfigShape(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	h := NewAdminHandler(s, auth.NewSessionManager())
	router := gin.New()
	router.POST("/admin/api/networks", h.CreateNetwork)

	resp := performRequest(t, router, http.MethodPost, "/admin/api/networks", `{"name":"home","token":"token-1","ddns_config":"[]"}`, "127.0.0.1:1234", map[string]string{
		"Content-Type": "application/json",
	})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusBadRequest, resp.Body.String())
	}
}

func TestUpdateNetworkReturnsBadRequestForUnsupportedProvider(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	network := &store.Network{
		Name:       "home",
		Token:      "token-1",
		DDNSConfig: "{}",
	}
	if err := s.CreateNetwork(network); err != nil {
		t.Fatalf("CreateNetwork(seed) error = %v", err)
	}

	h := NewAdminHandler(s, auth.NewSessionManager())
	router := gin.New()
	router.PUT("/admin/api/networks/:id", h.UpdateNetwork)

	resp := performRequest(t, router, http.MethodPut, "/admin/api/networks/1", `{"name":"home","token":"token-1","ddns_type":"cloudflare","ddns_config":"{}"}`, "127.0.0.1:1234", map[string]string{
		"Content-Type": "application/json",
	})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusBadRequest, resp.Body.String())
	}
}

func TestUpdateNetworkReturnsBadRequestForInvalidDNSPodConfig(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	network := &store.Network{
		Name:       "home",
		Token:      "token-1",
		DDNSConfig: "{}",
	}
	if err := s.CreateNetwork(network); err != nil {
		t.Fatalf("CreateNetwork(seed) error = %v", err)
	}

	h := NewAdminHandler(s, auth.NewSessionManager())
	router := gin.New()
	router.PUT("/admin/api/networks/:id", h.UpdateNetwork)

	resp := performRequest(t, router, http.MethodPut, "/admin/api/networks/1", `{"name":"home","token":"token-1","ddns_enabled":true,"ddns_type":"dnspod","ddns_config":"{\"domain\":\"example.com\",\"record\":\"\",\"id\":\"abc\",\"token\":\"def\"}"}`, "127.0.0.1:1234", map[string]string{
		"Content-Type": "application/json",
	})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusBadRequest, resp.Body.String())
	}
}
