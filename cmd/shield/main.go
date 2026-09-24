package main

import (
	"flag"
	"github.com/AnTengye/lol-shield/internal/pkg/historycache"
	"os"
	"path/filepath"

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

func main() {
	flag.Parse()
	// 初始化配置文件
	configs.Init(*configPath)
	syslog.Init()
	syslog.L.Infof("配置初始化完成")
	if !viper.GetBool(configs.MockLCUEnabled) {
		admin.MustRunWithAdmin(*tauriSidecar)
	}
	if *tauriSidecar {
		viper.Set(configs.WebAddr, "127.0.0.1:9365")
	}
	if viper.GetBool(configs.Dev) {
		syslog.L.Infof("当前为开发模式: 启动流程仍使用主程序自身提权。")
	}
	var lcuSvc lcuapi.Service
	if viper.GetBool(configs.MockLCUEnabled) {
		lcuSvc = lcuapi.NewHTTPService(viper.GetString(configs.MockLCUBaseURL))
	} else {
		lcuSvc = lcuapi.New()
	}
	shield := client.NewShieldWithLCU(lcuSvc)
	dir := *dataDir
	if dir == "" {
		base, err := os.UserCacheDir()
		if err != nil {
			syslog.L.Fatal(err)
		}
		dir = filepath.Join(base, "work.bigorange.lol-shield")
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
	if err := shield.Run(); err != nil {
		syslog.L.Fatal(err)
	}
}
