package mcpserver

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// fakeAPI embeds the client interface so only the methods under test need
// implementations; anything else panics.
type fakeAPI struct {
	nuon.Client
	user     *models.AppAccount
	org      *models.AppOrg
	apps     []*models.AppApp
	appComps map[string][]*models.AppComponent
}

func (f *fakeAPI) GetApp(ctx context.Context, appID string) (*models.AppApp, error) {
	for _, a := range f.apps {
		if a.ID == appID || a.Name == appID {
			return a, nil
		}
	}
	return nil, &fakeNotFound{}
}

func (f *fakeAPI) GetAppComponents(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppComponent, bool, error) {
	return f.appComps[appID], false, nil
}

type fakeNotFound struct{}

func (e *fakeNotFound) Error() string { return "not found" }

func (f *fakeAPI) GetCurrentUser(ctx context.Context) (*models.AppAccount, error) {
	return f.user, nil
}

func (f *fakeAPI) GetOrg(ctx context.Context) (*models.AppOrg, error) {
	return f.org, nil
}

func (f *fakeAPI) GetApps(ctx context.Context, query *models.GetPaginatedQuery) ([]*models.AppApp, bool, error) {
	return f.apps, false, nil
}

func connect(t *testing.T, api nuon.Client, allowWrites bool, cfg ...*config.Config) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()

	var c *config.Config
	if len(cfg) > 0 {
		c = cfg[0]
	}
	srv := New(api, c).buildServer(allowWrites)
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
	api := &fakeAPI{}

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
	api := &fakeAPI{
		user: &models.AppAccount{ID: "acc1", Email: "amit@nuon.co"},
		org:  &models.AppOrg{ID: "org1", Name: "test-org"},
	}

	cs := connect(t, api, false)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "whoami"})
	require.NoError(t, err)
	require.False(t, res.IsError)
	require.Len(t, res.Content, 1)

	text := res.Content[0].(*mcp.TextContent).Text
	require.Contains(t, text, "amit@nuon.co")
	require.Contains(t, text, "test-org")
}

func TestSelectedAppContext(t *testing.T) {
	api := &fakeAPI{
		user: &models.AppAccount{ID: "acc1", Email: "amit@nuon.co"},
		org:  &models.AppOrg{ID: "org1", Name: "test-org"},
		apps: []*models.AppApp{{ID: "app1", Name: "uptime-monitor"}},
		appComps: map[string][]*models.AppComponent{
			"app1": {{ID: "cmp1", Name: "api_image", AppID: "app1"}},
		},
	}
	cfg := &config.Config{AppID: "app1"}

	cs := connect(t, api, false, cfg)

	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "whoami"})
	require.NoError(t, err)
	text := res.Content[0].(*mcp.TextContent).Text
	require.Contains(t, text, "selected_app")
	require.Contains(t, text, "uptime-monitor")

	res, err = cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "list_components"})
	require.NoError(t, err)
	require.False(t, res.IsError)
	text = res.Content[0].(*mcp.TextContent).Text
	require.Contains(t, text, "api_image", "should default to selected app's components")
}

func TestListApps(t *testing.T) {
	api := &fakeAPI{
		apps: []*models.AppApp{{ID: "app1", Name: "uptime-monitor"}},
	}

	cs := connect(t, api, false)
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: "list_apps"})
	require.NoError(t, err)
	require.False(t, res.IsError)

	text := res.Content[0].(*mcp.TextContent).Text
	require.Contains(t, text, "uptime-monitor")
	require.NotContains(t, text, "app_sandbox_config", "summaries must stay trimmed")
}
