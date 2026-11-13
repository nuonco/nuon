package size

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type middleware struct {
	l   *zap.Logger
	cfg *internal.Config
}

func (m *middleware) Name() string {
	return "size"
}

func (m *middleware) Handler() gin.HandlerFunc {
	return LimitRequestBody(m.cfg.MaxRequestSize, m.l)
}

// LimitRequestBody caps request bodies at limit and warns once a request reads past half of it, so a
// legitimate workload approaching the cap shows up before it is rejected.
func LimitRequestBody(limit int64, l *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request == nil || c.Request.Body == nil || limit <= 0 {
			c.Next()
			return
		}
		if c.Request.ContentLength > limit {
			warn(l, c, limit, c.Request.ContentLength, true)
			c.AbortWithStatus(http.StatusRequestEntityTooLarge)
			return
		}

		body := &countingReader{r: http.MaxBytesReader(c.Writer, c.Request.Body, limit)}
		c.Request.Body = body
		c.Next()

		if body.n > limit/2 {
			warn(l, c, limit, body.n, body.n >= limit)
		}
	}
}

func warn(l *zap.Logger, c *gin.Context, limit, size int64, rejected bool) {
	if l == nil {
		return
	}
	l.Warn("request body near size limit",
		zap.String("method", c.Request.Method),
		zap.String("route", c.FullPath()),
		zap.Int64("bytes", size),
		zap.Int64("limit", limit),
		zap.Bool("rejected", rejected),
	)
}

type countingReader struct {
	r io.ReadCloser
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

func (c *countingReader) Close() error {
	return c.r.Close()
}

type Params struct {
	fx.In

	Cfg *internal.Config
	L   *zap.Logger
}

func New(params Params) *middleware {
	return &middleware{
		l:   params.L,
		cfg: params.Cfg,
	}
}
