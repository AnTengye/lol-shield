package main

import (
	"encoding/json"
	"flag"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AnTengye/lol-shield/internal/client"
	"github.com/AnTengye/lol-shield/internal/pkg/windows/admin"
	"github.com/spf13/viper"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
)

type (
	VersionInfo struct {
		Version string `json:"Version"`
		Url     string `json:"url"`
		Force   bool   `json:"force"`
	}
)

var configPath = flag.String("c", "config.yaml", "配置文件路径")

func main() {
	flag.Parse()
	// 初始化配置文件
	configs.Init(*configPath)
	syslog.Init()
	syslog.L.Infof("配置初始化完成,正在检查更新中...")
	if viper.GetBool(configs.Dev) {
		err := admin.RunAsAdmin("lcu-info.exe", "")
		if err != nil {
			syslog.L.Fatal("请允许管理员权限启动，否则无法获取客户端信息:", err)
		} else {
			syslog.L.Infof("子程序已以管理员权限启动")
		}
	} else {
		admin.MustRunWithAdmin()
	}
	err := checkUpdate()
	if err != nil {
		syslog.L.Errorf("检查更新失败:%v", err)
	}
	shield := client.NewShield()
	if err = shield.Run(); err != nil {
		syslog.L.Fatal(err)
	}
}

func checkUpdate() error {
	http.DefaultClient.Timeout = time.Second * 3
	resp, err := http.Get(configs.UpdateApi)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bts, _ := io.ReadAll(resp.Body)
	updateInfo := &VersionInfo{}
	if err = json.Unmarshal(bts, updateInfo); err != nil {
		return err
	}
	if updateInfo.Version == "" || updateInfo.Url == "" {
		return nil
	}
	compare := strings.Compare(configs.Version, updateInfo.Version)
	if compare < 0 {
		if updateInfo.Force {
			syslog.L.Infof("存在新版本:%s，需要强制更新", updateInfo.Version)
			panic("force update")
		}
		syslog.L.Infof("存在新版本:%s，建议更新", updateInfo.Version)
	}
	return nil
}
