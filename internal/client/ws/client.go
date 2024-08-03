package ws

import (
	"fmt"
	"net/http"
	"time"

	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	TickerTime = 5 * time.Second
	Timeout    = 3 * time.Second
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebClient struct {
	uid  string
	conn *websocket.Conn
	send chan []byte
}

// getuid
func (c *WebClient) GetUid() string {
	return c.uid
}

func (c *WebClient) healthCheck() {
	ticker := time.NewTicker(TickerTime)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(Timeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				syslog.L.Infof("%s-ui已断开", c.uid)
				return
			}
		}
	}
}
func (c *WebClient) Write(data []byte) {
	_ = c.conn.SetWriteDeadline(time.Now().Add(Timeout))
	w, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return
	}
	_, err = w.Write(data)
	if err := w.Close(); err != nil {
		return
	}
}
func (c *WebClient) read() {
	defer func() {
		c.conn.Close()
	}()
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				syslog.L.Errorf("close err:%v", err)
			}
			break
		}
		fmt.Println(msg)
	}
}

func ServerWebsocket(w http.ResponseWriter, r *http.Request) *WebClient {
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		syslog.L.Errorf("close err:%v", err)
		return nil
	}
	u := uuid.New()
	c := &WebClient{uid: u.String(), conn: conn, send: make(chan []byte, 256)}

	go c.healthCheck()
	return c
}
