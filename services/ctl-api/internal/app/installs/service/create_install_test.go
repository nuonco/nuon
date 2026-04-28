package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
)

func (s *InstallsServiceTestSuite) TestCreateInstallV2Success() {
	s.expectQueueCreation()

	body := CreateInstallV2Request{
		AppID: s.testApp.ID,
		CreateInstallParams: helpers.CreateInstallParams{
			Name: "my-install",
			AWSAccount: &struct {
				Region string `json:"region"`
			}{Region: "us-west-2"},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))

	assert.NotEmpty(s.T(), install.ID)
	assert.Equal(s.T(), "my-install", install.Name)
	assert.Equal(s.T(), s.testApp.ID, install.AppID)

	captured := s.mockEvClient.GetSignals()
	require.GreaterOrEqual(s.T(), len(captured), 4)

	var signalTypes []string
	for _, c := range captured {
		if sig, ok := c.Signal.(*signals.Signal); ok {
			signalTypes = append(signalTypes, string(sig.Type))
		}
	}
	assert.Contains(s.T(), signalTypes, string(signals.OperationCreated))
	assert.Contains(s.T(), signalTypes, string(signals.OperationPollDependencies))
	assert.Contains(s.T(), signalTypes, string(signals.OperationSyncActionWorkflowTriggers))
	assert.Contains(s.T(), signalTypes, string(signals.OperationExecuteFlow))
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2WithInputs() {
	s.expectQueueCreation()

	region := "us-east-1"
	body := CreateInstallV2Request{
		AppID: s.testApp.ID,
		CreateInstallParams: helpers.CreateInstallParams{
			Name: "install-with-inputs",
			AWSAccount: &struct {
				Region string `json:"region"`
			}{Region: "us-west-2"},
			Inputs: map[string]*string{
				"region": &region,
			},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
	assert.Equal(s.T(), "install-with-inputs", install.Name)
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2AzureAccount() {
	s.expectQueueCreation()

	body := CreateInstallV2Request{
		AppID: s.testApp.ID,
		CreateInstallParams: helpers.CreateInstallParams{
			Name: "azure-install",
			AzureAccount: &struct {
				Location string `json:"location"`
			}{Location: "eastus"},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
	assert.Equal(s.T(), "azure-install", install.Name)
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2ValidationErrors() {
	testCases := []struct {
		name string
		body interface{}
	}{
		{
			name: "missing app_id",
			body: map[string]interface{}{
				"name":        "foo",
				"aws_account": map[string]string{"region": "us-west-2"},
			},
		},
		{
			name: "missing name",
			body: map[string]interface{}{
				"app_id":      s.testApp.ID,
				"aws_account": map[string]string{"region": "us-west-2"},
			},
		},
		{
			name: "no cloud account",
			body: map[string]interface{}{
				"app_id": s.testApp.ID,
				"name":   "foo",
			},
		},
		{
			name: "aws missing region",
			body: map[string]interface{}{
				"app_id":      s.testApp.ID,
				"name":        "foo",
				"aws_account": map[string]string{},
			},
		},
		{
			name: "empty body",
			body: map[string]interface{}{},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rr := s.makeRequest(http.MethodPost, "/v1/installs", tc.body)
			require.Equal(s.T(), http.StatusBadRequest, rr.Code,
				"expected 400 for %s, got %d: %s", tc.name, rr.Code, rr.Body.String())
		})
	}
}

func (s *InstallsServiceTestSuite) TestCreateInstallV2InvalidAppID() {
	body := CreateInstallV2Request{
		AppID: "app_nonexistent_000000000",
		CreateInstallParams: helpers.CreateInstallParams{
			Name: "bad-app",
			AWSAccount: &struct {
				Region string `json:"region"`
			}{Region: "us-west-2"},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/installs", body)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) TestCreateInstallDeprecatedRoute() {
	s.expectQueueCreation()

	body := helpers.CreateInstallParams{
		Name: "deprecated-route-install",
		AWSAccount: &struct {
			Region string `json:"region"`
		}{Region: "us-west-2"},
	}

	path := fmt.Sprintf("/v1/apps/%s/installs", s.testApp.ID)
	rr := s.makeRequest(http.MethodPost, path, body)
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var install app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &install))
	assert.Equal(s.T(), "deprecated-route-install", install.Name)
	assert.Equal(s.T(), s.testApp.ID, install.AppID)
}
