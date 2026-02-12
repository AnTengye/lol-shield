package app

import (
	"errors"
	"sync"
	"time"

	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu/models"
	"github.com/AnTengye/lol-shield/internal/v2/domain"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
)

var ErrGameNotInProgress = errors.New("game is not in progress")

type RunningProvider interface {
	QueryGameFlowSession() (*lcu.GameFlowSession, error)
	GetCurrSummoner() (*lcu.SummonerInfo, error)
	ListGamesByUID(uuid string, begin, limit int) (*lcu.GameListResp, error)
	GetRankedDataByPUUID(puuid string) (*lcu.RankedData, error)
}

type legacyRunningProvider struct{}

func (legacyRunningProvider) QueryGameFlowSession() (*lcu.GameFlowSession, error) {
	return lcu.QueryGameFlowSession()
}

func (legacyRunningProvider) GetCurrSummoner() (*lcu.SummonerInfo, error) {
	return lcu.GetCurrSummoner()
}

func (legacyRunningProvider) ListGamesByUID(uuid string, begin, limit int) (*lcu.GameListResp, error) {
	return lcu.ListGamesByUID(uuid, begin, limit)
}

func (legacyRunningProvider) GetRankedDataByPUUID(puuid string) (*lcu.RankedData, error) {
	return lcu.GetRankedDataByPUUID(puuid)
}

type RunningService struct {
	mu       sync.RWMutex
	lastAt   time.Time
	last     domain.RunningSnapshot
	cacheTTL time.Duration
	provider RunningProvider
	sf       singleflight.Group
}

func NewRunningService(cacheTTL time.Duration) *RunningService {
	return NewRunningServiceWithProvider(cacheTTL, legacyRunningProvider{})
}

func NewRunningServiceWithProvider(cacheTTL time.Duration, provider RunningProvider) *RunningService {
	return &RunningService{
		cacheTTL: cacheTTL,
		provider: provider,
	}
}

func (s *RunningService) Snapshot() (domain.RunningSnapshot, error) {
	s.mu.RLock()
	if !s.lastAt.IsZero() && time.Since(s.lastAt) < s.cacheTTL {
		defer s.mu.RUnlock()
		return s.last, nil
	}
	s.mu.RUnlock()

	result, err, _ := s.sf.Do(
		"running_snapshot",
		func() (interface{}, error) {
			return s.build()
		},
	)
	if err != nil {
		return domain.RunningSnapshot{}, err
	}
	snapshot, _ := result.(domain.RunningSnapshot)

	s.mu.Lock()
	s.last = snapshot
	s.lastAt = time.Now()
	s.mu.Unlock()
	return snapshot, nil
}

func (s *RunningService) Invalidate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.last = domain.RunningSnapshot{}
	s.lastAt = time.Time{}
}

func (s *RunningService) build() (domain.RunningSnapshot, error) {
	session, err := s.provider.QueryGameFlowSession()
	if err != nil {
		return domain.RunningSnapshot{}, err
	}
	if session.Phase != models.GameFlowInProgress {
		return domain.RunningSnapshot{}, ErrGameNotInProgress
	}
	self, err := s.provider.GetCurrSummoner()
	if err != nil {
		return domain.RunningSnapshot{}, err
	}

	selfUsers, enemyUsers, groups, skinMap := buildTeamsFromSession(self.SummonerId, session)
	allPlayers := append(append(make([]lcu.UserId, 0, len(selfUsers)+len(enemyUsers)), selfUsers...), enemyUsers...)
	playerMap, err := s.enrichPlayers(allPlayers)
	if err != nil {
		return domain.RunningSnapshot{}, err
	}

	selfTeam := make([]domain.RunningPlayer, 0, len(selfUsers))
	for _, id := range selfUsers {
		selfTeam = append(selfTeam, playerMap[id.Puuid])
	}
	enemyTeam := make([]domain.RunningPlayer, 0, len(enemyUsers))
	for _, id := range enemyUsers {
		enemyTeam = append(enemyTeam, playerMap[id.Puuid])
	}

	groupPuuids := make(map[int][]string, len(groups))
	for k, users := range groups {
		ps := make([]string, 0, len(users))
		for _, u := range users {
			ps = append(ps, u.Puuid)
		}
		groupPuuids[k] = ps
	}

	finalSkinMap := make(map[string]domain.SkinBinding, len(skinMap))
	for puuid, skin := range skinMap {
		finalSkinMap[puuid] = domain.SkinBinding{
			ChampionID: skin.ChampionId,
			SkinID:     skin.SkinId,
		}
	}

	return domain.RunningSnapshot{
		GameFlow:  string(session.Phase),
		QueueID:   session.GameData.Queue.Id,
		QueueName: session.GameData.Queue.Name,
		SelfTeam:  selfTeam,
		EnemyTeam: enemyTeam,
		Groups:    groupPuuids,
		SkinMap:   finalSkinMap,
	}, nil
}

