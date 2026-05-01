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

// KnockHandler records observed client IPs and triggers DDNS updates.
type KnockHandler struct {
	store      *store.Store
	trustProxy bool
}

// NewKnockHandler builds the /knock handler with the effective proxy policy.
func NewKnockHandler(s *store.Store, trustProxy bool) *KnockHandler {
	return &KnockHandler{store: s, trustProxy: trustProxy}
}

func (h *KnockHandler) Handle(c *gin.Context) {
	networkVal, exists := c.Get(string(auth.NetworkContextKey))
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "network not found in context"})
		return
	}
	network, ok := networkVal.(*store.Network)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid network in context"})
		return
	}

	clientIP := ip.ExtractIP(c, h.trustProxy)

	// PreviousIP on a new knock should point to the last observed IP, not the
	// previous knock's PreviousIP chain.
	previousIP, err := h.store.GetLatestIP(network.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get latest IP"})
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
