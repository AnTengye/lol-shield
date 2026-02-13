package client

import (
	"net/http"
	"strconv"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/client/resp"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu/models"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
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
		resp.WriteRespData(
			ctx, gin.H{
				"port":  p.port,
				"token": p.token,
			},
		)
	}
}

func GetAssets(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		data, err := p.lcuService.GetCustomAssets(ctx.Param("assets"))
		if err != nil {
			resp.WriteErrRes(ctx, resp.LcuConnectErr.WithField(err.Error()))
			return
		}
		ctx.Header("Cache-Control", "public, max-age=31536000")
		ctx.Data(http.StatusOK, "image/png", data)
	}
}

func GetUser(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		summoner := p.currSummoner
		data, err := p.lcuService.GetRankedData()
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
			Uuid:          summoner.Puuid,
		})
	}
}

func ListGames(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		uid := ctx.Param("uid")
		page := ctx.DefaultQuery("page", "0")
		pageSize := ctx.DefaultQuery("pageSize", "10")
		pageNum, _ := strconv.Atoi(page)
		pageSizeNum, _ := strconv.Atoi(pageSize)
		if pageSizeNum > 20 {
			resp.WriteErrRes(ctx, resp.InputDataErr.WithField("pageSize不能大于20"))
			return
		}
		if pageNum < 0 {
			resp.WriteErrRes(ctx, resp.InputDataErr)
			return
		}
		if uid == "" {
			resp.WriteErrRes(ctx, resp.InputDataErr)
			return
		}
		data, err := p.lcuService.ListGamesByUID(uid, pageNum*pageSizeNum, pageSizeNum)
		if err != nil {
			resp.WriteErrRes(ctx, resp.LcuConnectErr.WithField(err.Error()))
			return
		}
		if len(data.Games.Games) == 0 {
			resp.WriteErrRes(ctx, resp.DataNotFound.WithField("没有足够的数据"))
			return
		}
		respData := make([]resp.GameList, 0, len(data.Games.Games))
		for _, game := range data.Games.Games {
			if len(game.Participants) == 0 {
				syslog.L.Errorf("没有参与比赛的数据", zap.Int64("gameId", game.GameId))
				//resp.WriteErrRes(ctx, resp.DataNotFound.WithField("没有足够的队列"))
				continue
			}
			participant := game.Participants[0]
			respData = append(respData, resp.GameList{
				CreateTime: game.GameCreation,
				GameId:     game.GameId,
				GameMode:   string(game.GameMode),
				GameType:   string(game.GameType),
				ChampionId: int64(participant.ChampionId),
				Win:        participant.Stats.Win,
				Assists:    participant.Stats.Assists,
				Kills:      participant.Stats.Kills,
				Deaths:     participant.Stats.Deaths,
				QueueId:    int64(game.QueueId),
			})

		}
		resp.WriteRespData(ctx, respData)
	}
}

func GetGameDetail(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		gameId := ctx.Param("gameId")
		if gameId == "" {
			resp.WriteErrRes(ctx, resp.InputDataErr)
			return
		}
		gameIdNum, _ := strconv.ParseInt(gameId, 10, 64)
		data, err := p.lcuService.GetGameSummary(gameIdNum)
		if err != nil {
			resp.WriteErrRes(ctx, resp.LcuConnectErr.WithField(err.Error()))
			return
		}
		resp.WriteRespData(ctx, data)
	}
}

func GetRankHighest(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		puuid := ctx.Param("puuid")
		if puuid == "" {
			resp.WriteErrRes(ctx, resp.InputDataErr)
			return
		}
		data, err := p.lcuService.GetRankedDataByPUUID(puuid)
		if err != nil {
			resp.WriteErrRes(ctx, resp.LcuConnectErr.WithField(err.Error()))
			return
		}
		resp.WriteRespData(ctx, data.HighestRankedEntry)
	}
}

func GetMulRankHighest(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		puuids := ctx.QueryArray("puuid")
		if len(puuids) == 0 {
			resp.WriteErrRes(ctx, resp.InputDataErr)
			return
		}
		type Temp struct {
			Puuid string      `json:"puuid"`
			Data  interface{} `json:"data"`
		}
		result := make([]Temp, len(puuids))
		for i, puuid := range puuids {
			if puuid == "" {
				resp.WriteErrRes(ctx, resp.InputDataErr)
				return
			}
			data, err := p.lcuService.GetRankedDataByPUUID(puuid)
			if err != nil {
				resp.WriteErrRes(ctx, resp.LcuConnectErr.WithField(err.Error()))
				return
			}
			result[i] = Temp{
				Puuid: puuid,
				Data:  data.HighestRankedEntry,
			}
		}
		resp.WriteRespData(ctx, result)
	}
}

func GetGameRunning(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if p.getGameState() == models.GameFlowInProgress {
			if p.CurGame == nil {
				resp.WriteErrRes(ctx, resp.DataNotFound.WithField("获取数据失败"))
				return
			}
			resp.WriteRespData(ctx, *p.CurGame)
		} else {
			resp.WriteErrRes(ctx, resp.DataNotFound.WithField("未在比赛中"))
		}
	}
}

func GetSkinInfo(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp.WriteRespData(ctx, SkinInfo)
	}
}
