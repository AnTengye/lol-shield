package client

import (
	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu/models"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	"go.uber.org/zap"
)

const (
	defaultScore       = 100 // 默认分数
	minGameDurationSec = 15 * 60
)

var (
	SkinInfo map[int64]lcu.SkinUrl
)

func initSkin(summonerId int64) {
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
}

func getTeamUsers() (string, []lcu.UserId, error) {
	conversationID, err := lcu.GetCurrConversationID()
	if err != nil {
		return "", nil, err
	}
	msgList, err := lcu.ListConversationMsg(conversationID)
	if err != nil {
		return "", nil, err
	}
	summonerIDList := getIDListFromConversationMsgList(msgList)
	return conversationID, summonerIDList, nil
}
func getIDListFromConversationMsgList(msgList []lcu.ConversationMsg) []lcu.UserId {
	list := make([]lcu.UserId, 0, 5)
	for _, msg := range msgList {
		if msg.Type == lcu.ConversationMsgTypeSystem && msg.Body == lcu.JoinedRoomMsg {
			list = append(list, lcu.UserId{
				SummonerId: msg.FromSummonerId,
				Puuid:      msg.FromPid[:len(models.PUUIDNone)],
			})
		}
	}
	return list
}

func getAllUsersFromSession(selfID lcu.UserId, session *lcu.GameFlowSession) (
	selfTeam lcu.TeamInfo, enemyTeam lcu.TeamInfo, groups map[int][]lcu.UserId, skinInfoMap map[string]lcu.ChampionSkinInfo,
) {
	selfTeamUsers := make([]lcu.UserId, 0, 5)
	enemyTeamUsers := make([]lcu.UserId, 0, 5)
	selfTeamID := models.TeamIDNone
	for _, teamUser := range session.GameData.TeamOne {
		if selfID.SummonerId == teamUser.SummonerId {
			selfTeamID = models.TeamIDBlue
			break
		}
	}
	if selfTeamID == models.TeamIDNone {
		for _, teamUser := range session.GameData.TeamTwo {
			if selfID.SummonerId == teamUser.SummonerId {
				selfTeamID = models.TeamIDRed
				break
			}
		}
	}
	if selfTeamID == models.TeamIDNone {
		syslog.L.Errorf("获取用户队伍信息失败:%+v", session.GameData)
		return
	}
	skinMap := make(map[int]int, 10)
	for _, skin := range session.GameData.PlayerChampionSelections {
		skinMap[skin.ChampionId] = skin.SelectedSkinIndex
	}
	skinInfoMap = make(map[string]lcu.ChampionSkinInfo, 10)
	teamParticipants := make(map[lcu.UserId]int, 10)
	for _, user := range session.GameData.TeamOne {
		userID := lcu.UserId{
			SummonerId: user.SummonerId,
			Puuid:      user.Puuid,
		}
		if userID.SummonerId == 0 {
			// 人机
			break
		}
		if models.TeamIDBlue == selfTeamID {
			selfTeamUsers = append(selfTeamUsers, userID)
		} else {
			enemyTeamUsers = append(enemyTeamUsers, userID)
		}
		teamParticipants[userID] = user.TeamParticipantId
		if skinIndex, ok := skinMap[user.ChampionId]; ok {
			skinInfoMap[userID.Puuid] = lcu.ChampionSkinInfo{
				ChampionId: int64(user.ChampionId),
				SkinId:     int64(user.ChampionId*1000 + skinIndex),
			}
		}
	}
	for _, user := range session.GameData.TeamTwo {
		userID := lcu.UserId{
			SummonerId: user.SummonerId,
			Puuid:      user.Puuid,
		}
		if userID.SummonerId == 0 {
			// 人机
			break
		}
		if models.TeamIDRed == selfTeamID {
			selfTeamUsers = append(selfTeamUsers, userID)
		} else {
			enemyTeamUsers = append(enemyTeamUsers, userID)
		}
		teamParticipants[userID] = user.TeamParticipantId
		if skinIndex, ok := skinMap[user.ChampionId]; ok {
			skinInfoMap[userID.Puuid] = lcu.ChampionSkinInfo{
				ChampionId: int64(user.ChampionId),
				SkinId:     int64(user.ChampionId*1000 + skinIndex),
			}
		}
	}
	// 使用map根据teamId分组
	groups = make(map[int][]lcu.UserId)

	for userId, teamId := range teamParticipants {
		groups[teamId] = append(groups[teamId], userId)
	}
	enemyTeamID := models.TeamIDBlue
	if models.TeamIDBlue == selfTeamID {
		enemyTeamID = models.TeamIDRed
	}
	selfTeam = lcu.TeamInfo{
		UserList: selfTeamUsers,
		TeamId:   selfTeamID,
	}
	enemyTeam = lcu.TeamInfo{
		UserList: enemyTeamUsers,
		TeamId:   enemyTeamID,
	}
	return
}

// 通过对局信息，返回红蓝双方的队伍信息
func getTeamInfo(self lcu.UserId, game *lcu.GameSummary) (selfTeamUsers []lcu.UserId, enemyTeamUsers []lcu.UserId) {
	selfTeamUsers = make([]lcu.UserId, 0, 5)
	enemyTeamUsers = make([]lcu.UserId, 0, 5)
	partInfoMap := make(map[int]models.TeamID, 10) // key:participantId value:teamId
	for _, partInfo := range game.Participants {
		partInfoMap[partInfo.ParticipantId] = partInfo.TeamId
	}
	// 是否反转
	isReverse := false
	for _, teamUser := range game.ParticipantIdentities {
		userID := lcu.UserId{
			SummonerId: teamUser.Player.SummonerId,
			Puuid:      teamUser.Player.Puuid,
		}
		if teamId, ok := partInfoMap[teamUser.ParticipantId]; ok {
			if teamId == models.TeamIDBlue {
				selfTeamUsers = append(selfTeamUsers, userID)
			} else {
				if teamUser.Player.SummonerId == self.SummonerId {
					isReverse = true
				}
				enemyTeamUsers = append(enemyTeamUsers, userID)
			}
		}
	}
	if isReverse {
		selfTeamUsers, enemyTeamUsers = enemyTeamUsers, selfTeamUsers
	}
	return
}
