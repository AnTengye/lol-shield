package middleware

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func AllowedOrigin(origin string) bool {
	switch origin {
	case "tauri://localhost", "http://tauri.localhost", "https://tauri.localhost", "http://localhost:5173", "http://127.0.0.1:5173":
		return true
	default:
		return false
	}
}

func Cors() gin.HandlerFunc {
	return cors.New(
		cors.Config{
			AllowOriginFunc: AllowedOrigin,
			AllowMethods: []string{
				http.MethodGet,
				http.MethodHead,
				http.MethodPost,
				http.MethodPut,
				http.MethodPatch,
				http.MethodDelete,
				http.MethodConnect,
				http.MethodOptions,
				http.MethodTrace,
			},
			AllowHeaders:     []string{"content-type", "x-requested-with", "token", "locale"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		},
	)
}
