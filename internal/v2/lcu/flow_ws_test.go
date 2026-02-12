package lcu

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/websocket"
)

func TestParseFlowEventReplay(t *testing.T) {
	payloadPath := filepath.Join("fixtures", "on_json_api_event_gameflow.txt")
	raw, err := os.ReadFile(payloadPath)
	if err != nil {
		t.Fatalf("read fixture failed: %v", err)
	}
	flow, ok := parseFlowEvent(websocket.TextMessage, raw)
	if !ok {
		t.Fatalf("expected parse success")
	}
	if flow != "InProgress" {
		t.Fatalf("expected InProgress, got %s", flow)
	}
}

func TestParseFlowEventInvalid(t *testing.T) {
	_, ok := parseFlowEvent(websocket.TextMessage, []byte(`{"bad":"payload"}`))
	if ok {
		t.Fatalf("expected parse failure")
	}
}
