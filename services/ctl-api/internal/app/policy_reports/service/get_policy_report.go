package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetPolicyReport
// @Summary				get policy report
// @Description.markdown	get_policy_report.md
// @Param					report_id	path	string	true	"policy report ID"
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
// @Success				200	{object}	app.PolicyReport
// @Router					/v1/policy-reports/{report_id} [get]
func (s *service) GetPolicyReport(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		return
	}

	reportID := ctx.Param("report_id")
	report, err := s.getPolicyReport(ctx, org.ID, reportID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get policy report"))
		return
	}

	ctx.JSON(http.StatusOK, report)
}

func (s *service) getPolicyReport(ctx *gin.Context, orgID, reportID string) (*app.PolicyReport, error) {
	var report app.PolicyReport
	res := s.db.WithContext(ctx).
		Where("id = ? AND org_id = ?", reportID, orgID).
		First(&report)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get policy report")
	}

	return &report, nil
}
