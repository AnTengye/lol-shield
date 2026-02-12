package client

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed web/dist
var frontendEmbedFS embed.FS

func registerFrontendRoutes(r *gin.Engine) {
	distFS, err := fs.Sub(frontendEmbedFS, "web/dist")
	if err != nil {
		return
	}
	httpFS := http.FS(distFS)
	indexHTML, indexErr := fs.ReadFile(distFS, "index.html")
	if indexErr != nil {
		return
	}

	r.NoRoute(
		func(c *gin.Context) {
			reqPath := c.Request.URL.Path
			if isBackendPath(reqPath) {
				c.Status(http.StatusNotFound)
				return
			}

			cleanPath := strings.TrimPrefix(path.Clean(reqPath), "/")
			if cleanPath == "." || cleanPath == "" {
				c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
				return
			}

			file, openErr := distFS.Open(cleanPath)
			if openErr == nil {
				defer file.Close()
				stat, statErr := file.Stat()
				if statErr == nil && !stat.IsDir() {
					c.FileFromFS("/"+cleanPath, httpFS)
					return
				}
			}

			c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
		},
	)
}

func isBackendPath(reqPath string) bool {
	return reqPath == "/v1" ||
		reqPath == "/riot" ||
		reqPath == "/ws" ||
		strings.HasPrefix(reqPath, "/v1/") ||
		strings.HasPrefix(reqPath, "/riot/") ||
		strings.HasPrefix(reqPath, "/ws/")
}
