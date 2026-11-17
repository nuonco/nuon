package service

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	notebookaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/notebooks/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"gorm.io/gorm"
)

func (s *service) notebookResource(ctx *gin.Context, query authz.ResourceSet, id string) (permissions.Target, error) {
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

func (s *service) notebookInstallTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	install, err := s.installHelpers.GetInstall(ctx, orgID, ctx.Param("install_id"))
	if err != nil {
		return permissions.Target{}, err
	}
	return s.notebookResource(ctx, installaccess.Installs, install.ID)
}

func (s *service) notebookTarget(ctx *gin.Context) (permissions.Target, error) {
	install, err := s.notebookInstallTarget(ctx)
	if err != nil {
		return install, err
	}
	target, err := s.notebookResource(ctx, notebookaccess.Notebooks, ctx.Param("notebook_id"))
	if err == nil && target.InstallID != install.ID {
		err = gorm.ErrRecordNotFound
	}
	return target, err
}

func (s *service) notebookCellTarget(ctx *gin.Context) (permissions.Target, error) {
	notebook, err := s.notebookTarget(ctx)
	if err != nil {
		return notebook, err
	}
	var cell app.NotebookCell
	if err := s.db.WithContext(ctx).Where(app.NotebookCell{ID: ctx.Param("cell_id"), OrgID: notebook.OrgID, NotebookID: notebook.ID}).First(&cell).Error; err != nil {
		return notebook, err
	}
	return s.notebookResource(ctx, notebookaccess.Cells, cell.ID)
}

func (s *service) notebookExecutionTarget(ctx *gin.Context) (permissions.Target, error) {
	cell, err := s.notebookCellTarget(ctx)
	if err != nil {
		return cell, err
	}
	return s.notebookResource(ctx, installaccess.Installs, cell.InstallID)
}

func (s *service) notebookRunTarget(ctx *gin.Context) (permissions.Target, error) {
	notebook, err := s.notebookTarget(ctx)
	if err != nil {
		return notebook, err
	}
	var run app.NotebookCellRun
	if err := s.db.WithContext(ctx).Where(app.NotebookCellRun{ID: ctx.Param("run_id"), OrgID: notebook.OrgID, NotebookID: notebook.ID, InstallID: notebook.InstallID}).First(&run).Error; err != nil {
		return notebook, err
	}
	return s.notebookResource(ctx, notebookaccess.Runs, run.ID)
}

func (s *service) notebookReorderTarget(ctx *gin.Context) (permissions.Target, error) {
	target, err := s.notebookTarget(ctx)
	if err != nil {
		return target, err
	}
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return target, stderr.NewInvalidRequest(err)
	}
	ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
	var req ReorderCellsRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return target, stderr.NewInvalidRequest(err)
	}
	for _, id := range req.CellIDs {
		if id == "" {
			return target, gorm.ErrRecordNotFound
		}
		var cell app.NotebookCell
		if err := s.db.WithContext(ctx).Where(app.NotebookCell{ID: id, OrgID: target.OrgID, NotebookID: target.ID}).First(&cell).Error; err != nil {
			return target, err
		}
	}
	return target, nil
}

func (s *service) notebookListRule(query authz.ResourceSet) require.Rule {
	return require.Collection(permissions.NotebookInspect, func(ctx *gin.Context) (*authz.Resources, error) {
		orgID, err := cctx.OrgIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		return query.In(s.db, orgID)
	})
}
