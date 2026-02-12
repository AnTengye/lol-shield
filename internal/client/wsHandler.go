package client

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/pkg/cache"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu/models"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	"github.com/AnTengye/lol-shield/internal/pkg/tree"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
)

var (
	SkinInfo map[int64]lcu.SkinUrl
)

func (p *Shield) initSkin(summonerId int64) {
	infos, err := lcu.GetSkinsBySummonerId(summonerId)
	if err != nil {
		syslog.L.Error("获取皮肤信息失败", zap.Error(err))
		return
	}
	SkinInfo = make(map[int64]lcu.SkinUrl, len(infos))
	for _, v := range infos {
		if len(v.Skins) == 0 {
			continue
		}
		for _, s := range v.Skins {
			lsp := s.LoadScreenPath[len("/lol-game-data/assets"):]
			SkinInfo[int64(s.Id)] = lcu.SkinUrl{
				// /lol-game-data/assets/ASSETS/Characters/Ahri/Skins/Base/AhriLoadscreen_0.jpg
				LoadScreenPath: lsp,
			}
			if len(s.Chromas) != 0 {
				for _, chromas := range s.Chromas {
					SkinInfo[int64(chromas.Id)] = lcu.SkinUrl{
						LoadScreenPath: lsp,
					}
				}
			}
		}
	}
	p.CurInfo.SkinSync = 1
	p.Notice()
}
func (p *Shield) HandlerLOLLeagueSessionToken(c *tree.Context) error {
	/*
		{
		  "numFriends": 0,
		  "numFriendsAvailable": 0,
		  "numFriendsAway": 0,
		  "numFriendsInChampSelect": 0,
		  "numFriendsInGame": 0,
		  "numFriendsInQueue": 0,
		  "numFriendsMobile": 0,
		  "numFriendsOnline": 0
		}
	*/
	return nil
}
func (p *Shield) HandlerFlowChange(c *tree.Context) error {
	gameFlow, ok := c.Data.(string)
	if !ok {
		return errors.New("flow data error")
	}
	p.onGameFlowUpdate(models.GameStatus(gameFlow))
	return nil
}

// 处理选人界面
func (p *Shield) HandlerOnChampSelectSessionUpdate(c *tree.Context) error {
	bts, err := json.Marshal(c.Data)
	if err != nil {
		return err
	}
	sessionInfo := &lcu.ChampSelectSessionInfo{}
	err = json.Unmarshal(bts, sessionInfo)
	if err != nil {
		return err
	}
	isSelfPick := false
	isSelfBan := false
	userActionID := 0
	if len(sessionInfo.Actions) == 0 {
		return nil
	}
loop:
	for _, actions := range sessionInfo.Actions {
		for _, action := range actions {
			if action.ActorCellId == sessionInfo.LocalPlayerCellId && action.IsInProgress {
				userActionID = action.Id
				if action.Type == lcu.ChampSelectPatchTypePick {
					isSelfPick = true
					break loop
				} else if action.Type == lcu.ChampSelectPatchTypeBan {
					isSelfBan = true
					break loop
				}
			}
		}
	}
	autoPickId := viper.GetInt(configs.GameAutoPick)
	autoBanId := viper.GetInt(configs.GameAutoBan)
	if autoPickId > 0 && isSelfPick {
		_ = lcu.PickChampion(autoPickId, userActionID)
	}
	if autoBanId > 0 && isSelfBan {
		_ = lcu.BanChampion(autoBanId, userActionID)
	}
	return nil
}

func (p *Shield) HandlerLolChatFriendCounts(c *tree.Context) error {
	bts, err := json.Marshal(c.Data)
	if err != nil {
		return err
	}
	syslog.L.Debugf("friends count:%s", string(bts))
	if viper.GetBool(configs.TempButton) {
		_, ok := cache.StaticMap.Load(fmt.Sprintf("friends_count:%s", "e8959476-e72f-5b05-b59a-a05688407ffc"))
		if !ok {
			info, err := lcu.GetFriendInfoByPUUID("e8959476-e72f-5b05-b59a-a05688407ffc")
			if err != nil {
				syslog.L.Errorf("friends err:%v", err)
				return err
			}
			if info.Availability == models.FriendStatusOnline {
				cache.StaticMap.Store(fmt.Sprintf("friends_count:%s", "e8959476-e72f-5b05-b59a-a05688407ffc"), 1)
				go func() {
					time.Sleep(10 * time.Second)
					err := lcu.SendConversationMsg("检测到目标人物上线，这里是智障助手的自动发言：请看到这条消息时，微信拍拍他，本人挂机中", "e8959476-e72f-5b05-b59a-a05688407ffc@pvp.net")
					if err != nil {
						syslog.L.Errorf("friends auto send err:%v", err)
					}
				}()
			}
		}
	}
	return nil
}

