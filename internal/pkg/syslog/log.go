package syslog

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var L *zap.SugaredLogger

func Init() error {
	level, err := zapcore.ParseLevel(viper.GetString(configs.LogLevel))
	if err != nil {
		return fmt.Errorf("日志级别无效: %w", err)
	}
	dir := viper.GetString(configs.LogFilepath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}
	filename := filepath.Join(dir, time.Now().Format("20060102")+".log")
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("打开日志失败: %w", err)
	}
	_ = file.Close()
	lumberJackLogger := &lumberjack.Logger{
		// 日志文件以日期命名
		Filename:   filename,
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
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		syncer,
		zap.NewAtomicLevelAt(level),
	)
	options := []zap.Option{}
	if viper.GetBool(configs.Dev) {
		options = append(options, zap.AddCaller(), zap.AddCallerSkip(1))
	}
	L = zap.New(core, options...).With(zap.Int("pid", os.Getpid())).Sugar()
	return nil
}
