package services

import (
	"context"

	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_instance_service.go -source=instance_service.go -package=services
type InstanceService interface {
	GetInstancesByInstall(context.Context, string) ([]*models.Instance, error)
}

var _ InstanceService = (*instanceService)(nil)

type instanceService struct {
	log  *zap.Logger
	repo repos.InstanceRepo
}

func NewInstanceService(db *gorm.DB, log *zap.Logger) *instanceService {
	return &instanceService{
		log:  log,
		repo: repos.NewInstanceRepo(db),
	}
}

func (i *instanceService) GetInstancesByInstall(ctx context.Context, installID string) ([]*models.Instance, error) {
	instances, err := i.repo.ListByInstall(ctx, installID)
	if err != nil {
		i.log.Error("failed to retrieve install's instances",
			zap.String("installID", installID),
			zap.String("error", err.Error()))
		return nil, err
	}

	return instances, nil
}
