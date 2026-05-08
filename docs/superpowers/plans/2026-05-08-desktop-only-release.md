# Desktop-Only Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove the browser-era embedded frontend delivery path, make Tauri the only release artifact, and launch the desktop sidecar backend without a visible console window.

**Architecture:** The Go backend will stop serving embedded SPA assets and will remain an HTTP API plus `/riot` asset server for the Tauri webview. Release automation will build only the sidecar binary and the Tauri bundle, while the Windows launch path will preserve elevation and sidecar arguments without surfacing a console window.

**Tech Stack:** Go, Gin, Viper, GitHub Actions, Tauri v2, Rust, tauri-plugin-shell, Windows ShellExecute

---

## File Map

- Modify: `.github/workflows/release.yml`
- Modify: `cmd/shield/main.go`
- Modify: `configs/config.go`
- Delete: `internal/client/browser.go`
- Delete: `internal/client/frontend.go`
- Delete: `internal/client/frontend_test.go`
- Modify: `internal/client/router.go`
- Modify: `internal/client/shield.go`
- Modify: `internal/client/shield_test.go`
- Modify: `internal/pkg/windows/admin/admin.go`
- Create: `internal/pkg/windows/admin/admin_test.go`
- Modify: `src-tauri/src/lib.rs`
- Modify: `src-tauri/Cargo.toml` only if a Windows-specific helper crate is required
- Delete: `internal/client/web/dist/index.html` if no compile-time embed remains
- Modify: `.gitignore` if the placeholder exception for `internal/client/web/dist/index.html` becomes obsolete
- Modify: `docs/mock-lcu.md` if local run instructions mention the old browser path

### Task 1: Remove Browser Runtime Coupling From The Go Backend

**Files:**
- Delete: `internal/client/browser.go`
- Delete: `internal/client/frontend.go`
- Delete: `internal/client/frontend_test.go`
- Modify: `internal/client/router.go`
- Modify: `internal/client/shield.go`
- Modify: `configs/config.go`
- Test: `internal/client/shield_test.go`

- [ ] **Step 1: Write the failing router/runtime tests**

Add or update `internal/client/shield_test.go` with assertions that the backend no longer serves a frontend root and no longer depends on browser auto-open behavior:

```go
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

	shield := NewShieldWithLCU(lcuapi.NewHTTPService("http://127.0.0.1:19365"))
	if shield.lcuService == nil {
		t.Fatal("expected LCU service to remain configured")
	}
}
```

- [ ] **Step 2: Run the focused tests to verify failure**

Run:

```bash
go test ./internal/client -run "TestBackendRootReturnsNotFoundWithoutEmbeddedFrontend|TestNewShieldWithMockModeStillInitializesLCUService" -v
```

Expected: the root-path assertion fails because `registerFrontendRoutes` still serves `index.html`.

- [ ] **Step 3: Remove the embedded frontend runtime path**

Apply the minimal code changes:

```go
// internal/client/router.go
func AddRouter(r *gin.Engine, p *Shield) {
	riotReq := r.Group("riot")
	riotReq.GET("*assets", GetAssets(p))
	v1 := r.Group("v1")
	// existing API routes stay unchanged...

	r.GET("ws", func(c *gin.Context) {
		client := ws.ServerWebsocket(c.Writer, c.Request)
		if client != nil {
			p.webWs = client
			syslog.L.Infof("成功连接到Web页面-ID：%s", client.GetUid())
			p.Notice()
		}
	})
}
```

```go
// internal/client/shield.go
func (p *Shield) notifyQuit() error {
	if viper.GetBool(configs.Dev) {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	webAddr := viper.GetString(configs.WebAddr)
	srv := NewServer(webAddr, p)
	p.httpSrv = srv
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	g, c := errgroup.WithContext(p.ctx)

	g.Go(func() error {
		err := p.httpSrv.ListenAndServe()
		if err != nil || !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	})

	g.Go(func() error {
		<-c.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		return p.httpSrv.Shutdown(ctx)
	})

	g.Go(func() error {
		for {
			select {
			case <-p.ctx.Done():
				return p.ctx.Err()
			case <-interrupt:
				_ = p.Stop()
			}
		}
	})

	err := g.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}
```

Also remove the `WebAutoOpen` default and any references to `registerFrontendRoutes`, `openWebPage`, `waitForWebReady`, and `normalizeWebURL`.

- [ ] **Step 4: Delete obsolete browser-only files and config entries**

Delete the unused files and stale config constant/default:

