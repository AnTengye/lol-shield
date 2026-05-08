package mocklcu

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadScenarioBuildsHistoryLookupKeys(t *testing.T) {
	scenarioDir := filepath.Join("fixtures", "testdata", "default")

	scenario, err := LoadScenario(scenarioDir)
	if err != nil {
		t.Fatalf("LoadScenario returned error: %v", err)
	}

	payload, ok := scenario.MatchHistory["de06293d-082d-59c2-83a6-273ab88164bc|0|8"]
	if !ok {
		t.Fatalf("expected history fixture key to exist")
	}
	if payload.Games.GameCount != 9 {
		t.Fatalf("expected gameCount 9, got %d", payload.Games.GameCount)
	}

	secondPayload, ok := scenario.MatchHistory["de06293d-082d-59c2-83a6-273ab88164bc|9|17"]
	if !ok {
		t.Fatalf("expected second history fixture key to exist")
	}
	if secondPayload.Games.GameCount != 4 {
		t.Fatalf("expected second fixture gameCount 4, got %d", secondPayload.Games.GameCount)
	}
}

func TestLoadScenarioLoadsDefaultRepositoryFixtures(t *testing.T) {
	scenario, err := LoadScenario(filepath.Join("fixtures", "default"))
	if err != nil {
		t.Fatalf("LoadScenario returned error: %v", err)
	}

	if len(scenario.GameflowPhase) == 0 {
		t.Fatalf("expected non-empty gameflow phase fixture")
	}
	if len(scenario.GameflowSession) == 0 {
		t.Fatalf("expected non-empty gameflow session fixture")
	}
	if len(scenario.CurrentSummoner) == 0 {
		t.Fatalf("expected non-empty current summoner fixture")
	}
	if len(scenario.Conversations) == 0 {
		t.Fatalf("expected non-empty conversations fixture")
	}
	if len(scenario.ConversationMessages["champ-select"]) == 0 {
		t.Fatalf("expected champ-select conversation messages fixture")
	}
	if len(scenario.RankedStats["de06293d-082d-59c2-83a6-273ab88164bc"]) == 0 {
		t.Fatalf("expected ranked stats fixture for local player")
	}
	if len(scenario.RankedStats["75126a7d-28e3-5dfa-8874-3a075c1805b1"]) == 0 {
		t.Fatalf("expected ranked stats fixture for secondary player")
	}
	if len(scenario.SummonersByPUUID["de06293d-082d-59c2-83a6-273ab88164bc"]) == 0 {
		t.Fatalf("expected summoner fixture for local player")
	}
	if len(scenario.SummonersByPUUID["75126a7d-28e3-5dfa-8874-3a075c1805b1"]) == 0 {
		t.Fatalf("expected summoner fixture for secondary player")
	}

	history, ok := scenario.MatchHistory["75126a7d-28e3-5dfa-8874-3a075c1805b1|0|9"]
	if !ok {
		t.Fatalf("expected history fixture for second player")
	}
	if len(history.Raw) == 0 {
		t.Fatalf("expected non-empty history fixture payload")
	}

	detail, ok := scenario.GameDetails[10913327389]
	if !ok {
		t.Fatalf("expected game detail fixture for 10913327389")
	}
	if len(detail) == 0 {
		t.Fatalf("expected non-empty game detail fixture payload")
	}

	if len(scenario.Assets["placeholder.png"]) == 0 {
		t.Fatalf("expected placeholder asset fixture")
	}
}

func TestLoadScenarioFailsOnUnexpectedFixtureLayout(t *testing.T) {
	scenarioDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(scenarioDir, "gameflow-phase.json"), []byte(`"InProgress"`), 0o644); err != nil {
		t.Fatalf("write phase fixture: %v", err)
	}

	badPath := filepath.Join(scenarioDir, "match-history", "products", "bad.json")
	if err := os.MkdirAll(filepath.Dir(badPath), 0o755); err != nil {
		t.Fatalf("mkdir bad fixture dir: %v", err)
	}
	if err := os.WriteFile(badPath, []byte(`{"games":{"gameCount":1}}`), 0o644); err != nil {
		t.Fatalf("write bad fixture: %v", err)
	}

	_, err := LoadScenario(scenarioDir)
	if err == nil {
		t.Fatalf("expected malformed fixture layout error")
	}
	if !strings.Contains(err.Error(), filepath.ToSlash(badPath)) && !strings.Contains(err.Error(), badPath) {
		t.Fatalf("expected error to mention offending path %q, got %v", badPath, err)
	}
}

