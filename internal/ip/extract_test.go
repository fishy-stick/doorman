package ip

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestExtractIPUsesFirstValidForwardedAddress(t *testing.T) {
	t.Parallel()

	c := newTestContext()
	c.Request.Header.Set("X-Forwarded-For", "invalid, 203.0.113.10, 203.0.113.11")
	c.Request.Header.Set("X-Real-IP", "203.0.113.20")

	if got := ExtractIP(c, true); got != "203.0.113.10" {
		t.Fatalf("ExtractIP() = %q, want %q", got, "203.0.113.10")
	}
}

func TestExtractIPIgnoresProxyHeadersWhenDisabled(t *testing.T) {
	t.Parallel()

	c := newTestContext()
	c.Request.Header.Set("X-Forwarded-For", "203.0.113.10")
	c.Request.Header.Set("X-Real-IP", "203.0.113.20")

	if got := ExtractIP(c, false); got != "198.51.100.50" {
		t.Fatalf("ExtractIP() = %q, want %q", got, "198.51.100.50")
	}
}

func TestExtractIPFallsBackToRealIPWhenForwardedForIsInvalid(t *testing.T) {
	t.Parallel()

	c := newTestContext()
	c.Request.Header.Set("X-Forwarded-For", "invalid")
	c.Request.Header.Set("X-Real-IP", "203.0.113.20")

	if got := ExtractIP(c, true); got != "203.0.113.20" {
		t.Fatalf("ExtractIP() = %q, want %q", got, "203.0.113.20")
	}
}

func newTestContext() *gin.Context {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "198.51.100.50:1234"
	c.Request = req

	return c
}
