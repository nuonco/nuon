package health

import (
	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"gorm.io/gorm"
)

type Service struct {
	cfg *internal.Config
	db  *gorm.DB
}

func (s *Service) RegisterRoutes(api *gin.Engine) error {
	api.GET("/livez", s.GetLivezHandler)
	api.GET("/readyz", s.GetReadyzHandler)
	api.GET("/version", s.GetVersionHandler)
	api.GET("/healthz", s.GetHealthzHandler)

	return nil
}

func (s *Service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/livez", s.GetLivezHandler)
	api.GET("/readyz", s.GetReadyzHandler)
	api.GET("/version", s.GetVersionHandler)
	api.GET("/healthz", s.GetHealthzHandler)

	return nil
}

func New(cfg *internal.Config, db *gorm.DB) (*Service, error) {
	return &Service{
		cfg: cfg,
		db:  db,
	}, nil
}
