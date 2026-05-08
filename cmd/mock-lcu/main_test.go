package main

import "testing"

func TestParseConfigUsesDefaults(t *testing.T) {
	cfg, err := parseConfig(nil)
	if err != nil {
		t.Fatalf("parseConfig returned error: %v", err)
	}

	if cfg.addr != "127.0.0.1:19365" {
		t.Fatalf("expected default addr, got %q", cfg.addr)
	}
	if cfg.scenarioDir != "internal/mocklcu/fixtures/default" {
		t.Fatalf("expected default scenario dir, got %q", cfg.scenarioDir)
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
