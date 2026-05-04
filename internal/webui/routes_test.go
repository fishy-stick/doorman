package webui

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutes(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	webFS := fstest.MapFS{
		"index.html":     {Data: []byte("<html>index</html>")},
		"assets/app.js":  {Data: []byte("console.log('ok')")},
		"favicon.svg":    {Data: []byte("<svg></svg>")},
		"assets/app.css": {Data: []byte("body{}")},
	}

	router := gin.New()
	adminAPI := router.Group("/admin/api")
	adminAPI.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	registerRoutes(router, fs.FS(webFS))

	tests := []struct {
		name       string
		path       string
		statusCode int
		body       string
	}{
		{
			name:       "admin root serves index",
			path:       "/admin",
			statusCode: http.StatusOK,
			body:       "<html>index</html>",
		},
		{
			name:       "admin slash serves index",
			path:       "/admin/",
			statusCode: http.StatusOK,
			body:       "<html>index</html>",
		},
		{
			name:       "spa path falls back to index",
			path:       "/admin/dashboard",
			statusCode: http.StatusOK,
			body:       "<html>index</html>",
		},
		{
			name:       "asset file is served",
			path:       "/admin/assets/app.js",
			statusCode: http.StatusOK,
			body:       "console.log('ok')",
		},
		{
			name:       "root static file is served",
			path:       "/admin/favicon.svg",
			statusCode: http.StatusOK,
			body:       "<svg></svg>",
		},
		{
			name:       "missing asset returns not found",
			path:       "/admin/assets/missing.js",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "api path is not swallowed by spa",
			path:       "/admin/api/missing",
			statusCode: http.StatusNotFound,
		},
		{
			name:       "registered api path still works",
			path:       "/admin/api/ping",
			statusCode: http.StatusOK,
			body:       "pong",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)

			router.ServeHTTP(recorder, req)

			if recorder.Code != tt.statusCode {
				t.Fatalf("status code = %d, want %d", recorder.Code, tt.statusCode)
			}

			if tt.body != "" && recorder.Body.String() != tt.body {
				t.Fatalf("body = %q, want %q", recorder.Body.String(), tt.body)
			}
		})
	}
}
