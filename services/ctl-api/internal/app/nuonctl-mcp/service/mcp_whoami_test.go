package service

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func TestMCPWhoami(t *testing.T) {
	acct := &app.Account{
		ID:         "acc_test",
		Email:      "dev@nuon.co",
		IsEmployee: true,
		Orgs: []*app.Org{
			{ID: "org_a", Name: "Development"},
		},
	}
	ctx := cctx.SetAccountContext(context.Background(), acct)
	ctx = context.WithValue(ctx, keys.OrgIDCtxKey, "org_a")

	result, _, err := New().mcpWhoami(ctx, &mcp.CallToolRequest{}, mcpWhoamiInput{})
	require.NoError(t, err)
	require.Len(t, result.Content, 1)

	text, ok := result.Content[0].(*mcp.TextContent)
	require.True(t, ok)
	assert.JSONEq(t, `{
		"account": {"id":"acc_test","email":"dev@nuon.co","employee":true},
		"orgs": [{"id":"org_a","name":"Development"}],
		"current_org": {"id":"org_a","name":"Development"}
	}`, text.Text)
}
