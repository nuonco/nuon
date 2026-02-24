package service

import (
	"net/http"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/stretchr/testify/require"
)

func (s *GeneralInternalTestSuite) TestAdminDeleteAccount() {
	testCases := []struct {
		name           string
		setupFunc      func() *app.Account
		requestFunc    func(acct *app.Account) AdminDeleteAccountRequest
		expectedStatus int
		validateFunc   func(acct *app.Account)
	}{
		{
			name: "deletes account by email",
			setupFunc: func() *app.Account {
				acct := s.service.Seeder.CreateAccount(s.ctx, s.T())
				return acct
			},
			requestFunc: func(acct *app.Account) AdminDeleteAccountRequest {
				return AdminDeleteAccountRequest{
					EmailOrSubjectOrID: acct.Email,
				}
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(acct *app.Account) {
				// Verify account is deleted
				var count int64
				err := s.service.DB.Model(&app.Account{}).Where("id = ?", acct.ID).Count(&count).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), int64(0), count, "account should be deleted")
			},
		},
		{
			name: "deletes account by subject",
			setupFunc: func() *app.Account {
				acct := s.service.Seeder.CreateAccount(s.ctx, s.T())
				return acct
			},
			requestFunc: func(acct *app.Account) AdminDeleteAccountRequest {
				return AdminDeleteAccountRequest{
					EmailOrSubjectOrID: acct.Subject,
				}
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(acct *app.Account) {
				// Verify account is deleted
				var count int64
				err := s.service.DB.Model(&app.Account{}).Where("id = ?", acct.ID).Count(&count).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), int64(0), count, "account should be deleted")
			},
		},
		{
			name: "deletes account by ID",
			setupFunc: func() *app.Account {
				acct := s.service.Seeder.CreateAccount(s.ctx, s.T())
				return acct
			},
			requestFunc: func(acct *app.Account) AdminDeleteAccountRequest {
				return AdminDeleteAccountRequest{
					EmailOrSubjectOrID: acct.ID,
				}
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(acct *app.Account) {
				// Verify account is deleted
				var count int64
				err := s.service.DB.Model(&app.Account{}).Where("id = ?", acct.ID).Count(&count).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), int64(0), count, "account should be deleted")
			},
		},
		{
			name: "fails with missing email_or_subject_or_id",
			setupFunc: func() *app.Account {
				return nil
			},
			requestFunc: func(acct *app.Account) AdminDeleteAccountRequest {
				return AdminDeleteAccountRequest{}
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Setup test data
			acct := tc.setupFunc()

			// Make request
			req := tc.requestFunc(acct)
			rr := s.makeRequest(http.MethodPost, "/v1/general/admin-delete-account", req)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Validate result
			if tc.validateFunc != nil && acct != nil {
				tc.validateFunc(acct)
			}
		})
	}
}
