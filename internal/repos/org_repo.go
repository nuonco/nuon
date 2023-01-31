package repos

import (
	"context"

	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrgRepo interface {
	Create(context.Context, *models.Org) (*models.Org, error)
	Delete(context.Context, uuid.UUID) (bool, error)
	Get(context.Context, uuid.UUID) (*models.Org, error)
	GetPageByUser(context.Context, string, *models.ConnectionOptions) ([]*models.Org, *utils.Page, error)
	SetWorkflowID(context.Context, uuid.UUID, string) error
	QueryAll(context.Context) *gorm.DB
}

var _ OrgRepo = (*orgRepo)(nil)

func NewOrgRepo(db *gorm.DB) orgRepo {
	return orgRepo{
		db: db,
	}
}

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_org_repo.go -source=org_repo.go -package=repos
type orgRepo struct {
	db *gorm.DB
}

func (o orgRepo) Create(ctx context.Context, org *models.Org) (*models.Org, error) {
	origID := org.ID
	err := o.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}).Create(&org).Error
	if err != nil {
		return nil, err
	}
	org.IsNew = org.ID != origID
	return org, nil
}

func (o orgRepo) Get(ctx context.Context, orgID uuid.UUID) (*models.Org, error) {
	var org models.Org
	err := o.db.WithContext(ctx).First(&org, "id = ?", orgID).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (o orgRepo) GetPageByUser(ctx context.Context, userID string, opts *models.ConnectionOptions) ([]*models.Org, *utils.Page, error) {
	var orgs []*models.Org
	tx := o.db.WithContext(ctx).Where("id IN (?)", o.db.Table("user_orgs").Select("org_id").Where("user_id = ?", userID)).Find(&orgs)
	pg, c, err := utils.NewPaginator(opts)
	if err != nil {
		return nil, nil, err
	}

	page, err := pg.Paginate(c, tx)
	if err != nil {
		return nil, nil, err
	}

	if err := page.Query(&orgs); err != nil {
		return nil, nil, err
	}

	return orgs, &page, nil
}

func (o orgRepo) Delete(ctx context.Context, orgID uuid.UUID) (bool, error) {
	var org models.Org
	if err := o.db.WithContext(ctx).Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).Delete(&org, "id = ?", orgID).Error; err != nil {
		return false, err
	}

	return org.ID != uuid.Nil, nil
}

func (o orgRepo) QueryAll(ctx context.Context) *gorm.DB {
	return o.db.WithContext(ctx).Model(&models.Org{})
}

func (o orgRepo) SetWorkflowID(ctx context.Context, orgID uuid.UUID, workflowID string) error {
	return o.db.WithContext(ctx).Model(&models.Org{}).Where("id = ?", orgID).Update("workflow_id", workflowID).Error
}
