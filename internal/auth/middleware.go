package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
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

type SessionManager struct {
	sessions map[string]time.Time
}

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
	sm.sessions[token] = time.Now().Add(SessionDuration)
	return token, nil
}

func (sm *SessionManager) Validate(token string) bool {
	expiry, exists := sm.sessions[token]
	if !exists {
		return false
	}
	if time.Now().After(expiry) {
		delete(sm.sessions, token)
		return false
	}
	return true
}

func (sm *SessionManager) Delete(token string) {
	delete(sm.sessions, token)
}

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

func AdminAuth(sm *SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie(SessionCookieName)
		if err != nil || cookie == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			c.Abort()
			return
		}

		if !sm.Validate(cookie) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func Login(c *gin.Context, adminPassword string, sm *SessionManager) {
	var req struct {
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Password != adminPassword {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	token, err := sm.Create()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	c.SetCookie(SessionCookieName, token, int(SessionDuration.Seconds()), "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "login successful"})
}

func Logout(c *gin.Context, sm *SessionManager) {
	cookie, err := c.Cookie(SessionCookieName)
	if err == nil && cookie != "" {
		sm.Delete(cookie)
	}

	c.SetCookie(SessionCookieName, "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}