```go
// configs/config.go
const (
	ShowVersion      = "show_version"
	Dev              = "dev"
	WebAddr          = "web.addr"
	LogFilepath      = "log.filepath"
	LCUTokenFromFile = "lcu.token_file"
	MockLCUEnabled   = "mock_lcu.enabled"
	MockLCUBaseURL   = "mock_lcu.base_url"
	MockLCUScenario  = "mock_lcu.scenario"
)

func Init(configPath string) {
	viper.SetDefault(ShowVersion, true)
	viper.SetDefault(Dev, false)
	viper.SetDefault(WebAddr, ":9365")
	viper.SetDefault(LCUTokenFromFile, false)
	viper.SetDefault(MockLCUEnabled, false)
	viper.SetDefault(MockLCUBaseURL, "http://127.0.0.1:19365")
	viper.SetDefault(MockLCUScenario, "default")
	// remaining defaults unchanged...
}
```

Remove `internal/client/web/dist/index.html` and the `.gitignore` allowance for that placeholder if `//go:embed web/dist` is gone.

- [ ] **Step 5: Run package tests to verify the cleanup**

Run:

```bash
go test ./internal/client ./configs -v
```

Expected: PASS with no embedded-frontend compile requirement and the new root-path behavior locked in.

- [ ] **Step 6: Commit**

```bash
git add internal/client/router.go internal/client/shield.go internal/client/shield_test.go configs/config.go .gitignore
git rm internal/client/browser.go internal/client/frontend.go internal/client/frontend_test.go internal/client/web/dist/index.html
git commit -m "refactor: remove embedded browser frontend runtime"
```

### Task 2: Make Desktop Sidecar Launch Without A Visible Console Window

**Files:**
- Modify: `cmd/shield/main.go`
- Modify: `internal/pkg/windows/admin/admin.go`
- Create: `internal/pkg/windows/admin/admin_test.go`
- Modify: `src-tauri/src/lib.rs`
- Test: `internal/pkg/windows/admin/admin_test.go`

- [ ] **Step 1: Write the failing Windows launch argument tests**

Create `internal/pkg/windows/admin/admin_test.go` to lock down sidecar-aware elevation behavior:

```go
func TestBuildElevatedArgsAppendsSidecarFlagWhenMissing(t *testing.T) {
	got := buildElevatedArgs([]string{"lol-shield.exe", "-c", "config.yaml"}, true)
	want := []string{"-c", "config.yaml", sidecarArg}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestBuildElevatedArgsKeepsExistingSidecarFlag(t *testing.T) {
	got := buildElevatedArgs([]string{"lol-shield.exe", sidecarArg, "-c", "config.yaml"}, true)
	want := []string{sidecarArg, "-c", "config.yaml"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
```

- [ ] **Step 2: Run the focused admin tests to verify failure**

Run:

```bash
go test ./internal/pkg/windows/admin -run "TestBuildElevatedArgs" -v
```

Expected: FAIL because the current worktree version of `admin.go` does not expose `buildElevatedArgs` or preserve the sidecar flag shape.

- [ ] **Step 3: Restore sidecar-aware elevation helpers and hide the elevated window**

Update `internal/pkg/windows/admin/admin.go` to use explicit helpers and a hidden show mode:

```go
const (
	sidecarEnv = "LOL_SHIELD_TAURI_SIDECAR"
	sidecarArg = "--tauri-sidecar"
)

func RunMeElevated() error {
	verb := "runas"
	exe, _ := os.Executable()
	cwd, _ := os.Getwd()
	args := strings.Join(buildElevatedArgs(os.Args, os.Getenv(sidecarEnv) == "1"), " ")
	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)
	var showCmd int32 = 0 // SW_HIDE
	err := windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, showCmd)
	if err != nil {
		return err
	}
	os.Exit(-2)
	return nil
}

func buildElevatedArgs(args []string, sidecar bool) []string {
	if len(args) <= 1 {
		if sidecar {
			return []string{sidecarArg}
		}
		return nil
	}

	elevatedArgs := append([]string{}, args[1:]...)
	if !sidecar {
		return elevatedArgs
	}
	for _, arg := range elevatedArgs {
		if arg == sidecarArg {
			return elevatedArgs
		}
	}
	return append(elevatedArgs, sidecarArg)
}
```

Update `cmd/shield/main.go` to keep explicit sidecar detection:

```go
var tauriSidecar = flag.Bool("tauri-sidecar", false, "以 Tauri sidecar 模式启动")

func main() {
	flag.Parse()
	configs.Init(*configPath)
	syslog.Init()
	isSidecar := isSidecarLaunch(*tauriSidecar, os.Args)
	if isSidecar {
		os.Setenv("LOL_SHIELD_TAURI_SIDECAR", "1")
	}
	if err := admin.RunMeElevated(); err != nil {
		syslog.L.Fatalf("请求管理员权限失败: %v", err)
	}
	// remaining startup unchanged...
}

func isSidecarLaunch(sidecarFlag bool, args []string) bool {
	return sidecarFlag || os.Getenv("LOL_SHIELD_TAURI_SIDECAR") == "1" || slices.Contains(args[1:], "--tauri-sidecar")
}
```

