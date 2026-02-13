package client

import (
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
	historyMap = map[string][]lcu.GameHistory{}
	userNameMap = make(map[string]lcu.UserName, 10)
	mu := sync.Mutex{}
	for _, summoner := range userList {
		puuid := summoner.Puuid
		g.Go(
			func() error {
				retry := 3
				for retry > 0 {
					retry--
					tmap, userName, err := gameHistorySync(lcuSvc, puuid)
					if err != nil {
						return err
					}
					if len(tmap) == 0 {
						continue
					}
					if userName.GameName == "" || userName.TagLine == "" {
						summonerInfo, err := lcuSvc.GetSummonerInfoByPUUID(puuid)
						if err != nil {
							return err
						}
						userName = lcu.UserName{
							GameName: summonerInfo.GameName,
							TagLine:  summonerInfo.TagLine,
						}
					}
					mu.Lock()
					defer mu.Unlock()
					maps.Copy(historyMap, tmap)
					userNameMap[puuid] = userName
					return nil
				}
				return nil
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
	if len(listResp.Games.Games) == 0 {
		return nil, userName, nil
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
