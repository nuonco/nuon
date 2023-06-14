package services

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/powertoolsdev/mono/services/api/internal/utils"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_org_service.go -source=org_service.go -package=services
type OrgService interface {
	DeleteOrg(context.Context, string) (bool, error)
	GetOrg(context.Context, string) (*models.Org, error)
	UpsertOrg(context.Context, models.OrgInput) (*models.Org, error)
	UserOrgs(context.Context, string, *models.ConnectionOptions) ([]*models.Org, *utils.Page, error)
}

var _ OrgService = (*orgService)(nil)

type orgService struct {
	log            *zap.Logger
	repo           repos.OrgRepo
	userOrgUpdater repos.UserRepo
}

func NewOrgService(db *gorm.DB, tc tclient.Client, log *zap.Logger) *orgService {
	orgRepo := repos.NewOrgRepo(db)
	userRepo := repos.NewUserRepo(db)
	return &orgService{
		log:            log,
		repo:           orgRepo,
		userOrgUpdater: userRepo,
	}
}

func (o orgService) DeleteOrg(ctx context.Context, inputID string) (bool, error) {
	deleted, err := o.repo.Delete(ctx, inputID)
	if err != nil {
		o.log.Error("failed to delete org",
			zap.String("orgID", inputID),
			zap.String("error", err.Error()))
		return false, err
	}
	if !deleted {
		return false, nil
	}

	return true, nil
}

func (o *orgService) GetOrg(ctx context.Context, inputID string) (*models.Org, error) {
	org, err := o.repo.Get(ctx, inputID)
	if err != nil {
		o.log.Error("failed to retrieve org",
			zap.String("orgID", inputID),
			zap.String("error", err.Error()))
		return nil, err
	}

	return org, nil
}

func (o *orgService) updateOrg(ctx context.Context, input models.OrgInput) (*models.Org, error) {
	org, err := o.GetOrg(ctx, *input.ID)
	if err != nil {
		o.log.Error("failed to retrieve org",
			zap.Any("input", input),
			zap.String("error", err.Error()))
		return nil, fmt.Errorf("failed to retrieve org: %w", err)
	}

	org.IsNew = false
	if input.GithubInstallID != nil {
		org.GithubInstallID = *input.GithubInstallID
	}
	if input.Name != "" {
		org.Name = input.Name
	}

	updatedOrg, err := o.repo.Update(ctx, org)
	if err != nil {
		o.log.Error("failed to update org",
			zap.Any("org", org),
			zap.String("error", err.Error()))
		return nil, err
	}
	return updatedOrg, nil
}

func (o *orgService) UpsertOrg(ctx context.Context, input models.OrgInput) (*models.Org, error) {
	if input.ID != nil {
		return o.updateOrg(ctx, input)
	}

	org := &models.Org{
		CreatedByID: input.OwnerID,
		IsNew:       true,
		Name:        input.Name,
	}
	org.ID = domains.NewOrgID()
	if input.GithubInstallID != nil {
		org.GithubInstallID = *input.GithubInstallID
	}

	// overrideID is used to override the ID set on this object, and should only be used locally during development.
	if input.OverrideID != nil {
		org.ID = *input.OverrideID
	}

	org, err := o.repo.Create(ctx, org)
	if err != nil {
		o.log.Error("failed to create org",
			zap.Any("org", org),
			zap.String("error", err.Error()))
		return nil, err
	}

	_, err = o.userOrgUpdater.UpsertUserOrg(ctx, input.OwnerID, org.ID)
	if err != nil {
		o.log.Error("failed to upsert org member",
			zap.String("userID", input.OwnerID),
			zap.String("orgID", org.ID),
			zap.String("error", err.Error()))
		return org, err
	}

	return org, err
}

func (o *orgService) UserOrgs(ctx context.Context, inputID string, options *models.ConnectionOptions) ([]*models.Org, *utils.Page, error) {
	org, pg, err := o.repo.GetPageByUser(ctx, inputID, options)
	if err != nil {
		o.log.Error("failed to get user's orgs",
			zap.String("userID", inputID),
			zap.Any("options", *options),
			zap.String("error", err.Error()))
		return nil, nil, err
	}
	return org, pg, nil
}
