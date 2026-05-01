package handler

import (
	"net/http"
	"time"

	"doorman/internal/auth"
	"doorman/internal/ddns"
	"doorman/internal/ip"
	"doorman/internal/store"
	"github.com/gin-gonic/gin"
)

type KnockHandler struct {
	store *store.Store
}

func NewKnockHandler(s *store.Store) *KnockHandler {
	return &KnockHandler{store: s}
}

func (h *KnockHandler) Handle(c *gin.Context) {
	networkVal, exists := c.Get(string(auth.NetworkContextKey))
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "network not found in context"})
		return
	}
	network := networkVal.(*store.Network)

	clientIP := ip.ExtractIP(c, true)

	previousIP, err := h.store.GetPreviousIP(network.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get previous IP"})
		return
	}

	ipChanged := previousIP == nil || *previousIP != clientIP

	ddnsStatus := "skipped"
	ddnsError := ""
	ddnsUpdated := false

	if network.DDNSEnabled && ipChanged {
		provider, err := ddns.GetProvider(network.DDNSType)
		if err != nil {
			ddnsStatus = "failed"
			ddnsError = err.Error()
		} else {
			result, err := provider.Update(network.DDNSConfig, clientIP)
			if err != nil {
				ddnsStatus = "failed"
				ddnsError = err.Error()
			} else if result.Updated {
				ddnsStatus = "success"
				ddnsUpdated = true
			} else {
				ddnsStatus = "skipped"
			}
		}
	}

	knock := &store.Knock{
		NetworkID:  network.ID,
		IP:         clientIP,
		PreviousIP: previousIP,
		IPChanged:  ipChanged,
		UserAgent:  c.GetHeader("User-Agent"),
		DDNSStatus: ddnsStatus,
		DDNSError:  ddnsError,
		CreatedAt:  time.Now(),
	}

	if err := h.store.InsertKnock(knock); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save knock"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"network":      network.Name,
		"ip":           clientIP,
		"changed":      ipChanged,
		"ddns_updated": ddnsUpdated,
		"timestamp":    knock.CreatedAt.Format(time.RFC3339),
	})
}
