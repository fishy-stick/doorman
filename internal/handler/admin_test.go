package handler

import (
	"net/http"
	"testing"

	"doorman/internal/auth"
	"doorman/internal/store"

	"github.com/gin-gonic/gin"
)

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
