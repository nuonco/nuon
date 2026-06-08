package service

import (
	"encoding/json"
	"testing"
)

// TestSlackInnerEventChannelParse covers Slack's polymorphic `channel`
// shape: object for channel_rename, bare string for channel_archive /
// channel_left. Both must round-trip through the same field reliably.
func TestSlackInnerEventChannelParse(t *testing.T) {
	cases := []struct {
		name     string
		jsonBody string
		wantID   string
		wantName string
	}{
		{
			name:     "channel_rename object shape",
			jsonBody: `{"type":"channel_rename","channel":{"id":"C1","name":"new-name","created":123}}`,
			wantID:   "C1",
			wantName: "new-name",
		},
		{
			name:     "channel_archive bare string",
			jsonBody: `{"type":"channel_archive","channel":"C2","user":"U1"}`,
			wantID:   "C2",
			wantName: "",
		},
		{
			name:     "channel_left bare string",
			jsonBody: `{"type":"channel_left","channel":"C3"}`,
			wantID:   "C3",
		},
		{
			name:     "member_joined_channel bare string",
			jsonBody: `{"type":"member_joined_channel","channel":"C4","user":"UBOT"}`,
			wantID:   "C4",
		},
		{
			name:     "missing channel returns empty",
			jsonBody: `{"type":"channel_rename"}`,
			wantID:   "",
		},
		{
			name:     "garbage channel field returns empty",
			jsonBody: `{"type":"channel_rename","channel":42}`,
			wantID:   "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var ev slackInnerEvent
			if err := json.Unmarshal([]byte(tc.jsonBody), &ev); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			ref := ev.parseChannelRef()
			if ref.ID != tc.wantID {
				t.Fatalf("id: got %q want %q", ref.ID, tc.wantID)
			}
			if ref.Name != tc.wantName {
				t.Fatalf("name: got %q want %q", ref.Name, tc.wantName)
			}
			if got := ev.channelIDFromEvent(); got != tc.wantID {
				t.Fatalf("channelIDFromEvent: got %q want %q", got, tc.wantID)
			}
		})
	}
}

// TestSlackInnerEventUserParse confirms the `user` field decodes off
// member_joined_channel events — it's how we tell our bot's own join apart
// from a human teammate joining (welcomeChannelOnBotJoin compares it to
// SlackInstallation.BotUserID).
func TestSlackInnerEventUserParse(t *testing.T) {
	var ev slackInnerEvent
	body := `{"type":"member_joined_channel","channel":"C4","user":"UBOT"}`
	if err := json.Unmarshal([]byte(body), &ev); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ev.User != "UBOT" {
		t.Fatalf("user: got %q want %q", ev.User, "UBOT")
	}
}
