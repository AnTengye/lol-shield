package client

import (
	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/client/resp"
	"github.com/AnTengye/lol-shield/internal/client/ws"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func AddRouter(r *gin.Engine, p *Shield) {
	v1 := r.Group("v1")
	// 获取所有配置
	v1.GET("config", p.GetConfig)
	// 更新配置
	v1.POST("config", p.UpdateConfig)
	// 获取lcu认证信息
	v1.GET("lcu", p.GetLcu)
	// 获取app信息
	v1.GET("version", p.GetVersion)

	r.GET(
		"ws", func(c *gin.Context) {
			client := ws.ServerWebsocket(c.Writer, c.Request)
			if client != nil {
				p.webWs = client
			}
		},
	)
}

func (p *Shield) GetConfig(ctx *gin.Context) {
	settings := viper.AllSettings()
	resp.WriteRespData(ctx, settings)
}

func (p *Shield) GetVersion(ctx *gin.Context) {
	resp.WriteRespData(
		ctx, gin.H{
			"version": configs.Version,
		},
	)
}

type ConfigReq struct {
	AutoConfirm bool `json:"auto_confirm"`
	AutoPick    int  `json:"auto_pick"`
	AutoBan     int  `json:"auto_ban"`
}

func (p *Shield) UpdateConfig(ctx *gin.Context) {
	var voReq ConfigReq
	err := ctx.ShouldBindJSON(&voReq)
	if err != nil {
		resp.WriteErrRes(ctx, resp.InputDataFormatErr.WithField(err.Error()))
		return
	}
	viper.Set(configs.GameAutoConfirm, voReq.AutoConfirm)
	viper.Set(configs.GameAutoPick, voReq.AutoPick)
	viper.Set(configs.GameAutoBan, voReq.AutoBan)
	err = viper.WriteConfig()
	if err != nil {
		resp.WriteErrRes(ctx, resp.FileOperationError.WithField(err.Error()))
		return
	}
	p.notice(ctx, voReq)
	resp.WriteRespData(ctx, nil)
}

func (p *Shield) notice(ctx *gin.Context, msg interface{}) {
	if p.webWs != nil {
		p.webWs.Write(ws.Message{
			Op:   1,
			Data: msg,
		})
	}
}

func (p *Shield) GetLcu(ctx *gin.Context) {
	port, token, err := lcu.GetLolClientApiInfo()
	if err != nil {
		resp.WriteErrRes(ctx, resp.LcuConnectErr.WithField(err.Error()))
		return
	}
	resp.WriteRespData(
		ctx, gin.H{
			"port":   port,
			"token":  token,
			"online": p.lcuActive,
		},
	)
}
