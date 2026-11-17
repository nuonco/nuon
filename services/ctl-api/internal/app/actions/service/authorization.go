package service

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	actionaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/access"
	appaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/access"
	installaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"gorm.io/gorm"
)

func actionRef(ctx *gin.Context) string {
	if id := ctx.Param("action_id"); id != "" {
		return id
	}
	return ctx.Param("action_workflow_id")
}

func (s *service) actionResource(ctx *gin.Context, query authz.ResourceSet, id string) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	resources, err := query.In(s.db, orgID)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx, id)
}

func (s *service) actionAppTarget(ctx *gin.Context) (permissions.Target, error) {
	return s.actionResource(ctx, appaccess.Apps, ctx.Param("app_id"))
}

func (s *service) actionTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	var row app.ActionWorkflow
	err = s.db.WithContext(ctx).Where(app.ActionWorkflow{OrgID: orgID}).
		Where(s.db.Where(app.ActionWorkflow{ID: actionRef(ctx)}).Or(app.ActionWorkflow{Name: actionRef(ctx)})).First(&row).Error
	if err != nil {
		return permissions.Target{}, err
	}
	return s.actionResource(ctx, actionaccess.Actions, row.ID)
}

func (s *service) actionIDTarget(ctx *gin.Context) (permissions.Target, error) {
	target, err := s.actionTarget(ctx)
	if err == nil && target.ID != actionRef(ctx) {
		err = gorm.ErrRecordNotFound
	}
	return target, err
}

func (s *service) appActionTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	row, err := s.findAppActionWorkflow(ctx, orgID, ctx.Param("app_id"), actionRef(ctx))
	if err != nil {
		return permissions.Target{}, err
	}
	return s.actionResource(ctx, actionaccess.Actions, row.ID)
}

func (s *service) actionConfigTarget(ctx *gin.Context) (permissions.Target, error) {
	id := ctx.Param("action_config_id")
	if id == "" {
		id = ctx.Param("action_workflow_config_id")
	}
	return s.actionResource(ctx, actionaccess.Configs, id)
}

func (s *service) actionLatestConfigTarget(ctx *gin.Context) (permissions.Target, error) {
	config, err := s.getActionWorkflowLatestConfig(ctx, actionRef(ctx))
	if err != nil {
		return permissions.Target{}, err
	}
	return s.actionResource(ctx, actionaccess.Configs, config.ID)
}

func (s *service) actionInstallTarget(ctx *gin.Context) (permissions.Target, error) {
	return s.actionResource(ctx, installaccess.Installs, ctx.Param("install_id"))
}

func (s *service) actionRunTarget(ctx *gin.Context) (permissions.Target, error) {
	return s.actionResource(ctx, actionaccess.Runs, ctx.Param("run_id"))
}

func (s *service) actionStepTarget(ctx *gin.Context) (permissions.Target, error) {
	target, err := s.actionResource(ctx, actionaccess.Steps, ctx.Param("step_id"))
	if err != nil {
		return target, err
	}
	var step app.InstallActionWorkflowRunStep
	err = s.db.WithContext(ctx).Where(app.InstallActionWorkflowRunStep{ID: target.ID, OrgID: target.OrgID, InstallActionWorkflowRunID: ctx.Param("run_id")}).First(&step).Error
	if err == nil && target.InstallID != ctx.Param("install_id") {
		err = gorm.ErrRecordNotFound
	}
	return target, err
}

func (s *service) installedActionTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	var row app.InstallActionWorkflow
	err = s.db.WithContext(ctx).Where(app.InstallActionWorkflow{OrgID: orgID, InstallID: ctx.Param("install_id"), ActionWorkflowID: actionRef(ctx)}).First(&row).Error
	if err != nil {
		return permissions.Target{}, err
	}
	return s.actionResource(ctx, actionaccess.InstallActions, row.ID)
}

func readActionAuthorizationBody(ctx *gin.Context, value any) error {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return stderr.NewInvalidRequest(err)
	}
	ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
	if err := json.Unmarshal(body, value); err != nil {
		return stderr.NewInvalidRequest(err)
	}
	return nil
}

