package handler

import (
	"encoding/json"
	"net/http"
	"testing"

	"doorman/internal/auth"
	"doorman/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	if body := resp.Body.String(); body != "{\"error\":\"Not authenticated.\"}" {
		t.Fatalf("body = %s, want %s", body, `{"error":"Not authenticated."}`)
	}
}

func TestSessionReturnsLocalizedUnauthorizedMessage(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	sm := auth.NewSessionManager()

	h := NewAdminHandler(s, sm)
	router := gin.New()
	protected := router.Group("/admin/api")
	protected.Use(auth.AdminAuth(sm))
	protected.GET("/session", h.Session)

	resp := performRequest(t, router, http.MethodGet, "/admin/api/session", "", "127.0.0.1:1234", map[string]string{
		"X-Doorman-Locale": "zh-CN",
	})
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusUnauthorized, resp.Body.String())
	}
	if body := resp.Body.String(); body != "{\"error\":\"未登录。\"}" {
		t.Fatalf("body = %s, want %s", body, `{"error":"未登录。"}`)
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

func TestCreateNetworkGeneratesUUIDToken(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	h := NewAdminHandler(s, auth.NewSessionManager())
	router := gin.New()
	router.POST("/admin/api/networks", h.CreateNetwork)

	resp := performRequest(t, router, http.MethodPost, "/admin/api/networks", `{"name":"home","ddns_config":"{}"}`, "127.0.0.1:1234", map[string]string{
		"Content-Type": "application/json",
	})
	if resp.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusCreated, resp.Body.String())
	}

	var body networkDetailResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if _, err := uuid.Parse(body.Token); err != nil {
		t.Fatalf("token = %q, want UUID: %v", body.Token, err)
	}
	if body.Commands.Curl == "" || body.Commands.Crontab == "" {
		t.Fatalf("commands = %+v, want generated commands", body.Commands)
	}
}

func TestCreateNetworkIgnoresSubmittedToken(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	h := NewAdminHandler(s, auth.NewSessionManager())
	router := gin.New()
	router.POST("/admin/api/networks", h.CreateNetwork)

	resp := performRequest(t, router, http.MethodPost, "/admin/api/networks", `{"name":"home","token":"manual-token","ddns_config":"{}"}`, "127.0.0.1:1234", map[string]string{
		"Content-Type": "application/json",
	})
	if resp.Code != http.StatusCreated {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusCreated, resp.Body.String())
	}

	var body networkDetailResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body.Token == "manual-token" {
		t.Fatal("token used submitted value, want generated UUID")
	}
	if _, err := uuid.Parse(body.Token); err != nil {
		t.Fatalf("token = %q, want UUID: %v", body.Token, err)
	}
}

func TestCreateNetworkReturnsBadRequestForUnsupportedProvider(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	h := NewAdminHandler(s, auth.NewSessionManager())
	router := gin.New()
	router.POST("/admin/api/networks", h.CreateNetwork)

	resp := performRequest(t, router, http.MethodPost, "/admin/api/networks", `{"name":"home","ddns_enabled":true,"ddns_type":"cloudflare","ddns_config":"{}"}`, "127.0.0.1:1234", map[string]string{
		"Content-Type": "application/json",
	})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusBadRequest, resp.Body.String())
	}
	if body := resp.Body.String(); body != "{\"error\":\"Unsupported DDNS provider \\\"cloudflare\\\". Supported providers: DNSPod.\"}" {
		t.Fatalf("body = %s", body)
	}
}

func TestCreateNetworkReturnsBadRequestForInvalidDDNSConfigShape(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	h := NewAdminHandler(s, auth.NewSessionManager())
	router := gin.New()
	router.POST("/admin/api/networks", h.CreateNetwork)

	resp := performRequest(t, router, http.MethodPost, "/admin/api/networks", `{"name":"home","ddns_enabled":true,"ddns_type":"dnspod","ddns_config":"[]"}`, "127.0.0.1:1234", map[string]string{
		"Content-Type": "application/json",
	})
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusBadRequest, resp.Body.String())
	}
}

