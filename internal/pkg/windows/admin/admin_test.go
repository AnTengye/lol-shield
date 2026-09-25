//go:build windows

package admin

import (
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestParentProcessHelper(t *testing.T) {
	if os.Getenv("SHIELD_PARENT_TEST") != "1" {
		return
	}
	_, _ = io.Copy(io.Discard, os.Stdin)
	os.Exit(0)
}

func TestWatchParentStopsWhenDesktopExits(t *testing.T) {
	command := exec.Command(os.Args[0], "-test.run=^TestParentProcessHelper$")
	command.Env = append(os.Environ(), "SHIELD_PARENT_TEST=1")
	input, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = command.Process.Kill(); _ = input.Close() })
	done, stopped := make(chan struct{}), make(chan struct{})
	defer close(done)
	if err := WatchParent(uint32(command.Process.Pid), done, func() { close(stopped) }); err != nil {
		t.Fatal(err)
	}
	_ = input.Close()
	if err := command.Wait(); err != nil {
		t.Fatal(err)
	}
	select {
	case <-stopped:
	case <-time.After(3 * time.Second):
		t.Fatal("桌面进程退出后后台未收到停止通知")
	}
}

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

func TestBuildElevatedArgsKeepsUserDataDirectory(t *testing.T) {
	args := buildElevatedArgs([]string{"--data-dir", `C:\Users\test user\AppData\Local\Shield`}, true)
	if !strings.Contains(args, `"C:\Users\test user\AppData\Local\Shield"`) || !strings.Contains(args, "--data-dir") {
		t.Fatalf("用户数据目录丢失: %s", args)
	}
}

func TestBuildElevatedArgsSkipsSidecarFlagWhenNotRequested(t *testing.T) {
	t.Parallel()

	args := buildElevatedArgs([]string{"--config", "config.yaml"}, false)

	if strings.Contains(args, "--tauri-sidecar") {
		t.Fatalf("expected no sidecar flag when not requested, args=%q", args)
	}
}
