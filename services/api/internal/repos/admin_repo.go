package repos

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"gorm.io/gorm"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_admin_repo.go -source=admin_repo.go -package=repos
type AdminRepo interface {
	GetLatestSandboxVersion(context.Context) (*models.SandboxVersion, error)
	GetSandboxVersionByID(context.Context, string) (*models.SandboxVersion, error)
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

func (a adminRepo) GetSandboxVersionByID(ctx context.Context, sandboxID string) (*models.SandboxVersion, error) {
	var sandboxVersion models.SandboxVersion

	// SELECT * from sandbox_versions WHERE ID = sandboxID
	if err := a.db.WithContext(ctx).
		First(&sandboxVersion, "id = ?", sandboxID).Error; err != nil {
		return nil, err
	}
	return &sandboxVersion, nil
}

func (a adminRepo) GetLatestSandboxVersion(ctx context.Context) (*models.SandboxVersion, error) {
	var sandboxVersion models.SandboxVersion

	// SELECT * FROM sandbox_versions ORDER BY created_at desc LIMIT 1;
	if err := a.db.WithContext(ctx).
		Order("created_at desc").
		Limit(1).
		Find(&sandboxVersion).Error; err != nil {
		return nil, err
	}
	return &sandboxVersion, nil
}

func (a adminRepo) UpsertSandboxVersion(ctx context.Context, sandboxVersion *models.SandboxVersion) (*models.SandboxVersion, error) {
	// if id was provided check that the sandbox exists
	if sandboxVersion.ID != "" {
		_, err := a.GetSandboxVersionByID(ctx, sandboxVersion.ID)
		if err != nil {
			return nil, err
		}
	} else {
		sandboxVersion.ID = domains.NewSandboxID()
	}

	// upsert sandbox
	if err := a.db.WithContext(ctx).Save(sandboxVersion).Find(&sandboxVersion).Error; err != nil {
		return nil, err
	}

	return sandboxVersion, nil
}
