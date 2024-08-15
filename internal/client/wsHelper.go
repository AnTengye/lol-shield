package client

import (
	"sync"

	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	"golang.org/x/exp/maps"
	"golang.org/x/sync/errgroup"
)

func getGameHistoryByUserList(userList []lcu.UserId) (historyMap map[string][]lcu.GameHistory, userNameMap map[string]lcu.UserName, err error) {
	g := errgroup.Group{}
	historyMap = map[string][]lcu.GameHistory{}
	userNameMap = make(map[string]lcu.UserName, 10)
	mu := sync.Mutex{}
	for _, summoner := range userList {
		puuid := summoner.Puuid
		g.Go(
			func() error {
				tmap, userName, err := gameHistorySync(puuid)
				if err != nil {
					return err
				}
				mu.Lock()
				defer mu.Unlock()
				maps.Copy(historyMap, tmap)
				userNameMap[puuid] = userName
				return nil
			},
		)
	}
	err = g.Wait()
	if err != nil {
		syslog.L.Errorf("查询用户得分失败:%v", err)
		return
	}
	return historyMap, userNameMap, nil
}

func gameHistorySync(puuid string) (historyMap map[string][]lcu.GameHistory, userName lcu.UserName, err error) {
	historyMap = make(map[string][]lcu.GameHistory, 1)
	listResp, err := lcu.ListGamesByUID(puuid, 0, 10)
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
