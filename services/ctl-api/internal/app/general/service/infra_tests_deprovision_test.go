package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *GeneralTemporalTestSuite) TestInfraTestsDeprovision_Success() {
	s.mockTC.EXPECT().ExecuteWorkflowInNamespace(
		gomock.Any(),
		"infra-tests",
		gomock.Any(),
		"Deprovision",
		gomock.Any(),
	).Return(nil, nil)

	reqBody := InfraTestsDeprovisionRequest{
		SandboxName:      "test-sandbox",
		SandboxRef:       "main",
		TerraformVersion: "1.5.0",
		Region:           "us-west-2",
		SandboxVars: map[string]interface{}{
			"key1": "value1",
		},
		OrgID:       s.testOrg.ID,
		CanaryID:    "canary-123",
		Directory:   "/tmp/test",
		ClusterName: "test-cluster",
		Account: map[string]interface{}{
			"account_id": "123456789",
		},
		Profile: "test-profile",
	}
	rr := s.makeRequest(http.MethodPost, "/v1/general/infra-tests/deprovision", reqBody)

	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "ok", response["status"])
}

func (s *GeneralTemporalTestSuite) TestInfraTestsDeprovision_TemporalError() {
	expectedErr := assert.AnError
	s.mockTC.EXPECT().ExecuteWorkflowInNamespace(
		gomock.Any(),
		"infra-tests",
		gomock.Any(),
		"Deprovision",
		gomock.Any(),
	).Return(nil, expectedErr)

	reqBody := InfraTestsDeprovisionRequest{
		SandboxName:      "test-sandbox",
		SandboxRef:       "main",
		TerraformVersion: "1.5.0",
		Region:           "us-west-2",
		SandboxVars: map[string]interface{}{
			"key1": "value1",
		},
		OrgID:       s.testOrg.ID,
		CanaryID:    "canary-123",
		Directory:   "/tmp/test",
		ClusterName: "test-cluster",
		Account: map[string]interface{}{
			"account_id": "123456789",
		},
		Profile: "test-profile",
	}
	rr := s.makeRequest(http.MethodPost, "/v1/general/infra-tests/deprovision", reqBody)

	if rr.Code == http.StatusCreated {
		s.T().Logf("Expected error but got success. Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusInternalServerError, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), "unable to provision infra-tests")
}

func (s *GeneralTemporalTestSuite) TestInfraTestsDeprovision_MalformedJSON() {
	req, err := http.NewRequest(http.MethodPost, "/v1/general/infra-tests/deprovision", bytes.NewBuffer([]byte("not valid json")))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
}

func (s *GeneralTemporalTestSuite) TestInfraTestsDeprovision_MinimalFields() {
	s.mockTC.EXPECT().ExecuteWorkflowInNamespace(
		gomock.Any(),
		"infra-tests",
		gomock.Any(),
		"Deprovision",
		gomock.Any(),
	).Return(nil, nil)

	reqBody := InfraTestsDeprovisionRequest{
		SandboxName: "test-sandbox",
		CanaryID:    "canary-123",
	}
	rr := s.makeRequest(http.MethodPost, "/v1/general/infra-tests/deprovision", reqBody)

	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "ok", response["status"])
}
