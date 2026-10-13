package nuon

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSetAgentClientHeader(t *testing.T) {
	c, err := New(WithURL("https://api.example.com"), WithAuthToken("token"))
	if err != nil {
		t.Fatal(err)
	}

	var gotAgent, gotCommand, gotUserAgent string
	c.appTransport.transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotAgent = req.Header.Get("X-Nuon-Agent")
		gotCommand = req.Header.Get("X-Nuon-Command")
		gotUserAgent = req.Header.Get("User-Agent")
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		}, nil
	})

	c.SetAgentClient("cursor")
	c.SetUserAgent("nuon-cli/1.2.3 (cursor)")
	c.SetCommand("nuon orgs list")
	req, err := http.NewRequest(http.MethodGet, "https://api.example.com/v1/x", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.appTransport.RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if gotAgent != "cursor" {
		t.Fatalf("X-Nuon-Agent = %q, want cursor", gotAgent)
	}
	if gotCommand != "nuon orgs list" {
		t.Fatalf("X-Nuon-Command = %q", gotCommand)
	}
	if gotUserAgent != "nuon-cli/1.2.3 (cursor)" {
		t.Fatalf("User-Agent = %q", gotUserAgent)
	}

	c.SetAgentClient("")
	c.SetUserAgent("")
	c.SetCommand("")
	gotAgent, gotCommand, gotUserAgent = "unset", "unset", "unset"
	req, err = http.NewRequest(http.MethodGet, "https://api.example.com/v1/x", nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.appTransport.RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if gotAgent != "" || gotCommand != "" || gotUserAgent != "" {
		t.Fatalf("headers = agent %q command %q user-agent %q, want empty", gotAgent, gotCommand, gotUserAgent)
	}
}
