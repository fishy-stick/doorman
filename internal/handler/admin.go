package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"doorman/internal/auth"
	"doorman/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminHandler struct {
	store *store.Store
	sm    *auth.SessionManager
}

// NewAdminHandler builds the admin API handler set.
func NewAdminHandler(s *store.Store, sm *auth.SessionManager) *AdminHandler {
	return &AdminHandler{store: s, sm: sm}
}

func (h *AdminHandler) Login(c *gin.Context) {
	var req struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	valid, err := h.store.VerifyPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	token, err := h.sm.Create()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	c.SetCookie(auth.SessionCookieName, token, int(auth.SessionDuration.Seconds()), "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "login successful"})
}

func (h *AdminHandler) Logout(c *gin.Context) {
	cookie, err := c.Cookie(auth.SessionCookieName)
	if err == nil && cookie != "" {
		h.sm.Delete(cookie)
	}

	c.SetCookie(auth.SessionCookieName, "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func (h *AdminHandler) Session(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"authenticated": true})
}

func (h *AdminHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	valid, err := h.store.VerifyPassword(req.OldPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid old password"})
		return
	}

	if err := h.store.UpdatePassword(req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	h.sm.Reset()
	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}

func (h *AdminHandler) ListNetworks(c *gin.Context) {
	networks, err := h.store.ListNetworks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list networks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"networks": networks})
}

func (h *AdminHandler) GetNetwork(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid network id"})
		return
	}

	network, err := h.store.GetNetwork(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get network"})
		return
	}

	if network == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "network not found"})
		return
	}

	h.respondNetworkDetail(c, http.StatusOK, network)
}

func (h *AdminHandler) respondNetworkDetail(c *gin.Context, status int, network *store.Network) {
	latestKnock, _ := h.store.GetLatestKnock(network.ID)

	var currentIP, previousIP, lastKnock, ddnsStatus interface{}
	if latestKnock != nil {
		currentIP = latestKnock.IP
		previousIP = latestKnock.PreviousIP
		lastKnock = latestKnock.CreatedAt
		ddnsStatus = latestKnock.DDNSStatus
	}

	curlCmd := fmt.Sprintf(`curl -H "Authorization: Bearer %s" http://your-server:8080/knock`, network.Token)
	crontabCmd := fmt.Sprintf(`*/5 * * * * curl -s -H "Authorization: Bearer %s" http://your-server:8080/knock > /dev/null 2>&1`, network.Token)

	c.JSON(status, gin.H{
		"id":           network.ID,
		"name":         network.Name,
		"token":        network.Token,
		"ddns_enabled": network.DDNSEnabled,
		"ddns_type":    network.DDNSType,
		"ddns_config":  network.DDNSConfig,
		"current_ip":   currentIP,
		"previous_ip":  previousIP,
		"last_knock":   lastKnock,
		"ddns_status":  ddnsStatus,
		"commands": gin.H{
			"curl":    curlCmd,
			"crontab": crontabCmd,
		},
	})
}

func (h *AdminHandler) CreateNetwork(c *gin.Context) {
	var network store.Network
	if err := c.ShouldBindJSON(&network); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	network.ID = 0
	network.Token = uuid.NewString()

	if err := h.store.CreateNetwork(&network); err != nil {
		h.respondStoreError(c, err, "failed to create network")
		return
	}

	h.respondNetworkDetail(c, http.StatusCreated, &network)
}

func (h *AdminHandler) UpdateNetwork(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid network id"})
		return
	}

	existing, err := h.store.GetNetwork(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get network"})
		return
	}

	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "network not found"})
		return
	}

	var network store.Network
	if err := c.ShouldBindJSON(&network); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	network.ID = id
	network.Token = existing.Token
	if err := h.store.UpdateNetwork(&network); err != nil {
		h.respondStoreError(c, err, "failed to update network")
		return
	}

	c.JSON(http.StatusOK, network)
}

func (h *AdminHandler) RegenerateNetworkToken(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid network id"})
		return
	}

	existing, err := h.store.GetNetwork(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get network"})
		return
	}

	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "network not found"})
		return
	}

	if err := h.store.UpdateNetworkToken(id, uuid.NewString()); err != nil {
		h.respondStoreError(c, err, "failed to regenerate token")
		return
	}

	updated, err := h.store.GetNetwork(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get network"})
		return
	}

	h.respondNetworkDetail(c, http.StatusOK, updated)
}

func (h *AdminHandler) DeleteNetwork(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid network id"})
		return
	}

	existing, err := h.store.GetNetwork(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get network"})
		return
	}

	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "network not found"})
		return
	}

	if err := h.store.DeleteNetwork(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete network"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "network deleted"})
}

func (h *AdminHandler) ListKnocks(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid network id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "50"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 50
	}

	knocks, total, err := h.store.ListKnocks(id, page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list knocks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   total,
		"page":    page,
		"size":    size,
		"records": knocks,
	})
}

func (h *AdminHandler) respondStoreError(c *gin.Context, err error, defaultMessage string) {
	switch {
	case errors.Is(err, store.ErrInvalidNetwork):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, store.ErrNetworkNameConflict), errors.Is(err, store.ErrNetworkTokenConflict):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": defaultMessage})
	}
}
