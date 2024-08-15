package syslog

import (
	"os"
	"time"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var L *zap.SugaredLogger

func Init() {
	lumberJackLogger := &lumberjack.Logger{
		// 日志文件以日期命名
		Filename:   viper.GetString(configs.LogFilepath) + "/" + time.Now().Format("20060102") + ".log",
		MaxSize:    viper.GetInt(configs.LogSize),
		MaxBackups: viper.GetInt(configs.LogBackups),
		MaxAge:     viper.GetInt(configs.LogAge),
		Compress:   viper.GetBool(configs.LogCompress),
		LocalTime:  true,
	}
	syncFile := zapcore.AddSync(lumberJackLogger) // 打印到文件
	syncConsole := zapcore.AddSync(os.Stderr)     // 打印到控制台
	syncer := zapcore.NewMultiWriteSyncer(syncFile, syncConsole)
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncodeDuration = zapcore.StringDurationEncoder
	level, err := zapcore.ParseLevel(viper.GetString(configs.LogLevel))
	if err != nil {
		panic("level error")
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		syncer,
		zap.NewAtomicLevelAt(level),
	)
	L = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
}
