package history

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AnTengye/lol-shield/internal/core/lcuapi"
	"github.com/AnTengye/lol-shield/internal/mocklcu"
	"github.com/AnTengye/lol-shield/internal/pkg/historycache"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
)

type fakeLCU struct {
	lcuapi.Service
	calls            atomic.Int32
	detail           *lcu.GameSummary
	list             *lcu.GameListResp
	err              error
	entered, release chan struct{}
}

func (f *fakeLCU) GetGameSummary(int64) (*lcu.GameSummary, error) {
	f.calls.Add(1)
	if f.entered != nil {
		f.entered <- struct{}{}
		<-f.release
	}
	return f.detail, f.err
}
func (f *fakeLCU) ListGamesByUID(string, int, int) (*lcu.GameListResp, error) {
	f.calls.Add(1)
	return f.list, f.err
}
func (f *fakeLCU) GetSummonerInfoByPUUID(p string) (*lcu.SummonerInfo, error) {
	f.calls.Add(1)
	return &lcu.SummonerInfo{Puuid: p, GameName: "测试玩家"}, f.err
}
func testService(t *testing.T) (*Service, *fakeLCU) {
	t.Helper()
	cache, err := historycache.Open(t.TempDir(), 2_000_000)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cache.Close() })
	var detail lcu.GameSummary
	if err = json.Unmarshal([]byte(`{"gameId":42,"platformId":"KR","gameDuration":1000,"participants":[{"participantId":1,"teamId":100,"stats":{"win":true}}],"participantIdentities":[{"participantId":1,"player":{"puuid":"p"}}],"teams":[{"teamId":100,"win":"Win"}]}`), &detail); err != nil {
		t.Fatal(err)
	}
	fake := &fakeLCU{detail: &detail}
	return New(fake, cache, "live", nil), fake
}
func TestDetailCacheOfflineRefreshAndScope(t *testing.T) {
	service, fake := testService(t)
	ctx := context.Background()
	first, err := service.Detail(ctx, 42, "live:KR", "")
	if err != nil || !first.Meta.Cached {
		t.Fatal(first, err)
	}
	second, err := service.Detail(ctx, 42, "live:KR", "")
	if err != nil || second.Meta.Source != "disk-cache" || fake.calls.Load() != 1 {
		t.Fatal(second, err, fake.calls.Load())
	}
	fake.err = errors.New("断开")
	fallback, err := service.Detail(ctx, 42, "live:KR", "refresh")
	if err != nil || !fallback.Meta.RefreshFailed || !fallback.Meta.Cached {
		t.Fatal(fallback, err)
	}
	before := fake.calls.Load()
	_, err = service.Detail(ctx, 100, "live:KR", "cache-only")
	if !errors.Is(err, historycache.ErrMiss) || fake.calls.Load() != before {
		t.Fatal(err)
	}
	fake.err = nil
	_, err = service.Detail(ctx, 42, "live:EUW", "refresh")
	if !errors.Is(err, ErrScope) {
		t.Fatal(err)
	}
	_, err = service.Detail(ctx, 42, "mock:default:KR", "")
	if !errors.Is(err, ErrScope) {
		t.Fatal(err)
	}
}
func TestScopeRejectsMockPrefixCollision(t *testing.T) {
	service, fake := testService(t)
	service.Source = "mock:foo"
	for _, scope := range []string{"mock:foobar:KR", "mock:foo:bar:KR", "mock:foo:", "live:KR"} {
		if _, err := service.Detail(context.Background(), 42, scope, ""); !errors.Is(err, ErrScope) {
			t.Fatalf("错误范围 %q: %v", scope, err)
		}
	}
	if fake.calls.Load() != 0 {
		t.Fatal("非法范围不得查询客户端")
	}
}

