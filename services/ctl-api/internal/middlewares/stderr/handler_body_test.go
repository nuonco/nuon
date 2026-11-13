package stderr

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type countingReader struct {
	r io.Reader
	n int
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += n
	return n, err
}

func newRouter(handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(New(zap.NewNop(), &internal.Config{}).Handler())
	r.POST("/", handler)
	return r
}

func TestErrorMiddlewareBuffersOnlyALoggingPrefix(t *testing.T) {
	body := &countingReader{r: bytes.NewReader(make([]byte, 4*maxLoggedRequestBody))}
	r := newRouter(func(c *gin.Context) {
		// a handler that rejects before reading, like a missing webhook signature
		c.Status(http.StatusUnauthorized)
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", body))

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.LessOrEqual(t, body.n, maxLoggedRequestBody)
}

func TestErrorMiddlewareHandlerSeesFullBody(t *testing.T) {
	want := bytes.Repeat([]byte("abc"), maxLoggedRequestBody)
	var got []byte
	r := newRouter(func(c *gin.Context) {
		got, _ = io.ReadAll(c.Request.Body)
		c.Status(http.StatusOK)
	})

	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(want)))
	assert.Equal(t, want, got)
}

func TestErrorMiddlewareMapsHandlerBodyLimitTo413(t *testing.T) {
	const routeCap = 1024
	r := newRouter(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, routeCap)
		if _, err := io.ReadAll(c.Request.Body); err != nil {
			c.Error(err)
			return
		}
		c.Status(http.StatusOK)
	})

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(make([]byte, 2*routeCap))))
	require.Equal(t, http.StatusRequestEntityTooLarge, rec.Code)
}
