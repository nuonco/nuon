package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type BackfillInstallCloudMetadataRequest struct {
	InstallID string `json:"install_id" validate:"required"`
}

type BackfillInstallCloudMetadataResponse struct {
	Updated bool `json:"updated"`
	// Identifier is what was written, for the caller's logs. Empty when nothing
	// changed.
	Identifier string `json:"identifier,omitempty"`
	SkipReason string `json:"skip_reason,omitempty"`
}

const (
	cloudMetadataSkipNoStackOutputs = "install has no stack outputs to backfill from"
	cloudMetadataSkipNoIdentifier   = "stack outputs carry no cloud account identifier"
	cloudMetadataSkipUserSupplied   = "target identifier was supplied by the user"
)

// BackfillInstallCloudMetadata populates an install's CloudPlatformMetadata from the
// identifier its stack already reported.
//
// This is the first half of onboarding an org to phone-home auth. Nothing downstream
// works without it: the secret's resource policy names a role in a specific account,
// and EnsureInstallPhoneHomeSecret skips any install with no target, so an install
// created before target_account_id existed can never be provisioned until this runs.
//
// Both observed_* and target_* are written, with target_source recording that the
// value came from here. The provenance matters and is the reason for the field: these
// identifiers arrived over *unauthenticated* phone homes, so a backfilled target is
// only as trustworthy as whichever account last called in. That is a deliberate trade
// for migrating live installs — it is strictly better than the status quo, which
// trusts every future phone home too — but it means target_* is a strong control only
// for installs created after phone-home auth shipped.
//
// Re-runnable: a target supplied by a user or derived from an account connection is
// never overwritten, so running this after someone has set the value by hand is a
// no-op rather than a silent downgrade.
//
// @temporal-gen-v2 activity
// @by-field InstallID
// @start-to-close-timeout 2m
func (a *Activities) BackfillInstallCloudMetadata(
	ctx context.Context, req *BackfillInstallCloudMetadataRequest,
) (*BackfillInstallCloudMetadataResponse, error) {
	if err := a.v.StructCtx(ctx, req); err != nil {
		return nil, fmt.Errorf("unable to validate request: %w", err)
	}

	var install app.Install
	if res := a.db.WithContext(ctx).
		Preload("InstallStack.InstallStackOutputs").
		Where(app.Install{ID: req.InstallID}).
		First(&install); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error)
	}

	cpm := install.CloudPlatformMetadata
	if cpm.TargetSource == app.CloudPlatformTargetSourceUser ||
		cpm.TargetSource == app.CloudPlatformTargetSourceConnection {
		return &BackfillInstallCloudMetadataResponse{SkipReason: cloudMetadataSkipUserSupplied}, nil
	}

	if install.InstallStack == nil {
		return &BackfillInstallCloudMetadataResponse{SkipReason: cloudMetadataSkipNoStackOutputs}, nil
	}
	outputs := install.InstallStack.InstallStackOutputs

	identifier := ""
	switch {
	case outputs.AWSStackOutputs != nil && outputs.AWSStackOutputs.AccountID != "":
		identifier = outputs.AWSStackOutputs.AccountID
		cpm.ObservedAccountID = identifier
		cpm.TargetAccountID = identifier
	case outputs.GCPStackOutputs != nil && outputs.GCPStackOutputs.ProjectID != "":
		identifier = outputs.GCPStackOutputs.ProjectID
		cpm.ObservedProjectID = identifier
		cpm.TargetProjectID = identifier
	case outputs.AzureStackOutputs != nil && outputs.AzureStackOutputs.SubscriptionID != "":
		identifier = outputs.AzureStackOutputs.SubscriptionID
		cpm.ObservedSubscriptionID = identifier
		cpm.TargetSubscriptionID = identifier
	default:
		return &BackfillInstallCloudMetadataResponse{SkipReason: cloudMetadataSkipNoIdentifier}, nil
	}

	if cpm == install.CloudPlatformMetadata {
		return &BackfillInstallCloudMetadataResponse{Identifier: identifier}, nil
	}
	cpm.TargetSource = app.CloudPlatformTargetSourceBackfill

	if res := a.db.WithContext(ctx).
		Model(&app.Install{ID: install.ID}).
		Update("cloud_platform_metadata", cpm); res.Error != nil {
		return nil, generics.TemporalGormError(res.Error)
	}

	a.l.Info("backfilled install cloud platform metadata",
		zap.String("install_id", install.ID),
		zap.String("org_id", install.OrgID),
		zap.String("identifier", identifier),
	)

	return &BackfillInstallCloudMetadataResponse{Updated: true, Identifier: identifier}, nil
}
