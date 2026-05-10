package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu/models"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type fakeAssetLCU struct {
	assetPath string
	response  *lcu.AssetResponse
	err       error
}

func (f *fakeAssetLCU) GetToken(bool) (int, string, error) { return 0, "", nil }
func (f *fakeAssetLCU) Init(int, string)                   {}
func (f *fakeAssetLCU) IsProcessNotFound(error) bool       { return false }
func (f *fakeAssetLCU) GetCurrSummoner() (*lcu.SummonerInfo, error) {
	return nil, nil
}
func (f *fakeAssetLCU) GetCurrentFlow() (models.GameStatus, error) {
	return models.GameFlowNone, nil
}
func (f *fakeAssetLCU) QueryGameFlowSession() (*lcu.GameFlowSession, error) {
	return nil, nil
}
func (f *fakeAssetLCU) GetSkinsBySummonerId(int64) ([]lcu.SkinInfo, error) {
	return nil, nil
}
func (f *fakeAssetLCU) GetCurrConversationID() (string, error) { return "", nil }
func (f *fakeAssetLCU) ListConversationMsg(string) ([]lcu.ConversationMsg, error) {
	return nil, nil
}
func (f *fakeAssetLCU) GetSummonerInfoByPUUID(string) (*lcu.SummonerInfo, error) {
	return nil, nil
}
func (f *fakeAssetLCU) ListGamesByUID(string, int, int) (*lcu.GameListResp, error) {
	return nil, nil
}
func (f *fakeAssetLCU) AcceptGame() error           { return nil }
func (f *fakeAssetLCU) PickChampion(int, int) error { return nil }
func (f *fakeAssetLCU) BanChampion(int, int) error  { return nil }
func (f *fakeAssetLCU) GetFriendInfoByPUUID(string) (*lcu.FriendInfo, error) {
	return nil, nil
}
func (f *fakeAssetLCU) SendConversationMsg(string, string) error { return nil }
func (f *fakeAssetLCU) GetCustomAsset(path string) (*lcu.AssetResponse, error) {
	f.assetPath = path
	return f.response, f.err
}
func (f *fakeAssetLCU) GetRankedData() (*lcu.RankedData, error) { return nil, nil }
func (f *fakeAssetLCU) GetGameSummary(int64) (*lcu.GameSummary, error) {
	return nil, nil
}
func (f *fakeAssetLCU) GetRankedDataByPUUID(string) (*lcu.RankedData, error) {
	return nil, nil
}

func TestGetAssetsPassesThroughUpstreamStatusAndContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	syslog.L = zap.NewNop().Sugar()
	lcuSvc := &fakeAssetLCU{
		response: &lcu.AssetResponse{
			StatusCode:  http.StatusNotFound,
			ContentType: "application/json",
			Body:        []byte(`{"errorCode":"RESOURCE_NOT_FOUND"}`),
		},
	}
	engine := gin.New()
	engine.GET("/riot/*assets", GetAssets(NewShieldWithLCU(lcuSvc)))

	req := httptest.NewRequest(http.MethodGet, "/riot/ASSETS/Items/Icons2D/does-not-exist.png", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if lcuSvc.assetPath != "/ASSETS/Items/Icons2D/does-not-exist.png" {
		t.Fatalf("expected asset path to be forwarded unchanged, got %q", lcuSvc.assetPath)
	}
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d with body %s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected content-type application/json, got %q", got)
	}
	if got := w.Body.String(); got != `{"errorCode":"RESOURCE_NOT_FOUND"}` {
		t.Fatalf("unexpected response body: %q", got)
	}
}

func TestGetAssetsDefaultsContentTypeForOpaqueImageBytes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	syslog.L = zap.NewNop().Sugar()
	lcuSvc := &fakeAssetLCU{
		response: &lcu.AssetResponse{
			StatusCode: http.StatusOK,
			Body:       []byte{0x89, 0x50, 0x4e, 0x47},
		},
	}
	engine := gin.New()
	engine.GET("/riot/*assets", GetAssets(NewShieldWithLCU(lcuSvc)))

	req := httptest.NewRequest(http.MethodGet, "/riot/v1/champion-icons/266.png", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if got := w.Header().Get("Content-Type"); got != "image/png" {
		t.Fatalf("expected detected image/png content-type, got %q", got)
	}
}
