package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfigSchemaSourceDeprecation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/v1/general/config-schema", (&service{}).GetConfigSchema)

	testCases := []struct {
		name              string
		queryParams       string
		expectDeprecation bool
	}{
		{"type param", "?type=runner", false},
		{"source param", "?source=runner", true},
		{"underscore source param", "?source=container_image", true},
		{"type wins over source", "?type=helm&source=runner", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/v1/general/config-schema"+tc.queryParams, nil)
			router.ServeHTTP(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)

			var result map[string]interface{}
			require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &result))

			if tc.expectDeprecation {
				assert.Equal(t, "true", rr.Header().Get("Deprecation"))
				assert.Contains(t, rr.Header().Get("Warning"), "deprecated")
				assert.Contains(t, result["$comment"], "deprecated")
			} else {
				assert.Empty(t, rr.Header().Get("Deprecation"))
				assert.Empty(t, rr.Header().Get("Warning"))
				assert.Empty(t, result["$comment"])
			}
		})
	}
}

func (s *GeneralPublicTestSuite) TestGetConfigSchema() {
	testCases := []struct {
		name              string
		queryParams       string
		expectedCode      int
		expectDeprecation bool
	}{
		{
			name:         "returns schema for action type",
			queryParams:  "?type=action",
			expectedCode: http.StatusOK,
		},
		{
			name:         "returns schema for helm type",
			queryParams:  "?type=helm",
			expectedCode: http.StatusOK,
		},
		{
			name:         "returns schema for terraform type",
			queryParams:  "?type=terraform",
			expectedCode: http.StatusOK,
		},
		{
			name:              "accepts source as a deprecated alias for type",
			queryParams:       "?source=runner",
			expectedCode:      http.StatusOK,
			expectDeprecation: true,
		},
		{
			name:              "accepts underscore variants of hyphenated types",
			queryParams:       "?source=container_image",
			expectedCode:      http.StatusOK,
			expectDeprecation: true,
		},
		{
			name:         "returns schema for job type",
			queryParams:  "?type=job",
			expectedCode: http.StatusOK,
		},
		{
			name:         "prefers type over source when both are provided",
			queryParams:  "?type=helm&source=unknown-type",
			expectedCode: http.StatusOK,
		},
		{
			name:         "returns error for unknown source",
			queryParams:  "?source=unknown-type",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "returns error for unknown full type",
			queryParams:  "?type=full",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "returns error when type is missing",
			queryParams:  "",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "returns error for unknown schema type",
			queryParams:  "?type=unknown-type",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			path := "/v1/general/config-schema" + tc.queryParams
			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				// Verify response is valid JSON - schema structure varies by type
				var result map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &result)
				require.NoError(s.T(), err)
				// Just verify it's a non-empty valid JSON object
				assert.NotEmpty(s.T(), result, "Schema response should be non-empty")

				if tc.expectDeprecation {
					assert.Equal(s.T(), "true", rr.Header().Get("Deprecation"))
					assert.Contains(s.T(), rr.Header().Get("Warning"), "deprecated")
					assert.Contains(s.T(), result["$comment"], "deprecated")
				} else {
					assert.Empty(s.T(), rr.Header().Get("Deprecation"))
					assert.Empty(s.T(), result["$comment"])
				}
			}
		})
	}
}