const authorizedActionRunKey = "nuon.actions.authorized_run"

func (s *service) actionExecutionTarget(ctx *gin.Context) (permissions.Target, error) {
	target, err := s.actionInstallTarget(ctx)
	if err != nil {
		return target, err
	}
	var req CreateInstallActionWorkflowRunRequest
	if err := readActionAuthorizationBody(ctx, &req); err != nil {
		return target, err
	}
	configTarget, err := s.actionResource(ctx, actionaccess.Configs, req.ActionWorkFlowConfigID)
	if err != nil {
		return target, err
	}
	if configTarget.AppID != target.AppID {
		return target, gorm.ErrRecordNotFound
	}
	prepared, err := s.prepareInstallActionWorkflowRun(ctx, target.ID, req)
	if err != nil {
		return target, err
	}
	pinned, err := s.actionResource(ctx, actionaccess.Configs, prepared.ActionWorkflowConfigID)
	if err == nil && pinned.AppID != target.AppID {
		err = gorm.ErrRecordNotFound
	}
	if err == nil {
		ctx.Set(authorizedActionRunKey, prepared)
	}
	return target, err
}

func (s *service) actionConfigWriteTarget(ctx *gin.Context) (permissions.Target, error) {
	target, err := s.actionIDTarget(ctx)
	if err != nil {
		return target, err
	}
	parent, err := s.findApp(ctx, target.OrgID, target.AppID)
	if err != nil {
		return target, err
	}
	if parent.ID != target.AppID || parent.OrgID != target.OrgID {
		return target, gorm.ErrRecordNotFound
	}
	var req CreateActionWorkflowConfigRequest
	if err := readActionAuthorizationBody(ctx, &req); err != nil {
		return target, err
	}
	if err := req.Validate(s.v); err != nil {
		return target, stderr.NewInvalidRequest(err)
	}
	var config app.AppConfig
	if err := s.db.WithContext(ctx).Where(app.AppConfig{ID: req.AppConfigID, OrgID: target.OrgID, AppID: target.AppID}).First(&config).Error; err != nil {
		return target, err
	}
	for _, ref := range req.Dependencies {
		if ref == "" {
			return target, gorm.ErrRecordNotFound
		}
		var component app.Component
		if err := s.db.WithContext(ctx).Where(app.Component{OrgID: target.OrgID, AppID: target.AppID}).
			Where(s.db.Where(app.Component{ID: ref}).Or(app.Component{Name: ref})).First(&component).Error; err != nil {
			return target, err
		}
	}
	for _, trigger := range req.Triggers {
		if trigger.ComponentName != "" {
			var component app.Component
			if err := s.db.WithContext(ctx).Where(app.Component{OrgID: target.OrgID, AppID: target.AppID, Name: trigger.ComponentName}).First(&component).Error; err != nil {
				return target, err
			}
		}
	}
	for _, step := range req.Steps {
		if step.ConnectedGithubVCSConfig != nil {
			authority, err := authz.AuthorityFromContext(ctx)
			if err != nil {
				return target, err
			}
			if !authority.Check(permissions.AppAutomationManage, target).Matched {
				return target, gorm.ErrRecordNotFound
			}
		}
	}
	return target, nil
}

func (s *service) actionLogsTarget(ctx *gin.Context) (permissions.Target, error) {
	run, err := s.actionRunTarget(ctx)
	if err != nil {
		return run, err
	}
	if run.InstallID != ctx.Param("install_id") {
		return run, gorm.ErrRecordNotFound
	}
	var stream app.LogStream
	if err := s.db.WithContext(ctx).Where(app.LogStream{OrgID: run.OrgID, OwnerID: run.ID, OwnerType: "install_action_workflow_runs"}).First(&stream).Error; err != nil {
		return run, err
	}
	return s.actionResource(ctx, installaccess.Installs, run.InstallID)
}

func (s *service) actionListRule(operation permissions.PermissionID, query authz.ResourceSet) require.Rule {
	return require.Collection(operation, func(ctx *gin.Context) (*authz.Resources, error) {
		orgID, err := cctx.OrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		return query.In(s.db, orgID)
	})
}
