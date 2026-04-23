package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// ---------------------------------------------------------------------------
// Admin Add Org Labels
// ---------------------------------------------------------------------------

// AdminAddOrgLabelsTestService holds all fx-injected dependencies.
type AdminAddOrgLabelsTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	OrgsService     *service
}

// AdminAddOrgLabelsTestSuite is the testify suite for the AdminAddOrgLabels endpoint.
type AdminAddOrgLabelsTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service AdminAddOrgLabelsTestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
}

func TestAdminAddOrgLabelsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AdminAddOrgLabelsTestSuite))
}

func (s *AdminAddOrgLabelsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	s.SetDB(s.service.DB)
}

func (s *AdminAddOrgLabelsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})

	err := s.service.OrgsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminAddOrgLabelsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminAddOrgLabelsTestSuite) setupTestData() {
	testAccID := domains.NewAccountID()
	testAcc := &app.Account{
		ID:          testAccID,
		Email:       fmt.Sprintf("%s@test.nuon.co", testAccID),
		Subject:     fmt.Sprintf("add-labels-%s", testAccID),
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, testAcc)

	orgID := domains.NewOrgID()
	testOrg := &app.Org{
		ID:          orgID,
		Name:        fmt.Sprintf("add-labels-%s", orgID),
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/test",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg
}

func (s *AdminAddOrgLabelsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(s.T(), err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminAddOrgLabelsTestSuite) createTestOrg(name string, existingLabels labels.Labels) *app.Org {
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	org := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        name,
		SandboxMode: true,
		Labeled:     labels.Labeled{Labels: existingLabels},
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/test",
		},
	}
	err := s.service.DB.WithContext(ctx).Create(org).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
	})
	return org
}

func (s *AdminAddOrgLabelsTestSuite) TestAdminAddOrgLabels() {
	s.Run("adds labels to org with no existing labels", func() {
		org := s.createTestOrg("test-add-labels", nil)

		reqBody := AdminAddOrgLabelsRequest{
			Labels: map[string]string{"env": "prod", "tier": "enterprise"},
		}
		rr := s.makeRequest(http.MethodPost, "/v1/orgs/"+org.ID+"/admin-labels", reqBody)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Org
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", response.Labels["env"])
		assert.Equal(s.T(), "enterprise", response.Labels["tier"])

		// Verify in DB
		var dbOrg app.Org
		err = s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", dbOrg.Labels["env"])
		assert.Equal(s.T(), "enterprise", dbOrg.Labels["tier"])
	})

	s.Run("merges labels with existing labels", func() {
		org := s.createTestOrg("test-merge-labels", labels.Labels{"env": "staging"})

		reqBody := AdminAddOrgLabelsRequest{
			Labels: map[string]string{"tier": "enterprise"},
		}
		rr := s.makeRequest(http.MethodPost, "/v1/orgs/"+org.ID+"/admin-labels", reqBody)

		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Org
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "staging", response.Labels["env"])
		assert.Equal(s.T(), "enterprise", response.Labels["tier"])
	})

	s.Run("overwrites existing key", func() {
		org := s.createTestOrg("test-overwrite-labels", labels.Labels{"env": "staging"})

		reqBody := AdminAddOrgLabelsRequest{
			Labels: map[string]string{"env": "prod"},
		}
		rr := s.makeRequest(http.MethodPost, "/v1/orgs/"+org.ID+"/admin-labels", reqBody)

		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Org
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", response.Labels["env"])
	})

	s.Run("fails with empty body", func() {
		rr := s.makeRequest(http.MethodPost, "/v1/orgs/"+s.testOrg.ID+"/admin-labels", map[string]interface{}{})
		require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	})

	s.Run("fails with invalid JSON", func() {
		req, err := http.NewRequest(http.MethodPost, "/v1/orgs/"+s.testOrg.ID+"/admin-labels", bytes.NewBufferString("{invalid"))
		require.NoError(s.T(), err)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		s.router.ServeHTTP(rr, req)
		require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	})

	s.Run("fails when org not found", func() {
		reqBody := AdminAddOrgLabelsRequest{
			Labels: map[string]string{"env": "prod"},
		}
		rr := s.makeRequest(http.MethodPost, "/v1/orgs/"+domains.NewOrgID()+"/admin-labels", reqBody)
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}

// ---------------------------------------------------------------------------
// Admin Remove Org Labels
// ---------------------------------------------------------------------------

// AdminRemoveOrgLabelsTestService holds all fx-injected dependencies.
type AdminRemoveOrgLabelsTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	OrgsHelpers     *orgshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	OrgsService     *service
}

