package mcpserver

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func connect(t *testing.T, api nuon.Client, allowWrites bool) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	srv := New(api, nil, allowWrites).buildServer()
	st, ct := mcp.NewInMemoryTransports()

	_, err := srv.Connect(ctx, st, nil)
	require.NoError(t, err)

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	require.NoError(t, err)
	t.Cleanup(func() { cs.Close() })

	return cs
}

func TestToolRegistration(t *testing.T) {
	ctrl := gomock.NewController(t)
	api := nuon.NewMockClient(ctrl)

	readOnly := connect(t, api, false)
	res, err := readOnly.ListTools(context.Background(), nil)
	require.NoError(t, err)

	names := map[string]bool{}
	for _, tool := range res.Tools {
		names[tool.Name] = true
	}
	for _, want := range []string{"whoami", "list_apps", "get_app", "list_installs", "get_install", "list_install_components", "list_components"} {
		require.True(t, names[want], "missing read tool %s", want)
	}
	require.False(t, names["create_install"], "write tool exposed without --allow-writes")
	require.False(t, names["deploy_component"], "write tool exposed without --allow-writes")

	writes := connect(t, api, true)
	res, err = writes.ListTools(context.Background(), nil)
	require.NoError(t, err)
	names = map[string]bool{}
	for _, tool := range res.Tools {
		names[tool.Name] = true
	}
	require.True(t, names["create_install"])
	require.True(t, names["deploy_component"])
}

func TestWhoami(t *testing.T) {
	ctrl := gomock.NewController(t)
	api := nuon.NewMockClient(ctrl)
	api.EXPECT().GetCurrentUser(gomock.Any()).Return(&models.AppAccount{ID: "acc1", Email: "amit@nuon.co"}, nil)
	api.EXPECT().GetOrg(gomock.Any()).Return(&models.AppOrg{ID: "org1", Name: "test-org"}, nil)

	cs := connect(t, api, false)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "whoami"})
	require.NoError(t, err)
	require.False(t, res.IsError)
	require.Len(t, res.Content, 1)

	text := res.Content[0].(*mcp.TextContent).Text
	require.Contains(t, text, "amit@nuon.co")
	require.Contains(t, text, "test-org")
}

func TestListApps(t *testing.T) {
	ctrl := gomock.NewController(t)
	api := nuon.NewMockClient(ctrl)
	api.EXPECT().GetApps(gomock.Any(), gomock.Any()).Return([]*models.AppApp{
		{ID: "app1", Name: "uptime-monitor"},
	}, false, nil)

	cs := connect(t, api, false)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "list_apps"})
	require.NoError(t, err)
	require.False(t, res.IsError)

	text := res.Content[0].(*mcp.TextContent).Text
	require.Contains(t, text, "uptime-monitor")
	require.NotContains(t, text, "app_sandbox_config", "summaries must stay trimmed")
}
