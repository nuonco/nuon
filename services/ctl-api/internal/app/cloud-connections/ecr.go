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

// appSyncDisabled reports whether the org has disable-app-sync, including
// when the flag is forced on for this deployment.
func appSyncDisabled(ctx context.Context, db *gorm.DB, orgID string) bool {
	if app.FeatureForced(app.OrgFeatureDisableAppSync) {
		return true
	}
	if db == nil {
		return false
	}
	var org app.Org
	if err := db.WithContext(ctx).First(&org, "id = ?", orgID).Error; err != nil {
		return false
	}
	return org.Features[string(app.OrgFeatureDisableAppSync)]
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
		msg := "ECR image configuration has no pinned cloud connection; run `nuon apps sync` or run your app branch before building"
		if appSyncDisabled(ctx, db, orgID) {
			msg = "ECR image configuration has no pinned cloud connection; run your app branch before building"
		}
		return nil, stderr.NewInvalidRequest(errors.New(msg))
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
