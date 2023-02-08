package repos

import (
	"context"

	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
	"gorm.io/gorm"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_user_repo.go -source=user_repo.go -package=repos
type AdminRepo interface {
	UpsertSandboxVersion(context.Context, *models.SandboxVersion) (*models.SandboxVersion, error)
}

var _ AdminRepo = (*adminRepo)(nil)

func NewAdminRepo(db *gorm.DB) adminRepo {
	return adminRepo{
		db: db,
	}
}

type adminRepo struct {
	db *gorm.DB
}

func (a adminRepo) UpsertSandboxVersion(ctx context.Context, sandboxVersion *models.SandboxVersion) (*models.SandboxVersion, error) {
	if sandboxVersion.ID != uuid.Nil {
		if err := a.db.WithContext(ctx).Updates(sandboxVersion).Find(&sandboxVersion).Error; err != nil {
			return nil, err
		}
	} else {
		if err := a.db.WithContext(ctx).Create(sandboxVersion).Find(&sandboxVersion).Error; err != nil {
			return nil, err
		}
	}

	return sandboxVersion, nil
}