func TestMissingDetailStatisticsRemainNullAcrossCache(t *testing.T) {
	service, fake := testService(t)
	zero := 0
	fake.detail.Participants[0].Stats.Kills = &zero
	for _, policy := range []string{"", "cache-only"} {
		result, err := service.Detail(context.Background(), 42, "live:KR", policy)
		if err != nil {
			t.Fatal(err)
		}
		body, err := json.Marshal(result.Data)
		if err != nil {
			t.Fatal(err)
		}
		var value struct {
			Participants []struct {
				Stats map[string]any `json:"stats"`
			} `json:"participants"`
		}
		if err = json.Unmarshal(body, &value); err != nil {
			t.Fatal(err)
		}
		stats := value.Participants[0].Stats
		if stats["kills"] != float64(0) || stats["deaths"] != nil || stats["goldEarned"] != nil || stats["visionScore"] != nil {
			t.Fatalf("真实零值与缺失值必须区分: %v", stats)
		}
	}
}

func TestLocalQueryLatency(t *testing.T) {
	if os.Getenv("SHIELD_CACHE_PERF") != "1" {
		t.Skip("设置 SHIELD_CACHE_PERF=1 测量本地查询 P95")
	}
	cache, err := historycache.Open(t.TempDir(), historycache.MaxBytes)
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()
	scenario, err := mocklcu.LoadScenario("../../mocklcu/fixtures/default")
	if err != nil {
		t.Fatal(err)
	}
	var detail lcu.GameSummary
	for _, raw := range scenario.GameDetails {
		if err = json.Unmarshal(raw, &detail); err != nil {
			t.Fatal(err)
		}
		break
	}
	for i := 0; i < 200; i++ {
		detail.GameId = int64(1000 + i)
		if err = cache.PutJSON("live:KR", "detail", fmt.Sprint(detail.GameId), detail, 0); err != nil {
			t.Fatal(err)
		}
	}
	for player := 0; player < 100; player++ {
		items := make([]historycache.Summary, 200)
		for i := range items {
			items[i] = historycache.Summary{GameId: int64(1000 + i), CreateTime: int64(i + 1), QueueId: 450, Win: i%2 == 0}
		}
		if err = cache.Index("live:KR", fmt.Sprint(player), fmt.Sprintf("测试玩家%d", player), "KR", items, 0); err != nil {
			t.Fatal(err)
		}
	}
	service := New(nil, cache, "live", func() bool { return false })
	measure := func(name string, budget time.Duration, query func(int) error) {
		samples := make([]time.Duration, 200)
		for i := range samples {
			start := time.Now()
			if err := query(i); err != nil {
				t.Fatal(err)
			}
			samples[i] = time.Since(start)
		}
		sort.Slice(samples, func(i, j int) bool { return samples[i] < samples[j] })
		p95 := samples[(len(samples)*95+99)/100-1]
		t.Logf("%s P95=%s（100玩家 / 20,000摘要 / 200详情，服务层，不含HTTP和渲染）", name, p95)
		if p95 > budget {
			t.Errorf("%s 超出 %s", name, budget)
		}
	}
	measure("详情", 100*time.Millisecond, func(i int) error {
		_, err := service.Detail(context.Background(), int64(1000+i), "live:KR", "cache-only")
		return err
	})
	measure("玩家查询", 150*time.Millisecond, func(i int) error {
		_, _, err := cache.Players("测试玩家", 0, 20)
		return err
	})
	measure("战绩筛选", 150*time.Millisecond, func(i int) error {
		queue, win := 450, true
		_, err := cache.History(historycache.Filter{Scope: "live:KR", Puuid: fmt.Sprint(i % 100), PageSize: 20, Queue: &queue, Win: &win, DetailsOnly: true})
		return err
	})
}

