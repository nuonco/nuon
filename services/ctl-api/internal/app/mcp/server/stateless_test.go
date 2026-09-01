package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type stubMCPService struct{}

func (stubMCPService) RegisterPublicRoutes(*gin.Engine) error         { return nil }
func (stubMCPService) RegisterRunnerRoutes(*gin.Engine) error         { return nil }
func (stubMCPService) RegisterAuthRoutes(*gin.Engine) error           { return nil }
func (stubMCPService) RegisterInternalRoutes(*gin.Engine) error       { return nil }
func (stubMCPService) RegisterAdminDashboardRoutes(*gin.Engine) error { return nil }
func (stubMCPService) RegisterSlackRoutes(*gin.Engine) error          { return nil }

func (stubMCPService) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, api.MCPReadTool(
		"echo_org",
		"Echo org",
		"returns the org resolved for the calling request",
	), func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return api.MCPJSONResult(map[string]string{"org_id": keys.OrgIDFromContext(ctx)})
	})
}

func withMCPAuth(ctx context.Context, orgID, accountID string) context.Context {
	ctx = context.WithValue(ctx, keys.AccountIDCtxKey, accountID)
	ctx = context.WithValue(ctx, keys.OrgIDCtxKey, orgID)
	return ctx
}

// newStatelessTestServer serves the real MCP handler behind a stub auth layer
// that injects the same context the production middleware would.
func newStatelessTestServer(t *testing.T, orgID string) *httptest.Server {
	t.Helper()

	s := &Server{services: []api.Service{stubMCPService{}}, schemaCache: mcp.NewSchemaCache()}

	handler := s.newMCPHandler()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := withMCPAuth(r.Context(), orgID, "acc_test")
		handler.ServeHTTP(w, r.WithContext(ctx))
	}))
	t.Cleanup(ts.Close)

	return ts
}

// TestStatelessRoundTrip drives the server with a real MCP client over the same
// transport the CLI proxy uses, so an SDK upgrade that breaks the handshake,
// tool listing, or tool calls fails here.
func TestStatelessRoundTrip(t *testing.T) {
	ts := newStatelessTestServer(t, "org_a")
	ctx := context.Background()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: ts.URL}, nil)
	require.NoError(t, err)
	defer session.Close()

	tools, err := session.ListTools(ctx, nil)
	require.NoError(t, err)
	require.Len(t, tools.Tools, 1)
	assert.Equal(t, "echo_org", tools.Tools[0].Name)

	res, err := session.CallTool(ctx, &mcp.CallToolParams{Name: "echo_org"})
	require.NoError(t, err)
	require.Len(t, res.Content, 1)
	text, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.JSONEq(t, `{"org_id":"org_a"}`, text.Text)
}

// TestStatelessIssuesNoSessionID guards the stateless contract: the server must
// not hand out an Mcp-Session-Id, because no replica retains session state.
func TestStatelessIssuesNoSessionID(t *testing.T) {
	ts := newStatelessTestServer(t, "org_a")

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{` +
		`"protocolVersion":"2025-06-18","capabilities":{},` +
		`"clientInfo":{"name":"test-client","version":"0.0.1"}}}`

	req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, resp.Header.Get("Mcp-Session-Id"), "stateless server must not issue a session id")
}

// TestStatelessRejectsGET documents that the server-initiated SSE stream is gone
// in stateless mode; clients must not depend on it.
func TestStatelessRejectsGET(t *testing.T) {
	ts := newStatelessTestServer(t, "org_a")

	req, err := http.NewRequest(http.MethodGet, ts.URL, http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}
