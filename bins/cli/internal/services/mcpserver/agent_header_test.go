package mcpserver

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/agentclient"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestAuthRoundTripperSetsAgentHeader(t *testing.T) {
	var gotAgent, gotCommand, gotUserAgent string
	transport := &authRoundTripper{
		token:     "token",
		orgID:     "org",
		agent:     agentclient.Cursor,
		userAgent: agentclient.UserAgent("1.2.3", agentclient.Cursor),
		command:   "nuon agents mcp",
		base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotAgent = req.Header.Get(agentclient.Header)
			gotCommand = req.Header.Get(agentclient.CommandHeader)
			gotUserAgent = req.Header.Get("User-Agent")
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	req, err := http.NewRequest(http.MethodPost, "https://mcp.example.com/mcp", nil)
	require.NoError(t, err)
	_, err = transport.RoundTrip(req)
	require.NoError(t, err)
	require.Equal(t, agentclient.Cursor, gotAgent)
	require.Equal(t, "nuon agents mcp", gotCommand)
	require.Equal(t, "nuon-cli/1.2.3 (cursor)", gotUserAgent)
}

func TestAuthRoundTripperOrgIDIsPerRequest(t *testing.T) {
	var gotOrg, gotBody string
	transport := &authRoundTripper{
		orgID: "org_config",
		base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			gotOrg = req.Header.Get("X-Nuon-Org-ID")
			body, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			gotBody = string(body)
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	send := func(body string) {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, "https://mcp.example.com/mcp", bytes.NewReader([]byte(body)))
		require.NoError(t, err)
		_, err = transport.RoundTrip(req)
		require.NoError(t, err)
	}

	send(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_apps","arguments":{"org_id":"org_a","app":"app_1"}}}`)
	require.Equal(t, "org_a", gotOrg)
	require.NotContains(t, gotBody, "org_id")
	require.Contains(t, gotBody, "app_1")

	send(`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_apps","arguments":{"org_id":"org_b"}}}`)
	require.Equal(t, "org_b", gotOrg)
	require.NotContains(t, gotBody, "org_a")

	send(`{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"whoami","arguments":{}}}`)
	require.Equal(t, "org_config", gotOrg)
}

func TestAuthRoundTripperOmitsEmptyAgent(t *testing.T) {
	var present bool
	transport := &authRoundTripper{
		base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			_, present = req.Header[http.CanonicalHeaderKey(agentclient.Header)]
			return &http.Response{
				StatusCode: http.StatusNoContent,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
			}, nil
		}),
	}

	req, err := http.NewRequest(http.MethodPost, "https://mcp.example.com/mcp", nil)
	require.NoError(t, err)
	_, err = transport.RoundTrip(req)
	require.NoError(t, err)
	require.False(t, present)
}
