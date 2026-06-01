package service

import (
	"encoding/json"
	"net/http"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals/promotion"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *GeneralInternalTestSuite) TestRestartGeneralEventLoop() {
	testCases := []struct {
		name           string
		expectedStatus int
		validateFunc   func(respBody string)
	}{
		{
			name:           "sends restart signal and returns ok",
			expectedStatus: http.StatusCreated,
			validateFunc: func(respBody string) {
				// Verify response structure
				var resp map[string]string
				err := json.Unmarshal([]byte(respBody), &resp)
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "ok", resp["status"], "status should be ok")
				assert.Equal(s.T(), string(promotion.SignalType), resp["type"], "type should be promotion")

				// Verify signal was sent
				capturedSignals := tests.GetQueueSignals(s.T(), s.service.DB)
				require.Len(s.T(), capturedSignals, 1, "expected exactly one signal to be sent")

				qs := capturedSignals[0]
				assert.Equal(s.T(), promotion.SignalType, qs.Type, "signal type should be promotion")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset mock before test

			// Make request
			rr := s.makeRequest(http.MethodPost, "/v1/general/restart-event-loop", map[string]interface{}{})

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Validate response and signal
			if tc.validateFunc != nil {
				tc.validateFunc(rr.Body.String())
			}
		})
	}
}
