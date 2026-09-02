package mcpserver

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/bins/cli/internal/config"
)

func fakeUpstream(t *testing.T) (*mcp.Server, *mcp.ClientSession) {
	t.Helper()

	server := mcp.NewServer(&mcp.Implementation{Name: "fake-upstream", Version: "0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "whoami",
		Description: "Get current user info.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: `{"user":"test@nuon.co","org":"test-org"}`}},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_apps",
		Description: "List all apps.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: `[{"id":"app1","name":"my-app"}]`}},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "deploy_component",
		Description: "WRITE OPERATION: Deploy a component.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: `{"status":"ok"}`}},
		}, nil, nil
	})

	st, ct := mcp.NewInMemoryTransports()
	_, err := server.Connect(context.Background(), st, nil)
	require.NoError(t, err)

	client := mcp.NewClient(&mcp.Implementation{Name: "test-connector", Version: "0"}, nil)
	session, err := client.Connect(context.Background(), ct, nil)
	require.NoError(t, err)
	t.Cleanup(func() { session.Close() })

	return server, session
}

func connectProxy(t *testing.T, upstream *mcp.ClientSession, allowWrites bool, opts ...Option) *mcp.ClientSession {
	t.Helper()

	svc := New(&config.Config{}, allowWrites, opts...)

	server, err := svc.buildProxyServer(context.Background(), upstream)
	require.NoError(t, err)

	st, ct := mcp.NewInMemoryTransports()
	_, err = server.Connect(context.Background(), st, nil)
	require.NoError(t, err)

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	cs, err := client.Connect(context.Background(), ct, nil)
	require.NoError(t, err)
	t.Cleanup(func() { cs.Close() })

	return cs
}

func TestProxyToolDiscovery(t *testing.T) {
	_, upstream := fakeUpstream(t)

	readOnly := connectProxy(t, upstream, false)
	res, err := readOnly.ListTools(context.Background(), nil)
	require.NoError(t, err)

	names := map[string]bool{}
	for _, tool := range res.Tools {
		names[tool.Name] = true
	}
	require.True(t, names["whoami"], "missing read tool whoami")
	require.True(t, names["list_apps"], "missing read tool list_apps")
	require.False(t, names["deploy_component"], "write tool exposed without --allow-writes")

	withWrites := connectProxy(t, upstream, true)
	res, err = withWrites.ListTools(context.Background(), nil)
	require.NoError(t, err)
	names = map[string]bool{}
	for _, tool := range res.Tools {
		names[tool.Name] = true
	}
	require.True(t, names["deploy_component"], "write tool should be exposed with --allow-writes")
}

func TestProxyToolForwarding(t *testing.T) {
	_, upstream := fakeUpstream(t)

	cs := connectProxy(t, upstream, false)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "whoami"})
	require.NoError(t, err)
	require.False(t, res.IsError)
	require.Len(t, res.Content, 1)

	text := res.Content[0].(*mcp.TextContent).Text
	require.Contains(t, text, "test@nuon.co")
	require.Contains(t, text, "test-org")
}

func TestProxyListApps(t *testing.T) {
	_, upstream := fakeUpstream(t)

	cs := connectProxy(t, upstream, false)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "list_apps"})
	require.NoError(t, err)
	require.False(t, res.IsError)

	text := res.Content[0].(*mcp.TextContent).Text
	require.Contains(t, text, "my-app")
}

func TestEndpointAndNameFromAPIURL(t *testing.T) {
	tests := []struct {
		apiURL   string
		endpoint string
		name     string
	}{
		{
			apiURL:   "https://api.nuon.co",
			endpoint: "https://mcp.nuon.co/mcp",
			name:     "nuon",
		},
		{
			apiURL:   "https://api.stage.nuon.co/",
			endpoint: "https://mcp.stage.nuon.co/mcp",
			name:     "nuon-stage",
		},
		{
			apiURL:   "http://localhost:8081",
			endpoint: "http://localhost:8088/mcp",
			name:     "nuon-local",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.endpoint, EndpointFromAPIURL(test.apiURL))
			require.Equal(t, test.name, NameFromAPIURL(test.apiURL))
		})
	}
}

func TestProxyServerNameOverride(t *testing.T) {
	_, upstream := fakeUpstream(t)

	session := connectProxy(t, upstream, false, WithName("custom-server"))
	require.Equal(t, "custom-server", session.InitializeResult().ServerInfo.Name)
}

func TestEndpointOverride(t *testing.T) {
	svc := New(
		&config.Config{APIURL: "https://api.stage.nuon.co"},
		false,
		WithEndpoint("https://example.com/mcp"),
	)

	require.Equal(t, "https://example.com/mcp", svc.mcpEndpoint())
}
