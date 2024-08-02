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

func getTeamUsers() (string, []int64, error) {
	conversationID, err := lcu.GetCurrConversationID()
	if err != nil {
		return "", nil, err
	}
	msgList, err := lcu.ListConversationMsg(conversationID)
	if err != nil {
		return "", nil, err
	}
	summonerIDList := getSummonerIDListFromConversationMsgList(msgList)
	return conversationID, summonerIDList, nil
}
func getSummonerIDListFromConversationMsgList(msgList []lcu.ConversationMsg) []int64 {
	summonerIDList := make([]int64, 0, 5)
	for _, msg := range msgList {
		if msg.Type == lcu.ConversationMsgTypeSystem && msg.Body == lcu.JoinedRoomMsg {
			summonerIDList = append(summonerIDList, msg.FromSummonerId)
		}
	}
	return summonerIDList
}

func GetUserScore(summonerID int64) (*lcu.UserScore, error) {
	userScoreInfo := &lcu.UserScore{
		SummonerID: summonerID,
		Score:      defaultScore,
	}
	return userScoreInfo, nil
}

func listGameHistory(summonerID int64) ([]lcu.GameInfo, error) {
	fmtList := make([]lcu.GameInfo, 0, 20)
	resp, err := lcu.ListGamesBySummonerID(summonerID, 0, 20)
	if err != nil {
		syslog.L.Error("查询用户战绩失败", zap.Error(err), zap.Int64("summonerID", summonerID))
		return nil, err
	}
	for _, gameItem := range resp.Games.Games {
		if gameItem.QueueId != models.NormalQueueID &&
			gameItem.QueueId != models.RankSoleQueueID &&
			gameItem.QueueId != models.ARAMQueueID &&
			gameItem.QueueId != models.RankFlexQueueID {
			continue
		}
		if gameItem.GameDuration < minGameDurationSec {
			continue
		}
		fmtList = append(fmtList, gameItem)
	}
	return fmtList, nil
}

func getAllUsersFromSession(selfID int64, session *lcu.GameFlowSession) (
	selfTeamUsers []int64,
	enemyTeamUsers []int64,
) {
	selfTeamUsers = make([]int64, 0, 5)
	enemyTeamUsers = make([]int64, 0, 5)
	selfTeamID := models.TeamIDNone
	for _, teamUser := range session.GameData.TeamOne {
		summonerID := int64(teamUser.SummonerId)
		if selfID == summonerID {
			selfTeamID = models.TeamIDBlue
			break
		}
	}
	if selfTeamID == models.TeamIDNone {
		for _, teamUser := range session.GameData.TeamTwo {
			summonerID := int64(teamUser.SummonerId)
			if selfID == summonerID {
				selfTeamID = models.TeamIDRed
				break
			}
		}
	}
	if selfTeamID == models.TeamIDNone {
		return
	}
	for _, user := range session.GameData.TeamOne {
		userID := int64(user.SummonerId)
		if userID <= 0 {
			return
		}
		if models.TeamIDBlue == selfTeamID {
			selfTeamUsers = append(selfTeamUsers, userID)
		} else {
			enemyTeamUsers = append(enemyTeamUsers, userID)
		}
	}
	for _, user := range session.GameData.TeamTwo {
		userID := int64(user.SummonerId)
		if userID <= 0 {
			return
		}
		if models.TeamIDRed == selfTeamID {
			selfTeamUsers = append(selfTeamUsers, userID)
		} else {
			enemyTeamUsers = append(enemyTeamUsers, userID)
		}
	}
	return
}
