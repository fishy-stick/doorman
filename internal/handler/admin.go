package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"doorman/internal/auth"
	"doorman/internal/i18n"
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
		writeLocalizedError(c, http.StatusBadRequest, i18n.MessageInvalidRequest, nil)
		return
	}

	valid, err := h.store.VerifyPassword(req.Password)
	if err != nil {
		writeLocalizedError(c, http.StatusInternalServerError, i18n.MessageInternalServerError, nil)
		return
	}

	if !valid {
		writeLocalizedError(c, http.StatusUnauthorized, i18n.MessageInvalidPassword, nil)
		return
	}

	token, err := h.sm.Create()
	if err != nil {
		writeLocalizedError(c, http.StatusInternalServerError, i18n.MessageFailedCreateSession, nil)
		return
	}

	c.SetCookie(auth.SessionCookieName, token, int(auth.SessionDuration.Seconds()), "/", "", false, true)
	writeLocalizedMessage(c, http.StatusOK, i18n.MessageLoginSuccessful, nil)
}

func (h *AdminHandler) Logout(c *gin.Context) {
	cookie, err := c.Cookie(auth.SessionCookieName)
	if err == nil && cookie != "" {
		h.sm.Delete(cookie)
	}

	c.SetCookie(auth.SessionCookieName, "", -1, "/", "", false, true)
	writeLocalizedMessage(c, http.StatusOK, i18n.MessageLogoutSuccessful, nil)
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
		writeLocalizedError(c, http.StatusBadRequest, i18n.MessageInvalidRequest, nil)
		return
	}

	valid, err := h.store.VerifyPassword(req.OldPassword)
	if err != nil {
		writeLocalizedError(c, http.StatusInternalServerError, i18n.MessageInternalServerError, nil)
		return
	}

	if !valid {
		writeLocalizedError(c, http.StatusUnauthorized, i18n.MessageInvalidOldPassword, nil)
		return
	}

	if err := h.store.UpdatePassword(req.NewPassword); err != nil {
		writeLocalizedError(c, http.StatusInternalServerError, i18n.MessageFailedUpdatePassword, nil)
		return
	}

	h.sm.Reset()
	writeLocalizedMessage(c, http.StatusOK, i18n.MessagePasswordUpdated, nil)
}

func (h *AdminHandler) ListNetworks(c *gin.Context) {
	networks, err := h.store.ListNetworks()
	if err != nil {
		writeLocalizedError(c, http.StatusInternalServerError, i18n.MessageFailedListNetworks, nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{"networks": networks})
}

func (h *AdminHandler) GetNetwork(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		writeLocalizedError(c, http.StatusBadRequest, i18n.MessageInvalidNetworkID, nil)
		return
	}

	network, err := h.store.GetNetwork(id)
	if err != nil {
		writeLocalizedError(c, http.StatusInternalServerError, i18n.MessageFailedGetNetwork, nil)
		return
	}

	if network == nil {
		writeLocalizedError(c, http.StatusNotFound, i18n.MessageNetworkNotFound, nil)
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
		writeLocalizedError(c, http.StatusBadRequest, i18n.MessageInvalidRequest, nil)
		return
	}
	network.ID = 0
	network.Token = uuid.NewString()

	if err := h.store.CreateNetwork(&network); err != nil {
		h.respondStoreError(c, err, i18n.MessageFailedCreateNetwork)
		return
	}

	h.respondNetworkDetail(c, http.StatusCreated, &network)
}

func (h *AdminHandler) UpdateNetwork(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		writeLocalizedError(c, http.StatusBadRequest, i18n.MessageInvalidNetworkID, nil)
		return
	}

	existing, err := h.store.GetNetwork(id)
	if err != nil {
		writeLocalizedError(c, http.StatusInternalServerError, i18n.MessageFailedGetNetwork, nil)
		return
	}

	if existing == nil {
		writeLocalizedError(c, http.StatusNotFound, i18n.MessageNetworkNotFound, nil)
		return
	}

	var network store.Network
	if err := c.ShouldBindJSON(&network); err != nil {
		writeLocalizedError(c, http.StatusBadRequest, i18n.MessageInvalidRequest, nil)
		return
	}

	network.ID = id
	network.Token = existing.Token
	if err := h.store.UpdateNetwork(&network); err != nil {
		h.respondStoreError(c, err, i18n.MessageFailedUpdateNetwork)
		return
	}

	updated, err := h.store.GetNetwork(id)
	if err != nil {
		writeLocalizedError(c, http.StatusInternalServerError, i18n.MessageFailedGetNetwork, nil)
		return
	}

	if updated == nil {
		writeLocalizedError(c, http.StatusInternalServerError, i18n.MessageNetworkDisappearedAfterSave, nil)
		return
	}

	h.respondNetworkDetail(c, http.StatusOK, updated)
}

func (h *AdminHandler) RegenerateNetworkToken(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		writeLocalizedError(c, http.StatusBadRequest, i18n.MessageInvalidNetworkID, nil)
		return
	}

	existing, err := h.store.GetNetwork(id)
	if err != nil {
		writeLocalizedError(c, http.StatusInternalServerError, i18n.MessageFailedGetNetwork, nil)
		return
	}

	if existing == nil {
		writeLocalizedError(c, http.StatusNotFound, i18n.MessageNetworkNotFound, nil)
		return
	}

	if err := h.store.UpdateNetworkToken(id, uuid.NewString()); err != nil {
		h.respondStoreError(c, err, i18n.MessageFailedRegenerateToken)
		return
	}

	updated, err := h.store.GetNetwork(id)
	if err != nil {
		writeLocalizedError(c, http.StatusInternalServerError, i18n.MessageFailedGetNetwork, nil)
		return
	}

	h.respondNetworkDetail(c, http.StatusOK, updated)
}

func (h *AdminHandler) DeleteNetwork(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		writeLocalizedError(c, http.StatusBadRequest, i18n.MessageInvalidNetworkID, nil)
		return
	}

	existing, err := h.store.GetNetwork(id)
	if err != nil {
		writeLocalizedError(c, http.StatusInternalServerError, i18n.MessageFailedGetNetwork, nil)
		return
	}

	if existing == nil {
		writeLocalizedError(c, http.StatusNotFound, i18n.MessageNetworkNotFound, nil)
		return
	}

	if err := h.store.DeleteNetwork(id); err != nil {
		writeLocalizedError(c, http.StatusInternalServerError, i18n.MessageFailedDeleteNetwork, nil)
		return
	}

	writeLocalizedMessage(c, http.StatusOK, i18n.MessageNetworkDeleted, nil)
}

func (h *AdminHandler) ListKnocks(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		writeLocalizedError(c, http.StatusBadRequest, i18n.MessageInvalidNetworkID, nil)
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
		writeLocalizedError(c, http.StatusInternalServerError, i18n.MessageFailedListKnocks, nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":   total,
		"page":    page,
		"size":    size,
		"records": knocks,
	})
}

func (h *AdminHandler) respondStoreError(c *gin.Context, err error, defaultMessage i18n.MessageKey) {
	locale := localeForContext(c)

	switch {
	case errors.Is(err, store.ErrInvalidNetwork):
		c.JSON(http.StatusBadRequest, gin.H{"error": networkErrorMessage(locale, err)})
	case errors.Is(err, store.ErrNetworkNameConflict), errors.Is(err, store.ErrNetworkTokenConflict):
		c.JSON(http.StatusConflict, gin.H{"error": networkErrorMessage(locale, err)})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.Text(locale, defaultMessage, nil)})
	}
}
