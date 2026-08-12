package service

import (
	"github.com/gin-gonic/gin"
)

// The handlers below delegate to their bare-route counterparts. They exist so
// each ancestor-scoped path gets its own swagger operation, and so a generated
// client method, rather than being reachable only by hand-built requests.
// Ancestor ownership is enforced by the group's guard, not here.

// @ID						GetInstallPolicyReport
// @Summary				get policy report
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the report, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
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
// @Router					/v1/installs/{install_id}/policy-reports/{report_id} [get]
func (s *service) GetInstallPolicyReport(ctx *gin.Context) {
	s.GetPolicyReport(ctx)
}

// @ID						ExportInstallPolicyReport
// @Summary				export policy report
// @Description			Ancestor-scoped alternative to the bare route: authorizes against the install that owns the report, so an identity scoped to that install can reach it.
// @Param					install_id	path	string	true	"install ID"
// @Param					report_id	path	string	true	"policy report ID"
// @Param					format		query	string	false	"export format: json, sarif, pdf (default: json)"
// @Tags					policy-reports
// @Accept					json
// @Produce				json
// @Produce				application/pdf
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	interface{}
// @Router					/v1/installs/{install_id}/policy-reports/{report_id}/export [get]
func (s *service) ExportInstallPolicyReport(ctx *gin.Context) {
	s.ExportPolicyReport(ctx)
}
