package service

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/access"
	installaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	runbookaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"gorm.io/gorm"
)

func (s *service) runbookResource(ctx *gin.Context, query authz.ResourceSet, id string) (permissions.Target, error) {
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

func (s *service) runbookAppTarget(ctx *gin.Context) (permissions.Target, error) {
	return s.runbookResource(ctx, appaccess.Apps, ctx.Param("app_id"))
}

func (s *service) runbookTarget(ctx *gin.Context) (permissions.Target, error) {
	return s.runbookResource(ctx, runbookaccess.Runbooks, ctx.Param("runbook_id"))
}

func (s *service) runbookReadTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	ref := ctx.Param("runbook_id")
	var row app.Runbook
	err = s.db.WithContext(ctx).Where(app.Runbook{OrgID: orgID, AppID: ctx.Param("app_id")}).
		Where(s.db.Where(app.Runbook{ID: ref}).Or(app.Runbook{Name: ref})).First(&row).Error
	if err != nil {
		return permissions.Target{}, err
	}
	return s.runbookResource(ctx, runbookaccess.Runbooks, row.ID)
}

func (s *service) installRunbookTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	var row app.InstallRunbook
	ref := ctx.Param("runbook_id")
	err = s.db.WithContext(ctx).
		Joins("JOIN runbooks ON runbooks.id = install_runbooks.runbook_id AND runbooks.deleted_at = 0").
		Where(app.InstallRunbook{OrgID: orgID, InstallID: ctx.Param("install_id")}).
		Where("install_runbooks.runbook_id = ? OR runbooks.name = ?", ref, ref).First(&row).Error
	if err != nil {
		return permissions.Target{}, err
	}
	return s.runbookResource(ctx, runbookaccess.InstallRunbooks, row.ID)
}

func (s *service) runbookRunTarget(ctx *gin.Context) (permissions.Target, error) {
	return s.runbookResource(ctx, runbookaccess.Runs, ctx.Param("run_id"))
}

func (s *service) runbookExecutionTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	install, err := s.resolveInstallRef(ctx, orgID, ctx.Param("install_id"))
	if err != nil {
		return permissions.Target{}, err
	}
	return s.runbookResource(ctx, installaccess.Installs, install.ID)
}

func (s *service) runbookConfigWriteTarget(ctx *gin.Context) (permissions.Target, error) {
	target, err := s.runbookTarget(ctx)
	if err != nil {
		return target, err
	}
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return target, stderr.NewInvalidRequest(err)
	}
	ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
	var req CreateRunbookConfigRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return target, stderr.NewInvalidRequest(err)
	}
	if req.AppConfigID != nil && *req.AppConfigID != "" {
		var config app.AppConfig
		if err := s.db.WithContext(ctx).Where(app.AppConfig{ID: *req.AppConfigID, OrgID: target.OrgID, AppID: target.AppID}).First(&config).Error; err != nil {
			return target, err
		}
	}
	for _, step := range req.Steps {
		if step == nil {
			return target, gorm.ErrRecordNotFound
		}
		if step.Type == string(app.RunbookStepTypeWaitForEvent) {
			authority, err := authz.AuthorityFromContext(ctx)
			if err != nil {
				return target, err
			}
			if !authority.Check(permissions.AppAutomationManage, target).Matched {
				return target, gorm.ErrRecordNotFound
			}
		}
		if step.ComponentName != "" {
			var component app.Component
			if err := s.db.WithContext(ctx).Where(app.Component{OrgID: target.OrgID, AppID: target.AppID, Name: step.ComponentName}).First(&component).Error; err != nil {
				return target, err
			}
		}
		if step.ActionName != "" {
			var action app.ActionWorkflow
			if err := s.db.WithContext(ctx).Where(app.ActionWorkflow{OrgID: target.OrgID, AppID: target.AppID, Name: step.ActionName}).First(&action).Error; err != nil {
				return target, err
			}
		}
	}
	return target, nil
}

func (s *service) runbookListRule(operation permissions.PermissionID, query authz.ResourceSet) require.Rule {
	return require.Collection(operation, func(ctx *gin.Context) (*authz.Resources, error) {
		orgID, err := cctx.OrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		return query.In(s.db, orgID)
	})
}
