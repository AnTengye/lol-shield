package mocklcu

import (
	"mime"
	"net/http"
	"path"
)

func NewServer(s *Scenario) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, contentType, status := ResolveRequest(s, r.URL.RequestURI())
		if status == http.StatusNotFound && looksLikeAssetPath(r.URL.Path) {
			if placeholder, ok := s.Assets["placeholder.png"]; ok {
				writeResponse(w, placeholder, assetContentType(r.URL.Path), http.StatusOK)
				return
			}
		}

		writeResponse(w, body, contentType, status)
	})
}

func looksLikeAssetPath(requestPath string) bool {
	return path.Ext(requestPath) != "" && path.Dir(requestPath) != "." && path.Clean(requestPath) != "." &&
		len(requestPath) > len("/lol-game-data/assets/") && requestPath[:len("/lol-game-data/assets/")] == "/lol-game-data/assets/"
}

func assetContentType(requestPath string) string {
	contentType := mime.TypeByExtension(path.Ext(requestPath))
	if contentType == "" {
		return "image/png"
	}
	return contentType
}

func writeResponse(w http.ResponseWriter, body []byte, contentType string, status int) {
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
