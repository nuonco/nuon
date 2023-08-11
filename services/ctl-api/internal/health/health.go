package health

import (
	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Service struct {
	gitRef string
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

func New(cfg *internal.Config) (*Service, error) {
	return &Service{
		gitRef: cfg.GitRef,
	}, nil
}
