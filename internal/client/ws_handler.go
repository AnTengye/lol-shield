package client

import (
	"encoding/json"
	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/pkg/tree"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	RouteMethodUpdate = "Update"
	RouteMethodCreate = "Create"
)
const (
	LOLContentTargeting          = "/lol-content-targeting/v1"
	WsLOLContentTargetingFilters = LOLContentTargeting + "/filters"

	LOLGameFlow            = "/lol-gameflow/v1"
	WsLOLGameFlowGamePhase = LOLGameFlow + "/gameflow-phase"

	LOLChampSelect          = "/lol-champ-select/v1"
	WsLOLChampSelectSession = LOLChampSelect + "/session"

	LOLLeagueSession        = "/lol-league-session/v1"
	WsLOLLeagueSessionToken = LOLLeagueSession + "/league-session-token"

	LOLHovercard                       = "/lol-hovercard/v1"
	WsLOLHovercardFriendInfoBySummoner = LOLHovercard + "/friend-info-by-summoner/:id"
	WsLOLHovercardFriendInfo           = LOLHovercard + "/friend-info"
)

func (p *Shield) AddStaticRoute() {
	p.wsRouter.RegisterRoute(RouteMethodUpdate, WsLOLGameFlowGamePhase, p.HandlerFlowChange)
	p.wsRouter.RegisterRoute(RouteMethodUpdate, WsLOLChampSelectSession, p.HandlerOnChampSelectSessionUpdate)
	p.wsRouter.RegisterRoute(RouteMethodUpdate, WsLOLLeagueSessionToken, p.HandlerLOLLeagueSessionToken)
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
	p.onGameFlowUpdate(gameFlow)
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
