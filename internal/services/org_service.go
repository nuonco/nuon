package services

import (
	"context"
	"fmt"

	"github.com/gosimple/slug"
	apicontext "github.com/powertoolsdev/api/internal/context"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/repos"
	"github.com/powertoolsdev/api/internal/utils"
	"github.com/powertoolsdev/api/internal/workflows"
	tclient "go.temporal.io/sdk/client"
	"gorm.io/gorm"
)

type OrgService struct {
	wkflowMgr      workflows.OrgWorkflowManager
	repo           repos.OrgRepo
	userOrgUpdater repos.UserRepo
}

func NewOrgService(db *gorm.DB, tc tclient.Client) *OrgService {
	orgRepo := repos.NewOrgRepo(db)
	userRepo := repos.NewUserRepo(db)
	return &OrgService{
		repo:           orgRepo,
		userOrgUpdater: userRepo,
		wkflowMgr:      workflows.NewOrgWorkflowManager(tc),
	}
}

func (o OrgService) DeleteOrg(ctx context.Context, inputID string) (bool, error) {
	orgID, err := parseID(inputID)
	if err != nil {
		return false, err
	}

	deleted, err := o.repo.Delete(ctx, orgID)
	if err != nil {
		return false, err
	}
	if !deleted {
		return false, nil
	}

	if err := o.wkflowMgr.Deprovision(ctx, orgID.String()); err != nil {
		return false, fmt.Errorf("unable to start deprovision workflow: %w", err)
	}
	return true, nil
}

func (o OrgService) GetOrg(ctx context.Context, id string) (*models.Org, error) {
	orgID, err := parseID(id)
	if err != nil {
		return nil, err
	}

	return o.repo.Get(ctx, orgID)
}

func (o OrgService) GetOrgBySlug(ctx context.Context, slug string) (*models.Org, error) {
	return o.repo.GetBySlug(ctx, slug)
}

func (o OrgService) UpsertOrg(ctx context.Context, input models.OrgInput) (*models.Org, error) {
	user, err := apicontext.GetUser(ctx)
	if err != nil {
		return nil, err
	}

	org := &models.Org{
		Name: input.Name,
		Slug: slug.Make(input.Name),
	}
	if input.ID != nil {
		orgID, er := parseID(*input.ID)
		if er != nil {
			return nil, er
		}
		org.ID = orgID
	} else {
		org.CreatedByID = user.ID
	}

	org, err = o.repo.Create(ctx, org)
	if err != nil {
		return nil, err
	}

	if !org.IsNew {
		return org, nil
	}

	err = o.wkflowMgr.Provision(ctx, org.ID.String())
	if err != nil {
		return nil, fmt.Errorf("unable to provision org: %w", err)
	}

	_, err = o.userOrgUpdater.UpsertUserOrg(ctx, user.ID, org.ID)
	if err != nil {
		return org, err
	}

	return org, err
}

func (o OrgService) UserOrgs(ctx context.Context, inputID string, options *models.ConnectionOptions) ([]*models.Org, *utils.Page, error) {
	userID, err := parseID(inputID)
	if err != nil {
		return nil, nil, err
	}

	return o.repo.GetPageByUser(ctx, userID, options)
}

func (o OrgService) Orgs(ctx context.Context, options *models.ConnectionOptions) ([]*models.Org, *utils.Page, error) {
	var orgs []*models.Org

	pg, c, err := utils.NewPaginator(options)

	if err != nil {
		return nil, nil, err
	}

	tx := o.repo.QueryAll(ctx)
	page, err := pg.Paginate(c, tx)
	if err != nil {
		return nil, nil, err
	}

	if err := page.Query(&orgs); err != nil {
		return nil, nil, err
	}

	return orgs, &page, nil
}
