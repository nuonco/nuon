package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/api/internal/models"
)

func (a *activities) InsertArtifact(ctx context.Context, artifact *models.Artifact) (*models.Artifact, error) {
	if err := artifact.NewID(); err != nil {
		return nil, fmt.Errorf("StartBuild.InsertArtifact unable to make nanoid: %w", err)
	}
	if err := a.db.WithContext(ctx).Create(artifact).Error; err != nil {
		return nil, fmt.Errorf("StartBuild.InsertArtifact error inserting artifact: %w", err)
	}
	return artifact, fmt.Errorf("not yet implemented")
}