func (p *Shield) ChampionSelectStart() {
	if p.CurLobby == nil {
		// 获取当前游戏进程
		session, err := lcu.QueryGameFlowSession()
		if err != nil {
			return
		}
		queueInfo := session.GameData.Queue
		p.CurLobby = &LobbyInfo{
			AllowPeople: len(queueInfo.AllowablePremadeSizes),
			GameMode:    models.GameMode(queueInfo.GameMode),
			QueueId:     models.GameQueueID(queueInfo.Id),
		}
	}
	switch p.CurLobby.GameMode {
	case models.GameModeStrawBerry:
		// 娱乐模式不处理用户信息
		return
	}
	var idList []lcu.UserId
	for i := 0; i < 3; i++ {
		// 获取队伍所有用户信息
		_, idList, _ = getTeamUsers()
		if len(idList) != p.CurLobby.AllowPeople {
			continue
		}
		time.Sleep(time.Second)
	}
	if len(idList) != p.CurLobby.AllowPeople {
		return // 未获取到队伍信息
	}
	var (
		historyMap  map[string][]lcu.GameHistory
		userNameMap map[string]lcu.UserName
		err         error
	)
	retry := 3
	for retry > 0 {
		retry--
		// 查询自己队伍的信息
		historyMap, userNameMap, err = getGameHistoryByUserList(idList)
		if err != nil {
			syslog.L.Errorf("查询用户信息失败:%v", err)
			return
		}
		if len(historyMap) == p.CurLobby.AllowPeople && len(userNameMap) == p.CurLobby.AllowPeople {
			break
		}
		time.Sleep(3 * time.Second)
	}
	p.CurGame = &GameInfo{
		AllGameHistory: historyMap,
		UserNameMap:    userNameMap,
	}
}
func (p Shield) AcceptGame() {
	_ = lcu.AcceptGame()
}
func (p *Shield) HandlerInProccessGame() {
	// 获取当前游戏进程
	session, err := lcu.QueryGameFlowSession()
	if err != nil {
		return
	}
	queueInfo := session.GameData.Queue
	p.CurLobby = &LobbyInfo{
		AllowPeople: len(queueInfo.AllowablePremadeSizes),
		GameMode:    models.GameMode(queueInfo.GameMode),
		QueueId:     models.GameQueueID(queueInfo.Id),
	}
	switch p.CurLobby.GameMode {
	case models.GameModeStrawBerry:
		// 娱乐模式不处理用户信息
		return
	}
	if session.Phase != models.GameFlowInProgress {
		return
	}
	p.CurInfo.GameStatus = GSStarted
	if p.currSummoner == nil {
		return
	}
	selfID := lcu.UserId{
		SummonerId: p.currSummoner.SummonerId,
		Puuid:      p.currSummoner.Puuid,
	}
	selfTeamUsers, enemyTeamUsers, groups, skinMap := getAllUsersFromSession(selfID, session)
	// 查询对面的信息
	historyMap, userNameMap, err := getGameHistoryByUserList(enemyTeamUsers.UserList)
	if err != nil {
		syslog.L.Errorf("查询用户信息失败:%v", err)
		return
	}
	if p.CurGame == nil || len(p.CurGame.AllGameHistory) == 0 {
		// 查询自己队伍的信息
		selfHistoryMap, selfUserNameMap, selfErr := getGameHistoryByUserList(selfTeamUsers.UserList)
		if selfErr != nil {
			syslog.L.Errorf("查询用户信息失败:%v", selfErr)
			return
		}
		p.CurGame = &GameInfo{
			AllGameHistory: selfHistoryMap,
			UserNameMap:    selfUserNameMap,
		}
	}
	p.CurGame.SelfTeamInfo = selfTeamUsers
	p.CurGame.EnemyTeamInfo = enemyTeamUsers
	p.CurGame.PreTeam = groups
	p.CurGame.SkinMap = skinMap
	p.CurGame.QueueId = p.CurLobby.QueueId
	p.CurGame.QueueName = queueInfo.Name
	maps.Copy(p.CurGame.AllGameHistory, historyMap)
	maps.Copy(p.CurGame.UserNameMap, userNameMap)
	p.Notice()
}

func (p *Shield) HandlerLobbyChange(c *tree.Context) error {
	marshal, err := json.Marshal(c.Data)
	if err != nil {
		return err
	}
	data := lcu.Lobby{}
	err = json.Unmarshal(marshal, &data)
	if err != nil {
		syslog.L.Errorf("unmarshal lobby data error:%v", err)
		return err
	}
	p.CurLobby = &LobbyInfo{
		AllowPeople: len(data.GameConfig.AllowablePremadeSizes),
		GameMode:    models.GameMode(data.GameConfig.GameMode),
		QueueId:     models.GameQueueID(data.GameConfig.QueueId),
	}
	return nil
}

func (p *Shield) HandlerGameFlowLock(c *tree.Context) error {
	marshal, err := json.Marshal(c.Data)
	if err != nil {
		return err
	}
	fmt.Println("GameFlowLock", string(marshal))
	return nil
}

func (p *Shield) HandlerLcuSettings(c *tree.Context) error {
	marshal, err := json.Marshal(c.Data)
	if err != nil {
		return err
	}
	fmt.Println("lcusettings", string(marshal))
	return nil
}

func (p *Shield) checkFlow() {
	flow, err := lcu.GetCurrentFlow()
	if err != nil {
		syslog.L.Error("获取游戏状态失败", zap.Error(err))
		return
	}
	p.onGameFlowUpdate(flow)
	return
}
