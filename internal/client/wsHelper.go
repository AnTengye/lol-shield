package client

import (
	"errors"
	"sync"
	"time"

	"github.com/AnTengye/lol-shield/internal/core/lcuapi"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
)

func getGameHistoryByUserList(
	lcuSvc lcuapi.Service, userList []lcu.UserId,
) (historyMap map[string][]lcu.GameHistory, userNameMap map[string]lcu.UserName, err error) {
	g := errgroup.Group{}
	g.SetLimit(4)
	historyMap = map[string][]lcu.GameHistory{}
	userNameMap = make(map[string]lcu.UserName, 10)
	mu := sync.Mutex{}
	for _, summoner := range userList {
		puuid := summoner.Puuid
		g.Go(
			func() error {
				tmap, userName, historyErr := gameHistorySync(lcuSvc, puuid)
				var profileErr error
				if userName.GameName == "" || userName.TagLine == "" {
					var profile *lcu.SummonerInfo
					profile, profileErr = lcuSvc.GetSummonerInfoByPUUID(puuid)
					if profileErr == nil && profile != nil {
						userName = lcu.UserName{GameName: profile.GameName, TagLine: profile.TagLine}
					}
				}
				mu.Lock()
				maps.Copy(historyMap, tmap)
				if userName.GameName != "" {
					userNameMap[puuid] = userName
				}
				mu.Unlock()
				return errors.Join(historyErr, profileErr)
			},
		)
		// 增加间隔，防止客户端崩溃
		time.Sleep(time.Second)
	}
	err = g.Wait()
	if err != nil {
		syslog.L.Errorf("查询用户得分失败:%v", err)
		return
	}
	return historyMap, userNameMap, nil
}

func gameHistorySync(lcuSvc lcuapi.Service, puuid string) (
	historyMap map[string][]lcu.GameHistory, userName lcu.UserName, err error,
) {
	historyMap = make(map[string][]lcu.GameHistory, 1)
	listResp, err := lcuSvc.ListGamesByUID(puuid, 0, 10)
	if err != nil {

		return nil, userName, err
	}
	if listResp == nil || listResp.ErrorCode != "" {
		return nil, userName, errors.New("客户端未返回有效战绩")
	}
	historyMap[puuid] = []lcu.GameHistory{}
	if len(listResp.Games.Games) == 0 {
		return historyMap, userName, nil
	}
	for _, game := range listResp.Games.Games {
		historyMap[puuid] = append(historyMap[puuid], game.ToGameHistory())
		for _, par := range game.ParticipantIdentities {
			if par.Player.Puuid == puuid {
				userName = lcu.UserName{
					GameName: par.Player.GameName,
					TagLine:  par.Player.TagLine,
				}
				break
			}
		}
	}
	return historyMap, userName, nil
}
