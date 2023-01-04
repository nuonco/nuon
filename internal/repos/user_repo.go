package repos

import (
	"context"

	"github.com/google/uuid"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_user_repo.go -source=user_repo.go -package=repos
type UserRepo interface {
	Upsert(context.Context, *models.User) (*models.User, error)
	Delete(context.Context, uuid.UUID) (bool, error)
	Get(context.Context, uuid.UUID) (*models.User, error)
	GetByOrg(context.Context, uuid.UUID) ([]*models.User, error)
	GetPageByOrg(context.Context, uuid.UUID, *models.ConnectionOptions) ([]*models.User, *utils.Page, error)
	GetPageAll(context.Context, *models.ConnectionOptions) ([]*models.User, *utils.Page, error)
	GetByEmail(context.Context, string) (*models.User, error)
	GetByExternalID(context.Context, string) (*models.User, error)
	UpsertUserOrg(context.Context, uuid.UUID, uuid.UUID) (*models.UserOrg, error)
}

var _ UserRepo = (*userRepo)(nil)

func NewUserRepo(db *gorm.DB) userRepo {
	return userRepo{
		db: db,
	}
}

type userRepo struct {
	db *gorm.DB
}

func (u userRepo) Delete(ctx context.Context, userID uuid.UUID) (bool, error) {
	user := models.User{
		Model: models.Model{
			ID: userID,
		},
	}
	if err := u.db.WithContext(ctx).Model(&user).Association("Orgs").Clear(); err != nil {
		return false, err
	}

	if err := u.db.WithContext(ctx).Delete(&user, "id = ?", userID).Error; err != nil {
		return false, err
	}

	return true, nil
}

// TODO: replace this with a get, update and create for better control flow in the services layer
func (u userRepo) Upsert(ctx context.Context, user *models.User) (*models.User, error) {
	err := u.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "email"}},
		UpdateAll: true,
	}).Create(user).Error

	// we don't care if it's duplicate
	if err != nil && utils.IsDuplicateKeyError(err) {
		return user, nil
	} else if err != nil {
		return nil, err
	}

	return user, err
}

func (u userRepo) GetByOrg(ctx context.Context, orgID uuid.UUID) ([]*models.User, error) {
	var users []*models.User
	if err := u.db.WithContext(ctx).Where("id IN (?)", u.db.Table("user_orgs").Select("user_id").Where("org_id = ?", orgID)).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (u userRepo) GetPageByOrg(ctx context.Context, orgID uuid.UUID, options *models.ConnectionOptions) ([]*models.User, *utils.Page, error) {
	var users []*models.User
	tx := u.db.WithContext(ctx).Where("id IN (?)", u.db.Table("user_orgs").Select("user_id").Where("org_id = ?", orgID)).Find(&users)
	pg, c, err := utils.NewPaginator(options)
	if err != nil {
		return nil, nil, err
	}

	page, err := pg.Paginate(c, tx)
	if err != nil {
		return nil, nil, err
	}

	if err := page.Query(&users); err != nil {
		return nil, nil, err
	}

	return users, &page, nil
}

func (u userRepo) GetPageAll(ctx context.Context, options *models.ConnectionOptions) ([]*models.User, *utils.Page, error) {
	var users []*models.User
	tx := u.db.WithContext(ctx).Find(&users)
	pg, c, err := utils.NewPaginator(options)
	if err != nil {
		return nil, nil, err
	}

	page, err := pg.Paginate(c, tx)
	if err != nil {
		return nil, nil, err
	}

	if err := page.Query(&users); err != nil {
		return nil, nil, err
	}

	return users, &page, nil
}

func (u userRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := u.db.WithContext(ctx).Preload("Orgs").First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u userRepo) GetByExternalID(ctx context.Context, externalID string) (*models.User, error) {
	var user models.User
	if err := u.db.WithContext(ctx).Preload("Orgs").First(&user, "external_id = ?", externalID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u userRepo) Get(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := u.db.WithContext(ctx).Preload("Orgs").First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (u userRepo) UpsertUserOrg(ctx context.Context, userID, orgID uuid.UUID) (*models.UserOrg, error) {
	var uo models.UserOrg
	uo.UserID = userID
	uo.OrgID = orgID
	if err := u.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&uo).Error; err != nil {
		return nil, err
	}
	return &uo, nil
}
