package mocklcu

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
)

type Scenario struct {
	GameflowPhase []byte
	MatchHistory  map[string]*HistoryFixture
}

type HistoryFixture struct {
	Raw   []byte
	Games lcu.GameList
}

func LoadScenario(dir string) (*Scenario, error) {
	phase, err := os.ReadFile(filepath.Join(dir, "gameflow-phase.json"))
	if err != nil {
		return nil, err
	}

	matchHistory := make(map[string]*HistoryFixture)
	productRoot := filepath.Join(dir, "match-history", "products")
	if err := filepath.WalkDir(productRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		rel, err := filepath.Rel(productRoot, path)
		if err != nil {
			return err
		}

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

	return &Scenario{
		GameflowPhase: phase,
		MatchHistory:  matchHistory,
	}, nil
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
