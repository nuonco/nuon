package service

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *AppCRUDTestSuite) createOtherOrg() *app.Org {
	ctx := cctx.SetAccountContext(context.Background(), s.testAcc)

	otherOrg := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        "other-org-" + domains.NewOrgID(),
		CreatedByID: s.testAcc.ID,
		OrgType:     app.OrgTypeDefault,
	}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(otherOrg).Error)

	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", otherOrg.ID)
	})

	return otherOrg
}

func (s *AppCRUDTestSuite) createSecretTestApp(orgID string) *app.App {
	testApp := &app.App{
		ID:          domains.NewAppID(),
		Name:        "secret-app-" + domains.NewAppID(),
		OrgID:       orgID,
		CreatedByID: s.testAcc.ID,
		Status:      app.AppStatusProvisioning,
	}
	require.NoError(s.T(), s.service.DB.Create(testApp).Error)

	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.App{}, "id = ?", testApp.ID)
	})

	return testApp
}

func (s *AppCRUDTestSuite) createTestSecret(a *app.App, name, value string) *app.AppSecret {
	secret := &app.AppSecret{
		ID:          domains.NewAppSecretID(),
		OrgID:       a.OrgID,
		AppID:       a.ID,
		CreatedByID: s.testAcc.ID,
		Name:        name,
		Value:       value,
	}
	require.NoError(s.T(), s.service.DB.Create(secret).Error)

	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.AppSecret{}, "id = ?", secret.ID)
	})

	return secret
}

func (s *AppCRUDTestSuite) findTestSecret(secretID string) (*app.AppSecret, error) {
	var secret app.AppSecret
	if err := s.service.DB.First(&secret, "id = ?", secretID).Error; err != nil {
		return nil, err
	}
	return &secret, nil
}

func (s *AppCRUDTestSuite) TestDeleteAppSecretScoping() {
	for _, route := range []string{"secrets", "secret"} {
		s.Run(route+"/own secret by app id", func() {
			a := s.createSecretTestApp(s.testOrg.ID)
			secret := s.createTestSecret(a, "own_secret", "value")

			rr := s.makeRequest(http.MethodDelete, "/v1/apps/"+a.ID+"/"+route+"/"+secret.ID, nil)
			require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

			_, err := s.findTestSecret(secret.ID)
			assert.Error(s.T(), err)
		})

		s.Run(route+"/own secret by app name", func() {
			a := s.createSecretTestApp(s.testOrg.ID)
			secret := s.createTestSecret(a, "own_secret", "value")

			rr := s.makeRequest(http.MethodDelete, "/v1/apps/"+a.Name+"/"+route+"/"+secret.ID, nil)
			require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

			_, err := s.findTestSecret(secret.ID)
			assert.Error(s.T(), err)
		})

		s.Run(route+"/other org secret through own app", func() {
			a := s.createSecretTestApp(s.testOrg.ID)
			otherApp := s.createSecretTestApp(s.createOtherOrg().ID)
			secret := s.createTestSecret(otherApp, "other_secret", "value")

			rr := s.makeRequest(http.MethodDelete, "/v1/apps/"+a.ID+"/"+route+"/"+secret.ID, nil)
			require.Equal(s.T(), http.StatusNotFound, rr.Code, rr.Body.String())

			_, err := s.findTestSecret(secret.ID)
			assert.NoError(s.T(), err)
		})

		s.Run(route+"/other org app", func() {
			otherApp := s.createSecretTestApp(s.createOtherOrg().ID)
			secret := s.createTestSecret(otherApp, "other_secret", "value")

			rr := s.makeRequest(http.MethodDelete, "/v1/apps/"+otherApp.ID+"/"+route+"/"+secret.ID, nil)
			require.Equal(s.T(), http.StatusNotFound, rr.Code, rr.Body.String())

			_, err := s.findTestSecret(secret.ID)
			assert.NoError(s.T(), err)
		})

		s.Run(route+"/secret from a different app in the same org", func() {
			a := s.createSecretTestApp(s.testOrg.ID)
			sibling := s.createSecretTestApp(s.testOrg.ID)
			secret := s.createTestSecret(sibling, "sibling_secret", "value")

			rr := s.makeRequest(http.MethodDelete, "/v1/apps/"+a.ID+"/"+route+"/"+secret.ID, nil)
			require.Equal(s.T(), http.StatusNotFound, rr.Code, rr.Body.String())

			_, err := s.findTestSecret(secret.ID)
			assert.NoError(s.T(), err)
		})

		s.Run(route+"/non-existent secret", func() {
			a := s.createSecretTestApp(s.testOrg.ID)

			rr := s.makeRequest(http.MethodDelete, "/v1/apps/"+a.ID+"/"+route+"/"+domains.NewAppSecretID(), nil)
			require.Equal(s.T(), http.StatusNotFound, rr.Code, rr.Body.String())
		})
	}
}

func (s *AppCRUDTestSuite) TestCreateAppSecretScoping() {
	for _, route := range []string{"secrets", "secret"} {
		s.Run(route+"/own app by id", func() {
			a := s.createSecretTestApp(s.testOrg.ID)

			rr := s.makeRequest(http.MethodPost, "/v1/apps/"+a.ID+"/"+route, CreateAppSecretRequest{Name: "own_secret", Value: "value"})
			require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

			var secret app.AppSecret
			require.NoError(s.T(), s.service.DB.First(&secret, "app_id = ? AND name = ?", a.ID, "own_secret").Error)
			assert.Equal(s.T(), s.testOrg.ID, secret.OrgID)
		})

		s.Run(route+"/own app by name stores the resolved app id", func() {
			a := s.createSecretTestApp(s.testOrg.ID)

			rr := s.makeRequest(http.MethodPost, "/v1/apps/"+a.Name+"/"+route, CreateAppSecretRequest{Name: "own_secret", Value: "value"})
			require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

			var secret app.AppSecret
			require.NoError(s.T(), s.service.DB.First(&secret, "app_id = ? AND name = ?", a.ID, "own_secret").Error)
		})

		s.Run(route+"/other org app is rejected without creating a secret", func() {
			otherApp := s.createSecretTestApp(s.createOtherOrg().ID)

			rr := s.makeRequest(http.MethodPost, "/v1/apps/"+otherApp.ID+"/"+route, CreateAppSecretRequest{Name: "planted_secret", Value: "value"})
			require.Equal(s.T(), http.StatusNotFound, rr.Code, rr.Body.String())

			var count int64
			require.NoError(s.T(), s.service.DB.Model(&app.AppSecret{}).Where("app_id = ?", otherApp.ID).Count(&count).Error)
			assert.Zero(s.T(), count)
		})

		s.Run(route+"/other org secret is not overwritten", func() {
			otherOrg := s.createOtherOrg()
			otherApp := s.createSecretTestApp(otherOrg.ID)
			secret := s.createTestSecret(otherApp, "other_secret", "original")

			rr := s.makeRequest(http.MethodPost, "/v1/apps/"+otherApp.ID+"/"+route, CreateAppSecretRequest{Name: "other_secret", Value: "overwritten"})
			require.Equal(s.T(), http.StatusNotFound, rr.Code, rr.Body.String())

			got, err := s.findTestSecret(secret.ID)
			require.NoError(s.T(), err)
			assert.Equal(s.T(), "original", got.Value)
			assert.Equal(s.T(), otherOrg.ID, got.OrgID)
		})
	}
}
