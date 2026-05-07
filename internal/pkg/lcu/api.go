package lcu

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AnTengye/lol-shield/internal/pkg/cache"

	"github.com/AnTengye/lol-shield/internal/pkg/lcu/models"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

var (
	queryGameSummaryLimiter = rate.NewLimiter(rate.Every(time.Second/50), 50)
)

// GetCurrSummoner 获取当前召唤师
func GetCurrSummoner() (*SummonerInfo, error) {
	bts, err := cli.httpGet("/lol-summoner/v1/current-summoner")
	if err != nil {
		return nil, err
	}
	data := &SummonerInfo{}
	err = json.Unmarshal(bts, data)
	if err != nil {
		syslog.L.Info("获取当前召唤师失败", zap.Error(err))
		return nil, err
	}
	if data.SummonerId == 0 {
		return nil, errors.New("获取当前召唤师失败")
	}
	return data, nil
}

// 获取会话组消息记录
func ListConversationMsg(conversationID string) ([]ConversationMsg, error) {
	bts, err := cli.httpGet(fmt.Sprintf("/lol-chat/v1/conversations/%s/messages", conversationID))
	if err != nil {
		return nil, err
	}
	list := make([]ConversationMsg, 0, 10)
	err = json.Unmarshal(bts, &list)
	if err != nil {
		syslog.L.Error("获取会话组消息记录失败", zap.Error(err))
		return nil, err
	}
	return list, nil
}

// 获取当前对局聊天组
func GetCurrConversationID() (string, error) {
	bts, err := cli.httpGet("/lol-chat/v1/conversations")
	if err != nil {
		return "", err
	}
	list := make([]Conversation, 0, 1)
	err = json.Unmarshal(bts, &list)
	if err != nil {
		syslog.L.Info("获取当前对局聊天组失败", zap.Error(err))
		return "", err
	}
	for _, conversation := range list {
		if conversation.Type == models.ChatTypeSelect {
			return conversation.Id, nil
		}
	}
	return "", errors.New("当前不在英雄选择阶段")
}

// 发送消息到聊天组
func SendConversationMsg(msg string, conversationID string) error {
	data := struct {
		Body string `json:"body"`
		Type string `json:"type"`
	}{
		Body: msg,
		Type: "chat",
	}
	_, err := cli.httpPost(fmt.Sprintf("/lol-chat/v1/conversations/%s/messages", conversationID), data)
	return err
}

// 申请加好友
func ApplyFriend(summonerID int64) error {
	data := struct {
		ID string `json:"id"`
	}{
		ID: strconv.FormatInt(summonerID, 10),
	}
	_, err := cli.httpPost("/lol-chat/v1/friend-requests", data)
	return err
}

// 取消加好友
func CancelApplyFriend(summonerID int64) error {
	_, err := cli.httpDel(fmt.Sprintf("/lol-chat/v1/friend-requests/%d", summonerID))
	return err
}

// 查询用户信息
func ListSummoner(summonerIDList []int64) ([]Summoner, error) {
	idStrList := make([]string, 0, len(summonerIDList))
	for _, id := range summonerIDList {
		idStrList = append(idStrList, strconv.FormatInt(id, 10))
	}
	bts, err := cli.httpGet(
		fmt.Sprintf(
			"/lol-summoner/v2/summoners?ids=[%s]",
			strings.Join(idStrList, ","),
		),
	)
	if len(bts) > 0 && bts[0] == '[' {
		list := make([]Summoner, 0, len(summonerIDList))
		err = json.Unmarshal(bts, &list)
		if err != nil {
			syslog.L.Info("查询用户信息失败", zap.Error(err))
			return nil, err
		}
		return list, err
	}
	data := &CommonResp{}
	err = json.Unmarshal(bts, data)
	if err != nil {
		syslog.L.Info("查询用户信息失败", zap.Error(err))
		return nil, err
	}
	return nil, errors.New(data.Message)
}

// 查询用户信息
func QuerySummoner(summonerID int64) (*Summoner, error) {
	list, err := ListSummoner([]int64{summonerID})
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, errors.New("获取召唤师信息失败 list == 0")
	}
	return &list[0], nil
}

// 接受对局
func AcceptGame() error {
	_, err := cli.httpPost("/lol-matchmaking/v1/ready-check/accept", nil)
	return err
}

