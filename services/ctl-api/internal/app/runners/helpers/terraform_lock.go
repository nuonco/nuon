package helpers

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

func (s *Helpers) GetWorkspaceLock(ctx context.Context, workspaceID string) (*app.TerraformLock, error) {
	tfs := &app.TerraformWorkspaceLockState{}

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

func (s *Helpers) LockWorkspace(ctx context.Context, workspaceID string, lock *app.TerraformLock) (*app.TerraformWorkspaceLockState, error) {
	tfs := &app.TerraformWorkspaceLockState{
		WorkspaceID: workspaceID,
		Lock:        lock,
	}

	res := s.db.WithContext(ctx).Create(tfs)
	if res.Error != nil {
		return nil, res.Error
	}
	return tfs, nil
}

func (s *Helpers) UnlockWorkspace(ctx context.Context, workspaceID string) error {
	res := s.db.WithContext(ctx).
		Where("workspace_id = ?", workspaceID).
		Delete(&app.TerraformWorkspaceLockState{})
	if res.Error != nil {
		return res.Error
	}
	return nil
}
