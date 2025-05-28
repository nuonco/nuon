package helpers

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

func (s *Helpers) GetTerraformStateJSON(ctx context.Context, workspaceID string) ([]byte, error) {
	tfs := &app.TerraformWorkspaceStateJSON{}

	res := s.db.WithContext(ctx).
		First(tfs, "workspace_id = ?", workspaceID)
	if res.Error != nil {
		// if no lock is found, return nil as the lock does not exist
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, res.Error
	}

	return tfs.Contents, nil
}

func (s *Helpers) UpdateStateJSON(ctx context.Context, workspaceID string, jobID *string, contents []byte) error {
	tfs := &app.TerraformWorkspaceStateJSON{
		WorkspaceID: workspaceID,
		RunnerJobID: jobID,
		Contents:    contents,
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
