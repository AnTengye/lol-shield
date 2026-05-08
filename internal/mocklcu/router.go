package mocklcu

import (
	"net/url"
	"strings"
)

func ResolveRequest(s *Scenario, rawURL string) ([]byte, string, int) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return []byte(`{"error":"bad request"}`), "application/json", 400
	}

	if puuid, ok := matchHistoryPUUID(u.Path); ok {
		key := puuid + "|" + u.Query().Get("begIndex") + "|" + u.Query().Get("endIndex")
		if fixture, ok := s.MatchHistory[key]; ok {
			return fixture.Raw, "application/json", 200
		}
	}

	if u.Path == "/lol-gameflow/v1/gameflow-phase" {
		return s.GameflowPhase, "application/json", 200
	}

	return []byte(`{"error":"not found"}`), "application/json", 404
}

func matchHistoryPUUID(path string) (string, bool) {
	const prefix = "/lol-match-history/v1/products/lol/"
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}

	parts := strings.Split(strings.TrimPrefix(path, prefix), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "matches" {
		return "", false
	}

	return parts[0], true
}
