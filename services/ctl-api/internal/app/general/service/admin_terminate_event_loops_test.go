package service

import (
	"net/http"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *GeneralInternalTestSuite) TestAdminTerminateEventLoops() {
	testCases := []struct {
		name           string
		expectedStatus int
		validateFunc   func()
	}{
		{
			name:           "sends terminate event loops signal",
			expectedStatus: http.StatusCreated,
			validateFunc: func() {
				capturedSignals := s.mockEvClient.GetSignals()
				require.Len(s.T(), capturedSignals, 1, "expected exactly one signal to be sent")

				signal := capturedSignals[0]
				assert.Equal(s.T(), "general", signal.ID, "signal should be sent to 'general' ID")

				// Type assert to get the actual signal
				genSignal, ok := signal.Signal.(*signals.Signal)
				require.True(s.T(), ok, "signal should be of type *signals.Signal")
				assert.Equal(s.T(), signals.OperationTerminateEventLoops, genSignal.Type, "signal type should be OperationTerminateEventLoops")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset mock before test
			s.mockEvClient.Reset()

			// Make request
			rr := s.makeRequest(http.MethodPost, "/v1/general/terminate-event-loops", map[string]interface{}{})

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Validate signal
			if tc.validateFunc != nil {
				tc.validateFunc()
			}
		})
	}
}
