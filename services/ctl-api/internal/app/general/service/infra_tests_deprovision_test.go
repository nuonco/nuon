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
	// Set up mock expectation for successful workflow execution
	s.mockTC.EXPECT().ExecuteWorkflowInNamespace(
		gomock.Any(),  // ctx
		"infra-tests", // namespace
		gomock.Any(),  // options
		"Deprovision", // workflow
		gomock.Any(),  // args
	).Return(nil, nil)

	// Make request with all required fields
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

	// Assert response
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	// Parse response
	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "ok", response["status"])
}

func (s *GeneralTemporalTestSuite) TestInfraTestsDeprovision_TemporalError() {
	// Set up mock expectation to return error
	expectedErr := assert.AnError
	s.mockTC.EXPECT().ExecuteWorkflowInNamespace(
		gomock.Any(),  // ctx
		"infra-tests", // namespace
		gomock.Any(),  // options
		"Deprovision", // workflow
		gomock.Any(),  // args
	).Return(nil, expectedErr)

	// Make request
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

	// Assert error response
	if rr.Code == http.StatusCreated {
		s.T().Logf("Expected error but got success. Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	assert.NotEqual(s.T(), http.StatusCreated, rr.Code)
	assert.Contains(s.T(), rr.Body.String(), "unable to provision infra-tests")
}

func (s *GeneralTemporalTestSuite) TestInfraTestsDeprovision_MalformedJSON() {
	// No mock expectation needed - should fail JSON parsing before calling temporal

	// Make request with malformed JSON (not using makeRequest to send raw body)
	req, err := http.NewRequest(http.MethodPost, "/v1/general/infra-tests/deprovision", bytes.NewBuffer([]byte("not valid json")))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	// Assert error response (should be 400 Bad Request or similar)
	assert.NotEqual(s.T(), http.StatusCreated, rr.Code)
}

func (s *GeneralTemporalTestSuite) TestInfraTestsDeprovision_MinimalFields() {
	// Set up mock expectation for successful workflow execution
	s.mockTC.EXPECT().ExecuteWorkflowInNamespace(
		gomock.Any(),  // ctx
		"infra-tests", // namespace
		gomock.Any(),  // options
		"Deprovision", // workflow
		gomock.Any(),  // args
	).Return(nil, nil)

	// Make request with minimal required fields (testing that optional fields work)
	reqBody := InfraTestsDeprovisionRequest{
		SandboxName: "test-sandbox",
		CanaryID:    "canary-123",
	}
	rr := s.makeRequest(http.MethodPost, "/v1/general/infra-tests/deprovision", reqBody)

	// Assert response
	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	// Parse response
	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "ok", response["status"])
}
