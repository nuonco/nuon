package helpers

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

func (s *Helpers) GetWorkspaceLock(ctx context.Context, workspaceID string) (*app.TerraformLock, error) {
	tfs := &app.TerraformWorkspaceLock{}

	res := s.db.WithContext(ctx).
		First(tfs, "workspace_id = ?", workspaceID)
	if res.Error != nil {
		// if no lock is found, return nil as the lock does not exist
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, res.Error
	}

	return tfs.Lock, nil
}

func (s *Helpers) LockWorkspace(ctx context.Context, workspaceID string, jobID *string, lock *app.TerraformLock) (*app.TerraformWorkspaceLock, error) {
	tfs := &app.TerraformWorkspaceLock{
		WorkspaceID: workspaceID,
		Lock:        lock,
		RunnerJobID: jobID,
	}

	workspace := &app.TerraformWorkspace{}
	resCheck := s.db.WithContext(ctx).
		First(workspace, "id = ?", workspaceID)
	if resCheck.Error != nil {
		return nil, resCheck.Error
	}

	tfs.OrgID = workspace.OrgID

	res := s.db.WithContext(ctx).Create(tfs)
	if res.Error != nil {
		return nil, res.Error
	}
	return tfs, nil
}

func (s *Helpers) UnlockWorkspace(ctx context.Context, workspaceID string) error {
	res := s.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Delete(&app.TerraformWorkspaceLock{})
	if res.Error != nil {
		return res.Error
	}
	return nil
}
