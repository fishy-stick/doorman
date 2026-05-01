package ip

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

func ExtractIP(c *gin.Context, trustProxy bool) string {
	if trustProxy {
		if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
			for _, part := range strings.Split(xff, ",") {
				ip := strings.TrimSpace(part)
				if ip != "" {
					return ip
				}
			}
		}

		if xri := c.GetHeader("X-Real-IP"); xri != "" {
			return strings.TrimSpace(xri)
		}
	}

	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return host
}
