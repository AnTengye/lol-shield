package admin

import (
	"strings"
	"testing"
)

func TestBuildElevatedArgsAddsSidecarFlagWhenRequested(t *testing.T) {
	t.Parallel()

	args := buildElevatedArgs([]string{"--config", "mock config.yaml"}, true)

	if !strings.Contains(args, "--tauri-sidecar") {
		t.Fatalf("expected sidecar flag to be propagated, args=%q", args)
	}
	if !strings.Contains(args, "\"mock config.yaml\"") {
		t.Fatalf("expected existing args to remain quoted, args=%q", args)
	}
}

func TestBuildElevatedArgsDoesNotDuplicateSidecarFlag(t *testing.T) {
	t.Parallel()

	args := buildElevatedArgs([]string{"--tauri-sidecar", "--config", "config.yaml"}, true)

	if count := strings.Count(args, "--tauri-sidecar"); count != 1 {
		t.Fatalf("expected sidecar flag once, got %d in %q", count, args)
	}
}

func TestBuildElevatedArgsSkipsSidecarFlagWhenNotRequested(t *testing.T) {
	t.Parallel()

	args := buildElevatedArgs([]string{"--config", "config.yaml"}, false)

	if strings.Contains(args, "--tauri-sidecar") {
		t.Fatalf("expected no sidecar flag when not requested, args=%q", args)
	}
}
