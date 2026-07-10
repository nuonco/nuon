package helpers

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func (s *Helpers) GetTerraformStateJSON(ctx context.Context, workspaceID string) ([]byte, error) {
	tfs := &app.TerraformWorkspaceStateJSON{}

	res := s.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Order("created_at DESC").
		First(tfs)
	if res.Error != nil {
		// if no lock is found, return nil as the lock does not exist
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, res.Error
	}

	contents, fromBlob := tfs.GetContents(blobstore.WithBlobService(ctx, s.blobSvc), s.cfg.BlobReadEnabled)
	if fromBlob {
		s.l.Debug("read terraform workspace state json contents from blob",
			zap.String("workspace_id", workspaceID),
			zap.String("state_id", tfs.ID),
			zap.Int("bytes", len(contents)))
	}

	return contents, nil
}

func (s *Helpers) CreateStateJSON(ctx context.Context, workspaceID string, jobID *string, contents []byte) error {
	workspace := &app.TerraformWorkspace{}
	resCheck := s.db.WithContext(ctx).
		First(workspace, "id = ?", workspaceID)
	if resCheck.Error != nil {
		return resCheck.Error
	}

	tfs := &app.TerraformWorkspaceStateJSON{
		WorkspaceID:  workspaceID,
		RunnerJobID:  jobID,
		Contents:     contents,
		ContentsBlob: &blobstore.Blob{},
		OrgID:        workspace.OrgID,
	}
	tfs.ContentsBlob.Set(string(contents))

	ctx = blobstore.WithBlobService(ctx, s.blobSvc)
	if keys.OrgIDFromContext(ctx) == "" {
		ctx = cctx.SetOrgIDContext(ctx, workspace.OrgID)
	}

	res := s.db.WithContext(ctx).Create(tfs)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

func (s *Helpers) DeleteStateJSON(ctx context.Context, workspaceID string) error {
	res := s.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Delete(&app.TerraformWorkspaceStateJSON{})
	if res.Error != nil {
		return res.Error
	}
	return nil
}
