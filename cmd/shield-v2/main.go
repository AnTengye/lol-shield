package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	"github.com/AnTengye/lol-shield/internal/v2/api"
	"github.com/AnTengye/lol-shield/internal/v2/api/middleware"
	"github.com/AnTengye/lol-shield/internal/v2/app"
	vlcu "github.com/AnTengye/lol-shield/internal/v2/lcu"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var configPath = flag.String("c", "config.yaml", "config file path")

func main() {
	flag.Parse()
	configs.Init(*configPath)
	syslog.Init()

	if viper.GetBool(configs.Dev) {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store := app.NewStore()
	adapter := vlcu.NewLegacyAdapter()
	engine := app.NewEngine(store, adapter, viper.GetBool(configs.Dev))
	runningService := app.NewRunningService(2 * time.Second)
	engine.SetFlowListener(
		func(flow string) {
			if flow != "InProgress" {
				runningService.Invalidate()
			}
		},
	)
	engine.Start(rootCtx)
	defer engine.Stop()

	r := gin.New()
	r.Use(middleware.GinLogger(syslog.L), middleware.GinRecovery(syslog.L, true))
	r.Use(middleware.Cors())
	api.RegisterV2Routes(r, engine, runningService)
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	webAddr := viper.GetString(configs.WebAddr)
	srv := &http.Server{
		Addr:    webAddr,
		Handler: r,
	}

	go func() {
		syslog.L.Infof("v2 server listening at %s", webAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			syslog.L.Fatalf("listen failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()
	_ = srv.Shutdown(shutdownCtx)
}
