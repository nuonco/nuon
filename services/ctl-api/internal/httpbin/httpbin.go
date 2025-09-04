package httpbin

import (
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
	if c.Request.URL.Path == "/httpbin/panic" {
		panic("HTTPBIN force panic")
	}
	s.httpbin.Handler().ServeHTTP(c.Writer, c.Request)
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
