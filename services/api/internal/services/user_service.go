package services

import (
	"context"

	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_user_service.go -source=user_service.go -package=services
type UserService interface {
	UpsertUserOrg(context.Context, models.UserOrgInput) (*models.UserOrg, error)
}

var _ UserService = (*userService)(nil)

type userService struct {
	log  *zap.Logger
	repo repos.UserRepo
}

func NewUserService(db *gorm.DB, log *zap.Logger) *userService {
	userRepo := repos.NewUserRepo(db)
	return &userService{
		log:  log,
		repo: userRepo,
	}
}

func (u userService) UpsertUserOrg(ctx context.Context, input models.UserOrgInput) (*models.UserOrg, error) {
	userOrg, err := u.repo.UpsertUserOrg(ctx, input.UserID, input.OrgID)
	if err != nil {
		u.log.Error("failed to upsert org member",
			zap.String("userID", input.UserID),
			zap.String("orgID", input.OrgID),
			zap.String("error", err.Error()))
		return nil, err
	}
	return userOrg, nil
}
