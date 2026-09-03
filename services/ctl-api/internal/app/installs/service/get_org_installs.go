package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetOrgInstalls
// @Summary				get all installs for an org
// @Description.markdown	get_org_installs.md
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param         q								 query	string	false	"search query to filter installs by name, ID, or branch name"
// @Param					labels						query	string	false	"label filter (key:value,key:value)"
// @Param					runner_id				query	string	false	"filter by runner ID"
// @Param					branches				query	string	false	"filter installs by branch name (comma-separated; use __none__ for installs with no branch)"
// @Param					include_components	query	bool	false	"include install components"	Default(true)
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
	lbls := labels.ParseLabelsQuery(ctx.Query("labels"))
	runnerID := ctx.Query("runner_id")
	branches := ctx.Query("branches")
	includeComponents := ctx.Query("include_components") != "false"

	install, err := s.getOrgInstalls(ctx, org.ID, q, lbls, runnerID, branches, includeComponents)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get installs for org %s: %w", org.ID, err))
		return
	}

	ctx.JSON(http.StatusOK, install)
}

const branchFilterNoneToken = "__none__"

func parseBranchesFilter(raw string) (names []string, none bool) {
	if raw == "" {
		return nil, false
	}
	for _, part := range strings.Split(raw, ",") {
		p := strings.TrimSpace(part)
		switch {
		case p == "":
			continue
		case p == branchFilterNoneToken:
			none = true
		default:
			names = append(names, p)
		}
	}
	return names, none
}

func (s *service) getOrgInstalls(ctx *gin.Context, orgID, q string, lbls labels.Labels, runnerID, branches string, includeComponents bool) ([]app.Install, error) {
	var installs []app.Install
	tx := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Scopes(labels.WithLabels(views.TableOrViewName(s.db, &app.Install{}, ".labels"), lbls)).
		Preload("OperatingModel").
		Preload("AppSandboxConfig").
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("GCPAccount").
		Preload("AppBranch").
		Preload("AppRunnerConfig").
		Preload("AppConfig", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "app_id", "app_branch_id")
		}).
		Preload("AppConfig.RunnerConfig").
		Preload("App").
		Preload("App.AppRunnerConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(scopes.WithOverrideTable(views.CustomViewName(s.db, &app.AppRunnerConfig{}, "latest_view_v1")))
		}).
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

	if runnerID != "" {
		viewName := views.TableOrViewName(s.db, &app.Install{}, "")
		tx = tx.
			Joins("JOIN runner_groups ON runner_groups.owner_id = "+viewName+".id AND runner_groups.owner_type = 'installs' AND runner_groups.deleted_at = 0").
			Joins("JOIN runners ON runners.runner_group_id = runner_groups.id AND runners.deleted_at = 0").
			Where("runners.id = ?", runnerID)
	}

	branchCol := views.TableOrViewName(s.db, &app.Install{}, ".app_branch_id")
	branchNames, branchNone := parseBranchesFilter(branches)

	if q != "" || len(branchNames) > 0 {
		tx = tx.Joins("LEFT JOIN app_branches ON app_branches.id = " + branchCol + " AND app_branches.deleted_at = 0")
	}

	if q != "" {
		nameCol := views.TableOrViewName(s.db, &app.Install{}, ".name")
		idCol := views.TableOrViewName(s.db, &app.Install{}, ".id")
		queryPattern := "%" + q + "%"
		tx = tx.Where(nameCol+" ILIKE ? OR "+idCol+" ILIKE ? OR app_branches.name ILIKE ?", queryPattern, queryPattern, queryPattern)
	}

	switch {
	case len(branchNames) > 0 && branchNone:
		tx = tx.Where("(app_branches.name IN ? OR "+branchCol+" IS NULL OR "+branchCol+" = '')", branchNames)
	case len(branchNames) > 0:
		tx = tx.Where("app_branches.name IN ?", branchNames)
	case branchNone:
		tx = tx.Where("(" + branchCol + " IS NULL OR " + branchCol + " = '')")
	}
	res := tx.Find(&installs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org installs: %w", res.Error)
	}

	if includeComponents {
		if err := s.loadOrgInstallsComponents(ctx, orgID, installs); err != nil {
			return nil, err
		}
	}

	installs, err := db.HandlePaginatedResponse(ctx, installs)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return installs, nil
}

func (s *service) loadOrgInstallsComponents(ctx *gin.Context, orgID string, installs []app.Install) error {
	if len(installs) == 0 {
		return nil
	}

	installIDs := make([]string, 0, len(installs))
	for _, install := range installs {
		installIDs = append(installIDs, install.ID)
	}

	componentsByInstallID := make(map[string][]app.InstallComponent, len(installs))
	var installComponentsBatch []app.InstallComponent
	tx := s.db.WithContext(ctx).
		Preload("InstallDeploys", func(db *gorm.DB) *gorm.DB {
			return db.
				Scopes(scopes.WithOverrideTable(views.CustomViewName(s.db, &app.InstallDeploy{}, "latest_view_v1"))).
				Order("install_deploys_latest_view_v1.created_at DESC")
		}).
		Preload("Component").
		Where(app.InstallComponent{OrgID: orgID}).
		Where(map[string]any{"install_id": installIDs}).
		FindInBatches(&installComponentsBatch, 10, func(_ *gorm.DB, _ int) error {
			for _, installComponent := range installComponentsBatch {
				componentsByInstallID[installComponent.InstallID] = append(componentsByInstallID[installComponent.InstallID], installComponent)
			}
			return nil
		})
	if tx.Error != nil {
		return fmt.Errorf("unable to get install components for org %s: %w", orgID, tx.Error)
	}

	for i := range installs {
		installs[i].InstallComponents = componentsByInstallID[installs[i].ID]
	}

	return nil
}
