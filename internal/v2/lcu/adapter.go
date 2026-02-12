package lcu

import (
	"context"

	oldlcu "github.com/AnTengye/lol-shield/internal/pkg/lcu"
)

type Adapter interface {
	Discover(debug bool) (port int, token string, err error)
	Connect(port int, token string)
	CurrentFlow() (string, error)
	WatchFlow(ctx context.Context, port int, token string, onFlow func(flow string), onError func(err error))
}

type LegacyAdapter struct{}

func NewLegacyAdapter() *LegacyAdapter {
	return &LegacyAdapter{}
}

func (a *LegacyAdapter) Discover(debug bool) (port int, token string, err error) {
	return oldlcu.GetLcuToken(debug)
}

func (a *LegacyAdapter) Connect(port int, token string) {
	oldlcu.InitCli(port, token)
}

func (a *LegacyAdapter) CurrentFlow() (string, error) {
	flow, err := oldlcu.GetCurrentFlow()
	if err != nil {
		return "", err
	}
	return string(flow), nil
}
