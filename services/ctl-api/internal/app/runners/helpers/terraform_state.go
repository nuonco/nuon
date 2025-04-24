package helpers

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (s *Helpers) GetTerraformState(ctx context.Context, workspaceID string) (*app.TerraformWorkspaceState, error) {
	tfState := &app.TerraformWorkspaceState{}

	res := s.db.WithContext(ctx).
		Order("revision DESC").
		First(tfState, "terraform_workspace_id = ?", workspaceID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get terraform state: %w", res.Error)
	}

	return tfState, nil
}

func (s *Helpers) InsertTerraformState(ctx context.Context, workspaceID string, contents []byte, data *app.TerraformStateData) (*app.TerraformWorkspaceState, error) {
	tfState := app.TerraformWorkspaceState{
		TerraformWorkspaceID: workspaceID,
		Data:                 data,
		Contents:             contents,
	}

	res := s.db.WithContext(ctx).Create(&tfState)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "failed to insert new terraform state")
	}

	return &tfState, nil
}

func (s *Helpers) GetTerraformStateByID(ctx context.Context, workspaceID, id string) (*app.TerraformWorkspaceState, error) {
	tfState := &app.TerraformWorkspaceState{}
	res := s.db.WithContext(ctx).
		Where("terraform_workspace_id = ? AND id = ?", workspaceID, id).
		First(tfState)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get terraform state: %w", res.Error)
	}

	return tfState, nil
}
