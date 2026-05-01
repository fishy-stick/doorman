package webui

import (
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

func Register(r *gin.RouterGroup) {
	webFS, ok := assetFS()
	if !ok {
		return
	}

	registerRoutes(r, webFS)
}

func registerRoutes(r *gin.RouterGroup, webFS fs.FS) {
	indexHTML, err := fs.ReadFile(webFS, "index.html")
	if err != nil {
		panic(fmt.Errorf("read embedded index.html: %w", err))
	}

	staticServer := http.StripPrefix("/admin", http.FileServer(http.FS(webFS)))

	serveIndex := func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	}

	r.GET("", serveIndex)
	r.GET("/*path", func(c *gin.Context) {
		relPath := strings.TrimPrefix(c.Param("path"), "/")

		if relPath == "" {
			serveIndex(c)
			return
		}

		if relPath == "api" || strings.HasPrefix(relPath, "api/") {
			c.Status(http.StatusNotFound)
			return
		}

		if ext := path.Ext(relPath); ext != "" {
			file, err := webFS.Open(relPath)
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}

			defer file.Close()

			info, statErr := file.Stat()
			if statErr != nil || info.IsDir() {
				c.Status(http.StatusNotFound)
				return
			}

			staticServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		serveIndex(c)
	})
}
