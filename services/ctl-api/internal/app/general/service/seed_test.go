package service

import (
	"net/http"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *GeneralInternalTestSuite) TestSeed() {
	testCases := []struct {
		name           string
		expectedStatus int
		validateFunc   func()
	}{
		{
			name:           "sends terminate and seed signals",
			expectedStatus: http.StatusOK, // Note: Handler does NOT return JSON, so 200 with no body
			validateFunc: func() {
				// Verify both signals were sent
				capturedSignals := s.mockEvClient.GetSignals()
				require.Len(s.T(), capturedSignals, 2, "expected exactly two signals to be sent")

				// First signal should be OperationTerminateEventLoops
				signal1 := capturedSignals[0]
				assert.Equal(s.T(), "general", signal1.ID, "first signal should be sent to 'general' ID")
				genSignal1, ok := signal1.Signal.(*signals.Signal)
				require.True(s.T(), ok, "first signal should be of type *signals.Signal")
				assert.Equal(s.T(), signals.OperationTerminateEventLoops, genSignal1.Type, "first signal type should be OperationTerminateEventLoops")

				// Second signal should be OperationSeed
				signal2 := capturedSignals[1]
				assert.Equal(s.T(), "general", signal2.ID, "second signal should be sent to 'general' ID")
				genSignal2, ok := signal2.Signal.(*signals.Signal)
				require.True(s.T(), ok, "second signal should be of type *signals.Signal")
				assert.Equal(s.T(), signals.OperationSeed, genSignal2.Type, "second signal type should be OperationSeed")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset mock before test
			s.mockEvClient.Reset()

			// Make request
			rr := s.makeRequest(http.MethodPost, "/v1/general/seed", map[string]interface{}{})

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Validate signals
			if tc.validateFunc != nil {
				tc.validateFunc()
			}
		})
	}
}
