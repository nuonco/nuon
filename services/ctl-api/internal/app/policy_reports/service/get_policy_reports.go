package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetPolicyReports
// @Summary				get policy reports
// @Description.markdown	get_policy_reports.md
// @Param					offset		query	int		false	"offset of results to return"	Default(0)
// @Param					limit		query	int		false	"limit of results to return"	Default(10)
// @Param					page		query	int		false	"page number of results to return"	Default(0)
// @Param					owner_type	query	string	false	"owner type (install_deploys, install_sandbox_runs, component_builds)"
// @Param					owner_id	query	string	false	"owner id"
// @Param					app_id		query	string	false	"app id"
// @Param					install_id	query	string	false	"install id"
// @Param					status		query	string	false	"report status"
// @Tags					policy-reports
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.PolicyReport
// @Router					/v1/policy-reports [get]
func (s *service) GetPolicyReports(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	ownerType := ctx.Query("owner_type")
	ownerID := ctx.Query("owner_id")
	appID := ctx.Query("app_id")
	installID := ctx.Query("install_id")
	status := ctx.Query("status")

	reports, err := s.getPolicyReports(ctx, org.ID, ownerType, ownerID, appID, installID, status)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get policy reports"))
		return
	}

	ctx.JSON(http.StatusOK, reports)
}

func (s *service) getPolicyReports(ctx *gin.Context, orgID, ownerType, ownerID, appID, installID, status string) ([]*app.PolicyReport, error) {
	var reports []*app.PolicyReport

	query := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Where("org_id = ?", orgID)

	if ownerType != "" {
		if ownerType != string(app.PolicyReportOwnerTypeInstallDeploy) &&
			ownerType != string(app.PolicyReportOwnerTypeInstallSandboxRun) &&
			ownerType != string(app.PolicyReportOwnerTypeComponentBuild) {
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("invalid owner_type: %s", ownerType),
				Description: "invalid owner_type",
			}
		}
		query = query.Where("owner_type = ?", ownerType)
	}

	if ownerID != "" {
		query = query.Where("owner_id = ?", ownerID)
	}

	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}

	if installID != "" {
		query = query.Where("install_id = ?", installID)
	}

	if status != "" {
		query = query.Where("status->>'status' = ?", status)
	}

	res := query.Order("created_at desc").Find(&reports)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get policy reports: %w", res.Error)
	}

	reports, err := db.HandlePaginatedResponse(ctx, reports)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return reports, nil
}
