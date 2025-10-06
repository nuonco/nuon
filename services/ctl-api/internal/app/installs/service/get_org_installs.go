package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetOrgInstalls
// @Summary				get all installs for an org
// @Description.markdown	get_org_installs.md
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param         q								 query	string	false	"search query to filter installs by name"
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.Install
// @Router					/v1/installs [GET]
func (s *service) GetOrgInstalls(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	q := ctx.Query("q")

	install, err := s.getOrgInstalls(ctx, org.ID, q)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get installs for org %s: %w", org.ID, err))
		return
	}

	ctx.JSON(http.StatusOK, install)
}

func (s *service) getOrgInstalls(ctx *gin.Context, orgID, q string) ([]app.Install, error) {
	var installs []app.Install
	tx := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("AppSandboxConfig").
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("AppRunnerConfig").
		Preload("App").
		Preload("App.AppRunnerConfigs").
		Preload("App.Org").
		Preload("AppSandboxConfig.PublicGitVCSConfig").
		Preload("AppSandboxConfig.ConnectedGithubVCSConfig").
		Preload("InstallSandboxRuns", func(db *gorm.DB) *gorm.DB {
			return db.
				Scopes(scopes.WithOverrideTable(views.CustomViewName(s.db, &app.InstallSandboxRun{}, "state_view_v1"))).
				Order("install_sandbox_runs_state_view_v1.created_at DESC")
		}).
		Preload("InstallSandboxRuns.AppSandboxConfig").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Joins(fmt.Sprintf("JOIN apps ON apps.id=%s", views.TableOrViewName(s.db, &app.Install{}, ".app_id"))).
		Joins("JOIN orgs ON orgs.id=apps.org_id").
		Where(views.TableOrViewName(s.db, &app.Install{}, ".org_id")+" = ?", orgID).
		Order("name ASC")

	if q != "" {
		tx = tx.Where(views.TableOrViewName(s.db, &app.Install{}, ".name")+" ILIKE ?", "%"+q+"%")
	}
	res := tx.Find(&installs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org installs: %w", res.Error)
	}

	for i := range installs {
		// WARN: (rb) Get install components in batches to avoid loading too many components into memory at once
		installComponents, err := s.getOrgInstallsComponentsInBatches(ctx, orgID, installs[i])
		if err != nil {
			return nil, fmt.Errorf("unable to get install components for org %s: %w", orgID, err)
		}
		installs[i].InstallComponents = installComponents
	}

	installs, err := db.HandlePaginatedResponse(ctx, installs)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return installs, nil
}

func (s *service) getOrgInstallsComponentsInBatches(ctx *gin.Context, orgID string, install app.Install) ([]app.InstallComponent, error) {
	installComponents := make([]app.InstallComponent, 0)
	batchSize := 10
	offset := 0
	hasMore := true

	for hasMore {
		var installComponentsBatch []app.InstallComponent
		tx := s.db.WithContext(ctx).
			Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
				return db.
					Scopes(scopes.WithOverrideTable(views.CustomViewName(s.db, &app.InstallDeploy{}, "latest_view_v1"))).
					Order("install_deploys_latest_view_v1.created_at DESC")
			}).
			Preload("Component").
			Where("install_id = ?", install.ID).
			Limit(batchSize).
			Offset(offset).
			Find(&installComponents)

		if tx.Error != nil {
			return nil, fmt.Errorf("unable to get install components for org %s: %w", orgID, tx.Error)
		}

		installComponents = append(installComponents, installComponentsBatch...)

		if len(installComponentsBatch) < batchSize {
			hasMore = false
		} else {
			offset += batchSize
		}
	}

	return installComponents, nil
}
