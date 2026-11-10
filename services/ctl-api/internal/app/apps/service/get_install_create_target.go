package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// InstallCreateTargetResponse describes the app config, and any app branch
// and group, that a new install with the given labels and app_branch_id
// would start on.
type InstallCreateTargetResponse struct {
	AppConfigID                    string                                    `json:"app_config_id"`
	AppBranchID                    string                                    `json:"app_branch_id,omitempty"`
	AppBranchName                  string                                    `json:"app_branch_name,omitempty"`
	AppBranchGroup                 string                                    `json:"app_branch_group,omitempty"`
	AppBranchGroupAssignmentSource app.InstallAppBranchGroupAssignmentSource `json:"app_branch_group_assignment_source,omitempty"`
}

// @ID				GetInstallCreateTarget
// @Summary		get install create target
// @Description	Returns the app config, and any app branch and group, that a new install with these labels and app_branch_id would start on. Does not create anything.
// @Tags			apps
// @Param			app_id			path	string	true	"app ID"
// @Param			labels			query	string	false	"comma-separated key=value labels"
// @Param			app_branch_id	query	string	false	"optional app branch ID"
// @Accept			json
// @Produce		json
// @Security		APIKey
// @Security		OrgID
// @Failure		400	{object}	stderr.ErrResponse
// @Failure		401	{object}	stderr.ErrResponse
// @Failure		403	{object}	stderr.ErrResponse
// @Failure		404	{object}	stderr.ErrResponse
// @Failure		500	{object}	stderr.ErrResponse
// @Success		200	{object}	InstallCreateTargetResponse
// @Router			/v1/apps/{app_id}/install-create-target [get]
func (s *service) GetInstallCreateTarget(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	lbls := labels.ParseLabelsQuery(ctx.Query("labels"))
	appBranchID := ctx.Query("app_branch_id")

	_ = org
	_ = appID
	_ = lbls
	_ = appBranchID

	ctx.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
