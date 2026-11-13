package size

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

const testLimit = 1024

func serve(t *testing.T, size int, chunked bool) (*httptest.ResponseRecorder, error, *observer.ObservedLogs) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	core, logs := observer.New(zap.WarnLevel)

	var readErr error
	r := gin.New()
	r.Use(LimitRequestBody(testLimit, zap.New(core)))
	r.POST("/", func(c *gin.Context) {
		_, readErr = io.ReadAll(c.Request.Body)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(make([]byte, size)))
	if chunked {
		req.ContentLength = -1
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec, readErr, logs
}

func TestLimitRequestBody(t *testing.T) {
	t.Run("small body passes without a warning", func(t *testing.T) {
		rec, err, logs := serve(t, testLimit/4, false)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Zero(t, logs.Len())
	})

	t.Run("body past half the cap passes and warns", func(t *testing.T) {
		rec, err, logs := serve(t, testLimit*3/4, true)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		require.Equal(t, 1, logs.Len())
		assert.Equal(t, false, logs.All()[0].ContextMap()["rejected"])
	})

	t.Run("declared length over cap is rejected before the handler", func(t *testing.T) {
		rec, err, logs := serve(t, testLimit+1, false)
		require.NoError(t, err) // handler never ran
		assert.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
		require.Equal(t, 1, logs.Len())
		assert.Equal(t, true, logs.All()[0].ContextMap()["rejected"])
	})

	t.Run("chunked body over cap stops at the limit", func(t *testing.T) {
		_, err, logs := serve(t, testLimit*4, true)
		var tooLarge *http.MaxBytesError
		require.True(t, errors.As(err, &tooLarge))
		require.Equal(t, 1, logs.Len())
		assert.Equal(t, true, logs.All()[0].ContextMap()["rejected"])
	})
}
