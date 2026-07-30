package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/types"
)

// TestUpdateInstallHasNoTargetAccountField pins the immutability invariant: the
// target cloud account is write-once, and it stays that way by there being no field
// on the update request to carry it. Runs without a database on purpose — this is
// the assertion most likely to be broken by someone innocently adding a field.
func TestUpdateInstallHasNoTargetAccountField(t *testing.T) {
	forbidden := []string{"account_id", "aws_account", "azure_account", "gcp_account",
		"subscription_id", "project_id", "cloud_platform_metadata"}

	var walk func(reflect.Type, string)
	walk = func(typ reflect.Type, path string) {
		for typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		if typ.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			name := strings.Split(field.Tag.Get("json"), ",")[0]
			for _, bad := range forbidden {
				if name == bad {
					t.Errorf("UpdateInstallRequest%s exposes %q — the target cloud account must be immutable after creation", path, name)
				}
			}
			walk(field.Type, path+"."+field.Name)
		}
	}
	walk(reflect.TypeOf(UpdateInstallRequest{}), "")
}

// setOrgFeatures turns the given feature flags on for the suite's test org. The
// features client reads the org row per request, so this has to hit the database
// rather than just mutating s.testOrg.
func (s *InstallsServiceTestSuite) setOrgFeatures(features ...app.OrgFeature) {
	var org app.Org
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).First(&org, "id = ?", s.testOrg.ID).Error)

	if org.Features == nil {
		org.Features = types.StringBoolMap{}
	}
	for _, feature := range features {
		org.Features[string(feature)] = true
	}

	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(&app.Org{ID: s.testOrg.ID}).
		Select("features").
		Updates(map[string]any{"features": org.Features}).Error)
}

// seedVerifiedAWSAccountConnection creates a connection in a state CreateInstall
// will accept, so the target account can be derived from it.
func (s *InstallsServiceTestSuite) seedVerifiedAWSAccountConnection(accountID string) *app.AWSAccountConnection {
	connection := &app.AWSAccountConnection{
		OrgID:              s.testOrg.ID,
		Name:               "test-connection",
		AccountID:          accountID,
		DefaultRegion:      "us-west-2",
		RoleARN:            "arn:aws:iam::" + accountID + ":role/nuon-connection",
		VerificationStatus: app.AWSAccountConnectionVerificationVerified,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(connection).Error)
	return connection
}

func (s *InstallsServiceTestSuite) createInstallWithAWSAccount(
	name string, awsAccount *helpers.CreateInstallAWSAccountParams,
) *httptest.ResponseRecorder {
	body := CreateInstallV2Request{
		AppID: s.testApp.ID,
		CreateInstallParams: helpers.CreateInstallParams{
			Name:       name,
			AWSAccount: awsAccount,
		},
	}
	return s.makeRequest(http.MethodPost, "/v1/installs", body)
}

// With the flag off the field stays advisory: absent is fine, and a malformed value
// is deliberately not rejected so organizations that never opted in see no change.
func (s *InstallsServiceTestSuite) TestCreateInstallTargetAccountOptionalWhenFlagOff() {
	s.expectQueueCreation()

	rr := s.createInstallWithAWSAccount("no-target-account",
		&helpers.CreateInstallAWSAccountParams{Region: "us-west-2"})
	require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
	assert.Empty(s.T(), install.CloudPlatformMetadata.TargetAccountID)
	assert.Empty(s.T(), install.CloudPlatformMetadata.TargetSource)
	assert.Empty(s.T(), install.ExpectedAccountID)
}

func (s *InstallsServiceTestSuite) TestCreateInstallTargetAccountNotFormatCheckedWhenFlagOff() {
	s.expectQueueCreation()

	rr := s.createInstallWithAWSAccount("garbage-target-account",
		&helpers.CreateInstallAWSAccountParams{Region: "us-west-2", AccountID: "not-an-account"})
	require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
	assert.Equal(s.T(), "not-an-account", install.CloudPlatformMetadata.TargetAccountID)
	assert.Equal(s.T(), app.CloudPlatformTargetSourceUser, install.CloudPlatformMetadata.TargetSource)
}

func (s *InstallsServiceTestSuite) TestCreateInstallTargetAccountPersistedWhenFlagOff() {
	s.expectQueueCreation()

	rr := s.createInstallWithAWSAccount("with-target-account",
		&helpers.CreateInstallAWSAccountParams{Region: "us-west-2", AccountID: "123456789012"})
	require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
	assert.Equal(s.T(), "123456789012", install.CloudPlatformMetadata.TargetAccountID)
	assert.Equal(s.T(), app.CloudPlatformTargetSourceUser, install.CloudPlatformMetadata.TargetSource)

	// Round-trip through the database so AfterQuery runs and derives the coalesced field.
	var reloaded app.Install
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).First(&reloaded, "id = ?", install.ID).Error)
	assert.Equal(s.T(), "123456789012", reloaded.CloudPlatformMetadata.TargetAccountID)
	assert.Equal(s.T(), "123456789012", reloaded.ExpectedAccountID)
}

