package repos

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_user_repo.go -source=user_repo.go -package=repos
type UserRepo interface {
	UpsertUserOrg(context.Context, string, string) (*models.UserOrg, error)
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

func (u userRepo) UpsertUserOrg(ctx context.Context, userID string, orgID string) (*models.UserOrg, error) {
	var uo models.UserOrg
	uo.UserID = userID
	uo.OrgID = orgID
	uo.ID = domains.NewUserID()

	fmt.Println(ctx.Value(temporal.ContextKey{}))
	if err := u.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&uo).Error; err != nil {
		return nil, err
	}
	return &uo, nil
}
