package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/core/lcuapi"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func TestBackendRootReturnsNotFoundWithoutEmbeddedFrontend(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	AddRouter(engine, NewShield())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404 for removed browser frontend, got %d", w.Code)
	}
}

func TestNewShieldWithMockModeStillInitializesLCUService(t *testing.T) {
	viper.Set(configs.MockLCUEnabled, true)
	viper.Set(configs.MockLCUBaseURL, "http://127.0.0.1:19365")
	t.Cleanup(func() {
		viper.Set(configs.MockLCUEnabled, false)
		viper.Set(configs.MockLCUBaseURL, "http://127.0.0.1:19365")
	})

	svc := lcuapi.NewHTTPService("http://127.0.0.1:19365")
	shield := NewShieldWithLCU(svc)

	if shield.lcuService == nil {
		t.Fatal("expected LCU service to remain configured")
	}
}
