package client

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/AnTengye/lol-shield/internal/core/lcuapi"
	"github.com/AnTengye/lol-shield/internal/pkg/lcu"
)

func TestGetIDListFromConversationMsgList(t *testing.T) {
	const puuid = "de06293d-082d-59c2-83a6-273ab88164bc"
	joined := func(pid string) lcu.ConversationMsg {
		return lcu.ConversationMsg{
			Type:           lcu.ConversationMsgTypeSystem,
			Body:           lcu.JoinedRoomMsg,
			FromSummonerId: 42,
			FromPid:        pid,
		}
	}
	validUser := []lcu.UserId{{SummonerId: 42, Puuid: puuid}}
	tests := []struct {
		name     string
		messages []lcu.ConversationMsg
		want     []lcu.UserId
	}{
		{name: "no messages", want: []lcu.UserId{}},
		{name: "empty player ID", messages: []lcu.ConversationMsg{joined("")}, want: []lcu.UserId{}},
		{name: "short player ID", messages: []lcu.ConversationMsg{joined("system")}, want: []lcu.UserId{}},
		{name: "one byte short", messages: []lcu.ConversationMsg{joined(puuid[:len(puuid)-1])}, want: []lcu.UserId{}},
		{name: "bare PUUID", messages: []lcu.ConversationMsg{joined(puuid)}, want: validUser},
		{name: "chat address suffix", messages: []lcu.ConversationMsg{joined(puuid + "@pvp.net")}, want: validUser},
		{
			name: "mixed room messages",
			messages: []lcu.ConversationMsg{
				joined(""),
				{Type: "chat", Body: lcu.JoinedRoomMsg, FromPid: puuid},
				{Type: lcu.ConversationMsgTypeSystem, Body: "left_room", FromPid: puuid},
				joined(strings.Repeat("x", 35)),
				joined(puuid + "@pvp.net"),
			},
			want: validUser,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := getIDListFromConversationMsgList(test.messages); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("players = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestGetTeamUsersSkipsJoinedRoomMessagesWithoutPlayerIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/lol-chat/v1/conversations":
			_, _ = w.Write([]byte(`[{"id":"champ-select","type":"championSelect"}]`))
		case "/lol-chat/v1/conversations/champ-select/messages":
			_, _ = w.Write([]byte(`[
				{"type":"system","body":"joined_room","fromSummonerId":0},
				{"type":"system","body":"joined_room","fromPid":"","fromSummonerId":1},
				{"type":"system","body":"joined_room","fromPid":"short","fromSummonerId":2},
				{"type":"system","body":"joined_room","fromPid":"de06293d-082d-59c2-83a6-273ab88164bc@pvp.net","fromSummonerId":42}
			]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	conversationID, users, err := getTeamUsers(lcuapi.NewHTTPService(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	want := []lcu.UserId{{SummonerId: 42, Puuid: "de06293d-082d-59c2-83a6-273ab88164bc"}}
	if conversationID != "champ-select" || !reflect.DeepEqual(users, want) {
		t.Fatalf("conversation = %q, players = %#v, want champ-select and %#v", conversationID, users, want)
	}
}
