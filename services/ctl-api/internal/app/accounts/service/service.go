package service

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
)

type Params struct {
	fx.In

	DB *gorm.DB `name:"psql"`
}

type service struct {
	db *gorm.DB
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		db: params.DB,
	}
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	// accounts
	api.GET("/v1/account", s.GetCurrentAccount)

	// user journeys
	api.GET("/v1/account/user-journeys", s.GetUserJourneys)
	api.POST("/v1/account/user-journeys", s.CreateUserJourney)
	api.PATCH("/v1/account/user-journeys/:journey_name/steps/:step_name", s.UpdateUserJourneyStep)
	api.POST("/v1/account/user-journeys/:journey_name/reset", s.ResetUserJourney)
	api.POST("/v1/account/user-journeys/:journey_name/complete", s.CompleteUserJourney)
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	// No internal routes for accounts service at this time
	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	// No runner routes for accounts service at this time
	return nil
}

func (s *service) getAccount(ctx *gin.Context, accountID string) (*app.Account, error) {
	var account app.Account

	res := s.db.WithContext(ctx).
		Preload("Roles").
		Preload("Roles.Policies").
		Preload("Roles.Org").
		Where("id = ?", accountID).
		First(&account)

	if res.Error != nil {
		return nil, res.Error
	}

	return &account, nil
}
