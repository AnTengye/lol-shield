package configs

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

const (
	Version = "v1.0.0"
)
const (
	ShowVersion      = "show_version"
	Dev              = "dev"
	WebAddr          = "web.addr"
	LogFilepath      = "log.filepath"
	LogSize          = "log.size"
	LogBackups       = "log.backups"
	LogAge           = "log.age"
	LogCompress      = "log.compress"
	LogLevel         = "log.level"
	GameAutoConfirm  = "game.auto_confirm"
	GameAutoPick     = "game.auto_pick"
	GameAutoBan      = "game.auto_ban"
	TempButton       = "temp.test"
	LCUTokenFromFile = "lcu.token_file"
	MockLCUEnabled   = "mock_lcu.enabled"
	MockLCUBaseURL   = "mock_lcu.base_url"
	MockLCUScenario  = "mock_lcu.scenario"
	// GameAutoConfirm = "game.auto_confirm"
)

func Init(configPath string) error {
	viper.SetDefault(ShowVersion, true)
	viper.SetDefault(Dev, false)
	viper.SetDefault(LogFilepath, "./log")
	viper.SetDefault(LogSize, 1024)
	viper.SetDefault(LogBackups, 7)
	viper.SetDefault(LogAge, 7)
	viper.SetDefault(LogCompress, true)
	viper.SetDefault(LogLevel, "info")
	viper.SetDefault(GameAutoConfirm, true)
	viper.SetDefault(GameAutoPick, 0)
	viper.SetDefault(GameAutoBan, 0)
	viper.SetDefault(WebAddr, ":9365")
	viper.SetDefault(LCUTokenFromFile, false)
	viper.SetDefault(MockLCUEnabled, false)
	viper.SetDefault(MockLCUBaseURL, "http://127.0.0.1:19365")
	viper.SetDefault(MockLCUScenario, "default")
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("读取配置 %s 失败: %w", configPath, err)
		}
		// 只创建缺失的配置，解析失败或权限错误不能覆盖用户原文件。
		if err := viper.SafeWriteConfigAs(configPath); err != nil {
			return fmt.Errorf("创建配置 %s 失败: %w", configPath, err)
		}
	}
	return nil
}
