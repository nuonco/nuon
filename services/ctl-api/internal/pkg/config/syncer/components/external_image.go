package components

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/validation"
)

// SyncExternalImageComponent creates an ExternalImageComponentConfig and its
// ComponentConfigConnection for a component of type external_image.
//
// Mirrors creation logic from
// services/ctl-api/internal/app/components/service/create_external_image_config.go.
func SyncExternalImageComponent(ctx context.Context, db *gorm.DB, comp *config.Component, componentID, appID, appConfigID string) (string, string, error) {
	if err := validateExternalImageComponent(comp); err != nil {
		return "", "", sync.SyncErr{
			Resource:    fmt.Sprintf("component-%s", comp.Name),
			Description: fmt.Sprintf("validation failed: %v", err),
		}
	}

	imageURL, tag, awsECR, gcpGAR, err := extractExternalImageSource(comp.ExternalImage)
	if err != nil {
		return "", "", sync.SyncErr{
			Resource:    fmt.Sprintf("component-%s", comp.Name),
			Description: err.Error(),
		}
	}

	// Build operation roles map.
	operationRoles := make(pgtype.Hstore)
	for _, role := range comp.OperationRoles {
		role := role
		operationRoles[string(role.Operation)] = &role.RoleName
	}

	// References.
	references := make([]string, 0, len(comp.References))
	for _, ref := range comp.References {
		references = append(references, ref.String())
	}

	cfg := app.ExternalImageComponentConfig{
		ImageURL:          imageURL,
		Tag:               tag,
		AWSECRImageConfig: awsECR,
		GCPGARImageConfig: gcpGAR,
	}

	componentConfigConnection := app.ComponentConfigConnection{
		ExternalImageComponentConfig: &cfg,
		ComponentID:                  componentID,
		AppConfigID:                  appConfigID,
		ComponentDependencyIDs:       pq.StringArray{},
		References:                   pq.StringArray(references),
		Checksum:                     comp.Checksum,
		BuildTimeout:                 comp.ExternalImage.BuildTimeout,
		DeployTimeout:                comp.ExternalImage.DeployTimeout,
		OperationRoles:               operationRoles,
	}

	if res := db.WithContext(ctx).Create(&componentConfigConnection); res.Error != nil {
		return "", "", sync.SyncInternalErr{
			Description: fmt.Sprintf("unable to create external image component config for %s", comp.Name),
			Err:         res.Error,
		}
	}

	return componentConfigConnection.ID, componentConfigConnection.Checksum, nil
}

// extractExternalImageSource pulls the image URL, tag, and registry-specific
// auth configs from the user-supplied ExternalImage config block.
func extractExternalImageSource(ext *config.ExternalImageComponentConfig) (string, string, *app.AWSECRImageConfig, *app.GCPGARImageConfig, error) {
	switch {
	case ext.AWSECRImageConfig != nil:
		ecr := ext.AWSECRImageConfig
		return ecr.ImageURL, ecr.Tag, &app.AWSECRImageConfig{
			IAMRoleARN: ecr.IAMRoleARN,
			AWSRegion:  ecr.AWSRegion,
		}, nil, nil
	case ext.GCPGARImageConfig != nil:
		gar := ext.GCPGARImageConfig
		return gar.ImageURL, gar.Tag, nil, &app.GCPGARImageConfig{
			GCPProjectID:             gar.GCPProjectID,
			GCPRegion:                gar.GCPRegion,
			ServiceAccountEmail:      gar.ServiceAccountEmail,
			WorkloadIdentityProvider: gar.WorkloadIdentityProvider,
		}, nil
	case ext.PublicImageConfig != nil:
		pub := ext.PublicImageConfig
		return pub.ImageURL, pub.Tag, nil, nil, nil
	default:
		return "", "", nil, nil, fmt.Errorf("external_image requires aws_ecr, gcp_gar, or public source config")
	}
}

// validateExternalImageComponent validates the external_image component config
// shape. Mirrors validation from
// services/ctl-api/internal/app/components/service/create_external_image_config.go.
func validateExternalImageComponent(comp *config.Component) error {
	if comp.ExternalImage == nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("external_image config is required"),
			Description: fmt.Sprintf("Component '%s' is missing external_image configuration", comp.Name),
		}
	}

	sources := 0
	if comp.ExternalImage.AWSECRImageConfig != nil {
		sources++
	}
	if comp.ExternalImage.GCPGARImageConfig != nil {
		sources++
	}
	if comp.ExternalImage.PublicImageConfig != nil {
		sources++
	}
	if sources == 0 {
		return stderr.ErrUser{
			Err:         fmt.Errorf("image_source_required"),
			Code:        "image_source_required",
			Description: fmt.Sprintf("Component '%s' requires one of aws_ecr, gcp_gar, or public image source configuration", comp.Name),
		}
	}
	if sources > 1 {
		return stderr.ErrUser{
			Err:         fmt.Errorf("multiple_image_sources"),
			Code:        "multiple_image_sources",
			Description: fmt.Sprintf("Component '%s' must specify exactly one of aws_ecr, gcp_gar, or public image source configuration", comp.Name),
		}
	}

	if comp.ExternalImage.BuildTimeout != "" {
		if err := validation.ValidateBuildTimeout(comp.ExternalImage.BuildTimeout); err != nil {
			return err
		}
	}
	if comp.ExternalImage.DeployTimeout != "" {
		if err := validation.ValidateDeployTimeout(comp.ExternalImage.DeployTimeout); err != nil {
			return err
		}
	}

	return nil
}
