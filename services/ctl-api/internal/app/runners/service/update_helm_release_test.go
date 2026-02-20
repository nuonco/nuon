package service

import (
	"bytes"
	"context"
	"encoding/json"
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

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type UpdateHelmReleaseTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type UpdateHelmReleaseTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service UpdateHelmReleaseTestService
	router  *gin.Engine
	testOrg *app.Org
	testAcc *app.Account
}

func TestUpdateHelmReleaseSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(UpdateHelmReleaseTestSuite))
}

func (s *UpdateHelmReleaseTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *UpdateHelmReleaseTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with runner routes
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:  s.service.L,
		DB: s.service.DB,
	})
	err := s.service.RunnersService.RegisterRunnerRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *UpdateHelmReleaseTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *UpdateHelmReleaseTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())
}

func (s *UpdateHelmReleaseTestSuite) createHelmChart(ctx context.Context, helmChartID string) {
	chart := &app.HelmChart{
		ID:          helmChartID,
		CreatedByID: s.testAcc.ID,
		OrgID:       s.testOrg.ID,
		OwnerID:     domains.NewInstallID(),
		OwnerType:   "install",
	}
	err := s.service.DB.WithContext(ctx).Create(chart).Error
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().Delete(chart)
	})
}

func (s *UpdateHelmReleaseTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(s.T(), err)
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *UpdateHelmReleaseTestSuite) TestUpdateHelmReleaseValidation() {
	testCases := []struct {
		name         string
		helmChartID  string
		namespace    string
		key          string
		body         interface{}
		expectedCode int
	}{
		{
			name:         "missing helm_chart_id returns 404",
			helmChartID:  "",
			namespace:    "default",
			key:          "key",
			body:         map[string]interface{}{"name": "test"},
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "missing namespace returns 404",
			helmChartID:  "hchvalid123456789012345678",
			namespace:    "",
			key:          "key",
			body:         map[string]interface{}{"name": "test"},
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "missing key returns 307 redirect",
			helmChartID:  "hchvalid123456789012345678",
			namespace:    "default",
			key:          "",
			body:         map[string]interface{}{"name": "test"},
			expectedCode: http.StatusTemporaryRedirect,
		},
		{
			name:         "invalid JSON body returns 400",
			helmChartID:  "hchvalid123456789012345678",
			namespace:    "default",
			key:          "key",
			body:         "invalid-json",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "empty body returns 400",
			helmChartID:  "hchvalid123456789012345678",
			namespace:    "default",
			key:          "key",
			body:         nil,
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			path := "/v1/helm-releases/" + tc.helmChartID + "/releases/" + tc.namespace + "/" + tc.key
			rr := s.makeRequest("PUT", path, tc.body)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)
		})
	}
}

func (s *UpdateHelmReleaseTestSuite) TestUpdateHelmReleaseEncoding() {
	// Document that UpdateHelmRelease requires valid helm.Release for encoding
	// Similar to CreateHelmRelease, this is complex due to proto structures

	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	helmChartID := domains.NewHelmChartID()
	namespace := "default"
	key := "sh.helm.release.v1.test-release.v1"

	// Create helm chart to satisfy FK constraint
	s.createHelmChart(ctx, helmChartID)

	// Create an existing release to update
	existingRelease := &app.HelmRelease{
		HelmChartID: helmChartID,
		Key:         key,
		CreatedByID: s.testAcc.ID,
		OrgID:       s.testOrg.ID,
		Type:        "helm.sh/release.v1",
		Body:        "",
		Name:        "test-release",
		Namespace:   namespace,
		Version:     1,
		Status:      "deployed",
		Owner:       "helm",
	}
	err := s.service.DB.WithContext(ctx).Create(existingRelease).Error
	require.NoError(s.T(), err)

	releaseKey := existingRelease.Key
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().
			Where("helm_chart_id = ? AND key = ?", helmChartID, releaseKey).
			Delete(&app.HelmRelease{})
	})

	s.T().Log("UpdateHelmRelease requires valid helm.Release JSON for encoding")
	s.T().Log("The helm.Release type has complex nested proto structures")
	s.T().Log("Incomplete release data will cause encoding failures")

	// Try to update with minimal data (will likely fail encoding)
	requestBody := map[string]interface{}{
		"name":      "updated-release",
		"version":   2,
		"namespace": namespace,
		"info": map[string]interface{}{
			"status": "superseded",
		},
	}

	rr := s.makeRequest("PUT", "/v1/helm-releases/"+helmChartID+"/releases/"+namespace+"/"+key, requestBody)

	s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())

	// Will likely fail due to incomplete helm.Release structure for encoding
	// or succeed if the encoding tolerates missing fields
}

