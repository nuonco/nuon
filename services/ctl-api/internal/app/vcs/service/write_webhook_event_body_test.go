package service

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadBodyCapsWebhookPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for name, tc := range map[string]struct {
		size          int
		contentLength int64
		wantErr       error
	}{
		"at the cap is read":                   {size: maxWebhookBody, contentLength: -1},
		"declared length over cap is rejected": {size: 1, contentLength: maxWebhookBody + 1, wantErr: errWebhookBodyTooLarge},
		"streamed body over cap is rejected":   {size: maxWebhookBody + 1, contentLength: -1, wantErr: errWebhookBodyTooLarge},
	} {
		t.Run(name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(make([]byte, tc.size)))
			ctx.Request.ContentLength = tc.contentLength

			body, err := readBody(ctx)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, body, tc.size)
		})
	}
}

func TestWriteWebhookEventRejectsMissingSignatureBeforeLookup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{}`)))
	ctx.Request.Header.Set("X-Hub-Signature-256", "00")

	// a nil db would panic if the handler reached the subscription lookup
	require.NotPanics(t, func() { (&service{}).WriteWebhookEvent(ctx) })
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
