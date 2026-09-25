package client

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/core/lcuapi"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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

func TestShutdownEndpointReturnsAck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	syslog.L = zap.NewNop().Sugar()
	engine := gin.New()
	AddRouter(engine, NewShield())

	req := httptest.NewRequest(http.MethodPost, "/v1/shutdown", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"shutting_down":true`) {
		t.Fatalf("expected shutdown ack payload, got %s", w.Body.String())
	}
}

// 验证 /v1/shutdown 端到端链路：桌面壳更新前经此端点请求提权 sidecar
// 自行退出，服务应优雅结束且 Run 不返回错误（否则 main 会以非零码退出）。
func TestShutdownEndpointStopsServerGracefully(t *testing.T) {
	gin.SetMode(gin.TestMode)
	syslog.L = zap.NewNop().Sugar()

	// 预约一个空闲端口供本地服务监听，避免与其它测试或开发环境冲突
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	base := "http://" + listener.Addr().String()
	_ = listener.Close()

	viper.Set(configs.WebAddr, strings.TrimPrefix(base, "http://"))
	viper.Set(configs.MockLCUEnabled, true)
	t.Cleanup(func() {
		viper.Set(configs.WebAddr, ":9365")
		viper.Set(configs.MockLCUEnabled, false)
	})

	// 后台监控轮询会访问 LCU 服务，用一个空服务承接即可，不影响退出路径
	mockLcu := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockLcu.Close()

	shield := NewShieldWithLCU(lcuapi.NewHTTPService(mockLcu.URL))
	done := make(chan error, 1)
	go func() { done <- shield.Run() }()

	var res *http.Response
	for i := 0; i < 50; i++ {
		res, err = http.Get(base + "/v1/version")
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatalf("local server did not start: %v", err)
	}
	res.Body.Close()

	shutdownRes, err := http.Post(base+"/v1/shutdown", "application/json", nil)
	if err != nil {
		t.Fatalf("shutdown request failed: %v", err)
	}
	shutdownRes.Body.Close()
	if shutdownRes.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 from shutdown, got %d", shutdownRes.StatusCode)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected graceful shutdown without error, got %v", err)
		}
	case <-time.After(5 * time.Second):
		_ = shield.Stop()
		t.Fatal("shield.Run did not return after shutdown request")
	}
}

// 端口已被旧实例占用时（sidecar 重复启动），监听失败必须让进程立即
// 以错误退出，而不是在 g.Wait 上永久挂起变成僵尸进程。
func TestNotifyQuitExitsWhenPortOccupied(t *testing.T) {
	gin.SetMode(gin.TestMode)
	syslog.L = zap.NewNop().Sugar()

	// 占住端口：模拟旧 sidecar 仍在运行、新实例重复启动的场景
	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer occupied.Close()

	viper.Set(configs.WebAddr, occupied.Addr().String())
	viper.Set(configs.MockLCUEnabled, true)
	t.Cleanup(func() {
		viper.Set(configs.WebAddr, ":9365")
		viper.Set(configs.MockLCUEnabled, false)
	})

	mockLcu := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockLcu.Close()

	shield := NewShieldWithLCU(lcuapi.NewHTTPService(mockLcu.URL))
	done := make(chan error, 1)
	go func() { done <- shield.Run() }()

	select {
	case err := <-done:
		_ = shield.Stop()
		if err == nil {
			t.Fatal("expected listen error when port is occupied, got nil")
		}
	case <-time.After(5 * time.Second):
		_ = shield.Stop()
		t.Fatal("shield.Run hung when listening port was already occupied")
	}
}
