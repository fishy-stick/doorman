package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"doorman/internal/store"
	"github.com/gin-gonic/gin"
)

const (
	SessionCookieName = "doorman_session"
	SessionDuration   = 24 * time.Hour
)

type contextKey string

const NetworkContextKey contextKey = "network"

// SessionManager keeps short-lived admin sessions in memory.
type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]time.Time
}

// NewSessionManager creates an empty in-memory session store.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]time.Time),
	}
}

func (sm *SessionManager) Create() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)
	sm.mu.Lock()
	sm.sessions[token] = time.Now().Add(SessionDuration)
	sm.mu.Unlock()
	return token, nil
}

func (sm *SessionManager) Validate(token string) bool {
	sm.mu.RLock()
	expiry, exists := sm.sessions[token]
	sm.mu.RUnlock()
	if !exists {
		return false
	}
	if time.Now().After(expiry) {
		sm.mu.Lock()
		delete(sm.sessions, token)
		sm.mu.Unlock()
		return false
	}
	return true
}

func (sm *SessionManager) Delete(token string) {
	sm.mu.Lock()
	delete(sm.sessions, token)
	sm.mu.Unlock()
}

// Reset invalidates every active admin session.
func (sm *SessionManager) Reset() {
	sm.mu.Lock()
	sm.sessions = make(map[string]time.Time)
	sm.mu.Unlock()
}

// KnockAuth resolves the bearer token to a managed network.
func KnockAuth(s *store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "empty token"})
			c.Abort()
			return
		}

		network, err := s.GetNetworkByToken(token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			c.Abort()
			return
		}

		if network == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set(string(NetworkContextKey), network)
		c.Next()
	}
}

// AdminAuth verifies the admin session cookie before protected requests.
func AdminAuth(sm *SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(SessionCookieName)
		if err != nil || cookie == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			c.Abort()
			return
		}

		if !sm.Validate(cookie) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			c.Abort()
			return
		}

		c.Next()
	}
}
