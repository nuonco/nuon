package mcpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
)

func statelessUpstream(t *testing.T) *mcp.ClientSession {
	t.Helper()

	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		server := mcp.NewServer(&mcp.Implementation{Name: "fake-ctl-api", Version: "0"}, nil)
		mcp.AddTool(server, &mcp.Tool{
			Name:        "whoami",
			Description: "Get current user info.",
		}, func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: `{"user":"test@nuon.co"}`}},
			}, nil, nil
		})
		mcp.AddTool(server, &mcp.Tool{
			Name:        "deploy_component",
			Description: "WRITE OPERATION: Deploy a component.",
		}, func(context.Context, *mcp.CallToolRequest, struct{}) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: `{"status":"ok"}`}},
			}, nil, nil
		})
		return server
	}, &mcp.StreamableHTTPOptions{Stateless: true})

	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	client := mcp.NewClient(&mcp.Implementation{Name: "nuon-cli-proxy", Version: "0"}, nil)
	session, err := client.Connect(context.Background(), &mcp.StreamableClientTransport{Endpoint: ts.URL}, nil)
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })

	return session
}

func TestProxyOverStatelessUpstream(t *testing.T) {
	upstream := statelessUpstream(t)
	ctx := context.Background()

	readOnly := connectProxy(t, upstream, false)

	tools, err := readOnly.ListTools(ctx, nil)
	require.NoError(t, err)
	names := map[string]bool{}
	for _, tool := range tools.Tools {
		names[tool.Name] = true
	}
	require.True(t, names["whoami"], "read tool missing over stateless upstream")
	require.False(t, names["deploy_component"], "write tool exposed without --allow-writes")

	res, err := readOnly.CallTool(ctx, &mcp.CallToolParams{Name: "whoami"})
	require.NoError(t, err)
	require.Len(t, res.Content, 1)
	text, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	require.JSONEq(t, `{"user":"test@nuon.co"}`, text.Text)

	for range 3 {
		_, err := readOnly.CallTool(ctx, &mcp.CallToolParams{Name: "whoami"})
		require.NoError(t, err)
	}

	withWrites := connectProxy(t, upstream, true)
	tools, err = withWrites.ListTools(ctx, nil)
	require.NoError(t, err)
	names = map[string]bool{}
	for _, tool := range tools.Tools {
		names[tool.Name] = true
	}
	require.True(t, names["deploy_component"], "write tool should be exposed with --allow-writes")
}
