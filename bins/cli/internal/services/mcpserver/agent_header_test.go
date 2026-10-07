package mcpserver

import (
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
