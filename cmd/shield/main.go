package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"

	"github.com/AnTengye/lol-shield/internal/pkg/historycache"

	"github.com/AnTengye/lol-shield/internal/client"
	"github.com/AnTengye/lol-shield/internal/core/lcuapi"
	"github.com/AnTengye/lol-shield/internal/pkg/windows/admin"
	"github.com/spf13/viper"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
)

var configPath = flag.String("c", "config.yaml", "配置文件路径")
var tauriSidecar = flag.Bool("tauri-sidecar", false, "标记当前进程由 Tauri sidecar 启动")
var dataDir = flag.String("data-dir", "", "本地用户数据目录")
var desktopSession = flag.String("desktop-session", "", "桌面启动会话")
var parentPID = flag.Uint("parent-pid", 0, "负责本次启动的桌面进程")

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (result error) {
	dir := *dataDir
	if dir == "" {
		base, err := os.UserCacheDir()
		if err != nil {
			return err
		}
		dir = filepath.Join(base, "work.bigorange.lol-shield")
	}
	dir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	logs := filepath.Join(dir, "logs")
	if err := os.MkdirAll(logs, 0700); err != nil {
		return fmt.Errorf("创建启动日志目录失败: %w", err)
	}
	file, err := os.OpenFile(filepath.Join(logs, "startup-go.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("创建启动日志失败: %w", err)
	}
	defer file.Close()
	startup := log.New(io.MultiWriter(file, os.Stderr), "", log.LstdFlags|log.Lmicroseconds)
	_ = debug.SetCrashOutput(file, debug.CrashOptions{})
	defer debug.SetCrashOutput(nil, debug.CrashOptions{})
	defer func() {
		if value := recover(); value != nil {
			result = fmt.Errorf("启动异常: %v", value)
			startup.Printf("%s\n%s", result, debug.Stack())
		}
		if result != nil {
			startup.Printf("pid=%d 退出失败: %v", os.Getpid(), result)
		} else {
			startup.Printf("pid=%d 正常退出", os.Getpid())
		}
	}()
	cwd, _ := os.Getwd()
	startup.Printf("入口 pid=%d parent=%d cwd=%q data=%q", os.Getpid(), *parentPID, cwd, dir)
	path := *configPath
	if *tauriSidecar {
		if path == "config.yaml" {
			path = filepath.Join(dir, "config.yaml")
		}
		if err := migrateConfig(path); err != nil {
			return err
		}
		if *desktopSession == "" || *parentPID == 0 {
			return fmt.Errorf("桌面启动缺少会话或父进程信息，请从桌面主程序启动")
		}
	}
	startup.Printf("读取配置 %q", path)
	if err := configs.Init(path); err != nil {
		return err
	}
	if *tauriSidecar {
		viper.Set(configs.LogFilepath, logs)
	}
	if err := syslog.Init(); err != nil {
		return err
	}
	defer syslog.L.Sync()
	syslog.L.Infof("配置初始化完成")
	if !viper.GetBool(configs.MockLCUEnabled) {
		if *tauriSidecar {
			// 桌面壳直接启动最终提权进程，禁止再次自我提权而丢失进程归属。
			if err := admin.RequireElevated(); err != nil {
				return err
			}
		} else {
			admin.MustRunWithAdmin(false)
		}
	}
	startup.Printf("权限检查通过 pid=%d", os.Getpid())
	if *tauriSidecar {
		viper.Set(configs.WebAddr, "127.0.0.1:9365")
	}
	if viper.GetBool(configs.Dev) {
		syslog.L.Infof("当前为开发模式")
	}
	var lcuSvc lcuapi.Service
	if viper.GetBool(configs.MockLCUEnabled) {
		lcuSvc = lcuapi.NewHTTPService(viper.GetString(configs.MockLCUBaseURL))
	} else {
		lcuSvc = lcuapi.New()
	}
	shield := client.NewShieldWithLCU(lcuSvc)
	if *tauriSidecar {
		shield.ConfigureDesktop(*desktopSession)
		done := make(chan struct{})
		defer close(done)
		if err := admin.WatchParent(uint32(*parentPID), done, func() { _ = shield.Stop() }); err != nil {
			return fmt.Errorf("监控桌面进程失败: %w", err)
		}
	}
	cache, cacheErr := historycache.Open(filepath.Join(dir, "history-cache", "v1"), historycache.MaxBytes)
	if cacheErr != nil {
		syslog.L.Warnf("本地缓存不可用: %v", cacheErr)
	} else {
		defer cache.Close()
	}
	source := "live"
	if viper.GetBool(configs.MockLCUEnabled) {
		source = "mock:" + viper.GetString(configs.MockLCUScenario)
	}
	shield.ConfigureHistory(cache, source, cacheErr)
	startup.Printf("开始启动本地服务 pid=%d", os.Getpid())
	return shield.Run()
}

// 仅首次迁移安装目录中的旧配置；不覆盖已有用户配置。
func migrateConfig(destination string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return migrateConfigFrom(filepath.Join(filepath.Dir(exe), "config.yaml"), destination)
}

func migrateConfigFrom(source, destination string) error {
	if _, err := os.Stat(destination); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	if source == destination {
		return nil
	}
	data, err := os.ReadFile(source)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("读取旧配置失败: %w", err)
	}
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if os.IsExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	return err
}
