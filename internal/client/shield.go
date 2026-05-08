package client

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/AnTengye/lol-shield/configs"
	"github.com/AnTengye/lol-shield/internal/client/middleware"
	"github.com/AnTengye/lol-shield/internal/client/ws"
	"github.com/AnTengye/lol-shield/internal/core/lcuapi"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu/models"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	tree2 "github.com/AnTengye/lol-shield/internal/pkg/tree"
	"github.com/avast/retry-go"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type StatusType int
type GameStatusType int

const (
	STOnline StatusType = iota + 1
	STWaiting
)

const (
	GSWaiting GameStatusType = iota + 1
	GSStarted
)

type (
	Shield struct {
		ctx          context.Context
		httpSrv      *http.Server
		CurInfo      StatusInfo
		currSummoner *lcu.SummonerInfo
		cancel       func()
		mu           *sync.Mutex
		GameState    models.GameStatus
		CurGame      *GameInfo
		CurLobby     *LobbyInfo
		wsRouter     *tree2.Engine
		webWs        *ws.WebClient
		port         int
		token        string
		lcuService   lcuapi.Service
	}
	wsMsg struct {
		Data      interface{} `json:"data"`
		EventType string      `json:"eventType"`
		Uri       string      `json:"uri"`
	}
	StatusInfo struct {
		Status     StatusType     // 客户端状态
		GameStatus GameStatusType //游戏状态
		Uid        int64          //summonerId
		Uuid       string         //puuid
		SkinSync   int            //皮肤数据是否已同步
	}
)

type GameInfo struct {
	//GameSession lcu.GameFlowSession
	SelfTeamInfo   lcu.TeamInfo                    `json:"selfTeamInfo"`
	EnemyTeamInfo  lcu.TeamInfo                    `json:"enemyTeamInfo"`
	PreTeam        map[int][]lcu.UserId            `json:"preTeam"`
	AllGameHistory map[string][]lcu.GameHistory    `json:"allGameHistory"`
	UserNameMap    map[string]lcu.UserName         `json:"userNameMap"`
	SkinMap        map[string]lcu.ChampionSkinInfo `json:"skinMap"`
	QueueId        models.GameQueueID              `json:"queueId"`
	QueueName      string                          `json:"queueName"`
}

type LobbyInfo struct {
	AllowPeople int // 允许的人数
	GameMode    models.GameMode
	QueueId     models.GameQueueID
}

func (m wsMsg) ConvertToContext() *tree2.Context {
	return &tree2.Context{
		Data:   m.Data,
		Path:   m.Uri,
		Method: m.EventType,
	}
}

func (m StatusInfo) ToData() []byte {
	b, _ := json.Marshal(m)
	return b
}

const (
	// api事件前缀
	onJsonApiEventPrefixLen = len(`[8,"OnJsonApiEvent",`)
)

func NewShield() *Shield {
	return NewShieldWithLCU(lcuapi.New())
}

func NewShieldWithLCU(lcuSvc lcuapi.Service) *Shield {
	ctx, cancel := context.WithCancel(context.Background())
	if lcuSvc == nil {
		lcuSvc = lcuapi.New()
	}
	p := &Shield{
		ctx:       ctx,
		cancel:    cancel,
		mu:        &sync.Mutex{},
		GameState: models.GameFlowNone,
		CurInfo: StatusInfo{
			Status:     STWaiting,
			GameStatus: GSWaiting,
		},
		wsRouter:   tree2.NewEngine(),
		lcuService: lcuSvc,
	}
	p.RegisterStaticRoute()
	return p
}

func NewServer(addr string, p *Shield) *http.Server {
	engine := gin.New()
	//engine.Use(gin.Recovery())
	engine.Use(middleware.GinLogger(syslog.L), middleware.GinRecovery(syslog.L, true))
	engine.Use(middleware.Cors())
	AddRouter(engine, p)

	return &http.Server{
		Addr:    addr,
		Handler: engine,
	}
}

