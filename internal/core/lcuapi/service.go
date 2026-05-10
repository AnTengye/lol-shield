package lcuapi

import (
	"errors"

	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu/models"
)

type Service interface {
	GetToken(fromFile bool) (port int, token string, err error)
	Init(port int, token string)
	IsProcessNotFound(err error) bool

	GetCurrSummoner() (*lcu.SummonerInfo, error)
	GetCurrentFlow() (models.GameStatus, error)
	QueryGameFlowSession() (*lcu.GameFlowSession, error)
	GetSkinsBySummonerId(summonerID int64) ([]lcu.SkinInfo, error)

	GetCurrConversationID() (string, error)
	ListConversationMsg(conversationID string) ([]lcu.ConversationMsg, error)
	GetSummonerInfoByPUUID(puuid string) (*lcu.SummonerInfo, error)
	ListGamesByUID(uuid string, begin, limit int) (*lcu.GameListResp, error)

	AcceptGame() error
	PickChampion(championID, actionID int) error
	BanChampion(championID, actionID int) error
	GetFriendInfoByPUUID(puuid string) (*lcu.FriendInfo, error)
	SendConversationMsg(msg string, conversationID string) error

	GetCustomAsset(path string) (*lcu.AssetResponse, error)
	GetRankedData() (*lcu.RankedData, error)
	GetGameSummary(gameID int64) (*lcu.GameSummary, error)
	GetRankedDataByPUUID(puuid string) (*lcu.RankedData, error)
}

type defaultService struct{}

func New() Service {
	return &defaultService{}
}

func (s *defaultService) GetToken(fromFile bool) (int, string, error) {
	return lcu.GetLcuToken(fromFile)
}

func (s *defaultService) Init(port int, token string) {
	lcu.InitCli(port, token)
}

func (s *defaultService) IsProcessNotFound(err error) bool {
	return errors.Is(err, lcu.ErrLolProcessNotFound)
}

func (s *defaultService) GetCurrSummoner() (*lcu.SummonerInfo, error) {
	return lcu.GetCurrSummoner()
}

func (s *defaultService) GetCurrentFlow() (models.GameStatus, error) {
	return lcu.GetCurrentFlow()
}

func (s *defaultService) QueryGameFlowSession() (*lcu.GameFlowSession, error) {
	return lcu.QueryGameFlowSession()
}

func (s *defaultService) GetSkinsBySummonerId(summonerID int64) ([]lcu.SkinInfo, error) {
	return lcu.GetSkinsBySummonerId(summonerID)
}

func (s *defaultService) GetCurrConversationID() (string, error) {
	return lcu.GetCurrConversationID()
}

func (s *defaultService) ListConversationMsg(conversationID string) ([]lcu.ConversationMsg, error) {
	return lcu.ListConversationMsg(conversationID)
}

func (s *defaultService) GetSummonerInfoByPUUID(puuid string) (*lcu.SummonerInfo, error) {
	return lcu.GetSummonerInfoByPUUID(puuid)
}

func (s *defaultService) ListGamesByUID(uuid string, begin, limit int) (*lcu.GameListResp, error) {
	return lcu.ListGamesByUID(uuid, begin, limit)
}

func (s *defaultService) AcceptGame() error {
	return lcu.AcceptGame()
}

func (s *defaultService) PickChampion(championID, actionID int) error {
	return lcu.PickChampion(championID, actionID)
}

func (s *defaultService) BanChampion(championID, actionID int) error {
	return lcu.BanChampion(championID, actionID)
}

func (s *defaultService) GetFriendInfoByPUUID(puuid string) (*lcu.FriendInfo, error) {
	return lcu.GetFriendInfoByPUUID(puuid)
}

func (s *defaultService) SendConversationMsg(msg string, conversationID string) error {
	return lcu.SendConversationMsg(msg, conversationID)
}

func (s *defaultService) GetCustomAsset(path string) (*lcu.AssetResponse, error) {
	return lcu.GetCustomAsset(path)
}

func (s *defaultService) GetRankedData() (*lcu.RankedData, error) {
	return lcu.GetRankedData()
}

func (s *defaultService) GetGameSummary(gameID int64) (*lcu.GameSummary, error) {
	return lcu.GetGameSummary(gameID)
}

func (s *defaultService) GetRankedDataByPUUID(puuid string) (*lcu.RankedData, error) {
	return lcu.GetRankedDataByPUUID(puuid)
}
