package client

import (
	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu/models"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
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

	LOLChat               = "/lol-chat/v1"
	WsLolChatFriendCounts = LOLChat + "/friend-counts"

	// 大厅
	LOLLobby        = "/lol-lobby/v2"
	WsLolLobbyLobby = LOLLobby + "/lobby"

	// 数据记录
	LOLDataStore                     = "/data-store/v1"
	WsLOLInstallSettings             = LOLDataStore + "/install-settings"
	WsLOLInstallSettingsGameFlowLock = WsLOLInstallSettings + "/gameflow-patcher-lock"
	WsLOLInstallSettingsLcuSettings  = WsLOLInstallSettings + "/lcu-settings"
)

func (p *Shield) RegisterStaticRoute() {
	p.wsRouter.RegisterRoute(RouteMethodUpdate, WsLOLGameFlowGamePhase, p.HandlerFlowChange)
	p.wsRouter.RegisterRoute(RouteMethodUpdate, WsLolLobbyLobby, p.HandlerLobbyChange)
	p.wsRouter.RegisterRoute(RouteMethodUpdate, WsLOLChampSelectSession, p.HandlerOnChampSelectSessionUpdate)
	p.wsRouter.RegisterRoute(RouteMethodUpdate, WsLOLLeagueSessionToken, p.HandlerLOLLeagueSessionToken)
	p.wsRouter.RegisterRoute(RouteMethodUpdate, WsLolChatFriendCounts, p.HandlerLolChatFriendCounts)

	//p.wsRouter.RegisterRoute(RouteMethodUpdate, WsLOLInstallSettingsGameFlowLock, p.HandlerGameFlowLock)
	//p.wsRouter.RegisterRoute(RouteMethodUpdate, WsLOLInstallSettingsLcuSettings, p.HandlerLcuSettings)
}

// 状态变更
func (p *Shield) onGameFlowUpdate(gameFlow models.GameStatus) {
	syslog.L.Infof("游戏状态:%s", gameFlow)
	p.updateGameState(gameFlow)
	switch gameFlow {
	case models.GameFlowChampionSelect:
		go p.ChampionSelectStart()
	case models.GameFlowNone:
		p.reset()
		go p.Notice()
	case models.GameFlowInProgress:
		go p.HandlerInProccessGame()
	case models.GameFlowReadyCheck:
		if viper.GetBool(configs.GameAutoConfirm) {
			go p.AcceptGame()
		}
	case models.GameFlowEndOfGame:
		p.CurGame = nil
		p.CurInfo.GameStatus = GSWaiting
		go p.Notice()
	default:
	}
}