// 获取选人会话
func GetChampSelectSession() (*ChampSelectSessionInfo, error) {
	bts, err := cli.httpGet("/lol-champ-select/v1/session")
	if err != nil {
		return nil, err
	}
	data := &ChampSelectSessionInfo{}
	err = json.Unmarshal(bts, data)
	if err != nil {
		syslog.L.Error("查询选人会话详情失败", zap.Error(err))
		return nil, err
	}
	if data.CommonResp.ErrorCode != "" {
		return nil, errors.New(fmt.Sprintf("查询选人会话详情失败 :%s", data.CommonResp.Message))
	}
	return data, nil
}

func ChampSelectPatchAction(
	championID, actionID int, patchType ChampSelectPatchType,
	completed bool,
) error {
	body := struct {
		Completed  bool                 `json:"completed"`
		Type       ChampSelectPatchType `json:"type"`
		ChampionID int                  `json:"championId"`
	}{
		Completed:  completed,
		Type:       patchType,
		ChampionID: championID,
	}
	bts, err := cli.httpPatch(fmt.Sprintf("/lol-champ-select/v1/session/actions/%d", actionID), body)
	if err != nil {
		return err
	}
	data := &CommonResp{}
	err = json.Unmarshal(bts, data)
	if err != nil {
		syslog.L.Errorf("ChampSelectPatchAction详情失败:%v", zap.Error(err))
		return err
	}
	if data.ErrorCode != "" {
		return errors.New(fmt.Sprintf("ChampSelectPatchAction失败 :%s", data.Message))
	}
	return nil
}

// 选择英雄
func PickChampion(championID, actionID int) error {
	return ChampSelectPatchAction(championID, actionID, ChampSelectPatchTypePick, true)
}

// ban英雄
func BanChampion(championID, actionID int) error {
	return ChampSelectPatchAction(championID, actionID, ChampSelectPatchTypeBan, true)
}

// 查询游戏会话
func QueryGameFlowSession() (*GameFlowSession, error) {
	bts, err := cli.httpGet("/lol-gameflow/v1/session")
	if err != nil {
		return nil, err
	}
	data := &GameFlowSession{}
	err = json.Unmarshal(bts, data)
	if err != nil {
		syslog.L.Info("查询游戏会话失败", zap.Error(err))
		return nil, err
	}
	if data.CommonResp.ErrorCode != "" {
		return nil, errors.New(fmt.Sprintf("查询游戏会话失败 :%s", data.CommonResp.Message))
	}
	return data, nil
}

// 获取自定义资源
func GetCustomAssets(path string) ([]byte, error) {
	// check cli
	if cli == nil {
		return nil, errors.New("lcu client not init")
	}
	assetsData, err := cli.httpGet("/lol-game-data/assets" + path)
	return assetsData, err
}

// 获取排位数据
func GetRankedData() (*RankedData, error) {
	bts, err := cli.httpGet("/lol-ranked/v1/current-ranked-stats")
	if err != nil {
		return nil, err
	}
	data := &RankedData{}
	err = json.Unmarshal(bts, data)
	if err != nil {
		syslog.L.Info("获取排位数据", zap.Error(err))
		return nil, err
	}
	return data, nil
}

// 获取比赛记录
func ListGamesByUID(uuid string, begin, limit int) (*GameListResp, error) {
	// 因为api是全包，所以要-1
	end := begin + limit - 1
	url := fmt.Sprintf(
		"/lol-match-history/v1/products/lol/%s/matches?begIndex=%d&endIndex=%d",
		uuid, begin, end,
	)
	syslog.L.Infow(
		"LCU战绩请求",
		"puuid", uuid,
		"begIndex", begin,
		"endIndex", end,
		"url", url,
	)
	bts, err := cli.httpGet(
		url,
	)
	if err != nil {
		syslog.L.Errorw(
			"LCU战绩请求失败",
			"puuid", uuid,
			"begIndex", begin,
			"endIndex", end,
			"error", err,
		)
		return nil, err
	}
	syslog.L.Infow(
		"LCU战绩原始返回",
		"puuid", uuid,
		"begIndex", begin,
		"endIndex", end,
		"responseBytes", len(bts),
		"response", truncateLogString(string(bts), 12000),
	)
	data := &GameListResp{}
	err = json.Unmarshal(bts, data)
	if err != nil {
		syslog.L.Info("获取比赛记录", zap.Error(err), zap.String("response", truncateLogString(string(bts), 12000)))
		return nil, err
	}
	return data, nil
}

