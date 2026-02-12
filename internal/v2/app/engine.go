package app

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	"github.com/AnTengye/lol-shield/internal/v2/domain"
	vlcu "github.com/AnTengye/lol-shield/internal/v2/lcu"
)

type Engine struct {
	store   *Store
	adapter vlcu.Adapter
	debug   bool

	mu             sync.Mutex
	cancel         context.CancelFunc
	watchCancel    context.CancelFunc
	watcherRunning bool
	wg             sync.WaitGroup
	lastPort       int
	lastToken      string
	initialized    bool
	onFlow         func(flow string)
}

func NewEngine(store *Store, adapter vlcu.Adapter, debug bool) *Engine {
	return &Engine{
		store:   store,
		adapter: adapter,
		debug:   debug,
	}
}

func (e *Engine) Start(ctx context.Context) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.cancel != nil {
		return
	}
	loopCtx, cancel := context.WithCancel(ctx)
	e.cancel = cancel
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		e.loop(loopCtx)
	}()
}

func (e *Engine) Stop() {
	e.mu.Lock()
	cancel := e.cancel
	watchCancel := e.watchCancel
	e.cancel = nil
	e.watchCancel = nil
	e.watcherRunning = false
	e.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if watchCancel != nil {
		watchCancel()
	}
	e.wg.Wait()
}

func (e *Engine) Snapshot() domain.StateSnapshot {
	return e.store.Snapshot()
}

func (e *Engine) SetFlowListener(listener func(flow string)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onFlow = listener
}

func (e *Engine) loop(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.reconcile()
		}
	}
}

func (e *Engine) reconcile() {
	now := time.Now()
	port, token, err := e.adapter.Discover(e.debug)
	if err != nil {
		if !errors.Is(err, lcu.ErrLolProcessNotFound) {
			syslog.L.Warnf("v2 discover lcu failed: %v", err)
		}
		e.store.Update(
			func(current domain.StateSnapshot) domain.StateSnapshot {
				current.ConnectionStatus = domain.ConnectionOffline
				current.GameFlow = "None"
				current.Port = 0
				current.LastError = err.Error()
				current.LastPollAt = now
				return current
			},
		)
		return
	}

	if !e.initialized || e.lastPort != port || e.lastToken != token {
		e.initialized = true
		e.lastPort = port
		e.lastToken = token
		e.adapter.Connect(port, token)
		e.startFlowWatcher(port, token)
		e.store.Update(
			func(current domain.StateSnapshot) domain.StateSnapshot {
				current.ConnectionStatus = domain.ConnectionConnecting
				current.Port = port
				current.LastError = ""
				current.WatcherStatus = "connecting"
				current.LastPollAt = now
				return current
			},
		)
	} else if !e.isWatcherRunning() {
		e.startFlowWatcher(port, token)
		e.store.Update(
			func(current domain.StateSnapshot) domain.StateSnapshot {
				current.WatcherStatus = "reconnecting"
				current.ReconnectCount++
				return current
			},
		)
	}

	flow, err := e.adapter.CurrentFlow()
	if err != nil {
		e.store.Update(
			func(current domain.StateSnapshot) domain.StateSnapshot {
				current.ConnectionStatus = domain.ConnectionConnecting
				current.GameFlow = "None"
				current.Port = port
				current.LastError = err.Error()
				current.LastPollAt = now
				return current
			},
		)
		return
	}

	e.store.Update(
		func(current domain.StateSnapshot) domain.StateSnapshot {
			current.ConnectionStatus = domain.ConnectionOnline
			current.GameFlow = flow
			current.Port = port
			current.LastError = ""
			current.WatcherStatus = "running"
			current.LastPollAt = now
			return current
		},
	)
	e.emitFlow(flow)
}

func (e *Engine) startFlowWatcher(port int, token string) {
	e.mu.Lock()
	if e.watchCancel != nil {
		e.watchCancel()
	}
	watchCtx, watchCancel := context.WithCancel(context.Background())
	e.watchCancel = watchCancel
	e.watcherRunning = true
	e.wg.Add(1)
	e.mu.Unlock()

	go func() {
		defer e.wg.Done()
		e.adapter.WatchFlow(
			watchCtx,
			port,
			token,
			func(flow string) {
				e.store.Update(
					func(current domain.StateSnapshot) domain.StateSnapshot {
						current.ConnectionStatus = domain.ConnectionOnline
						current.GameFlow = flow
						current.Port = port
						current.LastError = ""
						current.WatcherStatus = "running"
						current.LastFlowEventAt = time.Now()
						return current
					},
				)
				e.emitFlow(flow)
			},
			func(err error) {
				e.mu.Lock()
				e.watcherRunning = false
				e.mu.Unlock()
				e.store.Update(
					func(current domain.StateSnapshot) domain.StateSnapshot {
						current.ConnectionStatus = domain.ConnectionConnecting
						current.LastError = err.Error()
						current.WatcherStatus = "stopped"
						return current
					},
				)
			},
		)
	}()
}

func (e *Engine) isWatcherRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.watcherRunning
}

func (e *Engine) emitFlow(flow string) {
	e.mu.Lock()
	listener := e.onFlow
	e.mu.Unlock()
	if listener != nil {
		listener(flow)
	}
}
