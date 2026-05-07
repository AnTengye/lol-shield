package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu/models"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type fakeHistoryLCU struct {
	begin     int
	limit     int
	gameCount int
	games     int
}

func (f *fakeHistoryLCU) GetToken(bool) (int, string, error) { return 0, "", nil }
func (f *fakeHistoryLCU) Init(int, string)                   {}
func (f *fakeHistoryLCU) IsProcessNotFound(error) bool       { return false }
func (f *fakeHistoryLCU) GetCurrSummoner() (*lcu.SummonerInfo, error) {
	return nil, nil
}
func (f *fakeHistoryLCU) GetCurrentFlow() (models.GameStatus, error) {
	return models.GameFlowNone, nil
}
func (f *fakeHistoryLCU) QueryGameFlowSession() (*lcu.GameFlowSession, error) {
	return nil, nil
}
func (f *fakeHistoryLCU) GetSkinsBySummonerId(int64) ([]lcu.SkinInfo, error) {
	return nil, nil
}
func (f *fakeHistoryLCU) GetCurrConversationID() (string, error) { return "", nil }
func (f *fakeHistoryLCU) ListConversationMsg(string) ([]lcu.ConversationMsg, error) {
	return nil, nil
}
func (f *fakeHistoryLCU) GetSummonerInfoByPUUID(string) (*lcu.SummonerInfo, error) {
	return nil, nil
}
func (f *fakeHistoryLCU) ListGamesByUID(_ string, begin, limit int) (*lcu.GameListResp, error) {
	f.begin = begin
	f.limit = limit
	gameCount := f.gameCount
	if gameCount == 0 {
		gameCount = 42
	}
	games := f.games
	if games == 0 {
		games = 1
	}
	body := `{
		"games": {
			"gameCount": ` + strconv.Itoa(gameCount) + `,
			"gameIndexBegin": 18,
			"gameIndexEnd": 26,
			"games": [`
	for i := 0; i < games; i++ {
		if i > 0 {
			body += `,`
		}
		body += `{
				"gameCreation": 1710000000000,
				"gameId": ` + strconv.Itoa(12345+i) + `,
				"gameMode": "CLASSIC",
				"gameType": "MATCHED_GAME",
				"queueId": 420,
				"participants": [{
					"championId": 99,
					"stats": {
						"win": true,
						"assists": 7,
						"kills": 8,
						"deaths": 1
					}
				}]
			}`
	}
	body += `]
		}
	}`
	data := &lcu.GameListResp{}
	if err := json.Unmarshal([]byte(body), data); err != nil {
		return nil, err
	}
	return data, nil
}
func (f *fakeHistoryLCU) AcceptGame() error           { return nil }
func (f *fakeHistoryLCU) PickChampion(int, int) error { return nil }
func (f *fakeHistoryLCU) BanChampion(int, int) error  { return nil }
func (f *fakeHistoryLCU) GetFriendInfoByPUUID(string) (*lcu.FriendInfo, error) {
	return nil, nil
}
func (f *fakeHistoryLCU) SendConversationMsg(string, string) error { return nil }
func (f *fakeHistoryLCU) GetCustomAssets(string) ([]byte, error)   { return nil, nil }
func (f *fakeHistoryLCU) GetRankedData() (*lcu.RankedData, error)  { return nil, nil }
func (f *fakeHistoryLCU) GetGameSummary(int64) (*lcu.GameSummary, error) {
	return nil, nil
}
func (f *fakeHistoryLCU) GetRankedDataByPUUID(string) (*lcu.RankedData, error) {
	return nil, nil
}

func TestListGamesReturnsPaginationMetadata(t *testing.T) {
	gin.SetMode(gin.TestMode)
	syslog.L = zap.NewNop().Sugar()
	lcuSvc := &fakeHistoryLCU{}
	engine := gin.New()
	engine.GET("/history/:uid", ListGames(NewShieldWithLCU(lcuSvc)))

	req := httptest.NewRequest(http.MethodGet, "/history/test-puuid?page=2&pageSize=9", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	if lcuSvc.begin != 18 || lcuSvc.limit != 9 {
		t.Fatalf("expected begin=18 limit=9, got begin=%d limit=%d", lcuSvc.begin, lcuSvc.limit)
	}
	var got struct {
		Code string `json:"code"`
		Data struct {
			List     []map[string]interface{} `json:"list"`
			Page     int                      `json:"page"`
			PageSize int                      `json:"pageSize"`
			Total    int                      `json:"total"`
			HasNext  bool                     `json:"hasNext"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Code != "0" {
		t.Fatalf("expected success code, got %q", got.Code)
	}
	if got.Data.Page != 2 || got.Data.PageSize != 9 || got.Data.Total != 42 {
		t.Fatalf("unexpected pagination data: %+v", got.Data)
	}
	if len(got.Data.List) != 1 {
		t.Fatalf("expected one game in list, got %d", len(got.Data.List))
	}
}

func TestListGamesAllowsNextPageWhenCurrentPageIsFull(t *testing.T) {
	gin.SetMode(gin.TestMode)
	syslog.L = zap.NewNop().Sugar()
	lcuSvc := &fakeHistoryLCU{
		gameCount: 9,
		games:     9,
	}
	engine := gin.New()
	engine.GET("/history/:uid", ListGames(NewShieldWithLCU(lcuSvc)))

	req := httptest.NewRequest(http.MethodGet, "/history/test-puuid?page=0&pageSize=9", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var got struct {
		Data struct {
			HasNext bool `json:"hasNext"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !got.Data.HasNext {
		t.Fatalf("expected hasNext=true when LCU returned a full page")
	}
}
