package main

import (
	"encoding/json"
	"flag"
	"github.com/AnTengye/lol-shield/internal/pkg/windows/admin"
	"io"
	"net/http"
	"strings"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/client"
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
	admin.MustRunWithAdmin()
	// 初始化配置文件
	configs.Init(*configPath)
	syslog.Init()
	err := checkUpdate()
	if err != nil {
		syslog.L.Errorf("检查更新失败:%v", err)
	}
	syslog.L.Infof("程序初始化完成")
	shield := client.NewShield()
	if err = shield.Run(); err != nil {
		syslog.L.Fatal(err)
	}
}

func checkUpdate() error {
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
			panic("force")
		}
		syslog.L.Infof("存在新版本:%s，建议更新", updateInfo.Version)
	}
	return nil
}
