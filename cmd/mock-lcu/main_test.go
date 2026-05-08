package main

import (
	"errors"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/AnTengye/lol-shield/internal/mocklcu"
)

func TestParseConfigUsesDefaults(t *testing.T) {
	cfg, err := parseConfig(nil)
	if err != nil {
		t.Fatalf("parseConfig returned error: %v", err)
	}

	if cfg.addr != "127.0.0.1:19365" {
		t.Fatalf("expected default addr, got %q", cfg.addr)
	}
	if !filepath.IsAbs(cfg.scenarioDir) {
		t.Fatalf("expected absolute default scenario dir, got %q", cfg.scenarioDir)
	}
	if filepath.Base(cfg.scenarioDir) != "default" {
		t.Fatalf("expected default scenario dir suffix, got %q", cfg.scenarioDir)
	}
}

func TestParseConfigAppliesArgs(t *testing.T) {
	cfg, err := parseConfig([]string{"-addr", "127.0.0.1:20001", "-scenario-dir", "fixtures/custom"})
	if err != nil {
		t.Fatalf("parseConfig returned error: %v", err)
	}

	if cfg.addr != "127.0.0.1:20001" {
		t.Fatalf("expected addr override, got %q", cfg.addr)
	}
	if cfg.scenarioDir != "fixtures/custom" {
		t.Fatalf("expected scenario dir override, got %q", cfg.scenarioDir)
	}
}

func TestRunLoadsScenarioAndStartsServer(t *testing.T) {
	originalLoadScenario := loadScenario
	originalNewServer := newServer
	originalListenAndServe := listenAndServe
	t.Cleanup(func() {
		loadScenario = originalLoadScenario
		newServer = originalNewServer
		listenAndServe = originalListenAndServe
	})

	wantScenarioDir := filepath.Join("fixtures", "default")
	scenario := &mocklcu.Scenario{}
	serverSentinel := errors.New("server started")

	var gotAddr string
	var gotScenarioDir string
	var gotHandler http.Handler

	loadScenario = func(dir string) (*mocklcu.Scenario, error) {
		gotScenarioDir = dir
		return scenario, nil
	}
	newServer = func(s *mocklcu.Scenario) http.Handler {
		if s != scenario {
			t.Fatalf("expected newServer to receive loaded scenario")
		}
		return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	}
	listenAndServe = func(addr string, handler http.Handler) error {
		gotAddr = addr
		gotHandler = handler
		return serverSentinel
	}

	err := run([]string{"-addr", "127.0.0.1:20001", "-scenario-dir", wantScenarioDir})
	if !errors.Is(err, serverSentinel) {
		t.Fatalf("expected server sentinel error, got %v", err)
	}
	if gotAddr != "127.0.0.1:20001" {
		t.Fatalf("expected addr 127.0.0.1:20001, got %q", gotAddr)
	}
	if gotScenarioDir != wantScenarioDir {
		t.Fatalf("expected scenario dir %q, got %q", wantScenarioDir, gotScenarioDir)
	}
	if gotHandler == nil {
		t.Fatalf("expected non-nil handler")
	}
}
