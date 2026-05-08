package mocklcu

import (
	"encoding/json"
	"mime"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
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
	if u.Path == "/lol-gameflow/v1/session" && len(s.GameflowSession) != 0 {
		return s.GameflowSession, "application/json", 200
	}
	if u.Path == "/lol-summoner/v1/current-summoner" && len(s.CurrentSummoner) != 0 {
		return s.CurrentSummoner, "application/json", 200
	}
	if u.Path == "/lol-chat/v1/conversations" && len(s.Conversations) != 0 {
		return s.Conversations, "application/json", 200
	}
	if conversationID, ok := conversationMessageID(u.Path); ok {
		if body := s.ConversationMessages[conversationID]; len(body) != 0 {
			return body, "application/json", 200
		}
	}
	if puuid, ok := summonerByPUUID(u.Path); ok {
		if body := s.SummonersByPUUID[puuid]; len(body) != 0 {
			return body, "application/json", 200
		}
	}
	if gameID, ok := matchHistoryGameID(u.Path); ok {
		if body := s.GameDetails[gameID]; len(body) != 0 {
			return body, "application/json", 200
		}
	}
	if puuid, ok := rankedStatsPUUID(u.Path); ok {
		if body := s.RankedStats[puuid]; len(body) != 0 {
			return body, "application/json", 200
		}
	}
	if u.Path == "/lol-ranked/v1/current-ranked-stats" {
		if body := currentRankedStats(s); len(body) != 0 {
			return body, "application/json", 200
		}
	}
	if assetKey, ok := assetPath(u.Path); ok {
		if body := s.Assets[assetKey]; len(body) != 0 {
			return body, assetContentTypeFromPath(u.Path), 200
		}
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

func conversationMessageID(requestPath string) (string, bool) {
	const prefix = "/lol-chat/v1/conversations/"
	if !strings.HasPrefix(requestPath, prefix) {
		return "", false
	}

	parts := strings.Split(strings.TrimPrefix(requestPath, prefix), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "messages" {
		return "", false
	}
	return parts[0], true
}

func summonerByPUUID(requestPath string) (string, bool) {
	const prefix = "/lol-summoner/v2/summoners/puuid/"
	if !strings.HasPrefix(requestPath, prefix) {
		return "", false
	}

	puuid := strings.TrimPrefix(requestPath, prefix)
	if puuid == "" || strings.Contains(puuid, "/") {
		return "", false
	}
	return puuid, true
}

func matchHistoryGameID(requestPath string) (int64, bool) {
	const prefix = "/lol-match-history/v1/games/"
	if !strings.HasPrefix(requestPath, prefix) {
		return 0, false
	}

	rawID := strings.TrimPrefix(requestPath, prefix)
	if rawID == "" || strings.Contains(rawID, "/") {
		return 0, false
	}
	gameID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		return 0, false
	}
	return gameID, true
}

func rankedStatsPUUID(requestPath string) (string, bool) {
	const prefix = "/lol-ranked/v1/ranked-stats/"
	if !strings.HasPrefix(requestPath, prefix) {
		return "", false
	}

	puuid := strings.TrimPrefix(requestPath, prefix)
	if puuid == "" || strings.Contains(puuid, "/") {
		return "", false
	}
	return puuid, true
}

func assetPath(requestPath string) (string, bool) {
	const prefix = "/lol-game-data/assets/"
	if !strings.HasPrefix(requestPath, prefix) {
		return "", false
	}
	assetKey := strings.TrimPrefix(requestPath, prefix)
	if assetKey == "" {
		return "", false
	}
	return assetKey, true
}

func assetContentTypeFromPath(requestPath string) string {
	contentType := mime.TypeByExtension(path.Ext(requestPath))
	if contentType == "" {
		return "image/png"
	}
	return contentType
}

func currentRankedStats(s *Scenario) []byte {
	if len(s.CurrentSummoner) == 0 {
		return nil
	}

	var current lcu.SummonerInfo
	if err := json.Unmarshal(s.CurrentSummoner, &current); err != nil {
		return nil
	}
	return s.RankedStats[current.Puuid]
}
