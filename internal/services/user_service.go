package services

import (
	"context"

	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/repos"
	"gorm.io/gorm"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_user_service.go -source=user_service.go -package=services
type UserService interface {
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

func (u userService) UpsertUserOrg(ctx context.Context, input models.UserOrgInput) (*models.UserOrg, error) {
	orgID, err := parseID(input.OrgID)
	if err != nil {
		return nil, err
	}

	return u.repo.UpsertUserOrg(ctx, input.UserID, orgID)
}
