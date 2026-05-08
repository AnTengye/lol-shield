package lcuapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPServiceListGamesByUIDUsesConfiguredBaseURL(t *testing.T) {
	var requestedPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.RequestURI()
		_, _ = w.Write([]byte(`{"games":{"gameCount":9,"gameIndexBegin":0,"gameIndexEnd":8,"games":[]}}`))
	}))
	defer srv.Close()

	svc := NewHTTPService(srv.URL)
	_, err := svc.ListGamesByUID("de06293d-082d-59c2-83a6-273ab88164bc", 0, 9)
	if err != nil {
		t.Fatalf("ListGamesByUID returned error: %v", err)
	}
	if requestedPath != "/lol-match-history/v1/products/lol/de06293d-082d-59c2-83a6-273ab88164bc/matches?begIndex=0&endIndex=8" {
		t.Fatalf("unexpected request path: %s", requestedPath)
	}
}
