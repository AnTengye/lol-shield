package mocklcu

import (
	"path/filepath"
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
