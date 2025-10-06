package httpbin

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mccutchen/go-httpbin/v2/httpbin"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
)

type Service struct {
	cfg     *internal.Config
	l       *zap.Logger
	httpbin *httpbin.HTTPBin
}

var _ api.Service = (*Service)(nil)

func (s *Service) RegisterPublicRoutes(api *gin.Engine) error {
	return s.registerRoutes(api)
}

func (s *Service) RegisterInternalRoutes(api *gin.Engine) error {
	return s.registerRoutes(api)
}

func (s *Service) RegisterRunnerRoutes(api *gin.Engine) error {
	return s.registerRoutes(api)
}

func (s *Service) registerRoutes(api *gin.Engine) error {
	if s.cfg.EnableHttpBinDebugEndpoints {
		httpbinGroup := api.Group("/httpbin")
		httpbinGroup.Any("/*any", s.Proxy)

		s.l.Info("registered httpbin routes", zap.String("prefix", "/httpbin"))
	}
	return nil
}

func (s *Service) Proxy(c *gin.Context) {

	switch c.Request.URL.Path {
	case "/httpbin/panic":
		panic("HTTPBIN force panic")
	case "/httpbin/oom":
		s.l.Info("Generating out-of-memory error...")
		// Generate up to 10GiB of data in 100MiB chunks
		var data [][]byte
		for i := range 100 {
			chunk := make([]byte, 100*1024*1024) // 100MiB
			for j := 0; j < len(chunk); j += 4096 {
				chunk[j] = byte(i % 256)
			}
			data = append(data, chunk)
			s.l.Info(fmt.Sprintf("Allocated %dMiB", (i+1)*100))
		}
	default:
		s.httpbin.Handler().ServeHTTP(c.Writer, c.Request)
	}
}

type Params struct {
	fx.In

	Cfg *internal.Config
	L   *zap.Logger
}

func New(params Params) (*Service, error) {
	// Create a new httpbin instance with default options
	h := httpbin.New(
		httpbin.WithPrefix("/httpbin"),
	)

	return &Service{
		cfg:     params.Cfg,
		l:       params.L,
		httpbin: h,
	}, nil
}
