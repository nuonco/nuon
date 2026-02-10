package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// TestCreateAppBranch tests the POST /v1/apps/:app_id/branches endpoint.
func (s *AppBranchesTestSuite) TestCreateAppBranch() {
	testCases := []struct {
		name         string
		setupFunc    func() *CreateAppBranchRequest
		expectedCode int
		validateFunc func(*app.AppBranch)
	}{
		{
			name: "validation error when name is missing",
			setupFunc: func() *CreateAppBranchRequest {
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

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.ConnectedGithubVCSConfig{}, "id = ?", vcsConfig.ID)
					s.service.DB.Unscoped().Delete(&app.VCSConnection{}, "id = ?", vcsConn.ID)
				})

				return &CreateAppBranchRequest{
					Name:                       "",
					ConnectedGithubVCSConfigID: vcsConfig.ID,
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "validation error when connected_github_vcs_config_id is missing",
			setupFunc: func() *CreateAppBranchRequest {
				return &CreateAppBranchRequest{
					Name:                       "test-branch",
					ConnectedGithubVCSConfigID: "",
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "successfully creates branch with valid input",
			setupFunc: func() *CreateAppBranchRequest {
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

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(&app.AppBranch{}, "app_id = ?", s.testApp.ID)
					s.service.DB.Unscoped().Delete(&app.ConnectedGithubVCSConfig{}, "id = ?", vcsConfig.ID)
					s.service.DB.Unscoped().Delete(&app.VCSConnection{}, "id = ?", vcsConn.ID)
				})

				return &CreateAppBranchRequest{
					Name:                       "main-branch",
					ConnectedGithubVCSConfigID: vcsConfig.ID,
				}
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(branch *app.AppBranch) {
				assert.NotEmpty(s.T(), branch.ID)
				assert.Equal(s.T(), "main-branch", branch.Name)
				assert.Equal(s.T(), s.testApp.ID, branch.AppID)
				assert.Equal(s.T(), s.testOrg.ID, branch.OrgID)

				var dbBranch app.AppBranch
				err := s.service.DB.First(&dbBranch, "id = ?", branch.ID).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "main-branch", dbBranch.Name)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req := tc.setupFunc()
			rr := s.makeRequestWithBody(http.MethodPost, "/v1/apps/"+s.testApp.ID+"/branches", req)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusCreated && tc.validateFunc != nil {
				var response app.AppBranch
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(s.T(), err)
				tc.validateFunc(&response)
			}
		})
	}
}
