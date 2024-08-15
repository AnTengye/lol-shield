package configs

import "github.com/spf13/viper"

const (
	Version   = "v1.0.0"
	UpdateApi = "https://oss.bigorange.work/lol/latest.json"
)
const (
	ShowVersion     = "show_version"
	Dev             = "dev"
	WebAddr         = "web.addr"
	LogFilepath     = "log.filepath"
	LogSize         = "log.size"
	LogBackups      = "log.backups"
	LogAge          = "log.age"
	LogCompress     = "log.compress"
	LogLevel        = "log.level"
	GameAutoConfirm = "game.auto_confirm"
	GameAutoPick    = "game.auto_pick"
	GameAutoBan     = "game.auto_ban"
	TempButton      = "temp.test"
	// GameAutoConfirm = "game.auto_confirm"
)

func Init(configPath string) {
	viper.SetDefault(ShowVersion, true)
	viper.SetDefault(Dev, true)
	viper.SetDefault(LogFilepath, "./log")
	viper.SetDefault(LogSize, 1024)
	viper.SetDefault(LogBackups, 7)
	viper.SetDefault(LogAge, 7)
	viper.SetDefault(LogCompress, true)
	viper.SetDefault(LogLevel, "debug")
	viper.SetDefault(GameAutoConfirm, true)
	viper.SetDefault(GameAutoPick, 0)
	viper.SetDefault(GameAutoBan, 0)
	viper.SetDefault(WebAddr, ":9365")
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		err := viper.WriteConfig()
		if err != nil {
			panic(err)
		}
	}
}