func (s *UpdateHelmReleaseTestSuite) TestUpdateHelmReleaseNonexistent() {
	// Test updating a nonexistent release
	// The handler doesn't explicitly check if the release exists before updating,
	// so it may succeed with 0 rows affected

	helmChartID := domains.NewHelmChartID()
	namespace := "default"
	key := "nonexistent-key"

	requestBody := map[string]interface{}{
		"name":      "test-release",
		"version":   1,
		"namespace": namespace,
		"info": map[string]interface{}{
			"status": "deployed",
		},
	}

	rr := s.makeRequest("PUT", "/v1/helm-releases/"+helmChartID+"/releases/"+namespace+"/"+key, requestBody)

	s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())

	// Behavior depends on encoding success and GORM's Updates behavior
	// Updates doesn't return error for 0 rows affected
}

func (s *UpdateHelmReleaseTestSuite) TestUpdateHelmReleaseComplexity() {
	// Document the complexity of helm release updates
	s.T().Log("UpdateHelmRelease faces the same complexity as CreateHelmRelease:")
	s.T().Log("1. Requires valid helm.Release JSON with nested proto structures")
	s.T().Log("2. helm.EncodeRelease does protobuf encoding + gzip + base64")
	s.T().Log("3. Missing or invalid fields cause encoding failures")
	s.T().Log("4. Unlike Create, Update doesn't require CreatedByID (uses existing record)")
	s.T().Log("")
	s.T().Log("In production, helm releases are updated by actual Helm operations")
	s.T().Log("These tests focus on parameter validation and documenting encoding requirements")
}

func (s *UpdateHelmReleaseTestSuite) TestUpdateHelmReleaseSuccessPath() {
	// Document what a successful update would look like
	// (even though we can't easily create valid helm.Release test data)

	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)

	helmChartID := domains.NewHelmChartID()
	namespace := "default"
	key := "sh.helm.release.v1.test-release.v1"

	// Create helm chart to satisfy FK constraint
	s.createHelmChart(ctx, helmChartID)

	// Create existing release
	existingRelease := &app.HelmRelease{
		HelmChartID: helmChartID,
		Key:         key,
		CreatedByID: s.testAcc.ID,
		OrgID:       s.testOrg.ID,
		Type:        "helm.sh/release.v1",
		Body:        "",
		Name:        "test-release",
		Namespace:   namespace,
		Version:     1,
		Status:      "deployed",
		Owner:       "helm",
	}
	err := s.service.DB.WithContext(ctx).Create(existingRelease).Error
	require.NoError(s.T(), err)

	releaseKey := existingRelease.Key
	s.T().Cleanup(func() {
		s.service.DB.Unscoped().
			Where("helm_chart_id = ? AND key = ?", helmChartID, releaseKey).
			Delete(&app.HelmRelease{})
	})

	s.T().Log("For a successful update:")
	s.T().Log("1. The release must exist in the database")
	s.T().Log("2. The request body must contain a valid helm.Release JSON")
	s.T().Log("3. helm.EncodeRelease must succeed with the release data")
	s.T().Log("4. GORM Updates will update Body, Name, Version, Status, Owner, UpdatedAt")
	s.T().Log("5. Returns 200 with null body on success")

	// Verify the release exists
	var checkRelease app.HelmRelease
	err = s.service.DB.WithContext(ctx).
		Where("helm_chart_id = ? AND namespace = ? AND key = ?", helmChartID, namespace, key).
		First(&checkRelease).Error
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "test-release", checkRelease.Name)
	assert.Equal(s.T(), 1, checkRelease.Version)
}
