package activities

import (
	"context"
	"fmt"

	"github.com/hashicorp/waypoint/pkg/server/gen"
)

type UpsertProjectRequest struct {
	OrgID     string `validate:"required"`
	ProjectID string `validate:"required"`
}

func (a *Activities) UpsertProject(ctx context.Context, req UpsertProjectRequest) error {
	_, err := a.wpClient.UpsertProject(ctx, req.OrgID, &gen.UpsertProjectRequest{
		Project: &gen.Project{
			Name: req.ProjectID,
		},
	})

	if err != nil {
		return fmt.Errorf("unable to upsert project: %w", err)
	}

	return nil
}
