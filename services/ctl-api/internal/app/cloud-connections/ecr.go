package cloudconnections

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

var ecrImagePattern = regexp.MustCompile(`^([0-9]{12})\.dkr\.ecr\.([a-z0-9-]+)\.amazonaws\.com/(.+)$`)

func ecrHasLegacyAuth(cfg *app.AWSECRImageConfig) bool {
	return cfg.IAMRoleARN != ""
}

func ECRConnection(ctx context.Context, db *gorm.DB, orgID string, cfg *app.AWSECRImageConfig, imageURL string, legacy bool) (*app.CloudConnection, error) {
	if cfg == nil {
		return nil, fmt.Errorf("ECR image configuration is required")
	}
	if ecrHasLegacyAuth(cfg) {
		if cfg.CloudConnectionID != "" {
			return nil, stderr.NewInvalidRequest(errors.New("ECR cloud connection access cannot be combined with iam_role_arn"))
		}
		return legacyImageAuth(legacy, ErrLegacyImageAuthDisabled)
	}
	if cfg.ID != "" && cfg.CloudConnectionID == "" {
		return nil, stderr.NewInvalidRequest(errors.New("ECR image configuration has no pinned cloud connection; sync the app configuration before building"))
	}
	parts := ecrImagePattern.FindStringSubmatch(imageURL)
	if parts == nil || parts[2] != cfg.AWSRegion {
		return nil, stderr.NewInvalidRequest(errors.New("ECR image URL must name an exact repository in the configured AWS region"))
	}
	repository := fmt.Sprintf("arn:aws:ecr:%s:%s:repository/%s", parts[2], parts[1], parts[3])
	connection, err := imageConnection(ctx, db, app.CloudConnection{
		ID: cfg.CloudConnectionID, OrgID: orgID, Platform: app.CloudPlatformAWS,
		Preset: app.CloudConnectionPresetECR, TargetID: parts[1], DefaultRegion: cfg.AWSRegion,
	}, repository)
	if err != nil {
		return nil, err
	}
	cfg.CloudConnectionID = connection.ID
	return connection, nil
}