func TestResolveJSONRouteMatchesHistoryRequest(t *testing.T) {
	scenario := &Scenario{
		GameflowPhase: []byte(`"InProgress"`),
		MatchHistory: map[string]*HistoryFixture{
			"de06293d-082d-59c2-83a6-273ab88164bc|0|8": {
				Raw: []byte(`{"games":{"gameCount":9}}`),
			},
		},
	}

	body, contentType, status := ResolveRequest(scenario, "/lol-match-history/v1/products/lol/de06293d-082d-59c2-83a6-273ab88164bc/matches?begIndex=0&endIndex=8")
	if status != 200 {
		t.Fatalf("expected status 200, got %d", status)
	}
	if contentType != "application/json" {
		t.Fatalf("expected json content type, got %q", contentType)
	}
	if string(body) != `{"games":{"gameCount":9}}` {
		t.Fatalf("unexpected body: %s", string(body))
	}
}

func TestResolveJSONRouteMatchesGameflowPhaseRequest(t *testing.T) {
	scenario := &Scenario{
		GameflowPhase: []byte(`"InProgress"`),
	}

	body, contentType, status := ResolveRequest(scenario, "/lol-gameflow/v1/gameflow-phase")
	if status != 200 {
		t.Fatalf("expected status 200, got %d", status)
	}
	if contentType != "application/json" {
		t.Fatalf("expected json content type, got %q", contentType)
	}
	if string(body) != `"InProgress"` {
		t.Fatalf("unexpected body: %s", string(body))
	}
}

func TestResolveJSONRouteMatchesDefaultScenarioRealtimeRoutes(t *testing.T) {
	scenario, err := LoadScenario(filepath.Join("fixtures", "default"))
	if err != nil {
		t.Fatalf("LoadScenario returned error: %v", err)
	}

	cases := []struct {
		name     string
		rawURL   string
		wantCode int
	}{
		{name: "gameflow session", rawURL: "/lol-gameflow/v1/session", wantCode: 200},
		{name: "current summoner", rawURL: "/lol-summoner/v1/current-summoner", wantCode: 200},
		{name: "conversation messages", rawURL: "/lol-chat/v1/conversations/champ-select/messages", wantCode: 200},
		{name: "summoner by puuid", rawURL: "/lol-summoner/v2/summoners/puuid/75126a7d-28e3-5dfa-8874-3a075c1805b1", wantCode: 200},
		{name: "game detail", rawURL: "/lol-match-history/v1/games/10913327389", wantCode: 200},
		{name: "ranked stats by puuid", rawURL: "/lol-ranked/v1/ranked-stats/de06293d-082d-59c2-83a6-273ab88164bc", wantCode: 200},
		{name: "current ranked stats", rawURL: "/lol-ranked/v1/current-ranked-stats", wantCode: 200},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body, contentType, status := ResolveRequest(scenario, tc.rawURL)
			if status != tc.wantCode {
				t.Fatalf("expected status %d, got %d", tc.wantCode, status)
			}
			if contentType != "application/json" {
				t.Fatalf("expected json content type, got %q", contentType)
			}
			if len(body) == 0 {
				t.Fatalf("expected non-empty response body")
			}
		})
	}
}

func TestResolveJSONRouteRejectsNonMatchesHistoryRequest(t *testing.T) {
	scenario := &Scenario{
		MatchHistory: map[string]*HistoryFixture{
			"de06293d-082d-59c2-83a6-273ab88164bc|0|8": {
				Raw: []byte(`{"games":{"gameCount":9}}`),
			},
		},
	}

	_, contentType, status := ResolveRequest(scenario, "/lol-match-history/v1/products/lol/de06293d-082d-59c2-83a6-273ab88164bc/profile?begIndex=0&endIndex=8")
	if status != 404 {
		t.Fatalf("expected status 404, got %d", status)
	}
	if contentType != "application/json" {
		t.Fatalf("expected json content type, got %q", contentType)
	}
}
