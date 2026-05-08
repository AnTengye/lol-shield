package client

import (
	"testing"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/core/lcuapi"
	"github.com/spf13/viper"
)

func TestNewShieldWithMockModeStartsWithoutTokenPolling(t *testing.T) {
	viper.Set(configs.MockLCUEnabled, true)
	viper.Set(configs.MockLCUBaseURL, "http://127.0.0.1:19365")
	t.Cleanup(func() {
		viper.Set(configs.MockLCUEnabled, false)
		viper.Set(configs.MockLCUBaseURL, "http://127.0.0.1:19365")
	})

	svc := lcuapi.NewHTTPService("http://127.0.0.1:19365")
	shield := NewShieldWithLCU(svc)

	if shield.lcuService == nil {
		t.Fatalf("expected lcu service")
	}
	if shield.CurInfo.Status != STWaiting {
		t.Fatalf("expected waiting status, got %v", shield.CurInfo.Status)
	}
	if shield.CurInfo.GameStatus != GSWaiting {
		t.Fatalf("expected waiting game status, got %v", shield.CurInfo.GameStatus)
	}
}
