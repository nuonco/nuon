package repos

import (
	"context"

	"github.com/powertoolsdev/mono/services/api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_instance_repo.go -source=instance_repo.go -package=repos
type InstanceRepo interface {
	Get(context.Context, string) (*models.Instance, error)
	Create(context.Context, []*models.Instance) ([]*models.Instance, error)
	Delete(context.Context, string) (bool, error)
}

var _ InstanceRepo = (*instanceRepo)(nil)

func NewInstanceRepo(db *gorm.DB) instanceRepo {
	return instanceRepo{
		db: db,
	}
}

type instanceRepo struct {
	db *gorm.DB
}

func (i instanceRepo) Get(ctx context.Context, instanceID string) (*models.Instance, error) {
	var instance models.Instance
	if err := i.db.WithContext(ctx).
		Preload(clause.Associations).
		First(&instance, "id = ?", instanceID).Error; err != nil {
		return nil, err
	}
	return &instance, nil
}

func (i instanceRepo) Create(ctx context.Context, instances []*models.Instance) ([]*models.Instance, error) {
	if err := i.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "component_id"}, {Name: "install_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"deployment_id", "install_id"}),
		}).Create(instances).Error; err != nil {
		return nil, err
	}

	return instances, nil
}

func (i instanceRepo) Delete(ctx context.Context, instanceID string) (bool, error) {
	var instance models.Instance
	if err := i.db.WithContext(ctx).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Delete(&instance, "id = ?", instanceID).Error; err != nil {
		return false, err
	}
	return instance.ID != "", nil
}