// AdminRemoveOrgLabelsTestSuite is the testify suite for the AdminRemoveOrgLabels endpoint.
type AdminRemoveOrgLabelsTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service AdminRemoveOrgLabelsTestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
}

func TestAdminRemoveOrgLabelsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(AdminRemoveOrgLabelsTestSuite))
}

func (s *AdminRemoveOrgLabelsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	s.SetDB(s.service.DB)
}

func (s *AdminRemoveOrgLabelsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestAcc: s.testAcc,
	})

	err := s.service.OrgsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminRemoveOrgLabelsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminRemoveOrgLabelsTestSuite) setupTestData() {
	testAccID := domains.NewAccountID()
	testAcc := &app.Account{
		ID:          testAccID,
		Email:       fmt.Sprintf("%s@test.nuon.co", testAccID),
		Subject:     fmt.Sprintf("rm-labels-%s", testAccID),
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, testAcc)

	orgID := domains.NewOrgID()
	testOrg := &app.Org{
		ID:          orgID,
		Name:        fmt.Sprintf("rm-labels-%s", orgID),
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/test",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg
}

func (s *AdminRemoveOrgLabelsTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(s.T(), err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminRemoveOrgLabelsTestSuite) createTestOrg(name string, existingLabels labels.Labels) *app.Org {
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	org := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        name,
		SandboxMode: true,
		Labeled:     labels.Labeled{Labels: existingLabels},
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/test",
		},
	}
	err := s.service.DB.WithContext(ctx).Create(org).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(&app.Org{}, "id = ?", org.ID)
	})
	return org
}

func (s *AdminRemoveOrgLabelsTestSuite) TestAdminRemoveOrgLabels() {
	s.Run("removes specified keys", func() {
		org := s.createTestOrg("test-remove-labels", labels.Labels{"env": "prod", "team": "platform", "region": "us-west-2"})

		reqBody := AdminRemoveOrgLabelsRequest{
			Keys: []string{"team"},
		}
		rr := s.makeRequest(http.MethodDelete, "/v1/orgs/"+org.ID+"/admin-labels", reqBody)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Org
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", response.Labels["env"])
		assert.Equal(s.T(), "us-west-2", response.Labels["region"])
		_, hasTeam := response.Labels["team"]
		assert.False(s.T(), hasTeam)

		// Verify in DB
		var dbOrg app.Org
		err = s.service.DB.First(&dbOrg, "id = ?", org.ID).Error
		require.NoError(s.T(), err)
		_, hasTeam = dbOrg.Labels["team"]
		assert.False(s.T(), hasTeam)
	})

	s.Run("removing non-existent key succeeds silently", func() {
		org := s.createTestOrg("test-nonexistent-key", labels.Labels{"env": "prod"})

		reqBody := AdminRemoveOrgLabelsRequest{
			Keys: []string{"nonexistent"},
		}
		rr := s.makeRequest(http.MethodDelete, "/v1/orgs/"+org.ID+"/admin-labels", reqBody)

		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Org
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "prod", response.Labels["env"])
	})

	s.Run("removes all labels", func() {
		org := s.createTestOrg("test-remove-all", labels.Labels{"a": "1", "b": "2"})

		reqBody := AdminRemoveOrgLabelsRequest{
			Keys: []string{"a", "b"},
		}
		rr := s.makeRequest(http.MethodDelete, "/v1/orgs/"+org.ID+"/admin-labels", reqBody)

		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.Org
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Empty(s.T(), response.Labels)
	})

	s.Run("fails with empty body", func() {
		rr := s.makeRequest(http.MethodDelete, "/v1/orgs/"+s.testOrg.ID+"/admin-labels", map[string]interface{}{})
		require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	})

	s.Run("fails with invalid JSON", func() {
		req, err := http.NewRequest(http.MethodDelete, "/v1/orgs/"+s.testOrg.ID+"/admin-labels", bytes.NewBufferString("{invalid"))
		require.NoError(s.T(), err)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		s.router.ServeHTTP(rr, req)
		require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	})

	s.Run("fails when org not found", func() {
		reqBody := AdminRemoveOrgLabelsRequest{
			Keys: []string{"env"},
		}
		rr := s.makeRequest(http.MethodDelete, "/v1/orgs/"+domains.NewOrgID()+"/admin-labels", reqBody)
		require.Equal(s.T(), http.StatusNotFound, rr.Code)
	})
}
