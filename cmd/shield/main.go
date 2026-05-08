package main

import (
	"flag"

	"github.com/AnTengye/lol-shield/internal/client"
	"github.com/AnTengye/lol-shield/internal/core/lcuapi"
	"github.com/AnTengye/lol-shield/internal/pkg/windows/admin"
	"github.com/spf13/viper"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
)

var configPath = flag.String("c", "config.yaml", "配置文件路径")
var tauriSidecar = flag.Bool("tauri-sidecar", false, "标记当前进程由 Tauri sidecar 启动")

func main() {
	flag.Parse()
	// 初始化配置文件
	configs.Init(*configPath)
	syslog.Init()
	syslog.L.Infof("配置初始化完成")
	admin.MustRunWithAdmin(*tauriSidecar)
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
	if err := shield.Run(); err != nil {
		syslog.L.Fatal(err)
	}
}