func (s *RunningService) enrichPlayers(users []lcu.UserId) (map[string]domain.RunningPlayer, error) {
	result := make(map[string]domain.RunningPlayer, len(users))
	mu := sync.Mutex{}
	g := errgroup.Group{}
	for _, user := range users {
		u := user
		g.Go(func() error {
			historyResp, err := s.provider.ListGamesByUID(u.Puuid, 0, 5)
			if err != nil {
				return err
			}
			player := domain.RunningPlayer{
				Puuid:      u.Puuid,
				SummonerID: u.SummonerId,
				GameName:   "Unknown",
				TagLine:    "",
				History:    make([]domain.MatchHistory, 0, 5),
			}
			for _, game := range historyResp.Games.Games {
				for _, pid := range game.ParticipantIdentities {
					if pid.Player.Puuid == u.Puuid {
						if pid.Player.GameName != "" {
							player.GameName = pid.Player.GameName
						}
						player.TagLine = pid.Player.TagLine
						break
					}
				}
				h := game.ToGameHistory()
				player.History = append(player.History, domain.MatchHistory{
					GameID:     h.GameId,
					CreateTime: h.CreateTime,
					QueueID:    h.QueueId,
					ChampionID: h.ChampionId,
					Kills:      h.Kills,
					Deaths:     h.Deaths,
					Assists:    h.Assists,
					Win:        h.Win,
					GameMode:   h.GameMode,
				})
			}
			ranked, err := s.provider.GetRankedDataByPUUID(u.Puuid)
			if err == nil {
				player.Highest = domain.RankedBriefInfo{
					Tier:      ranked.HighestRankedEntry.Tier,
					Division:  ranked.HighestRankedEntry.Division,
					QueueType: ranked.HighestRankedEntry.QueueType,
				}
			}
			mu.Lock()
			result[u.Puuid] = player
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return result, nil
}

func buildTeamsFromSession(
	selfSummonerID int64, session *lcu.GameFlowSession,
) (
	selfTeamUsers []lcu.UserId,
	enemyTeamUsers []lcu.UserId,
	groups map[int][]lcu.UserId,
	skinInfoMap map[string]lcu.ChampionSkinInfo,
) {
	selfTeamUsers = make([]lcu.UserId, 0, 5)
	enemyTeamUsers = make([]lcu.UserId, 0, 5)
	selfTeamID := models.TeamIDNone
	for _, teamUser := range session.GameData.TeamOne {
		if selfSummonerID == teamUser.SummonerId {
			selfTeamID = models.TeamIDBlue
			break
		}
	}
	if selfTeamID == models.TeamIDNone {
		for _, teamUser := range session.GameData.TeamTwo {
			if selfSummonerID == teamUser.SummonerId {
				selfTeamID = models.TeamIDRed
				break
			}
		}
	}
	if selfTeamID == models.TeamIDNone {
		return
	}
	skinMap := make(map[int]int, 10)
	for _, skin := range session.GameData.PlayerChampionSelections {
		skinMap[skin.ChampionId] = skin.SelectedSkinIndex
	}
	skinInfoMap = make(map[string]lcu.ChampionSkinInfo, 10)
	teamParticipants := make(map[lcu.UserId]int, 10)
	for _, user := range session.GameData.TeamOne {
		userID := lcu.UserId{SummonerId: user.SummonerId, Puuid: user.Puuid}
		if userID.SummonerId == 0 {
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
		userID := lcu.UserId{SummonerId: user.SummonerId, Puuid: user.Puuid}
		if userID.SummonerId == 0 {
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
	groups = make(map[int][]lcu.UserId)
	for userID, teamID := range teamParticipants {
		groups[teamID] = append(groups[teamID], userID)
	}
	return
}
