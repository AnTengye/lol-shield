package mocklcu

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServerServesHistoryAndAssets(t *testing.T) {
	scenario := &Scenario{
		MatchHistory: map[string]*HistoryFixture{
			"de06293d-082d-59c2-83a6-273ab88164bc|0|8": {
				Raw: []byte(`{"games":{"gameCount":9}}`),
			},
		},
		Assets: map[string][]byte{
			"placeholder.png": []byte("placeholder-bytes"),
		},
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
	if string(historyBody) != `{"games":{"gameCount":9}}` {
		t.Fatalf("unexpected history body: %s", string(historyBody))
	}

	assetResp, err := http.Get(server.URL + "/lol-game-data/assets/v1/champion-icons/266.png")
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
	if string(assetBody) != "placeholder-bytes" {
		t.Fatalf("unexpected asset body: %q", string(assetBody))
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
