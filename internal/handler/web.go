package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RegisterWebRoutes(r *gin.RouterGroup) {
	r.GET("/assets/*filepath", func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})

	r.GET("/*path", func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/admin") {
			c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(indexHTML))
			return
		}

		c.Status(http.StatusNotFound)
	})
}

const indexHTML = `<!DOCTYPE html>
<html>
<head><title>Doorman</title></head>
<body><div id="root"></div></body>
</html>`
