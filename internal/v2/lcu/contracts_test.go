package lcu

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	oldlcu "github.com/AnTengye/lol-shield/internal/pkg/lcu"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu/models"
)

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	p := filepath.Join("fixtures", name)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read fixture %s failed: %v", name, err)
	}
	return b
}

func TestContract_GameflowSessionReplay(t *testing.T) {
	b := readFixture(t, "gameflow_session.json")
	var session oldlcu.GameFlowSession
	if err := json.Unmarshal(b, &session); err != nil {
		t.Fatalf("unmarshal gameflow session failed: %v", err)
	}
	if session.Phase != models.GameFlowInProgress {
		t.Fatalf("expected InProgress, got %s", session.Phase)
	}
	if session.GameData.Queue.Id != 420 {
		t.Fatalf("expected queue 420, got %d", session.GameData.Queue.Id)
	}
}

func TestContract_RankedStatsReplay(t *testing.T) {
	b := readFixture(t, "ranked_stats.json")
	var ranked oldlcu.RankedData
	if err := json.Unmarshal(b, &ranked); err != nil {
		t.Fatalf("unmarshal ranked failed: %v", err)
	}
	if ranked.HighestRankedEntry.Tier != "EMERALD" {
		t.Fatalf("expected EMERALD, got %s", ranked.HighestRankedEntry.Tier)
	}
}

func TestContract_MatchHistoryReplay(t *testing.T) {
	b := readFixture(t, "match_history.json")
	var history oldlcu.GameListResp
	if err := json.Unmarshal(b, &history); err != nil {
		t.Fatalf("unmarshal match history failed: %v", err)
	}
	if len(history.Games.Games) != 1 {
		t.Fatalf("expected 1 game, got %d", len(history.Games.Games))
	}
	if history.Games.Games[0].GameId != 1234567890 {
		t.Fatalf("unexpected game id")
	}
}
