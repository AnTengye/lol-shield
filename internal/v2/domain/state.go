package domain

import "time"

type ConnectionStatus string

const (
	ConnectionOffline    ConnectionStatus = "offline"
	ConnectionConnecting ConnectionStatus = "connecting"
	ConnectionOnline     ConnectionStatus = "online"
)

type StateSnapshot struct {
	ConnectionStatus ConnectionStatus `json:"connectionStatus"`
	GameFlow         string           `json:"gameFlow"`
	Port             int              `json:"port,omitempty"`
	LastError        string           `json:"lastError,omitempty"`
	WatcherStatus    string           `json:"watcherStatus"`
	ReconnectCount   int              `json:"reconnectCount"`
	LastFlowEventAt  time.Time        `json:"lastFlowEventAt,omitempty"`
	LastPollAt       time.Time        `json:"lastPollAt,omitempty"`
	UpdatedAt        time.Time        `json:"updatedAt"`
}

func NewInitialSnapshot() StateSnapshot {
	return StateSnapshot{
		ConnectionStatus: ConnectionOffline,
		GameFlow:         "None",
		WatcherStatus:    "idle",
		UpdatedAt:        time.Now(),
	}
}
