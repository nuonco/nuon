package services

import (
	"context"

	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/repos"
	"github.com/powertoolsdev/api/internal/utils"
	"gorm.io/gorm"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_user_service.go -source=user_service.go -package=services
type UserService interface {
	UpsertUser(context.Context, models.UserInput) (*models.User, error)
	DeleteUser(context.Context, string) (bool, error)
	GetOrgUsers(context.Context, string, *models.ConnectionOptions) ([]*models.User, *utils.Page, error)
	GetUserByEmail(context.Context, string) (*models.User, error)
	GetUserByExternalID(context.Context, string) (*models.User, error)
	GetAllUsers(context.Context, *models.ConnectionOptions) ([]*models.User, *utils.Page, error)
	GetUser(context.Context, string) (*models.User, error)
	UpsertUserOrg(context.Context, models.UserOrgInput) (*models.UserOrg, error)
}

var _ UserService = (*userService)(nil)

type userService struct {
	repo repos.UserRepo
}

func NewUserService(db *gorm.DB) *userService {
	userRepo := repos.NewUserRepo(db)
	return &userService{
		repo: userRepo,
	}
}

func (u userService) DeleteUser(ctx context.Context, inputID string) (bool, error) {
	userID, err := parseID(inputID)
	if err != nil {
		return false, err
	}

	return u.repo.Delete(ctx, userID)
}

// UpsertUser: upsert a single user
func (u *userService) UpsertUser(ctx context.Context, input models.UserInput) (*models.User, error) {
	var user models.User
	user.FirstName = input.FirstName
	user.LastName = input.LastName
	user.Email = input.Email
	user.ExternalID = input.ExternalID

	if input.ID != nil {
		userID, err := parseID(*input.ID)
		if err != nil {
			return nil, err
		}
		user.ID = userID
	}
	return u.repo.Upsert(ctx, &user)
}

// GetOrgUsers: return all users for an org
func (u *userService) GetOrgUsers(ctx context.Context, inputID string, options *models.ConnectionOptions) ([]*models.User, *utils.Page, error) {
	orgID, err := parseID(inputID)
	if err != nil {
		return nil, nil, err
	}
	return u.repo.GetPageByOrg(ctx, orgID, options)
}

func (u userService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return u.repo.GetByEmail(ctx, email)
}

func (u userService) GetUserByExternalID(ctx context.Context, externalID string) (*models.User, error) {
	return u.repo.GetByExternalID(ctx, externalID)
}

func (u userService) GetAllUsers(ctx context.Context, options *models.ConnectionOptions) ([]*models.User, *utils.Page, error) {
	return u.repo.GetPageAll(ctx, options)
}

func (u userService) GetUser(ctx context.Context, id string) (*models.User, error) {
	userID, err := parseID(id)
	if err != nil {
		return nil, err
	}
	return u.repo.Get(ctx, userID)
}

func (u userService) UpsertUserOrg(ctx context.Context, input models.UserOrgInput) (*models.UserOrg, error) {
	userID, err := parseID(input.UserID)
	if err != nil {
		return nil, err
	}
	orgID, err := parseID(input.OrgID)
	if err != nil {
		return nil, err
	}

	return u.repo.UpsertUserOrg(ctx, userID, orgID)
}
