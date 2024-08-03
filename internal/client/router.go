package client

import (
	"github.com/AnTengye/lol-shield/internal/client/ws"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	"github.com/gin-gonic/gin"
)

func AddRouter(r *gin.Engine, p *Shield) {
	riotReq := r.Group("riot")
	riotReq.GET("*assets", GetAssets(p))
	v1 := r.Group("v1")
	// 获取所有配置
	v1.GET("config", GetConfig(p))
	// 更新配置
	v1.POST("config", UpdateConfig(p))
	// 获取lcu认证信息
	v1.GET("lcu", GetLcu(p))
	// 获取app信息
	v1.GET("version", GetVersion(p))
	// 获取用户信息
	v1.GET("user", GetUser(p))

	r.GET(
		"ws", func(c *gin.Context) {
			client := ws.ServerWebsocket(c.Writer, c.Request)
			if client != nil {
				p.webWs = client
				//go client.read() // 不需要读取web消息
				syslog.L.Infof("%s-ui已连接", client.GetUid())
				p.Notice()
			}
		},
	)
}
