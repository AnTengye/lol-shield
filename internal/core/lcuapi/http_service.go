package lcuapi

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu/models"
)

type httpService struct {
	baseURL string
	client  *http.Client
}

func NewHTTPService(baseURL string) Service {
	return &httpService{
		baseURL: strings.TrimRight(baseURL, "/"),
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (s *httpService) GetToken(bool) (int, string, error) { return 0, "", nil }
func (s *httpService) Init(int, string)                   {}
func (s *httpService) IsProcessNotFound(error) bool       { return false }

func (s *httpService) GetCurrSummoner() (*lcu.SummonerInfo, error) {
	data := &lcu.SummonerInfo{}
	if err := s.getJSON("/lol-summoner/v1/current-summoner", data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *httpService) GetCurrentFlow() (models.GameStatus, error) {
	body, err := s.get("/lol-gameflow/v1/gameflow-phase")
	if err != nil {
		return models.GameFlowNone, err
	}
	return models.GameStatus(strings.Trim(string(body), "\"")), nil
}

func (s *httpService) QueryGameFlowSession() (*lcu.GameFlowSession, error) {
	data := &lcu.GameFlowSession{}
	if err := s.getJSON("/lol-gameflow/v1/session", data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *httpService) GetSkinsBySummonerId(summonerID int64) ([]lcu.SkinInfo, error) {
	var data []lcu.SkinInfo
	if err := s.getJSON(fmt.Sprintf("/lol-champions/v1/inventories/%d/champions", summonerID), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *httpService) GetCurrConversationID() (string, error) {
	var list []lcu.Conversation
	if err := s.getJSON("/lol-chat/v1/conversations", &list); err != nil {
		return "", err
	}
	for _, conversation := range list {
		if conversation.Type == models.ChatTypeSelect {
			return conversation.Id, nil
		}
	}
	return "", errors.New("current champion-select conversation not found")
}

func (s *httpService) ListConversationMsg(conversationID string) ([]lcu.ConversationMsg, error) {
	var list []lcu.ConversationMsg
	if err := s.getJSON(fmt.Sprintf("/lol-chat/v1/conversations/%s/messages", conversationID), &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (s *httpService) GetSummonerInfoByPUUID(puuid string) (*lcu.SummonerInfo, error) {
	data := &lcu.SummonerInfo{}
	if err := s.getJSON(fmt.Sprintf("/lol-summoner/v2/summoners/puuid/%s", puuid), data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *httpService) ListGamesByUID(uuid string, begin, limit int) (*lcu.GameListResp, error) {
	end := begin + limit - 1
	data := &lcu.GameListResp{}
	if err := s.getJSON(
		fmt.Sprintf("/lol-match-history/v1/products/lol/%s/matches?begIndex=%d&endIndex=%d", uuid, begin, end),
		data,
	); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *httpService) AcceptGame() error           { return nil }
func (s *httpService) PickChampion(int, int) error { return nil }
func (s *httpService) BanChampion(int, int) error  { return nil }
func (s *httpService) SendConversationMsg(string, string) error {
	return nil
}

func (s *httpService) GetFriendInfoByPUUID(puuid string) (*lcu.FriendInfo, error) {
	data := &lcu.FriendInfo{}
	if err := s.getJSON(fmt.Sprintf("/lol-hovercard/v1/friend-info/%s", puuid), data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *httpService) GetCustomAssets(path string) ([]byte, error) {
	return s.get("/lol-game-data/assets" + path)
}

func (s *httpService) GetRankedData() (*lcu.RankedData, error) {
	data := &lcu.RankedData{}
	if err := s.getJSON("/lol-ranked/v1/current-ranked-stats", data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *httpService) GetGameSummary(gameID int64) (*lcu.GameSummary, error) {
	data := &lcu.GameSummary{}
	if err := s.getJSON(fmt.Sprintf("/lol-match-history/v1/games/%d", gameID), data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *httpService) GetRankedDataByPUUID(puuid string) (*lcu.RankedData, error) {
	data := &lcu.RankedData{}
	if err := s.getJSON(fmt.Sprintf("/lol-ranked/v1/ranked-stats/%s", puuid), data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *httpService) getJSON(requestPath string, out any) error {
	body, err := s.get(requestPath)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}

func (s *httpService) get(requestPath string) ([]byte, error) {
	resp, err := s.client.Get(s.baseURL + requestPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("GET %s failed: status %d", requestPath, resp.StatusCode)
	}
	return body, nil
}
