package ip

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// ExtractIP returns the caller IP using trusted proxy headers when enabled.
func ExtractIP(c *gin.Context, trustProxy bool) string {
	if trustProxy {
		if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
			for _, part := range strings.Split(xff, ",") {
				ip := normalizeIP(part)
				if ip != "" {
					return ip
				}
			}
		}

		if ip := normalizeIP(c.GetHeader("X-Real-IP")); ip != "" {
			return ip
		}
	}

	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		if ip := normalizeIP(c.Request.RemoteAddr); ip != "" {
			return ip
		}
		return c.Request.RemoteAddr
	}
	if ip := normalizeIP(host); ip != "" {
		return ip
	}
	return host
}

func normalizeIP(raw string) string {
	ip := strings.TrimSpace(raw)
	if ip == "" {
		return ""
	}

	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}

	return parsed.String()
}
