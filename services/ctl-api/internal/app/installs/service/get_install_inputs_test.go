package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func (s *InstallsServiceTestSuite) TestGetInstallInputsEmpty() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/inputs", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp []app.InstallInputs
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Empty(s.T(), resp)
}

func (s *InstallsServiceTestSuite) TestGetInstallInputsCurrentNoInputs() {
	install := s.createTestInstall()

	path := fmt.Sprintf("/v1/installs/%s/inputs/current", install.ID)
	rr := s.makeRequest(http.MethodGet, path, nil)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) TestMCPGetInstallInputsRedactsSensitiveValues() {
	install := s.createTestInstall()

	var inputCfg app.AppInputConfig
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Preload("AppInputs").
		Where(app.AppInputConfig{AppConfigID: s.testAppConfig.ID}).
		First(&inputCfg).Error)
	require.NotEmpty(s.T(), inputCfg.AppInputs)

	sensitiveInput := &app.AppInput{
		AppInputConfigID: inputCfg.ID,
		AppInputGroupID:  inputCfg.AppInputs[0].AppInputGroupID,
		Name:             "api_token",
		DisplayName:      "API token",
		Type:             app.AppInputTypeString,
		Sensitive:        true,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(sensitiveInput).Error)

	s.deps.Seeder.CreateInstallInputs(s.ctx, s.T(), install.ID, inputCfg.ID, map[string]*string{
		"region":    strPtr("us-west-2"),
		"api_token": strPtr("secret-value"),
	})

	ctx := cctx.SetAccountContext(context.Background(), &app.Account{
		ID: s.testAcc.ID,
		AllPermissions: permissions.Set{
			s.testOrg.ID: permissions.PermissionRead,
		},
	})
	ctx = context.WithValue(ctx, keys.OrgIDCtxKey, s.testOrg.ID)
	ctx = keys.WithTokenRole(ctx, string(app.RoleTypeOrgReadOnly))

	result, _, err := s.installsService.mcpGetInstallInputs(ctx, &mcp.CallToolRequest{}, mcpGetInstallInputsInput{
		Install: install.ID,
	})
	require.NoError(s.T(), err)
	require.Len(s.T(), result.Content, 1)

	text, ok := result.Content[0].(*mcp.TextContent)
	require.True(s.T(), ok)
	require.NotContains(s.T(), text.Text, "secret-value")

	var response mcpGetInstallInputsResult
	require.NoError(s.T(), json.Unmarshal([]byte(text.Text), &response))
	assert.Equal(s.T(), "us-west-2", response.Values["region"])
	assert.Equal(s.T(), "********", response.Values["api_token"])
}
