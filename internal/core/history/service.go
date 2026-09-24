package history

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AnTengye/lol-shield/internal/core/lcuapi"
	"github.com/AnTengye/lol-shield/internal/pkg/historycache"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"golang.org/x/sync/singleflight"
)

var ErrInput = errors.New("无效查询参数")
var ErrOffline = errors.New("客户端未连接，请查看已保存的离线记录")
var ErrScope = errors.New("对局大区与请求不一致，无法显示或缓存")

type Meta struct {
	Source        string    `json:"source"`
	Scope         string    `json:"scope"`
	Cached        bool      `json:"cached"`
	Stale         bool      `json:"stale"`
	FetchedAt     time.Time `json:"fetchedAt"`
	RefreshFailed bool      `json:"refreshFailed"`
	CacheError    string    `json:"cacheError,omitempty"`
}
type Result struct {
	Data json.RawMessage
	Meta Meta
}
type Service struct {
	LCU      lcuapi.Service
	Cache    *historycache.Store
	Source   string
	Online   func() bool
	requests singleflight.Group
}

func New(lcu lcuapi.Service, cache *historycache.Store, source string, online func() bool) *Service {
	return &Service{LCU: lcu, Cache: cache, Source: source, Online: online}
}
func (s *Service) scope(platform string) string {
	if platform == "" {
		return ""
	}
	return s.Source + ":" + platform
}
func (s *Service) validScope(scope string) bool {
	if scope == "" {
		return true
	}
	platform, ok := strings.CutPrefix(scope, s.Source+":")
	return ok && platform != "" && !strings.Contains(platform, ":")
}
func (s *Service) generation() uint64 {
	if s.Cache == nil {
		return 0
	}
	return s.Cache.Generation()
}
func (s *Service) query(ctx context.Context, scope, kind, id, policy string, ttl time.Duration, generation uint64, load func() (any, string, bool, error)) (Result, error) {
	if len(scope) > 200 || id == "" || len(id) > 200 {
		return Result{}, ErrInput
	}
	if !s.validScope(scope) {
		return Result{}, ErrScope
	}
	if policy != "" && policy != "cache-first" && policy != "cache-only" && policy != "refresh" {
		return Result{}, fmt.Errorf("%w：查询策略", ErrInput)
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	key := fmt.Sprintf("%d:%s:%s:%s:%s", generation, scope, kind, id, policy)
	channel := s.requests.DoChan(key, func() (any, error) {
		var cached historycache.Record
		cacheErr := historycache.ErrMiss
		if s.Cache != nil && scope != "" {
			cached, cacheErr = s.Cache.Read(scope, kind, id)
		}
		stale := ttl > 0 && time.Since(cached.FetchedAt) > ttl
		if cacheErr == nil && (policy == "cache-only" || (policy != "refresh" && !stale)) {
			return Result{cached.Body, Meta{Source: "disk-cache", Scope: scope, Cached: true, Stale: stale, FetchedAt: cached.FetchedAt}}, nil
		}
		if policy == "cache-only" {
			return Result{}, cacheErr
		}
		var data any
		var actualScope string
		var complete bool
		var err error
		if s.Online != nil && !s.Online() {
			err = ErrOffline
		} else {
			data, actualScope, complete, err = load()
		}
		if err == nil && scope != "" && actualScope != "" && actualScope != scope {
			err = ErrScope
		}
		if errors.Is(err, ErrScope) {
			return Result{}, err
		}
		if err != nil {
			if cacheErr == nil {
				return Result{cached.Body, Meta{Source: "disk-cache", Scope: scope, Cached: true, Stale: true, FetchedAt: cached.FetchedAt, RefreshFailed: true}}, nil
			}
			return Result{}, err
		}
		body, err := json.Marshal(data)
		if err != nil {
			return Result{}, err
		}
		result := Result{body, Meta{Source: "lcu", Scope: actualScope, FetchedAt: time.Now()}}
		if complete && actualScope != "" && s.Cache != nil {
			err = s.Cache.Put(actualScope, kind, id, body, "application/json", generation)
			result.Meta.Cached = err == nil
			if err != nil {
				result.Meta.CacheError = err.Error()
			}
		}
		return result, nil
	})
	select {
	case <-ctx.Done():
		return Result{}, ctx.Err()
	case value := <-channel:
		if value.Err != nil {
			return Result{}, value.Err
		}
		return value.Val.(Result), nil
	}
}
func (s *Service) Detail(ctx context.Context, gameId int64, scope, policy string) (Result, error) {
	if gameId <= 0 {
		return Result{}, fmt.Errorf("%w：对局 ID", ErrInput)
	}
	return s.query(ctx, scope, "detail", strconv.FormatInt(gameId, 10), policy, 0, s.generation(), func() (any, string, bool, error) {
		detail, err := s.LCU.GetGameSummary(gameId)
		if err != nil {
			return nil, "", false, err
		}
		if detail == nil || detail.GameId != gameId || detail.CommonResp.ErrorCode != "" {
			return nil, "", false, errors.New("客户端未返回有效详情")
		}
		complete := detail.GameDuration > 0 && len(detail.Participants) > 0 && len(detail.Teams) > 0 && len(detail.ParticipantIdentities) == len(detail.Participants)
		for _, team := range detail.Teams {
			if team.Win != "Win" && team.Win != "Fail" {
				complete = false
			}
		}
		return detail, s.scope(detail.PlatformId), complete, nil
	})
}
func (s *Service) List(ctx context.Context, puuid, scope, policy string, page, size int) (Result, error) {
	if puuid == "" || page < 0 || page > 100000 || size < 1 || size > 20 {
		return Result{}, fmt.Errorf("%w：战绩分页", ErrInput)
	}
	generation := s.generation()
	var name, tag string
	result, err := s.query(ctx, scope, "page", historycache.PageID(puuid, page, size), policy, time.Minute, generation, func() (any, string, bool, error) {
		raw, err := s.LCU.ListGamesByUID(puuid, page*size, size)
		if err != nil {
			return nil, "", false, err
		}
		if raw == nil || raw.ErrorCode != "" {
			return nil, "", false, errors.New("客户端未返回战绩")
		}
		data := historycache.Page{List: []historycache.Summary{}, Page: page, PageSize: size, Total: raw.Games.GameCount}
		begin := page*size - raw.Games.GameIndexBegin
		if begin < 0 {
			begin = 0
		}
		end := begin + size
		if end > len(raw.Games.Games) {
			end = len(raw.Games.Games)
		}
		actualScope := scope
		for i := begin; i < end; i++ {
			game := raw.Games.Games[i]
			if len(game.Participants) == 0 {
				continue
			}
			gameScope := s.scope(game.PlatformId)
			if gameScope != "" {
				if actualScope != "" && actualScope != gameScope {
					return nil, "", false, ErrScope
				}
				actualScope = gameScope
			}
			participant := game.Participants[0]
			for _, identity := range game.ParticipantIdentities {
				if identity.Player.Puuid == puuid {
					name = identity.Player.GameName
					tag = identity.Player.TagLine
					for _, p := range game.Participants {
						if p.ParticipantId == identity.ParticipantId {
							participant = p
							break
						}
					}
					break
				}
			}
			stats := participant.Stats
			data.List = append(data.List, historycache.Summary{Scope: gameScope, GameId: game.GameId, CreateTime: game.GameCreation, GameDuration: game.GameDuration, GameMode: string(game.GameMode), GameType: string(game.GameType), QueueId: int(game.QueueId), ChampionId: int(participant.ChampionId), Kills: stats.Kills, Deaths: stats.Deaths, Assists: stats.Assists, Win: stats.Win})
		}
		data.HasNext = (page+1)*size < data.Total
		complete := actualScope != ""
		for _, item := range data.List {
			if item.Scope == "" {
				complete = false
			}
		}
		return data, actualScope, complete, nil
	})
	if err != nil {
		return result, err
	}
	var data historycache.Page
	if err = json.Unmarshal(result.Data, &data); err != nil {
		return Result{}, err
	}
	if s.Cache != nil {
		if result.Meta.Cached && result.Meta.Source == "lcu" {
			if e := s.Cache.Index(result.Meta.Scope, puuid, name, tag, data.List, generation); e != nil {
				result.Meta.Cached = false
				result.Meta.CacheError = e.Error()
			}
		}
		s.Cache.MarkAvailability(&data)
	}
	result.Data, err = json.Marshal(data)
	return result, err
}
func (s *Service) Player(ctx context.Context, puuid, scope, policy string) (Result, error) {
	generation := s.generation()
	result, err := s.query(ctx, scope, "profile", puuid, policy, 10*time.Minute, generation, func() (any, string, bool, error) {
		player, err := s.LCU.GetSummonerInfoByPUUID(puuid)
		if err != nil {
			return nil, "", false, err
		}
		if player == nil || player.Puuid != puuid {
			return nil, "", false, errors.New("召唤师资料不可用")
		}
		return player, scope, true, nil
	})
	if err == nil && s.Cache != nil && scope != "" {
		var profile lcu.SummonerInfo
		if json.Unmarshal(result.Data, &profile) == nil {
			_ = s.Cache.Index(scope, puuid, profile.GameName, profile.TagLine, nil, generation)
		}
	}
	return result, err
}
func (s *Service) Rank(ctx context.Context, puuid, scope, policy string) (Result, error) {
	return s.query(ctx, scope, "rank", puuid, policy, 10*time.Minute, s.generation(), func() (any, string, bool, error) {
		data, err := s.LCU.GetRankedDataByPUUID(puuid)
		if err != nil {
			return nil, "", false, err
		}
		if data == nil {
			return nil, "", false, errors.New("段位信息不可用")
		}
		return data.HighestRankedEntry, scope, true, nil
	})
}
func (s *Service) Asset(path string) (*lcu.AssetResponse, error) {
	generation := s.generation()
	value, err, _ := s.requests.Do(fmt.Sprintf("asset:%d:%s", generation, path), func() (any, error) { return s.asset(path, generation) })
	if err != nil {
		return nil, err
	}
	response := *value.(*lcu.AssetResponse)
	return &response, nil
}
func (s *Service) asset(path string, generation uint64) (*lcu.AssetResponse, error) {
	cacheable := historycache.SafeAssetPath(path)
	scope := s.Source + ":resources-unversioned"
	var cached historycache.Record
	var cacheErr error = historycache.ErrMiss
	if cacheable && s.Cache != nil {
		cached, cacheErr = s.Cache.Read(scope, "asset", path)
		if cacheErr == nil && time.Since(cached.FetchedAt) < 24*time.Hour {
			return &lcu.AssetResponse{StatusCode: 200, ContentType: cached.ContentType, Body: cached.Body}, nil
		}
	}
	var data *lcu.AssetResponse
	var err error
	if s.Online != nil && !s.Online() {
		err = ErrOffline
	} else {
		data, err = s.LCU.GetCustomAsset(path)
	}
	if err != nil || data == nil || data.StatusCode != 200 {
		if cacheErr == nil {
			return &lcu.AssetResponse{StatusCode: 200, ContentType: cached.ContentType, Body: cached.Body}, nil
		}
		if err == nil && data != nil {
			return data, nil
		}
		if err == nil {
			err = errors.New("图片资源不可用")
		}
		return nil, err
	}
	if data.ContentType == "" {
		data.ContentType = lcu.DetectAssetContentType(path, data.Body)
	}
	if cacheable && s.Cache != nil && strings.HasPrefix(data.ContentType, "image/") {
		_ = s.Cache.Put(scope, "asset", path, data.Body, data.ContentType, generation)
	}
	return data, nil
}
