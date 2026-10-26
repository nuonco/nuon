package agentclient

import "testing"

func TestFromUserAgent(t *testing.T) {
	tests := []struct {
		name string
		ua   string
		want string
		ok   bool
	}{
		{name: "agent with version", ua: "claude-code/2.1.3", want: "claude-code", ok: true},
		{name: "agent with comment", ua: "Cursor/1.7.2 (darwin arm64)", want: "cursor", ok: true},
		{name: "agent without version", ua: "opencode", want: "opencode", ok: true},
		{name: "empty", ua: "", ok: false},
		{name: "browser", ua: "Mozilla/5.0 (Macintosh) AppleWebKit/537.36", ok: false},
		{name: "http library", ua: "node-fetch/1.0", ok: false},
		{name: "go default", ua: "Go-http-client/1.1", ok: false},
		{name: "nuon cli", ua: "nuon-cli/0.19.1263 (cursor)", ok: false},
		{name: "not a slug", ua: "my_agent/1.0", ok: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := FromUserAgent(tt.ua)
			if ok != tt.ok || got != tt.want {
				t.Fatalf("FromUserAgent(%q) = %q, %v; want %q, %v", tt.ua, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	if got := Normalize(" Cursor "); got != "cursor" {
		t.Fatalf("Normalize = %q, want cursor", got)
	}
}
