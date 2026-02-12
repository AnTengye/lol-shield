//go:build !with_frontend || no_frontend

package api

import "github.com/gin-gonic/gin"

func RegisterFrontendRoutes(r *gin.Engine) {
	// backend-only build: do not register SPA routes
}
