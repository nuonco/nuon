package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *AppConfigsTestSuite) TestGetAppConfigDiffLoadsIntermediateConfigBlob() {
	intermediate, err := json.Marshal(config.AppConfig{
		Branch: &config.AppBranchConfig{Name: "main"},
	})
	require.NoError(s.T(), err)

	ctx := cctx.SetAccountContext(context.Background(), s.testAcc)
	ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
	appConfig, err := s.service.AppsService.createAppConfig(ctx, s.testOrg.ID, s.testApp.ID, &CreateAppConfigRequest{
		IntermediateConfigJSON: string(intermediate),
	})
	require.NoError(s.T(), err)

	path := fmt.Sprintf("/v1/apps/%s/configs/%s/diff", s.testApp.ID, appConfig.ID)
	rr := s.makeGetRequest(http.MethodGet, path)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var response AppConfigDiffResponse
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &response))
	require.Equal(s.T(), appConfig.ID, response.ConfigID)
}