func (s *InstallsServiceTestSuite) TestCreateInstallTargetAccountRequiredWhenFlagOn() {
	s.setOrgFeatures(app.OrgFeaturePhoneHomeAuth)

	rr := s.createInstallWithAWSAccount("missing-target-account",
		&helpers.CreateInstallAWSAccountParams{Region: "us-west-2"})
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Contains(s.T(), rr.Body.String(), "account_id")
}

func (s *InstallsServiceTestSuite) TestCreateInstallTargetAccountFormatCheckedWhenFlagOn() {
	s.setOrgFeatures(app.OrgFeaturePhoneHomeAuth)

	for _, malformed := range []string{"12345", "not-an-account", "1234567890123", " 123456789012"} {
		s.Run(malformed, func() {
			rr := s.createInstallWithAWSAccount("malformed-"+malformed,
				&helpers.CreateInstallAWSAccountParams{Region: "us-west-2", AccountID: malformed})
			require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
		})
	}
}

func (s *InstallsServiceTestSuite) TestCreateInstallTargetAccountAcceptedWhenFlagOn() {
	s.setOrgFeatures(app.OrgFeaturePhoneHomeAuth)
	s.expectQueueCreation()

	rr := s.createInstallWithAWSAccount("valid-target-account",
		&helpers.CreateInstallAWSAccountParams{Region: "us-west-2", AccountID: "123456789012"})
	require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
	assert.Equal(s.T(), "123456789012", install.CloudPlatformMetadata.TargetAccountID)
	assert.Equal(s.T(), app.CloudPlatformTargetSourceUser, install.CloudPlatformMetadata.TargetSource)
}

func (s *InstallsServiceTestSuite) TestCreateInstallTargetAccountDerivedFromConnection() {
	s.setOrgFeatures(app.OrgFeaturePhoneHomeAuth, app.OrgFeatureAWSAccountConnections)
	s.expectQueueCreation()

	connection := s.seedVerifiedAWSAccountConnection("210987654321")

	rr := s.createInstallWithAWSAccount("connection-target-account",
		&helpers.CreateInstallAWSAccountParams{Region: "us-west-2", ConnectionID: connection.ID})
	require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
	assert.Equal(s.T(), "210987654321", install.CloudPlatformMetadata.TargetAccountID)
	assert.Equal(s.T(), app.CloudPlatformTargetSourceConnection, install.CloudPlatformMetadata.TargetSource)
}

// An explicit account ID may agree with the connection but never contradict it.
func (s *InstallsServiceTestSuite) TestCreateInstallTargetAccountConflictsWithConnection() {
	s.setOrgFeatures(app.OrgFeaturePhoneHomeAuth, app.OrgFeatureAWSAccountConnections)

	connection := s.seedVerifiedAWSAccountConnection("210987654321")

	rr := s.createInstallWithAWSAccount("conflicting-target-account",
		&helpers.CreateInstallAWSAccountParams{
			Region:       "us-west-2",
			ConnectionID: connection.ID,
			AccountID:    "123456789012",
		})
	require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
	assert.Contains(s.T(), rr.Body.String(), "does not match")
}

func (s *InstallsServiceTestSuite) TestCreateInstallTargetAccountAgreesWithConnection() {
	s.setOrgFeatures(app.OrgFeaturePhoneHomeAuth, app.OrgFeatureAWSAccountConnections)
	s.expectQueueCreation()

	connection := s.seedVerifiedAWSAccountConnection("210987654321")

	rr := s.createInstallWithAWSAccount("agreeing-target-account",
		&helpers.CreateInstallAWSAccountParams{
			Region:       "us-west-2",
			ConnectionID: connection.ID,
			AccountID:    "210987654321",
		})
	require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
	assert.Equal(s.T(), app.CloudPlatformTargetSourceConnection, install.CloudPlatformMetadata.TargetSource)
}

// GCP and Azure gain the same requirement. The suite's app is AWS-typed, so these
// drive the other branches through their own app fixtures.
func (s *InstallsServiceTestSuite) TestCreateInstallTargetProjectAndSubscriptionRequiredWhenFlagOn() {
	s.setOrgFeatures(app.OrgFeaturePhoneHomeAuth)

	for _, tc := range []struct {
		name       string
		runnerType app.AppRunnerType
		body       map[string]any
		wantField  string
	}{
		{
			name:       "gcp missing project_id",
			runnerType: app.AppRunnerTypeGCP,
			body:       map[string]any{"gcp_account": map[string]string{"region": "us-central1"}},
			wantField:  "project_id",
		},
		{
			name:       "azure missing subscription_id",
			runnerType: app.AppRunnerTypeAzure,
			body:       map[string]any{"azure_account": map[string]string{"location": "eastus"}},
			wantField:  "subscription_id",
		},
	} {
		s.Run(tc.name, func() {
			targetApp := s.deps.Seeder.CreateApp(s.ctx, s.T())
			cfg := s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), targetApp.ID)
			require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
				Model(&app.AppRunnerConfig{}).
				Where("app_id = ? AND app_config_id = ?", targetApp.ID, cfg.ID).
				Update("type", tc.runnerType).Error)

			body := map[string]any{"app_id": targetApp.ID, "name": tc.name}
			for k, v := range tc.body {
				body[k] = v
			}

			rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
			require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
			assert.Contains(s.T(), rr.Body.String(), tc.wantField)
		})
	}
}
