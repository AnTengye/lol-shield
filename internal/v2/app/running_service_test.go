package app

import (
	"testing"
	"time"

	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu/models"
)

type mockRunningProvider struct {
	session *lcu.GameFlowSession
}

func (m mockRunningProvider) QueryGameFlowSession() (*lcu.GameFlowSession, error) {
	return m.session, nil
}

func (m mockRunningProvider) GetCurrSummoner() (*lcu.SummonerInfo, error) {
	return &lcu.SummonerInfo{SummonerId: 1, Puuid: "self"}, nil
}

func (m mockRunningProvider) ListGamesByUID(uuid string, begin, limit int) (*lcu.GameListResp, error) {
	return &lcu.GameListResp{}, nil
}

func (m mockRunningProvider) GetRankedDataByPUUID(puuid string) (*lcu.RankedData, error) {
	return &lcu.RankedData{}, nil
}

func TestRunningService_NotInProgress(t *testing.T) {
	session := &lcu.GameFlowSession{}
	session.Phase = models.GameFlowLobby
	svc := NewRunningServiceWithProvider(time.Second, mockRunningProvider{session: session})
	_, err := svc.Snapshot()
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != ErrGameNotInProgress {
		t.Fatalf("expected ErrGameNotInProgress, got %v", err)
	}
}

func TestRunningService_InProgress(t *testing.T) {
	session := &lcu.GameFlowSession{}
	session.Phase = models.GameFlowInProgress
	session.GameData.Queue.Id = 420
	session.GameData.Queue.Name = "Ranked Solo"
	svc := NewRunningServiceWithProvider(time.Second, mockRunningProvider{session: session})
	snapshot, err := svc.Snapshot()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if snapshot.QueueID != 420 {
		t.Fatalf("expected queue id 420, got %d", snapshot.QueueID)
	}
}
