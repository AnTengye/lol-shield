package client

import (
	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/client/resp"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
)

type ConfigReq struct {
	AutoConfirm bool `json:"auto_confirm"`
	AutoPick    int  `json:"auto_pick"`
	AutoBan     int  `json:"auto_ban"`
}

func GetConfig(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		settings := viper.AllSettings()
		resp.WriteRespData(ctx, settings)
	}
}

func GetVersion(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp.WriteRespData(
			ctx, gin.H{
				"version": configs.Version,
			},
		)
	}
}

func UpdateConfig(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
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
		//p.notice(ctx, voReq)
	}
}

func GetLcu(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		port, token, err := lcu.GetLolClientApiInfo()
		if err != nil {
			resp.WriteErrRes(ctx, resp.LcuConnectErr.WithField(err.Error()))
			return
		}
		resp.WriteRespData(
			ctx, gin.H{
				"port":  port,
				"token": token,
			},
		)
	}
}

func GetAssets(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		data, err := lcu.GetCustomAssets(ctx.Param("assets"))
		if err != nil {
			resp.WriteErrRes(ctx, resp.LcuConnectErr.WithField(err.Error()))
			return
		}
		ctx.Data(http.StatusOK, "image/png", data)
	}
}

func GetUser(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		summoner := p.currSummoner
		data, err := lcu.GetRankedData()
		if err != nil {
			resp.WriteErrRes(ctx, resp.LcuConnectErr.WithField(err.Error()))
			return
		}
		resp.WriteRespData(ctx, resp.User{
			AccountId:     summoner.AccountId,
			GameName:      summoner.GameName,
			ProfileIconId: summoner.ProfileIconId,
			Level:         summoner.SummonerLevel,
			TagLine:       summoner.TagLine,
			Tier:          data.QueueMap.RANKEDSOLO5X5.Tier,
			Division:      data.QueueMap.RANKEDSOLO5X5.Division,
			IsProvisional: data.QueueMap.RANKEDSOLO5X5.IsProvisional,
		})
	}
}