func TestIncompleteAndUnscopedDetailsAreNotPersisted(t *testing.T) {
	service, fake := testService(t)
	fake.detail.PlatformId = ""
	result, err := service.Detail(context.Background(), 42, "", "")
	if err != nil || result.Meta.Cached {
		t.Fatal(result, err)
	}
	fake.detail.PlatformId = "KR"
	fake.detail.Teams = nil
	result, err = service.Detail(context.Background(), 42, "live:KR", "")
	if err != nil || result.Meta.Cached {
		t.Fatal(result, err)
	}
	stats, err := service.Cache.Stats()
	if err != nil || stats.Details != 0 {
		t.Fatal(stats, err)
	}
}
func TestClearWhileNetworkInFlightDoesNotResurrect(t *testing.T) {
	service, fake := testService(t)
	fake.entered = make(chan struct{}, 1)
	fake.release = make(chan struct{})
	done := make(chan Result, 1)
	go func() { result, _ := service.Detail(context.Background(), 42, "live:KR", ""); done <- result }()
	<-fake.entered
	if err := service.Cache.Clear("all"); err != nil {
		t.Fatal(err)
	}
	close(fake.release)
	result := <-done
	if result.Meta.Cached || service.Cache.Has("live:KR", "detail", "42") {
		t.Fatal(result)
	}
}
func TestConcurrentDetailQueriesAreCoalesced(t *testing.T) {
	service, fake := testService(t)
	fake.entered = make(chan struct{}, 32)
	fake.release = make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := service.Detail(context.Background(), 42, "live:KR", "")
			if err != nil {
				t.Error(err)
			}
		}()
	}
	<-fake.entered
	close(fake.release)
	wg.Wait()
	if fake.calls.Load() != 1 {
		t.Fatal(fake.calls.Load())
	}
}
func TestRepositoryFixturesArchiveAcrossRestart(t *testing.T) {
	scenario, err := mocklcu.LoadScenario("../../mocklcu/fixtures/default")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	cache, err := historycache.Open(dir, 10_000_000)
	if err != nil {
		t.Fatal(err)
	}
	fake := &fakeLCU{}
	service := New(fake, cache, "mock:default", nil)
	var puuid, scope string
	var gameID int64
	for key, fixture := range scenario.MatchHistory {
		for i, c := range key {
			if c == '|' {
				puuid = key[:i]
				break
			}
		}
		fake.list = &lcu.GameListResp{Games: fixture.Games}
		result, err := service.List(context.Background(), puuid, "", "", 0, 20)
		if err != nil || !result.Meta.Cached {
			t.Fatalf("列表 fixture 缺少平台或保存失败: %+v %v", result.Meta, err)
		}
		scope = result.Meta.Scope
		if len(fixture.Games.Games) > 0 {
			gameID = fixture.Games.Games[0].GameId
		}
		break
	}
	for id, raw := range scenario.GameDetails {
		var detail lcu.GameSummary
		if err = json.Unmarshal(raw, &detail); err != nil {
			t.Fatal(err)
		}
		if detail.PlatformId == "" {
			t.Fatal("详情 fixture 缺少平台")
		}
		gameID = id
		fake.detail = &detail
		result, err := service.Detail(context.Background(), id, service.scope(detail.PlatformId), "")
		if err != nil || !result.Meta.Cached {
			t.Fatal(result.Meta, err)
		}
		scope = result.Meta.Scope
		break
	}
	_, err = service.Player(context.Background(), puuid, scope, "")
	if err != nil {
		t.Fatal(err)
	}
	if err = cache.Close(); err != nil {
		t.Fatal(err)
	}
	cache, err = historycache.Open(dir, 10_000_000)
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()
	offline := New(nil, cache, "mock:default", func() bool { return false })
	result, err := offline.Detail(context.Background(), gameID, scope, "cache-only")
	if err != nil || result.Meta.Source != "disk-cache" {
		t.Fatal(result.Meta, err)
	}
	players, total, err := cache.Players("", 0, 20)
	if err != nil || total != 1 || players[0].Summaries == 0 {
		t.Fatal(players, total, err)
	}
	_, err = offline.Player(context.Background(), puuid, scope, "cache-only")
	if err != nil {
		t.Fatal(err)
	}
}
