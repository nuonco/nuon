package repos

import (
	"context"

	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_admin_repo.go -source=admin_repo.go -package=repos
type AdminRepo interface {
	GetSandboxVersion(context.Context, uuid.UUID) (*models.SandboxVersion, error)
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

func (a adminRepo) GetSandboxVersion(ctx context.Context, sandboxID uuid.UUID) (*models.SandboxVersion, error) {
	var sandboxVersion models.SandboxVersion
	if err := a.db.WithContext(ctx).
		Preload(clause.Associations).
		First(&sandboxVersion, "id = ?", sandboxID).Error; err != nil {
		return nil, err
	}
	return &sandboxVersion, nil
}

func (a adminRepo) UpsertSandboxVersion(ctx context.Context, sandboxVersion *models.SandboxVersion) (*models.SandboxVersion, error) {
	// if id was provided check that the sandbox exists
	if sandboxVersion.ID != uuid.Nil {
		_, err := a.GetSandboxVersion(ctx, sandboxVersion.ID)
		if err != nil {
			return nil, err
		}
	}

	// upsert sandbox
	if err := a.db.WithContext(ctx).Save(sandboxVersion).Find(&sandboxVersion).Error; err != nil {
		return nil, err
	}

	return sandboxVersion, nil
}
