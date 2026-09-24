package client

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/AnTengye/lol-shield/internal/client/middleware"
	"github.com/AnTengye/lol-shield/internal/client/resp"
	"github.com/AnTengye/lol-shield/internal/core/history"
	"github.com/AnTengye/lol-shield/internal/pkg/historycache"
	"github.com/gin-gonic/gin"
)

func writeHistoryResult(ctx *gin.Context, result history.Result, err error) {
	if err != nil {
		code, status := "LCU_UNAVAILABLE", http.StatusBadGateway
		switch {
		case errors.Is(err, historycache.ErrMiss):
			code, status = "CACHE_MISS", 404
		case errors.Is(err, historycache.ErrCorrupt):
			code, status = "CACHE_CORRUPT", 409
		case errors.Is(err, history.ErrInput):
			code, status = "INPUT_ERROR", 400
		case errors.Is(err, history.ErrScope):
			code, status = "SCOPE_MISMATCH", 409
		case errors.Is(err, history.ErrOffline):
			code, status = "LCU_OFFLINE", 503
		}
		ctx.JSON(status, gin.H{"code": code, "message": err.Error()})
		return
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.JSON(200, gin.H{"code": "0", "message": "OK", "data": result.Data, "meta": result.Meta})
}
func GetStatus(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) { resp.WriteRespData(ctx, p.statusSnapshot()) }
}
func GetPlayer(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := p.historyService.Player(ctx.Request.Context(), ctx.Param("puuid"), ctx.Query("scope"), ctx.Query("policy"))
		writeHistoryResult(ctx, result, err)
	}
}
func cacheAvailable(ctx *gin.Context, p *Shield) bool {
	if p.historyService.Cache == nil {
		ctx.JSON(503, gin.H{"code": "CACHE_UNAVAILABLE", "message": "本地缓存不可用：" + p.cacheError})
		return false
	}
	return true
}
func archivePage(ctx *gin.Context) (int, int, bool) {
	page, e1 := strconv.Atoi(ctx.DefaultQuery("page", "0"))
	size, e2 := strconv.Atoi(ctx.DefaultQuery("pageSize", "20"))
	if e1 != nil || e2 != nil || page < 0 || page > 100000 || size < 1 || size > 100 {
		ctx.JSON(400, gin.H{"code": "INPUT_ERROR", "message": "无效分页参数"})
		return 0, 0, false
	}
	return page, size, true
}
func ArchivePlayers(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !cacheAvailable(ctx, p) {
			return
		}
		page, size, ok := archivePage(ctx)
		if !ok {
			return
		}
		scope, puuid := ctx.Query("scope"), ctx.Query("puuid")
		if (scope == "") != (puuid == "") || len(scope) > 200 || len(puuid) > 200 {
			ctx.JSON(400, gin.H{"code": "INPUT_ERROR", "message": "无效玩家范围"})
			return
		}
		list, total, err := p.historyService.Cache.FindPlayers(ctx.Query("search"), scope, puuid, page, size)
		if err != nil {
			ctx.JSON(500, gin.H{"code": "CACHE_UNAVAILABLE", "message": err.Error()})
			return
		}
		resp.WriteRespData(ctx, gin.H{"list": list, "total": total, "page": page, "pageSize": size})
	}
}
func ArchiveHistory(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !cacheAvailable(ctx, p) {
			return
		}
		page, size, ok := archivePage(ctx)
		if !ok {
			return
		}
		filter := historycache.Filter{Scope: ctx.Query("scope"), Puuid: ctx.Query("puuid"), Page: page, PageSize: size, DetailsOnly: ctx.Query("detailsOnly") == "true"}
		if filter.Scope == "" || filter.Puuid == "" {
			ctx.JSON(400, gin.H{"code": "INPUT_ERROR", "message": "请选择大区和召唤师"})
			return
		}
		if raw := ctx.Query("queueId"); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil {
				ctx.AbortWithStatus(400)
				return
			}
			filter.Queue = &n
		}
		if raw := ctx.Query("win"); raw != "" {
			n, err := strconv.ParseBool(raw)
			if err != nil {
				ctx.AbortWithStatus(400)
				return
			}
			filter.Win = &n
		}
		for _, field := range []struct {
			name  string
			value *int64
		}{{"from", &filter.From}, {"to", &filter.To}} {
			if raw := ctx.Query(field.name); raw != "" {
				n, err := strconv.ParseInt(raw, 10, 64)
				if err != nil || n < 0 {
					ctx.AbortWithStatus(400)
					return
				}
				*field.value = n
			}
		}
		if filter.To > 0 && filter.From > filter.To {
			ctx.JSON(400, gin.H{"code": "INPUT_ERROR", "message": "开始日期不能晚于结束日期"})
			return
		}
		data, err := p.historyService.Cache.History(filter)
		if err != nil {
			ctx.JSON(500, gin.H{"code": "CACHE_UNAVAILABLE", "message": err.Error()})
			return
		}
		ctx.JSON(200, gin.H{"code": "0", "data": data, "meta": gin.H{"source": "disk-cache", "scope": filter.Scope, "cached": true}})
	}
}
func CacheStats(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if p.historyService.Cache == nil {
			resp.WriteRespData(ctx, historycache.Stats{MaxBytes: historycache.MaxBytes, Healthy: false, Message: p.cacheError})
			return
		}
		stats, err := p.historyService.Cache.Stats()
		if err != nil {
			ctx.JSON(500, gin.H{"code": "CACHE_UNAVAILABLE", "message": err.Error()})
			return
		}
		resp.WriteRespData(ctx, stats)
	}
}
func ClearCache(p *Shield) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !middleware.AllowedOrigin(ctx.GetHeader("Origin")) {
			ctx.AbortWithStatus(403)
			return
		}
		if !cacheAvailable(ctx, p) {
			return
		}
		var input struct {
			Kind string `json:"kind"`
		}
		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 1024)
		if ctx.ShouldBindJSON(&input) != nil || (input.Kind != "assets" && input.Kind != "all") {
			ctx.JSON(400, gin.H{"code": "INPUT_ERROR", "message": "无效清理范围"})
			return
		}
		before, err := p.historyService.Cache.Stats()
		if err != nil {
			ctx.AbortWithStatus(500)
			return
		}
		if err = p.historyService.Cache.Clear(input.Kind); err != nil {
			ctx.JSON(500, gin.H{"code": "CACHE_CLEAR_FAILED", "message": err.Error()})
			return
		}
		after, err := p.historyService.Cache.Stats()
		if err != nil {
			ctx.AbortWithStatus(500)
			return
		}
		resp.WriteRespData(ctx, gin.H{"stats": after, "freedBytes": before.UsedBytes - after.UsedBytes, "clearedAt": time.Now()})
	}
}
