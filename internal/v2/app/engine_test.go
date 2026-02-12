package app

import (
	"context"
	"errors"
	"testing"

	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/v2/domain"
)

type mockAdapter struct {
	port    int
	token   string
	flow    string
	disErr  error
	flowErr error
}

func (m *mockAdapter) Discover(debug bool) (int, string, error) {
	return m.port, m.token, m.disErr
}

func (m *mockAdapter) Connect(port int, token string) {}

func (m *mockAdapter) CurrentFlow() (string, error) {
	return m.flow, m.flowErr
}

func (m *mockAdapter) WatchFlow(
	ctx context.Context,
	port int,
	token string,
	onFlow func(flow string),
	onError func(err error),
) {
}

func TestEngineReconcileOffline(t *testing.T) {
	store := NewStore()
	engine := NewEngine(store, &mockAdapter{disErr: lcu.ErrLolProcessNotFound}, false)
	engine.reconcile()

	s := engine.Snapshot()
	if s.ConnectionStatus != domain.ConnectionOffline {
		t.Fatalf("expected offline, got %s", s.ConnectionStatus)
	}
}

func TestEngineReconcileOnline(t *testing.T) {
	store := NewStore()
	engine := NewEngine(
		store,
		&mockAdapter{
			port:  12345,
			token: "abc",
			flow:  "InProgress",
		},
		false,
	)
	engine.reconcile()

	s := engine.Snapshot()
	if s.ConnectionStatus != domain.ConnectionOnline {
		t.Fatalf("expected online, got %s", s.ConnectionStatus)
	}
	if s.GameFlow != "InProgress" {
		t.Fatalf("expected flow InProgress, got %s", s.GameFlow)
	}
}

func TestEngineReconcileConnectingWhenFlowFailed(t *testing.T) {
	store := NewStore()
	engine := NewEngine(
		store,
		&mockAdapter{
			port:    12345,
			token:   "abc",
			flowErr: errors.New("flow failed"),
		},
		false,
	)
	engine.reconcile()

	s := engine.Snapshot()
	if s.ConnectionStatus != domain.ConnectionConnecting {
		t.Fatalf("expected connecting, got %s", s.ConnectionStatus)
	}
}
