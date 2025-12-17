package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateInstallActionWorkflowRunRequest struct {
	ActionWorkFlowConfigID string            `json:"action_workflow_config_id"`
	RunEnvVars             map[string]string `json:"run_env_vars"`
}

func (c *CreateInstallActionWorkflowRunRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateInstallActionRun
// @Summary					create an action run for an install
// @Description.markdown	create_install_action_run.md
// @Tags					actions
// @Accept					json
// @Param					install_id	path	string									true	"install ID"
// @Param					req			body	CreateInstallActionWorkflowRunRequest	true	"Input"
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					201	{string}	ok
// @Router					/v1/installs/{install_id}/actions/runs [post]
func (s *service) CreateInstallActionRun(ctx *gin.Context) {
	s.CreateInstallActionWorkflowRun(ctx)
}
