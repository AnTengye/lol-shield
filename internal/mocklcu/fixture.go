package mocklcu

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
)

type Scenario struct {
	GameflowPhase        []byte
	GameflowSession      []byte
	CurrentSummoner      []byte
	Conversations        []byte
	ConversationMessages map[string][]byte
	RankedStats          map[string][]byte
	SummonersByPUUID     map[string][]byte
	MatchHistory         map[string]*HistoryFixture
	GameDetails          map[int64][]byte
	Assets               map[string][]byte
}

type HistoryFixture struct {
	Raw   []byte
	Games lcu.GameList
}

func LoadScenario(dir string) (*Scenario, error) {
	requireDefaultFixtureSet := isRepositoryDefaultScenario(dir)

	phase, err := os.ReadFile(filepath.Join(dir, "gameflow-phase.json"))
	if err != nil {
		return nil, err
	}
	if !json.Valid(phase) {
		return nil, fmt.Errorf("invalid json fixture: %s", filepath.ToSlash(filepath.Join(dir, "gameflow-phase.json")))
	}

	gameflowSession, err := loadOptionalJSONFile(filepath.Join(dir, "gameflow-session.json"))
	if err != nil {
		return nil, err
	}
	if requireDefaultFixtureSet && len(gameflowSession) == 0 {
		return nil, fmt.Errorf("missing required fixture: %s", filepath.ToSlash(filepath.Join(dir, "gameflow-session.json")))
	}

	currentSummoner, err := loadOptionalJSONFile(filepath.Join(dir, "current-summoner.json"))
	if err != nil {
		return nil, err
	}
	if requireDefaultFixtureSet && len(currentSummoner) == 0 {
		return nil, fmt.Errorf("missing required fixture: %s", filepath.ToSlash(filepath.Join(dir, "current-summoner.json")))
	}

	conversations, err := loadOptionalJSONFile(filepath.Join(dir, "conversations.json"))
	if err != nil {
		return nil, err
	}
	if requireDefaultFixtureSet && len(conversations) == 0 {
		return nil, fmt.Errorf("missing required fixture: %s", filepath.ToSlash(filepath.Join(dir, "conversations.json")))
	}

	conversationMessages, err := loadOptionalRawMap(filepath.Join(dir, "conversation-messages"), func(name string) (string, bool) {
		return jsonStemKey(name)
	})
	if err != nil {
		return nil, err
	}
	if requireDefaultFixtureSet && len(conversationMessages) == 0 {
		return nil, fmt.Errorf("missing required fixtures under: %s", filepath.ToSlash(filepath.Join(dir, "conversation-messages")))
	}

	rankedStats, err := loadOptionalRawMap(filepath.Join(dir, "ranked-stats"), func(name string) (string, bool) {
		return jsonStemKey(name)
	})
	if err != nil {
		return nil, err
	}
	if requireDefaultFixtureSet && len(rankedStats) == 0 {
		return nil, fmt.Errorf("missing required fixtures under: %s", filepath.ToSlash(filepath.Join(dir, "ranked-stats")))
	}

	summonersByPUUID, err := loadOptionalRawMap(filepath.Join(dir, "summoners", "by-puuid"), func(name string) (string, bool) {
		return jsonStemKey(name)
	})
	if err != nil {
		return nil, err
	}
	if requireDefaultFixtureSet && len(summonersByPUUID) == 0 {
		return nil, fmt.Errorf("missing required fixtures under: %s", filepath.ToSlash(filepath.Join(dir, "summoners", "by-puuid")))
	}

	matchHistory := make(map[string]*HistoryFixture)
	productRoot := filepath.Join(dir, "match-history", "products")
	if err := loadOptionalWalk(productRoot, func(path string, rel string) error {
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) != 2 {
			return fmt.Errorf("unexpected match-history fixture path: %s", filepath.ToSlash(path))
		}

		rangeKey, ok := historyRangeKey(parts[1])
		if !ok {
			return fmt.Errorf("unexpected match-history fixture filename: %s", filepath.ToSlash(path))
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var parsed lcu.GameListResp
		if err := json.Unmarshal(raw, &parsed); err != nil {
			return err
		}

		matchHistory[parts[0]+"|"+rangeKey] = &HistoryFixture{
			Raw:   raw,
			Games: parsed.Games,
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if requireDefaultFixtureSet && len(matchHistory) == 0 {
		return nil, fmt.Errorf("missing required fixtures under: %s", filepath.ToSlash(productRoot))
	}

	gameDetails, err := loadOptionalNumericRawMap(filepath.Join(dir, "match-history", "games"), func(name string) (int64, bool) {
		stem, ok := jsonStemKey(name)
		if !ok {
			return 0, false
		}
		gameID, convErr := strconv.ParseInt(stem, 10, 64)
		if convErr != nil {
			return 0, false
		}
		return gameID, true
	})
	if err != nil {
		return nil, err
	}
	if requireDefaultFixtureSet && len(gameDetails) == 0 {
		return nil, fmt.Errorf("missing required fixtures under: %s", filepath.ToSlash(filepath.Join(dir, "match-history", "games")))
	}

	assets, err := loadOptionalWalkBytes(filepath.Join(dir, "assets"))
	if err != nil {
		return nil, err
	}
	if requireDefaultFixtureSet && len(assets) == 0 {
		return nil, fmt.Errorf("missing required fixtures under: %s", filepath.ToSlash(filepath.Join(dir, "assets")))
	}

	return &Scenario{
		GameflowPhase:        phase,
		GameflowSession:      gameflowSession,
		CurrentSummoner:      currentSummoner,
		Conversations:        conversations,
		ConversationMessages: conversationMessages,
		RankedStats:          rankedStats,
		SummonersByPUUID:     summonersByPUUID,
		MatchHistory:         matchHistory,
		GameDetails:          gameDetails,
		Assets:               assets,
	}, nil
}

func loadOptionalJSONFile(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if !json.Valid(raw) {
		return nil, fmt.Errorf("invalid json fixture: %s", filepath.ToSlash(path))
	}
	return raw, nil
}

func loadOptionalRawMap(root string, keyFn func(name string) (string, bool)) (map[string][]byte, error) {
	values := make(map[string][]byte)
	err := loadOptionalWalk(root, func(path string, rel string) error {
		key, ok := keyFn(rel)
		if !ok {
			return fmt.Errorf("unexpected fixture path: %s", filepath.ToSlash(path))
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !json.Valid(raw) {
			return fmt.Errorf("invalid json fixture: %s", filepath.ToSlash(path))
		}
		values[key] = raw
		return nil
	})
	if err != nil {
		return nil, err
	}
	return values, nil
}

func loadOptionalNumericRawMap(root string, keyFn func(name string) (int64, bool)) (map[int64][]byte, error) {
	values := make(map[int64][]byte)
	err := loadOptionalWalk(root, func(path string, rel string) error {
		key, ok := keyFn(rel)
		if !ok {
			return fmt.Errorf("unexpected fixture path: %s", filepath.ToSlash(path))
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !json.Valid(raw) {
			return fmt.Errorf("invalid json fixture: %s", filepath.ToSlash(path))
		}
		values[key] = raw
		return nil
	})
	if err != nil {
		return nil, err
	}
	return values, nil
}

func loadOptionalWalkBytes(root string) (map[string][]byte, error) {
	values := make(map[string][]byte)
	err := loadOptionalWalk(root, func(path string, rel string) error {
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		values[filepath.ToSlash(rel)] = raw
		return nil
	})
	if err != nil {
		return nil, err
	}
	return values, nil
}

func jsonStemKey(name string) (string, bool) {
	clean := filepath.ToSlash(name)
	if strings.Contains(clean, "/") || filepath.Ext(clean) != ".json" {
		return "", false
	}
	return strings.TrimSuffix(clean, filepath.Ext(clean)), true
}

func isRepositoryDefaultScenario(dir string) bool {
	clean := filepath.Clean(dir)
	return filepath.Base(clean) == "default" && filepath.Base(filepath.Dir(clean)) == "fixtures"
}

func loadOptionalWalk(root string, visit func(path string, rel string) error) error {
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		return visit(path, rel)
	})
}

func historyRangeKey(name string) (string, bool) {
	base := strings.TrimSuffix(name, filepath.Ext(name))
	if !strings.HasPrefix(base, "beg-") {
		return "", false
	}
	parts := strings.Split(base, "-end-")
	if len(parts) != 2 {
		return "", false
	}
	return strings.TrimPrefix(parts[0], "beg-") + "|" + parts[1], true
}
