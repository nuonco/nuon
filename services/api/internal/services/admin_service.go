package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_admin_service.go -source=admin_service.go -package=services
type AdminService interface {
	UpsertSandboxVersion(context.Context, models.SandboxVersionInput) (*models.SandboxVersion, error)
}

var _ AdminService = (*adminService)(nil)

type adminService struct {
	log  *zap.Logger
	repo repos.AdminRepo
}

func NewAdminService(db *gorm.DB, log *zap.Logger) *adminService {
	adminRepo := repos.NewAdminRepo(db)
	return &adminService{
		log:  log,
		repo: adminRepo,
	}
}

func (a adminService) UpsertSandboxVersion(ctx context.Context, input models.SandboxVersionInput) (*models.SandboxVersion, error) {
	sandboxVersion := models.SandboxVersion{
		SandboxName:    input.SandboxName,
		SandboxVersion: input.SandboxVersion,
		TfVersion:      input.TfVersion,
	}
	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	sandboxVersionID, _ := uuid.Parse(input.ID)
	sandboxVersion.ID = sandboxVersionID

	updatedSandboxVersion, err := a.repo.UpsertSandboxVersion(ctx, &sandboxVersion)
	if err != nil {
		a.log.Error("failed to upsert sandbox version",
			zap.Any("sandboxVersion", sandboxVersion),
			zap.String("error", err.Error()))
		return nil, err
	}
	return updatedSandboxVersion, nil
}
