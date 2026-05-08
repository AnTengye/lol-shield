package client

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/AnTengye/lol-shield/internal/core/lcuapi"
	"github.com/AnTengye/lol-shield/internal/mocklcu"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestRunningAndHistoryHandlersWorkAgainstMockLCU(t *testing.T) {
	syslog.L = zap.NewNop().Sugar()

	scenario, err := mocklcu.LoadScenario(filepath.Join("..", "mocklcu", "fixtures", "default"))
	if err != nil {
		t.Fatalf("LoadScenario returned error: %v", err)
	}
	mockSrv := httptest.NewServer(mocklcu.NewServer(scenario))
	defer mockSrv.Close()

	shield := NewShieldWithLCU(lcuapi.NewHTTPService(mockSrv.URL))
	shield.bootstrapMockState()

	engine := gin.New()
	AddRouter(engine, shield)

	historyReq := httptest.NewRequest(http.MethodGet, "/v1/history/de06293d-082d-59c2-83a6-273ab88164bc?page=0&pageSize=9", nil)
	historyRes := httptest.NewRecorder()
	engine.ServeHTTP(historyRes, historyReq)
	if historyRes.Code != http.StatusOK {
		t.Fatalf("expected history 200, got %d: %s", historyRes.Code, historyRes.Body.String())
	}

	runningReq := httptest.NewRequest(http.MethodGet, "/v1/game/running", nil)
	runningRes := httptest.NewRecorder()
	engine.ServeHTTP(runningRes, runningReq)
	if runningRes.Code != http.StatusOK {
		t.Fatalf("expected running 200, got %d: %s", runningRes.Code, runningRes.Body.String())
	}
}
