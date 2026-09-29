package service

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"

	commonpb "go.temporal.io/api/common/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *GeneralInternalTestSuite) TestTemporalCodecDecode() {
	testCases := []struct {
		name           string
		requestBody    *commonpb.Payloads
		expectedStatus int
		validateFunc   func(rr *httptest.ResponseRecorder)
	}{
		{
			name: "decodes empty payloads successfully",
			requestBody: &commonpb.Payloads{
				Payloads: []*commonpb.Payload{},
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(s.T(), err)
				assert.NotNil(s.T(), resp, "response should be valid JSON")
			},
		},
		{
			name: "handles malformed request",
			requestBody: &commonpb.Payloads{
				Payloads: nil,
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				var resp commonpb.Payloads
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(s.T(), err)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			bodyBytes, err := json.Marshal(tc.requestBody)
			require.NoError(s.T(), err)

			req, err := http.NewRequest(http.MethodPost, "/v1/general/temporal-codec/decode", bytes.NewReader(bodyBytes))
			require.NoError(s.T(), err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			s.router.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			if tc.validateFunc != nil {
				tc.validateFunc(rr)
			}
		})
	}
}

func (s *GeneralInternalTestSuite) TestTemporalCodecDecodeRawJSON() {
	testCases := []struct {
		name           string
		requestJSON    string
		expectedStatus int
	}{
		{
			name:           "accepts raw JSON format",
			requestJSON:    `{"payloads":[]}`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req, err := http.NewRequest(http.MethodPost, "/v1/general/temporal-codec/decode", bytes.NewReader([]byte(tc.requestJSON)))
			require.NoError(s.T(), err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			s.router.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			if rr.Code == http.StatusOK {
				var resp map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(s.T(), err, "response should be valid JSON")
			}
		})
	}
}

func (s *GeneralInternalTestSuite) TestTemporalCodecDecodeInvalidBody() {
	testCases := []struct {
		name           string
		requestBody    io.Reader
		expectedStatus int
	}{
		{
			name:           "handles invalid JSON",
			requestBody:    bytes.NewReader([]byte(`{invalid json`)),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "handles empty body",
			requestBody:    bytes.NewReader([]byte(``)),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			req, err := http.NewRequest(http.MethodPost, "/v1/general/temporal-codec/decode", tc.requestBody)
			require.NoError(s.T(), err)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			s.router.ServeHTTP(rr, req)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)
		})
	}
}
