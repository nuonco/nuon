package health

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
)

type Service struct {
	cfg     *internal.Config
	db      *gorm.DB
	chDB    *gorm.DB
	tclient temporalclient.Client
	mw      metrics.Writer
}

var _ api.Service = (*Service)(nil)

func (s *Service) RegisterPublicRoutes(api *gin.Engine) error {
	api.GET("/livez", s.GetLivezHandler)
	api.GET("/readyz", s.GetReadyzHandler)
	api.GET("/version", s.GetVersionHandler)

	return nil
}

func (s *Service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/livez", s.GetLivezHandler)
	api.GET("/readyz", s.GetReadyzHandler)
	api.GET("/version", s.GetVersionHandler)

	return nil
}

func (s *Service) RegisterRunnerRoutes(api *gin.Engine) error {
	api.GET("/livez", s.GetLivezHandler)
	api.GET("/readyz", s.GetReadyzHandler)
	api.GET("/version", s.GetVersionHandler)

	return nil
}

type Params struct {
	fx.In

	Cfg     *internal.Config
	DB      *gorm.DB `name:"psql"`
	CHDB    *gorm.DB `name:"ch"`
	TClient temporalclient.Client
	MW      metrics.Writer
}

func New(params Params) (*Service, error) {
	return &Service{
		cfg:     params.Cfg,
		db:      params.DB,
		chDB:    params.CHDB,
		tclient: params.TClient,
		mw:      params.MW,
	}, nil
}
