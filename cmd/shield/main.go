package main

import (
	"flag"
	"os"

	"github.com/AnTengye/lol-shield/internal/client"
	"github.com/AnTengye/lol-shield/internal/pkg/windows/admin"
	"github.com/spf13/viper"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
)

var configPath = flag.String("c", "config.yaml", "配置文件路径")

func main() {
	flag.Parse()
	// 初始化配置文件
	configs.Init(*configPath)
	syslog.Init()
	syslog.L.Infof("配置初始化完成")
	if os.Getenv("LOL_SHIELD_TAURI_SIDECAR") == "1" {
		viper.Set(configs.WebAutoOpen, false)
	}
	admin.MustRunWithAdmin()
	if viper.GetBool(configs.Dev) {
		syslog.L.Infof("当前为开发模式: 启动流程仍使用主程序自身提权。")
	}
	shield := client.NewShield()
	if err := shield.Run(); err != nil {
		syslog.L.Fatal(err)
	}
}
