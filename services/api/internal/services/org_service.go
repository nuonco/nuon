package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/repos"
	"github.com/powertoolsdev/mono/services/api/internal/utils"
	"github.com/powertoolsdev/mono/services/api/internal/workflows"
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
	wkflowMgr      workflows.OrgWorkflowManager
}

func NewOrgService(db *gorm.DB, tc tclient.Client, log *zap.Logger) *orgService {
	orgRepo := repos.NewOrgRepo(db)
	userRepo := repos.NewUserRepo(db)
	return &orgService{
		log:            log,
		repo:           orgRepo,
		userOrgUpdater: userRepo,
		wkflowMgr:      workflows.NewOrgWorkflowManager(tc),
	}
}

func (o orgService) DeleteOrg(ctx context.Context, inputID string) (bool, error) {
	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	orgID, _ := uuid.Parse(inputID)

	deleted, err := o.repo.Delete(ctx, orgID)
	if err != nil {
		o.log.Error("failed to delete org",
			zap.String("orgID", orgID.String()),
			zap.String("error", err.Error()))
		return false, err
	}
	if !deleted {
		return false, nil
	}

	return true, nil
}

func (o *orgService) GetOrg(ctx context.Context, inputID string) (*models.Org, error) {
	// parsing the uuid while ignoring the error handling since we do this at protobuf level
	orgID, _ := uuid.Parse(inputID)

	org, err := o.repo.Get(ctx, orgID)
	if err != nil {
		o.log.Error("failed to retrieve org",
			zap.String("orgID", orgID.String()),
			zap.String("error", err.Error()))
		return nil, err
	}

	return org, nil
}

func (o *orgService) UpsertOrg(ctx context.Context, input models.OrgInput) (*models.Org, error) {
	org := &models.Org{
		Name: input.Name,
	}
	if input.ID != nil {
		// parsing the uuid while ignoring the error handling since we do this at protobuf level
		orgID, _ := uuid.Parse(*input.ID)
		org.ID = orgID
	} else {
		org.CreatedByID = input.OwnerID
	}

	if input.GithubInstallID != nil {
		org.GithubInstallID = *input.GithubInstallID
	}

	org, err := o.repo.Create(ctx, org)
	if err != nil {
		o.log.Error("failed to upsert org",
			zap.Any("org", org),
			zap.String("error", err.Error()))
		return nil, err
	}

	if !org.IsNew {
		return org, nil
	}

	_, err = o.userOrgUpdater.UpsertUserOrg(ctx, input.OwnerID, org.ID)
	if err != nil {
		o.log.Error("failed to upsert org member",
			zap.String("userID", input.OwnerID),
			zap.String("orgID", org.ID.String()),
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
