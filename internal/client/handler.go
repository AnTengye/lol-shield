package client

import (
	"golang.org/x/sync/errgroup"
	"strconv"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/client/resp"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu/models"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
	"strings"
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
		resp.WriteRespData(ctx, gin.H{"saved": true})
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
		assetPath := ctx.Param("assets")
		if strings.Contains(assetPath, "..") || strings.ContainsAny(assetPath, "\\\\\x00") {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		asset, err := p.historyService.Asset(assetPath)
		if err != nil {
			syslog.L.Errorw("资源代理请求失败", "path", assetPath, "error", err)
			resp.WriteErrRes(ctx, resp.LcuConnectErr.WithField(err.Error()))
			return
		}
		ctx.Header("Cache-Control", "no-store")
		if asset.ContentType == "" {
			asset.ContentType = lcu.DetectAssetContentType(assetPath, asset.Body)
		}
		if asset.ContentType != "" {
			ctx.Header("Content-Type", asset.ContentType)
		}
		ctx.Status(asset.StatusCode)
		_, _ = ctx.Writer.Write(asset.Body)
	}
}

func GetUser(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		p.stateMu.RLock()
		summoner := p.currSummoner
		p.stateMu.RUnlock()
		if summoner == nil {
			resp.WriteErrRes(ctx, resp.DataNotFound.WithField("客户端未连接"))
			return
		}
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
		pageNum, err := strconv.Atoi(page)
		if err != nil {
			resp.WriteErrRes(ctx, resp.InputDataErr.WithField("page"))
			return
		}
		pageSizeNum, err := strconv.Atoi(pageSize)
		if err != nil {
			resp.WriteErrRes(ctx, resp.InputDataErr.WithField("pageSize"))
			return
		}
		if pageSizeNum <= 0 {
			resp.WriteErrRes(ctx, resp.InputDataErr.WithField("pageSize"))
			return
		}
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
		result, err := p.historyService.List(ctx.Request.Context(), uid, ctx.Query("scope"), ctx.Query("policy"), pageNum, pageSizeNum)
		writeHistoryResult(ctx, result, err)
	}
}

func sliceRequestedGames(games []lcu.GameInfo, windowBegin, requestedBegin, pageSize int) []lcu.GameInfo {
	if len(games) == 0 || pageSize <= 0 {
		return nil
	}
	start := requestedBegin - windowBegin
	if start < 0 {
		start = 0
	}
	if start >= len(games) {
		return nil
	}
	end := start + pageSize
	if end > len(games) {
		end = len(games)
	}
	return games[start:end]
}

func GetGameDetail(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		gameId := ctx.Param("gameId")
		if gameId == "" {
			resp.WriteErrRes(ctx, resp.InputDataErr)
			return
		}
		gameIdNum, err := strconv.ParseInt(gameId, 10, 64)
		if err != nil || gameIdNum <= 0 {
			ctx.JSON(400, gin.H{"code": "INPUT_ERROR", "message": "无效对局 ID"})
			return
		}
		result, err := p.historyService.Detail(ctx.Request.Context(), gameIdNum, ctx.Query("scope"), ctx.Query("policy"))
		writeHistoryResult(ctx, result, err)
	}
}

func GetRankHighest(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		puuid := ctx.Param("puuid")
		if puuid == "" {
			resp.WriteErrRes(ctx, resp.InputDataErr)
			return
		}
		result, err := p.historyService.Rank(ctx.Request.Context(), puuid, ctx.Query("scope"), ctx.Query("policy"))
		writeHistoryResult(ctx, result, err)
	}
}

func GetMulRankHighest(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		puuids := ctx.QueryArray("puuid")
		if len(puuids) == 0 || len(puuids) > 20 {
			resp.WriteErrRes(ctx, resp.InputDataErr)
			return
		}
		type Temp struct {
			Puuid string      `json:"puuid"`
			Data  interface{} `json:"data"`
		}
		result := make([]Temp, len(puuids))
		group := new(errgroup.Group)
		group.SetLimit(4)
		scope, policy := ctx.Query("scope"), ctx.Query("policy")
		for _, puuid := range puuids {
			if puuid == "" || len(puuid) > 100 {
				resp.WriteErrRes(ctx, resp.InputDataErr)
				return
			}
		}
		for i, puuid := range puuids {
			group.Go(func() error {
				data, err := p.historyService.Rank(ctx.Request.Context(), puuid, scope, policy)
				result[i].Puuid = puuid
				if err == nil {
					result[i].Data = data.Data
				}
				return nil
			})
		}
		_ = group.Wait()
		resp.WriteRespData(ctx, result)
	}
}

func GetGameRunning(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		p.stateMu.RLock()
		game := p.CurGame
		p.stateMu.RUnlock()
		if p.getGameState() == models.GameFlowInProgress {
			if game == nil {
				resp.WriteErrRes(ctx, resp.DataNotFound.WithField("获取数据失败"))
				return
			}
			resp.WriteRespData(ctx, *game)
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

// Shutdown 请求本地服务优雅退出。
//
// sidecar 在 Windows 上会以管理员权限独立运行（见 internal/pkg/windows/admin），
// 桌面壳无法直接结束它，只能通过本端点请求其自行退出，
// 以释放可执行文件占用，供覆盖安装更新。
func Shutdown(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resp.WriteRespData(ctx, gin.H{"shutting_down": true})
		// 异步退出，避免 httpSrv.Shutdown 等待本响应时死等超时
		go p.Stop()
	}
}
