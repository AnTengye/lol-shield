package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/AnTengye/lol-shield/internal/pkg/historycache"
	"github.com/gin-gonic/gin"
)

func TestOfflineArchiveAndManagementBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cache, err := historycache.Open(t.TempDir(), 2_000_000)
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()
	p := NewShieldWithLCU(&fakeAssetLCU{})
	p.ConfigureHistory(cache, "live", nil)
	if err = cache.Index("live:KR", "player", "玩家", "KR", []historycache.Summary{{GameId: 42, CreateTime: 100, Win: true}}, 0); err != nil {
		t.Fatal(err)
	}
	if err = cache.PutJSON("live:KR", "detail", "42", map[string]int{"gameId": 42}, 0); err != nil {
		t.Fatal(err)
	}
	engine := gin.New()
	AddRouter(engine, p)
	request := func(method, path, origin, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		recorder := httptest.NewRecorder()
		engine.ServeHTTP(recorder, req)
		return recorder
	}
	for _, path := range []string{"/v1/archive/players", "/v1/archive/history?scope=live:KR&puuid=player", "/v1/game/42?scope=live:KR&policy=cache-only"} {
		res := request("GET", path, "", "")
		if res.Code != 200 {
			t.Fatal(path, res.Code, res.Body.String())
		}
		var body struct {
			Code string `json:"code"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil || body.Code != "0" {
			t.Fatalf("成功业务码必须为字符串 0：%s", res.Body.String())
		}
	}
	for _, suffix := range []string{"scope=live:KR&puuid=player", "scope=live:NA&puuid=player"} {
		res := request("GET", "/v1/archive/players?"+suffix, "", "")
		var body struct {
			Data struct {
				Total int `json:"total"`
			} `json:"data"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		want := 0
		if strings.Contains(suffix, "KR") {
			want = 1
		}
		if body.Data.Total != want {
			t.Fatal(body)
		}
	}
	for _, path := range []string{"/v1/game/nope", "/v1/game/-2", "/v1/game/42?policy=bogus", "/v1/archive/history?page=-1", "/v1/archive/history?scope=live:KR&puuid=player&from=10&to=1"} {
		if res := request("GET", path, "", ""); res.Code != 400 {
			t.Fatal(path, res.Code, res.Body.String())
		}
	}
	if res := request("GET", "/v1/game/43?scope=live:KR&policy=cache-only", "", ""); res.Code != 404 {
		t.Fatal(res.Code, res.Body.String())
	}
	for _, origin := range []string{"", "https://evil.example", "http://localhost:5173.evil.example", "null"} {
		if res := request("POST", "/v1/cache/clear", origin, `{"kind":"all"}`); res.Code != http.StatusForbidden {
			t.Fatal(origin, res.Code)
		}
	}
	if !cache.Has("live:KR", "detail", "42") {
		t.Fatal("未授权清理不得改变数据")
	}
	res := request("POST", "/v1/cache/clear", "http://localhost:5173", `{"kind":"../../"}`)
	if res.Code != 400 {
		t.Fatal(res.Code)
	}
	res = request("POST", "/v1/cache/clear", "http://localhost:5173", `{"kind":"all"}`)
	if res.Code != 200 || cache.Has("live:KR", "detail", "42") {
		t.Fatal(res.Code, res.Body.String())
	}
}
func TestStatusSnapshotAndDisconnectAreRaceSafe(t *testing.T) {
	p := NewShieldWithLCU(&fakeAssetLCU{})
	var group sync.WaitGroup
	for i := 0; i < 4; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for j := 0; j < 50; j++ {
				p.disconnect()
				p.isLcuActive()
				p.statusSnapshot()
			}
		}()
	}
	group.Wait()
	res := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(res)
	GetStatus(p)(ctx)
	var body struct {
		Code string     `json:"code"`
		Data StatusInfo `json:"data"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != "0" || body.Data.Status != STWaiting || body.Data.GameStatus != GSWaiting {
		t.Fatal(body)
	}
}
