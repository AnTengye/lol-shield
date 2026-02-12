package api

import (
	"errors"
	"net/http"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/v2/app"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type ConfigPatchReq struct {
	AutoConfirm bool `json:"autoConfirm"`
	AutoPick    int  `json:"autoPick"`
	AutoBan     int  `json:"autoBan"`
}

func RegisterV2Routes(r *gin.Engine, engine *app.Engine, running *app.RunningService) {
	g := r.Group("/api/v2")
	g.GET("/system/state", func(c *gin.Context) {
		c.JSON(
			http.StatusOK,
			gin.H{
				"code":    0,
				"message": "OK",
				"data":    engine.Snapshot(),
			},
		)
	})

	g.GET("/config", func(c *gin.Context) {
		c.JSON(
			http.StatusOK,
			gin.H{
				"code":    0,
				"message": "OK",
				"data": gin.H{
					"autoConfirm": viper.GetBool(configs.GameAutoConfirm),
					"autoPick":    viper.GetInt(configs.GameAutoPick),
					"autoBan":     viper.GetInt(configs.GameAutoBan),
				},
			},
		)
	})

	g.GET("/user", func(c *gin.Context) {
		summoner, err := lcu.GetCurrSummoner()
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"code": "C001", "message": "lcu request failed", "field": err.Error()})
			return
		}
		ranked, err := lcu.GetRankedData()
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"code": "C001", "message": "lcu request failed", "field": err.Error()})
			return
		}
		c.JSON(
			http.StatusOK,
			gin.H{
				"code":    0,
				"message": "OK",
				"data": gin.H{
					"profileIconId": summoner.ProfileIconId,
					"gameName":      summoner.GameName,
					"tagLine":       summoner.TagLine,
					"summonerLevel": summoner.SummonerLevel,
					"uuid":          summoner.Puuid,
					"tier":          ranked.HighestRankedEntry.Tier,
					"division":      ranked.HighestRankedEntry.Division,
				},
			},
		)
	})

	g.GET("/skins", func(c *gin.Context) {
		summoner, err := lcu.GetCurrSummoner()
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"code": "C001", "message": "lcu request failed", "field": err.Error()})
			return
		}
		infos, err := lcu.GetSkinsBySummonerId(summoner.SummonerId)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"code": "C001", "message": "lcu request failed", "field": err.Error()})
			return
		}
		skinInfo := make(map[int64]lcu.SkinUrl, len(infos))
		for _, v := range infos {
			if len(v.Skins) == 0 {
				continue
			}
			for _, s := range v.Skins {
				lsp := s.LoadScreenPath
				if len(lsp) >= len("/lol-game-data/assets") && lsp[:len("/lol-game-data/assets")] == "/lol-game-data/assets" {
					lsp = lsp[len("/lol-game-data/assets"):]
				}
				skinInfo[int64(s.Id)] = lcu.SkinUrl{LoadScreenPath: lsp}
				for _, chroma := range s.Chromas {
					skinInfo[int64(chroma.Id)] = lcu.SkinUrl{LoadScreenPath: lsp}
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "OK", "data": skinInfo})
	})

	g.GET("/running/snapshot", func(c *gin.Context) {
		data, err := running.Snapshot()
		if err != nil {
			if errors.Is(err, app.ErrGameNotInProgress) {
				c.JSON(
					http.StatusConflict,
					gin.H{
						"code":    "B010",
						"message": "game not in progress",
						"field":   err.Error(),
					},
				)
				return
			}
			c.JSON(
				http.StatusBadGateway,
				gin.H{
					"code":    "C001",
					"message": "lcu request failed",
					"field":   err.Error(),
				},
			)
			return
		}
		c.JSON(
			http.StatusOK,
			gin.H{
				"code":    0,
				"message": "OK",
				"data":    data,
			},
		)
	})

	g.PATCH("/config", func(c *gin.Context) {
		var req ConfigPatchReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(
				http.StatusBadRequest,
				gin.H{
					"code":    "A003",
					"message": "invalid request",
					"field":   err.Error(),
				},
			)
			return
		}
		viper.Set(configs.GameAutoConfirm, req.AutoConfirm)
		viper.Set(configs.GameAutoPick, req.AutoPick)
		viper.Set(configs.GameAutoBan, req.AutoBan)
		if err := viper.WriteConfig(); err != nil {
			c.JSON(
				http.StatusInternalServerError,
				gin.H{
					"code":    "B003",
					"message": "write config failed",
					"field":   err.Error(),
				},
			)
			return
		}
		c.JSON(
			http.StatusOK,
			gin.H{
				"code":    0,
				"message": "OK",
				"data": gin.H{
					"autoConfirm": req.AutoConfirm,
					"autoPick":    req.AutoPick,
					"autoBan":     req.AutoBan,
				},
			},
		)
	})

	r.GET("/riot/*assets", func(c *gin.Context) {
		data, err := lcu.GetCustomAssets(c.Param("assets"))
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"code": "C001", "message": "lcu request failed", "field": err.Error()})
			return
		}
		c.Header("Cache-Control", "public, max-age=31536000")
		c.Data(http.StatusOK, "application/octet-stream", data)
	})
}
