package service

import (
	"errors"
	"net/http"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/oidc/githubauth"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func (s *VCSServiceTestSuite) TestCreateConnectionCallback_Success() {
	s.testAcc.AllPermissions = permissions.Set{s.testOrg.ID: permissions.PermissionAll}
	s.mockGH.EXPECT().VerifyInstallationOwnership(gomock.Any(), "code-1", "12345").Return(nil)

	req := CreateConnectionCallbackRequest{
		OrgID:           s.testOrg.ID,
		GithubInstallID: "12345",
		Code:            "code-1",
	}

	rr := s.makeRequest(http.MethodPost, "/v1/vcs/connection-callback", req)

	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var conn app.VCSConnection
	err := s.service.DB.Where("org_id = ? AND github_install_id = ?", s.testOrg.ID, "12345").First(&conn).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.testOrg.ID, conn.OrgID)
	assert.Equal(s.T(), "12345", conn.GithubInstallID)
	assert.Equal(s.T(), "12345", conn.GithubAccountID)
	assert.Equal(s.T(), "test-org", conn.GithubAccountName)
}

func (s *VCSServiceTestSuite) TestCreateConnectionCallback_NotOwned() {
	s.testAcc.AllPermissions = permissions.Set{s.testOrg.ID: permissions.PermissionAll}
	s.mockGH.EXPECT().VerifyInstallationOwnership(gomock.Any(), "code-1", "99999").
		Return(githubauth.ErrInstallationNotOwned)

	req := CreateConnectionCallbackRequest{
		OrgID:           s.testOrg.ID,
		GithubInstallID: "99999",
		Code:            "code-1",
	}

	rr := s.makeRequest(http.MethodPost, "/v1/vcs/connection-callback", req)
	require.Equal(s.T(), http.StatusForbidden, rr.Code, rr.Body.String())

	var conn app.VCSConnection
	err := s.service.DB.Where("org_id = ? AND github_install_id = ?", s.testOrg.ID, "99999").First(&conn).Error
	require.ErrorIs(s.T(), err, gorm.ErrRecordNotFound, "no connection row may be created")
}

func (s *VCSServiceTestSuite) TestCreateConnectionCallback_VerificationError() {
	s.testAcc.AllPermissions = permissions.Set{s.testOrg.ID: permissions.PermissionAll}
	s.mockGH.EXPECT().VerifyInstallationOwnership(gomock.Any(), "stale", "12345").
		Return(errors.New("github rejected oauth code: bad_verification_code"))

	req := CreateConnectionCallbackRequest{
		OrgID:           s.testOrg.ID,
		GithubInstallID: "12345",
		Code:            "stale",
	}

	rr := s.makeRequest(http.MethodPost, "/v1/vcs/connection-callback", req)
	require.Equal(s.T(), http.StatusInternalServerError, rr.Code, rr.Body.String())

	var conn app.VCSConnection
	err := s.service.DB.Where("org_id = ? AND github_install_id = ?", s.testOrg.ID, "12345").First(&conn).Error
	require.ErrorIs(s.T(), err, gorm.ErrRecordNotFound)
}

func (s *VCSServiceTestSuite) TestCreateConnectionCallback_MissingCode() {
	s.testAcc.AllPermissions = permissions.Set{s.testOrg.ID: permissions.PermissionAll}
	// No VerifyInstallationOwnership expectation: a missing code must fail
	// before GitHub is consulted.

	req := CreateConnectionCallbackRequest{
		OrgID:           s.testOrg.ID,
		GithubInstallID: "12345",
	}

	rr := s.makeRequest(http.MethodPost, "/v1/vcs/connection-callback", req)
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())

	var conn app.VCSConnection
	err := s.service.DB.Where("org_id = ? AND github_install_id = ?", s.testOrg.ID, "12345").First(&conn).Error
	require.ErrorIs(s.T(), err, gorm.ErrRecordNotFound)
}

func (s *VCSServiceTestSuite) TestCreateConnectionCallback_OwnershipCheckDisabled() {
	s.testAcc.AllPermissions = permissions.Set{s.testOrg.ID: permissions.PermissionAll}
	s.withOwnershipCheckDisabled()

	req := CreateConnectionCallbackRequest{
		OrgID:           s.testOrg.ID,
		GithubInstallID: "12345",
	}

	rr := s.makeRequest(http.MethodPost, "/v1/vcs/connection-callback", req)
	require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())
}

func (s *VCSServiceTestSuite) TestCreateConnectionCallback_MissingFields() {
	req := map[string]interface{}{} // Empty request missing required fields

	rr := s.makeRequest(http.MethodPost, "/v1/vcs/connection-callback", req)

	if rr.Code != http.StatusBadRequest {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
}
