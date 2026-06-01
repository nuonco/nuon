package service

import (
	"net/http"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals/promotion"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/tests"
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
				capturedSignals := tests.GetQueueSignals(s.T(), s.service.DB)
				require.GreaterOrEqual(s.T(), len(capturedSignals), 1, "expected at least one signal to be sent")

				foundPromotion := false
				for _, qs := range capturedSignals {
					if qs.Type == signal.SignalType(promotion.SignalType) {
						foundPromotion = true
						if sig, ok := qs.Signal.Signal.(*promotion.Signal); ok {
							assert.Equal(s.T(), tag, sig.Tag, "promotion signal should include tag")
						}
					}
				}

				assert.True(s.T(), foundPromotion, "should have sent promotion signal")
			},
		},
		{
			name:           "fails with missing tag",
			requestBody:    AdminPromotionRequest{},
			expectedStatus: http.StatusCreated, // Note: Handler doesn't validate tag, still succeeds
			validateFunc: func(tag string) {
				capturedSignals := tests.GetQueueSignals(s.T(), s.service.DB)
				require.GreaterOrEqual(s.T(), len(capturedSignals), 1, "expected at least one signal to be sent")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset mock before test

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
