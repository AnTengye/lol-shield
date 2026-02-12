package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestFrontendRootNoRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	AddRouter(engine, NewShield())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "" {
		t.Fatalf("expected no redirect location, got %q", loc)
	}
}
