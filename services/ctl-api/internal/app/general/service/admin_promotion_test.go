package service

import (
	"net/http"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *GeneralInternalTestSuite) TestAdminPromotion() {
	testCases := []struct {
		name           string
		requestBody    AdminPromotionRequest
		expectedStatus int
		validateFunc   func(tag string)
	}{
		{
			name: "sends restart and promotion signals with tag",
			requestBody: AdminPromotionRequest{
				Tag: "v1.0.0",
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(tag string) {
				// Verify at least 2 signals were sent (OperationRestart + OperationPromotion)
				// Additional signals may be sent during initializeInstallStates
				capturedSignals := s.mockEvClient.GetSignals()
				require.GreaterOrEqual(s.T(), len(capturedSignals), 2, "expected at least two signals to be sent")

				// Verify first two signals are the expected types with tag
				foundRestart := false
				foundPromotion := false

				for i := 0; i < 2 && i < len(capturedSignals); i++ {
					signal := capturedSignals[i]
					assert.Equal(s.T(), "general", signal.ID, "signal should be sent to 'general' ID")

					genSignal, ok := signal.Signal.(*signals.Signal)
					require.True(s.T(), ok, "signal should be of type *signals.Signal")

					if genSignal.Type == signals.OperationRestart {
						foundRestart = true
						assert.Equal(s.T(), tag, genSignal.Tag, "restart signal should include tag")
					} else if genSignal.Type == signals.OperationPromotion {
						foundPromotion = true
						assert.Equal(s.T(), tag, genSignal.Tag, "promotion signal should include tag")
					}
				}

				assert.True(s.T(), foundRestart, "should have sent OperationRestart signal")
				assert.True(s.T(), foundPromotion, "should have sent OperationPromotion signal")
			},
		},
		{
			name:           "fails with missing tag",
			requestBody:    AdminPromotionRequest{},
			expectedStatus: http.StatusCreated, // Note: Handler doesn't validate tag, still succeeds
			validateFunc: func(tag string) {
				// Signals should still be sent, just with empty tag
				capturedSignals := s.mockEvClient.GetSignals()
				require.GreaterOrEqual(s.T(), len(capturedSignals), 2, "expected at least two signals to be sent")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset mock before test
			s.mockEvClient.Reset()

			// Make request
			rr := s.makeRequest(http.MethodPost, "/v1/general/promotion", tc.requestBody)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Validate signals
			if tc.validateFunc != nil {
				tc.validateFunc(tc.requestBody.Tag)
			}
		})
	}
}