func (p *Shield) Run() error {
	if viper.GetBool(configs.MockLCUEnabled) {
		go p.bootstrapMockState()
	} else {
		go p.MonitorStart()
	}
	syslog.L.Infof("等待客户端连接中...")
	return p.notifyQuit()
}
func (p *Shield) isLcuActive() bool {
	return p.CurInfo.Status == STOnline
}
func (p *Shield) notifyQuit() error {
	if viper.GetBool(configs.Dev) {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	webAddr := viper.GetString(configs.WebAddr)
	srv := NewServer(webAddr, p)
	p.httpSrv = srv
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	g, c := errgroup.WithContext(p.ctx)
	// http
	g.Go(
		func() error {
			err := p.httpSrv.ListenAndServe()
			if err != nil || !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		},
	)
	if viper.GetBool(configs.WebAutoOpen) {
		go func() {
			if !waitForWebReady(webAddr, 3*time.Second) {
				syslog.L.Warnf("页面未在预期时间内就绪，跳过自动打开: %s", normalizeWebURL(webAddr))
				return
			}
			if err := openWebPage(webAddr); err != nil {
				syslog.L.Warnf("自动打开浏览器失败: %v", err)
				return
			}
			syslog.L.Infof("已自动打开页面: %s", normalizeWebURL(webAddr))
		}()
	}
	// http-shutdown
	g.Go(
		func() error {
			<-c.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			return p.httpSrv.Shutdown(ctx)
		},
	)
	// wait quit
	g.Go(
		func() error {
			for {
				select {
				case <-p.ctx.Done():
					return p.ctx.Err()
				case <-interrupt:
					_ = p.Stop()
				}
			}
		},
	)
	err := g.Wait()
	if err != nil && !errors.Is(err, context.Canceled) {
		return err
	}
	return nil
}
func (p *Shield) Stop() error {
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}
func (p *Shield) getTokenFromFile() {

}

// MonitorStart 启动客户端监控
func (p *Shield) MonitorStart() {
	for {
		time.Sleep(time.Second)
		if !p.isLcuActive() {
			port, token, err := p.lcuService.GetToken(viper.GetBool(configs.LCUTokenFromFile))
			if err != nil {
				if !p.lcuService.IsProcessNotFound(err) {
					syslog.L.Error("获取lcu info 失败", zap.Error(err))
				}
				continue
			}
			p.lcuService.Init(port, token)
			syslog.L.Debug("lcu info", zap.Int("port", port), zap.String("token", token))
			err = p.runMonitor(port, token)
			if err != nil {
				syslog.L.Debugf("客户端已断开:%v", zap.Error(err))
			}
			p.currSummoner = nil
			p.CurInfo = StatusInfo{
				Status: STWaiting,
			}
			p.Notice()
		}
	}
}

// 初始化客户端监控
func (p *Shield) runMonitor(port int, authPwd string) error {
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
	rawUrl := fmt.Sprintf("wss://127.0.0.1:%d/", port)
	header := http.Header{}
	authSecret := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("riot:%s", authPwd)))
	header.Set("Authorization", "Basic "+authSecret)
	u, _ := url.Parse(rawUrl)
	syslog.L.Debugf("connect to lcu %s", u.String())
	c, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		return err
	}
	defer c.Close()
	err = retry.Do(
		func() error {
			currSummoner, err := p.lcuService.GetCurrSummoner()
			if err == nil {
				p.currSummoner = currSummoner
			}
			return err
		}, retry.Attempts(5), retry.Delay(time.Second),
	)
	if err != nil {
		return errors.New("获取当前召唤师信息失败:" + err.Error())
	}
	p.CurInfo = StatusInfo{
		Status:     STOnline,
		GameStatus: GSWaiting,
		Uid:        p.currSummoner.SummonerId,
		Uuid:       p.currSummoner.Puuid,
	}
	p.Notice()
	go p.initSkin(p.currSummoner.SummonerId)
	go p.checkFlow()
	_ = c.WriteMessage(websocket.TextMessage, []byte("[5, \"OnJsonApiEvent\"]"))
	for {
		msgType, message, err := c.ReadMessage()
		if err != nil {
			syslog.L.Debug("lol事件监控读取消息失败", zap.Error(err))
			return err
		}
		msg := &wsMsg{}
		if msgType != websocket.TextMessage || len(message) < onJsonApiEventPrefixLen+1 {
			syslog.L.Debugf("msgType:%v, message:%v", msgType, string(message))
			continue
		}
		_ = json.Unmarshal(message[onJsonApiEventPrefixLen:len(message)-1], msg)
		go func(m *wsMsg) {
			routeHandlerErr := p.wsRouter.GetRoute(m.ConvertToContext())
			if routeHandlerErr != nil {
				syslog.L.Debug(routeHandlerErr)
			}
		}(msg)
	}
}

func (p *Shield) updateGameState(state models.GameStatus) {
	p.mu.Lock()
	p.GameState = state
	p.mu.Unlock()
}
func (p *Shield) getGameState() models.GameStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.GameState
}

func (p *Shield) Notice() {
	if p.webWs == nil {
		return
	}
	p.webWs.Write(p.CurInfo.ToData())
}

func (p *Shield) bootstrapMockState() {
	currSummoner, err := p.lcuService.GetCurrSummoner()
	if err != nil {
		syslog.L.Fatalf("mock lcu bootstrap failed: %v", err)
	}
	p.currSummoner = currSummoner
	p.CurInfo = StatusInfo{
		Status:     STOnline,
		GameStatus: GSStarted,
		Uid:        currSummoner.SummonerId,
		Uuid:       currSummoner.Puuid,
	}
	flow, err := p.lcuService.GetCurrentFlow()
	if err != nil {
		syslog.L.Warnf("mock flow bootstrap failed: %v", err)
		p.Notice()
		return
	}
	p.updateGameState(flow)
	if flow == models.GameFlowInProgress {
		p.HandlerInProccessGame()
	} else {
		p.onGameFlowUpdate(flow)
	}
	p.Notice()
}

func (p *Shield) reset() {
	p.CurGame = nil
	p.CurLobby = nil
	p.CurInfo.GameStatus = GSWaiting
}
