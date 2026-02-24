package service

import (
	"encoding/json"
	"net/http"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals"
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
				assert.Equal(s.T(), string(signals.OperationRestart), resp["type"], "type should be restart")

				// Verify signal was sent
				capturedSignals := s.mockEvClient.GetSignals()
				require.Len(s.T(), capturedSignals, 1, "expected exactly one signal to be sent")

				signal := capturedSignals[0]
				assert.Equal(s.T(), "general", signal.ID, "signal should be sent to 'general' ID")

				// Type assert to get the actual signal
				genSignal, ok := signal.Signal.(*signals.Signal)
				require.True(s.T(), ok, "signal should be of type *signals.Signal")
				assert.Equal(s.T(), signals.OperationRestart, genSignal.Type, "signal type should be OperationRestart")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset mock before test
			s.mockEvClient.Reset()

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
