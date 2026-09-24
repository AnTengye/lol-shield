package mocklcu

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
)

func TestServerServesHistoryAndAssets(t *testing.T) {
	scenario, err := LoadScenario(filepath.Join("fixtures", "default"))
	if err != nil {
		t.Fatalf("LoadScenario returned error: %v", err)
	}

	server := httptest.NewServer(NewServer(scenario))
	defer server.Close()

	historyResp, err := http.Get(server.URL + "/lol-match-history/v1/products/lol/de06293d-082d-59c2-83a6-273ab88164bc/matches?begIndex=0&endIndex=8")
	if err != nil {
		t.Fatalf("history request failed: %v", err)
	}
	defer historyResp.Body.Close()

	historyBody := readAll(t, historyResp.Body)
	if historyResp.StatusCode != http.StatusOK {
		t.Fatalf("expected history status 200, got %d", historyResp.StatusCode)
	}
	if got := historyResp.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected history content type application/json, got %q", got)
	}
	if len(historyBody) == 0 {
		t.Fatalf("expected non-empty history body")
	}

	assetResp, err := http.Get(server.URL + "/lol-game-data/assets/v1/champion-icons/126.png")
	if err != nil {
		t.Fatalf("asset request failed: %v", err)
	}
	defer assetResp.Body.Close()

	assetBody := readAll(t, assetResp.Body)
	if assetResp.StatusCode != http.StatusOK {
		t.Fatalf("expected asset status 200, got %d", assetResp.StatusCode)
	}
	if got := assetResp.Header.Get("Content-Type"); got != "image/png" {
		t.Fatalf("expected asset content type image/png, got %q", got)
	}
	if len(assetBody) == 0 {
		t.Fatalf("expected non-empty asset body")
	}
}

func TestServerDoesNotTreatUnknownJSONPathAsAsset(t *testing.T) {
	scenario, err := LoadScenario(filepath.Join("fixtures", "default"))
	if err != nil {
		t.Fatalf("LoadScenario returned error: %v", err)
	}

	server := httptest.NewServer(NewServer(scenario))
	defer server.Close()

	resp, err := http.Get(server.URL + "/missing.json")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestHistoryFallbackPagination(t *testing.T) {
	scenario, err := LoadScenario(filepath.Join("fixtures", "default"))
	if err != nil {
		t.Fatal(err)
	}
	const path = "/lol-match-history/v1/products/lol/de06293d-082d-59c2-83a6-273ab88164bc/matches"
	for _, query := range []string{"?begIndex=0&endIndex=19", "?begIndex=20&endIndex=39"} {
		body, _, status := ResolveRequest(scenario, path+query)
		if status != 200 {
			t.Fatal(status, string(body))
		}
		var page lcu.GameListResp
		if err := json.Unmarshal(body, &page); err != nil {
			t.Fatal(err)
		}
		if len(page.Games.Games) > 20 || page.Games.GameCount == 0 {
			t.Fatal(page)
		}
		for i := 1; i < len(page.Games.Games); i++ {
			a, b := page.Games.Games[i-1], page.Games.Games[i]
			if a.GameCreation < b.GameCreation || a.GameId == b.GameId {
				t.Fatal("窗口必须去重并倒序")
			}
		}
	}
	_, _, status := ResolveRequest(scenario, path+"?begIndex=-1&endIndex=19")
	if status != 400 {
		t.Fatal(status)
	}
}

func readAll(t *testing.T, r io.Reader) []byte {
	t.Helper()

	body, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}

	return body
}
