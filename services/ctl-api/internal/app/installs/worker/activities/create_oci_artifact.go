package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type CreateOCIArtifactRequest struct {
	OwnerType string `validate:"required"`
	OwnerID   string `validate:"required"`

	Outputs state.OCIArtifactOutputs `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) CreateOCIArtifact(ctx context.Context, req CreateOCIArtifactRequest) (*app.OCIArtifact, error) {
	ociArt := app.OCIArtifact{
		OwnerType: req.OwnerType,
		OwnerID:   req.OwnerID,

		Tag:          req.Outputs.Tag,
		Repository:   req.Outputs.Repository,
		MediaType:    req.Outputs.MediaType,
		Digest:       req.Outputs.Digest,
		Size:         req.Outputs.Size,
		URLs:         req.Outputs.URLs,
		Annotations:  generics.ToHstore(req.Outputs.Annotations),
		ArtifactType: req.Outputs.ArtifactType,

		// Platform fields
		Architecture: req.Outputs.Platform.Architecture,
		OS:           req.Outputs.Platform.OS,
		OSVersion:    req.Outputs.Platform.OSVersion,
		Variant:      req.Outputs.Platform.Variant,
		OSFeatures:   req.Outputs.Platform.OSFeatures,
	}

	res := a.db.WithContext(ctx).Create(&ociArt)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create oci artifact")
	}

	return &ociArt, nil
}
