package lcu

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const onJSONAPIPrefixLen = len(`[8,"OnJsonApiEvent",`)

type wsEvent struct {
	Data      interface{} `json:"data"`
	EventType string      `json:"eventType"`
	URI       string      `json:"uri"`
}

func parseFlowEvent(msgType int, message []byte) (string, bool) {
	message = bytes.TrimSpace(message)
	if msgType != websocket.TextMessage || len(message) <= onJSONAPIPrefixLen+1 {
		return "", false
	}
	ev := wsEvent{}
	if err := json.Unmarshal(message[onJSONAPIPrefixLen:len(message)-1], &ev); err != nil {
		return "", false
	}
	if ev.URI != "/lol-gameflow/v1/gameflow-phase" {
		return "", false
	}
	val, ok := ev.Data.(string)
	if !ok {
		return "", false
	}
	return strings.Trim(val, `"`), true
}

func (a *LegacyAdapter) WatchFlow(
	ctx context.Context,
	port int,
	token string,
	onFlow func(flow string),
	onError func(err error),
) {
	dialer := websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	rawURL := fmt.Sprintf("wss://127.0.0.1:%d/", port)
	u, _ := url.Parse(rawURL)
	header := http.Header{}
	authSecret := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("riot:%s", token)))
	header.Set("Authorization", "Basic "+authSecret)

	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		onError(err)
		return
	}
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte(`[5, "OnJsonApiEvent"]`)); err != nil {
		onError(err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_ = conn.SetReadDeadline(time.Now().Add(15 * time.Second))
		msgType, message, err := conn.ReadMessage()
		if err != nil {
			onError(err)
			return
		}
		flow, ok := parseFlowEvent(msgType, message)
		if ok {
			onFlow(flow)
		}
	}
}
