package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestManagedServiceProcess(t *testing.T) {
	if os.Getenv("LOL_SHIELD_PROCESS_TEST") != "1" {
		return
	}
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func TestManagedServiceRestartsWithOwnedSessionAndStableDataDirectory(t *testing.T) {
	probe, err := net.Listen("tcp", "127.0.0.1:9365")
	if err != nil {
		t.Skip("开发环境已有服务使用桌面固定端口，不干扰现有进程")
	}
	_ = probe.Close()
	mock := httptest.NewServer(http.NotFoundHandler())
	defer mock.Close()
	dir := filepath.Join(t.TempDir(), "用户数据 with spaces")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(dir, "config.yaml")
	original := []byte(fmt.Sprintf("mock_lcu:\n  enabled: true\n  base_url: %q\nlog:\n  level: error\ngame:\n  auto_confirm: false\n", mock.URL))
	if err := os.WriteFile(config, original, 0600); err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Timeout: 500 * time.Millisecond, Transport: &http.Transport{Proxy: nil}}
	defer client.CloseIdleConnections()
	for attempt := 0; attempt < 2; attempt++ {
		t.Run(strconv.Itoa(attempt), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			session := fmt.Sprintf("lifecycle-test-session-%d", attempt)
			command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestManagedServiceProcess$",
				"--tauri-sidecar", "--data-dir", dir, "-c", config,
				"--desktop-session", session, "--parent-pid", strconv.Itoa(os.Getpid()))
			command.Dir = t.TempDir()
			command.Env = append(os.Environ(), "LOL_SHIELD_PROCESS_TEST=1")
			if err := command.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan struct{})
			var exitErr error
			go func() { exitErr = command.Wait(); close(done) }()
			t.Cleanup(func() { cancel(); <-done })
			request := func(method, path, token string) (*http.Response, error) {
				req, err := http.NewRequestWithContext(ctx, method, "http://127.0.0.1:9365"+path, nil)
				if err != nil {
					return nil, err
				}
				req.Header.Set("X-Shield-Session", token)
				return client.Do(req)
			}
			for {
				response, err := request(http.MethodGet, "/v1/health", session)
				if err == nil {
					var health struct {
						Service       string
						Protocol, PID int
					}
					decodeErr := json.NewDecoder(response.Body).Decode(&health)
					response.Body.Close()
					if decodeErr == nil && response.StatusCode == http.StatusOK && health.Service == "lol-shield" && health.Protocol == 1 && health.PID == command.Process.Pid {
						break
					}
				}
				select {
				case <-done:
					t.Fatalf("服务未就绪就退出: %v", exitErr)
				case <-ctx.Done():
					t.Fatal("受管服务启动超时")
				case <-time.After(25 * time.Millisecond):
				}
			}
			for _, call := range []struct {
				token string
				code  int
			}{{"previous-session", http.StatusForbidden}, {session, http.StatusOK}} {
				response, err := request(http.MethodPost, "/v1/shutdown", call.token)
				if err != nil {
					t.Fatal(err)
				}
				response.Body.Close()
				if response.StatusCode != call.code {
					t.Fatalf("关闭接口状态码: %d", response.StatusCode)
				}
			}
			select {
			case <-done:
				if exitErr != nil {
					t.Fatalf("服务没有正常退出: %v", exitErr)
				}
			case <-time.After(6 * time.Second):
				t.Fatal("关闭 ACK 后进程未实际退出")
			}
		})
	}
	actual, err := os.ReadFile(config)
	if err != nil || string(actual) != string(original) {
		t.Fatal("重启后用户配置发生变化")
	}
	logs, err := os.ReadFile(filepath.Join(dir, "logs", "startup-go.log"))
	if err != nil || strings.Count(string(logs), "正常退出") != 2 || strings.Contains(string(logs), "lifecycle-test-session") {
		t.Fatalf("启动日志不完整或包含会话令牌: %v", err)
	}
}

func TestManagedStartupLogsConfigFailureBeforeServiceStart(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	oldConfig, oldData, oldSidecar, oldSession, oldParent := *configPath, *dataDir, *tauriSidecar, *desktopSession, *parentPID
	t.Cleanup(func() {
		*configPath, *dataDir, *tauriSidecar, *desktopSession, *parentPID = oldConfig, oldData, oldSidecar, oldSession, oldParent
	})
	dir := t.TempDir()
	*dataDir = filepath.Join(dir, "用户数据 with spaces")
	if err := os.MkdirAll(*dataDir, 0700); err != nil {
		t.Fatal(err)
	}
	*configPath = filepath.Join(*dataDir, "config.yaml")
	*tauriSidecar, *desktopSession, *parentPID = true, "test-session", uint(os.Getpid())
	badConfig := "game: [broken\n"
	if err := os.WriteFile(*configPath, []byte(badConfig), 0600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(t.TempDir())
	if err := run(); err == nil {
		t.Fatal("无效配置应在启动阶段返回错误")
	}
	data, err := os.ReadFile(filepath.Join(*dataDir, "logs", "startup-go.log"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "退出失败") || !strings.Contains(string(data), "读取配置") {
		t.Fatalf("缺少启动错误: %s", data)
	}
	if strings.Contains(string(data), "test-session") {
		t.Fatal("日志不能记录会话令牌")
	}
	original, _ := os.ReadFile(*configPath)
	if string(original) != badConfig {
		t.Fatal("初始化覆盖了损坏配置")
	}
}

func TestConfigMigrationPreservesUserSettingsAndSource(t *testing.T) {
	dir := t.TempDir()
	source, target := filepath.Join(dir, "old.yaml"), filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(source, []byte("game:\n  auto_confirm: false\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := migrateConfigFrom(source, target); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(target)
	if !strings.Contains(string(first), "false") {
		t.Fatal("旧设置没有迁移")
	}
	if err := os.WriteFile(target, []byte("user settings"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := migrateConfigFrom(source, target); err != nil {
		t.Fatal(err)
	}
	actual, _ := os.ReadFile(target)
	if string(actual) != "user settings" {
		t.Fatal("已有用户配置被覆盖")
	}
	if _, err := os.Stat(source); err != nil {
		t.Fatal("旧文件不应被删除")
	}
}