func TestCreateNetworkReturnsLocalizedValidationError(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	h := NewAdminHandler(s, auth.NewSessionManager())
	router := gin.New()
	router.POST("/admin/api/networks", h.CreateNetwork)

	resp := performRequest(
		t,
		router,
		http.MethodPost,
		"/admin/api/networks",
		`{"name":"home","ddns_enabled":true,"ddns_type":"dnspod","ddns_config":"{\"domain\":\"example.com\",\"record\":\"\",\"id\":\"abc\",\"token\":\"def\"}"}`,
		"127.0.0.1:1234",
		map[string]string{
			"Content-Type":     "application/json",
			"X-Doorman-Locale": "zh-CN",
		},
	)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusBadRequest, resp.Body.String())
	}
	if body := resp.Body.String(); body != "{\"error\":\"DNSPod 配置中的“record”字段不能为空。\"}" {
		t.Fatalf("body = %s, want %s", body, `{"error":"DNSPod 配置中的“record”字段不能为空。"}`)
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

	resp := performRequest(t, router, http.MethodPut, "/admin/api/networks/1", `{"name":"home","ddns_enabled":true,"ddns_type":"cloudflare","ddns_config":"{}"}`, "127.0.0.1:1234", map[string]string{
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

func TestUpdateNetworkPreservesToken(t *testing.T) {
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

	resp := performRequest(t, router, http.MethodPut, "/admin/api/networks/1", `{"name":"home-updated","token":"token-2","ddns_config":"{}"}`, "127.0.0.1:1234", map[string]string{
		"Content-Type": "application/json",
	})
	if resp.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusOK, resp.Body.String())
	}

	var body networkDetailResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body.Token != "token-1" {
		t.Fatalf("response token = %q, want %q", body.Token, "token-1")
	}
	if body.Commands.Curl == "" || body.Commands.Crontab == "" {
		t.Fatalf("commands = %+v, want generated commands", body.Commands)
	}

	got, err := s.GetNetwork(network.ID)
	if err != nil {
		t.Fatalf("GetNetwork() error = %v", err)
	}
	if got == nil || got.Token != "token-1" {
		t.Fatalf("stored token = %v, want token-1", got)
	}
}

func TestRegenerateNetworkTokenUpdatesTokenAndInvalidatesOldToken(t *testing.T) {
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
	knockHandler := NewKnockHandler(s, false)
	router := gin.New()
	router.POST("/admin/api/networks/:id/token", h.RegenerateNetworkToken)
	router.GET("/knock", auth.KnockAuth(s), knockHandler.Handle)

	resp := performRequest(t, router, http.MethodPost, "/admin/api/networks/1/token", "", "127.0.0.1:1234", nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusOK, resp.Body.String())
	}

	var body networkDetailResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body.Token == "token-1" {
		t.Fatal("regenerated token did not change")
	}
	if _, err := uuid.Parse(body.Token); err != nil {
		t.Fatalf("token = %q, want UUID: %v", body.Token, err)
	}

	oldTokenResp := performRequest(t, router, http.MethodGet, "/knock", "", "198.51.100.10:1000", map[string]string{
		"Authorization": "Bearer token-1",
	})
	if oldTokenResp.Code != http.StatusUnauthorized {
		t.Fatalf("old token status code = %d, want %d, body = %s", oldTokenResp.Code, http.StatusUnauthorized, oldTokenResp.Body.String())
	}

	newTokenResp := performRequest(t, router, http.MethodGet, "/knock", "", "198.51.100.11:1001", map[string]string{
		"Authorization": "Bearer " + body.Token,
	})
	if newTokenResp.Code != http.StatusOK {
		t.Fatalf("new token status code = %d, want %d, body = %s", newTokenResp.Code, http.StatusOK, newTokenResp.Body.String())
	}
}

func TestRegenerateNetworkTokenReturnsNotFound(t *testing.T) {
	t.Parallel()

	s := newTestStore(t)
	h := NewAdminHandler(s, auth.NewSessionManager())
	router := gin.New()
	router.POST("/admin/api/networks/:id/token", h.RegenerateNetworkToken)

	resp := performRequest(t, router, http.MethodPost, "/admin/api/networks/99/token", "", "127.0.0.1:1234", nil)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d, body = %s", resp.Code, http.StatusNotFound, resp.Body.String())
	}
}

type networkDetailResponse struct {
	Token    string `json:"token"`
	Commands struct {
		Curl    string `json:"curl"`
		Crontab string `json:"crontab"`
	} `json:"commands"`
}