- [ ] **Step 4: Pass sidecar launch metadata from Tauri explicitly**

Keep the Tauri sidecar launch aligned with the Go side:

```rust
let sidecar = app
    .shell()
    .sidecar("lol-shield")
    .expect("failed to create lol-shield sidecar command")
    .arg("--tauri-sidecar")
    .env("LOL_SHIELD_TAURI_SIDECAR", "1");
```

If the console window still appears after `SW_HIDE`, add the narrowest necessary Rust-side creation flag or Windows subsystem adjustment in the same task, and keep the rest of the sidecar lifecycle unchanged.

- [ ] **Step 5: Run tests for the elevation/startup seam**

Run:

```bash
go test ./internal/pkg/windows/admin ./cmd/shield -v
```

Expected: PASS with sidecar-aware elevation argument coverage.

- [ ] **Step 6: Commit**

```bash
git add cmd/shield/main.go internal/pkg/windows/admin/admin.go internal/pkg/windows/admin/admin_test.go src-tauri/src/lib.rs
git commit -m "fix: hide desktop sidecar console window"
```

### Task 3: Simplify Release To Tauri-Only Artifacts And Verify Mock Desktop Flow

**Files:**
- Modify: `.github/workflows/release.yml`
- Modify: `docs/mock-lcu.md`
- Test: `cmd/mock-lcu/main_test.go`

- [ ] **Step 1: Write the failing workflow expectation as a review checklist**

Capture the intended workflow shape in the plan execution notes by asserting these conditions during review:

```text
- release.yml must not contain a standalone "Build frontend" step
- release.yml must not contain a "Sync embedded frontend dist" step
- release.yml must still install frontend dependencies
- release.yml must still build the Go sidecar before tauri:build
- release.yml must still publish the NSIS executable and checksum files
```

- [ ] **Step 2: Update the release workflow to Tauri-only bundling**

Modify `.github/workflows/release.yml` to remove the duplicate frontend build and embedded dist sync:

```yaml
- name: Install frontend dependencies
  run: pnpm --dir frontend install --frozen-lockfile

- name: Build Go sidecar
  shell: pwsh
  run: ./scripts/build-tauri-sidecar.ps1

- name: Build Tauri NSIS bundle
  run: pnpm --dir frontend tauri:build
```

Leave the cache, checksum, and release publishing steps intact.

- [ ] **Step 3: Refresh local workflow documentation**

Update `docs/mock-lcu.md` so the desktop run path points to Tauri-first usage instead of implying a browser frontend:

```md
## Desktop-first local flow

1. Start `mock-lcu` on `127.0.0.1:19365`.
2. Start the desktop app or run the sidecar backend only for API verification.
3. Verify `/v1/game/running`, `/v1/history/:uid`, and `/riot/...` responses through the desktop-backed local server.

The project no longer ships or validates an embedded browser frontend release path.
```

- [ ] **Step 4: Run verification commands**

Run:

```bash
go test ./internal/mocklcu ./internal/client ./internal/pkg/windows/admin ./cmd/mock-lcu ./cmd/shield -v
pnpm --dir frontend install --frozen-lockfile
pnpm --dir frontend tauri:build
```

Expected:

- Go test commands PASS
- Tauri build completes without relying on `internal/client/web/dist`

- [ ] **Step 5: Perform one local desktop smoke test**

Run the desktop-oriented mock flow:

```bash
go run ./cmd/mock-lcu -addr 127.0.0.1:19365
go run ./cmd/shield -c mock.config.yaml --tauri-sidecar
```

Verify manually:

```text
- no visible backend console window remains after elevation completes
- GET http://127.0.0.1:9365/v1/game/running returns 200
- GET http://127.0.0.1:9365/v1/history/de06293d-082d-59c2-83a6-273ab88164bc?page=0&pageSize=9 returns 200
- GET http://127.0.0.1:9365/ returns 404
```

- [ ] **Step 6: Commit**

```bash
git add .github/workflows/release.yml docs/mock-lcu.md
git commit -m "build: switch release to desktop-only flow"
```

## Self-Review

- Spec coverage:
  - desktop-only release artifact: Task 3
  - no embedded browser runtime path: Task 1
  - no visible backend console window: Task 2
  - mock LCU desktop usability retained: Tasks 1 and 3
- Placeholder scan: no TODO/TBD markers remain in the tasks.
- Type consistency: `sidecarArg`, `buildElevatedArgs`, and `isSidecarLaunch` are defined before later tasks rely on them.
