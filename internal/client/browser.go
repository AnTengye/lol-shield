package client

import (
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func openWebPage(addr string) error {
	targetURL := normalizeWebURL(addr)
	switch runtime.GOOS {
	case "windows":
		return exec.Command("cmd", "/c", "start", "", targetURL).Start()
	case "darwin":
		return exec.Command("open", targetURL).Start()
	default:
		return exec.Command("xdg-open", targetURL).Start()
	}
}

func normalizeWebURL(addr string) string {
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return addr
	}
	trimmed := strings.TrimSpace(addr)
	if strings.HasPrefix(trimmed, ":") {
		return "http://127.0.0.1" + trimmed
	}
	if strings.HasPrefix(trimmed, "0.0.0.0:") {
		return "http://127.0.0.1:" + strings.TrimPrefix(trimmed, "0.0.0.0:")
	}
	return fmt.Sprintf("http://%s", trimmed)
}

func waitForWebReady(addr string, timeout time.Duration) bool {
	targetURL := normalizeWebURL(addr)
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 300 * time.Millisecond}
	for time.Now().Before(deadline) {
		resp, err := client.Get(targetURL)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 500 {
				return true
			}
		}
		time.Sleep(120 * time.Millisecond)
	}
	return false
}
