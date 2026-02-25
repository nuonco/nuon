package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *GeneralRunnerTestSuite) TestPublishMetrics_Success() {
	// Create a valid metrics request with one incr metric
	req := []PublishMetricInput{
		{
			Incr: &metrics.Incr{
				Name: "test.metric",
				Tags: []string{"env:test"},
			},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/general/metrics", req)

	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	// Verify response body
	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "ok", resp["status"])
}

func (s *GeneralRunnerTestSuite) TestPublishMetrics_MultipleMetricTypes() {
	// Create a request with multiple metric types
	req := []PublishMetricInput{
		{
			Incr: &metrics.Incr{
				Name: "test.counter.incr",
				Tags: []string{"env:test", "type:incr"},
			},
		},
		{
			Decr: &metrics.Decr{
				Name: "test.counter.decr",
				Tags: []string{"env:test", "type:decr"},
			},
		},
		{
			Timing: &metrics.Timing{
				Name:  "test.timing",
				Value: 1234,
				Tags:  []string{"env:test", "type:timing"},
			},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/general/metrics", req)

	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	// Verify response body
	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "ok", resp["status"])
}

func (s *GeneralRunnerTestSuite) TestPublishMetrics_EmptyArray() {
	// Empty array should succeed with no metrics to write
	req := []PublishMetricInput{}

	rr := s.makeRequest(http.MethodPost, "/v1/general/metrics", req)

	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	// Verify response body
	var resp map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "ok", resp["status"])
}

func (s *GeneralRunnerTestSuite) TestPublishMetrics_InvalidBody() {
	// Send invalid JSON
	req, err := http.NewRequest(http.MethodPost, "/v1/general/metrics", nil)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")
	req.Body = http.NoBody

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)

	// Should return error response
	require.Equal(s.T(), http.StatusBadRequest, rr.Code)
	s.T().Logf("Invalid body error - Status: %d, Body: %s", rr.Code, rr.Body.String())
}
