package health

import (
	"github.com/gin-gonic/gin"
	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"gorm.io/gorm"
)

type Service struct {
	cfg     *internal.Config
	db      *gorm.DB
	tclient temporalclient.Client
}

func (s *Service) RegisterRoutes(api *gin.Engine) error {
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

func New(cfg *internal.Config, db *gorm.DB, tclient temporalclient.Client,
) (*Service, error) {
	return &Service{
		cfg:     cfg,
		db:      db,
		tclient: tclient,
	}, nil
}
