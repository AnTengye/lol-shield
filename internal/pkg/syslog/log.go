package syslog

import (
	"os"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var L *zap.SugaredLogger

func Init() {
	writeSyncer := zapcore.AddSync(
		&lumberjack.Logger{
			Filename:   viper.GetString(configs.LogFilepath),
			MaxSize:    viper.GetInt(configs.LogSize),
			MaxBackups: viper.GetInt(configs.LogBackups),
			MaxAge:     viper.GetInt(configs.LogAge),
			Compress:   viper.GetBool(configs.LogCompress),
			LocalTime:  true,
		},
	)
	if viper.GetBool(configs.Dev) || viper.GetString(configs.LogLevel) == "debug" {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}
	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncodeDuration = zapcore.StringDurationEncoder
	level, err := zapcore.ParseLevel(viper.GetString(configs.LogLevel))
	if err != nil {
		panic("level error")
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		writeSyncer,
		zap.NewAtomicLevelAt(level),
	)
	L = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1)).Sugar()
}
