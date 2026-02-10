package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// TestGetAppBranchAppConfigs tests the GET /v1/apps/:app_id/branches/:app_branch_id/configs endpoint.
func (s *AppBranchesTestSuite) TestGetAppBranchAppConfigs() {
	testCases := []struct {
		name          string
		setupFunc     func() (string, []string)
		expectedCount int
		expectedCode  int
	}{
		{
			name: "returns empty array when no configs exist",
			setupFunc: func() (string, []string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				ctx = cctx.SetOrgContext(ctx, s.testOrg)

				vcsConn := &app.VCSConnection{
					OrgID:             s.testOrg.ID,
					GithubInstallID:   "test-install-" + domains.NewVCSConnectionID(),
					GithubAccountID:   "test-account",
					GithubAccountName: "test-account-name",
				}
				err := s.service.DB.WithContext(ctx).Create(vcsConn).Error
				require.NoError(s.T(), err)

				vcsConfig := &app.ConnectedGithubVCSConfig{
					OrgID:           s.testOrg.ID,
					VCSConnectionID: vcsConn.ID,
					Repo:            "test-repo",
				}
				err = s.service.DB.WithContext(ctx).Create(vcsConfig).Error
				require.NoError(s.T(), err)

				branch := &app.AppBranch{
					OrgID:                      s.testOrg.ID,
					AppID:                      s.testApp.ID,
					Name:                       "test-branch",
					ConnectedGithubVCSConfigID: vcsConfig.ID,
				}
				err = s.service.DB.WithContext(ctx).Create(branch).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.AppBranch{}, "id = ?", branch.ID)
					s.service.DB.Unscoped().Delete(&app.ConnectedGithubVCSConfig{}, "id = ?", vcsConfig.ID)
					s.service.DB.Unscoped().Delete(&app.VCSConnection{}, "id = ?", vcsConn.ID)
				})

				return branch.ID, []string{}
			},
			expectedCount: 0,
			expectedCode:  http.StatusOK,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			branchID, _ := tc.setupFunc()

			path := "/v1/apps/" + s.testApp.ID + "/branches/" + branchID + "/configs"
			rr := s.makeGetRequest(http.MethodGet, path)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			var response []app.AppConfig
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(s.T(), err)
			require.NotNil(s.T(), response)
			require.Len(s.T(), response, tc.expectedCount)
		})
	}
}
