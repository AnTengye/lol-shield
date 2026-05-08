package configs

import "github.com/spf13/viper"

const (
	Version = "v1.0.0"
)
const (
	ShowVersion      = "show_version"
	Dev              = "dev"
	WebAddr          = "web.addr"
	WebAutoOpen      = "web.auto_open"
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

func Init(configPath string) {
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
	viper.SetDefault(WebAutoOpen, true)
	viper.SetDefault(LCUTokenFromFile, false)
	viper.SetDefault(MockLCUEnabled, false)
	viper.SetDefault(MockLCUBaseURL, "http://127.0.0.1:19365")
	viper.SetDefault(MockLCUScenario, "default")
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		err := viper.WriteConfig()
		if err != nil {
			panic(err)
		}
	}
}