func truncateLogString(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "...<truncated>"
}

// 查询对局详情
func GetGameSummary(gameID int64) (*GameSummary, error) {
	_ = queryGameSummaryLimiter.Wait(context.Background())
	bts, err := cli.httpGet(fmt.Sprintf("/lol-match-history/v1/games/%d", gameID))
	if err != nil {
		return nil, err
	}
	data := &GameSummary{}
	err = json.Unmarshal(bts, data)
	if err != nil {
		syslog.L.Info("查询对局详情失败", zap.Error(err))
		return nil, err
	}
	if data.CommonResp.ErrorCode != "" {
		return nil, errors.New(fmt.Sprintf("查询对局详情失败 :%s ,gameID: %d", data.CommonResp.Message, gameID))
	}
	return data, nil
}

// 查询排位数据-指定puuid
func GetRankedDataByPUUID(puuid string) (*RankedData, error) {
	var (
		bts []byte
		err error
	)
	get, _ := cache.Cache.Get(fmt.Sprintf("GetRankedDataByPUUID:%s", puuid))
	if get != nil {
		bts = get
	} else {
		bts, err = cli.httpGet(fmt.Sprintf("/lol-ranked/v1/ranked-stats/%s", puuid))
		if err != nil {
			return nil, err
		}
		_ = cache.Cache.Set(fmt.Sprintf("GetRankedDataByPUUID:%s", puuid), bts)
	}
	data := &RankedData{}
	err = json.Unmarshal(bts, data)
	if err != nil {
		syslog.L.Info("获取排位数据", zap.Error(err))
		return nil, err
	}
	return data, nil
}

// 查询好友信息
func GetFriendInfoByPUUID(puuid string) (*FriendInfo, error) {
	bts, err := cli.httpGet(
		fmt.Sprintf(
			"/lol-hovercard/v1/friend-info/%s",
			puuid,
		),
	)
	if err != nil {
		return nil, err
	}
	data := &FriendInfo{}
	err = json.Unmarshal(bts, data)
	if err != nil {
		syslog.L.Info("获取好友信息", zap.Error(err))
		return nil, err
	}
	return data, nil
}

// 查询皮肤
func GetSkinsMiniBySummonerId(summonerId int64) ([]byte, error) {
	bts, err := cli.httpGet(
		fmt.Sprintf(
			"/lol-champions/v1/inventories/%d/skins-minimal",
			summonerId,
		),
	)
	if err != nil {
		return nil, err
	}
	return bts, nil
}

func GetSkinsBySummonerId(summonerId int64) ([]SkinInfo, error) {
	bts, err := cli.httpGet(
		fmt.Sprintf(
			"/lol-champions/v1/inventories/%d/champions",
			summonerId,
		),
	)
	if err != nil {
		return nil, err
	}
	data := []SkinInfo{}
	err = json.Unmarshal(bts, &data)
	if err != nil {
		syslog.L.Info("获取皮肤信息", zap.Error(err))
		return nil, err
	}
	return data, nil
}

func GetCurrentFlow() (models.GameStatus, error) {
	bts, err := cli.httpGet("/lol-gameflow/v1/gameflow-phase")
	if err != nil {
		return models.GameFlowNone, err
	}
	trimmedBts := strings.Trim(string(bts), "\"")
	return models.GameStatus(trimmedBts), nil
}

// 查询召唤师信息-指定puuid
func GetSummonerInfoByPUUID(puuid string) (*SummonerInfo, error) {
	var (
		bts []byte
		err error
	)
	get, _ := cache.Cache.Get(fmt.Sprintf("GetSummonerInfoByPUUID:%s", puuid))
	if get != nil {
		bts = get
	} else {
		bts, err = cli.httpGet("/lol-summoner/v1/current-summoner")
		if err != nil {
			return nil, err
		}
		_ = cache.Cache.Set(fmt.Sprintf("GetSummonerInfoByPUUID:%s", puuid), bts)
	}
	data := &SummonerInfo{}
	err = json.Unmarshal(bts, data)
	if err != nil {
		syslog.L.Info("获取当前召唤师失败", zap.Error(err), zap.String("puuid", puuid))
		return nil, err
	}
	if data.SummonerId == 0 {
		return nil, errors.New("获取当前召唤师失败:无效编号")
	}
	return data, nil
}
