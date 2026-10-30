package cloudconnections

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func ResolveImageConnection(ctx context.Context, db *gorm.DB, orgID string, componentConfig *app.ComponentConfigConnection) error {
	cfg := componentConfig.ExternalImageComponentConfig
	if cfg == nil {
		return nil
	}
	var connection *app.CloudConnection
	var err error
	switch {
	case cfg.AWSECRImageConfig != nil:
		connection, err = ECRConnection(ctx, db, orgID, cfg.AWSECRImageConfig, cfg.ImageURL)
	case cfg.GCPGARImageConfig != nil:
		connection, err = GARConnection(ctx, db, orgID, cfg.GCPGARImageConfig, cfg.ImageURL)
	case cfg.AzureACRImageConfig != nil:
		acr := cfg.AzureACRImageConfig
		if !acr.UsesCloudConnection() {
			return nil
		}
		connection, err = ACRConnection(ctx, db, orgID, acr, cfg.ImageURL)
	default:
		return nil
	}
	if err != nil {
		return err
	}
	if componentConfig.Checksum != "" {
		componentConfig.Checksum = fmt.Sprintf("%x", sha256.Sum256([]byte(componentConfig.Checksum+":"+connection.ID)))
	}
	return nil
}

func imageConnection(ctx context.Context, db *gorm.DB, filter app.CloudConnection, repository string) (*app.CloudConnection, error) {
	if filter.OrgID == "" {
		return nil, fmt.Errorf("image repository access requires an organization")
	}
	if err := ValidateImageRepository(&filter, repository); err != nil {
		return nil, err
	}
	filter.Status = app.CloudConnectionStatusVerified
	query := db.WithContext(ctx).Where(filter)
	var connections []app.CloudConnection
	if err := query.Limit(2).Find(&connections).Error; err != nil {
		return nil, fmt.Errorf("unable to resolve image repository cloud connection: %w", err)
	}
	if len(connections) == 0 {
		return nil, fmt.Errorf("no verified %s image connection in this organization matches the cloud target for %s; create and verify a connection in Settings > Cloud connections, then sync the app configuration", filter.Platform, repository)
	}
	if len(connections) > 1 {
		return nil, fmt.Errorf("multiple verified %s image connections in this organization match the cloud target for %s; keep one image connection for this target in Settings > Cloud connections, then sync the app configuration", filter.Platform, repository)
	}
	return &connections[0], nil
}

var imageRepositoryNamePattern = regexp.MustCompile(`^[a-z0-9]+(?:(?:[._]|__|-+)[a-z0-9]+)*(?:/[a-z0-9]+(?:(?:[._]|__|-+)[a-z0-9]+)*)*$`)
var garRepositoryResourcePattern = regexp.MustCompile(`^projects/([^/]+)/locations/([a-z][a-z0-9-]*)/repositories/([a-z][a-z0-9_-]{0,62})$`)

func ValidateImageRepository(connection *app.CloudConnection, repository string) error {
	switch connection.Preset {
	case app.CloudConnectionPresetECR:
		resource, err := arn.Parse(repository)
		name := strings.TrimPrefix(resource.Resource, "repository/")
		if err != nil || resource.Partition != "aws" || resource.Service != "ecr" || resource.AccountID != connection.TargetID || resource.Region != connection.DefaultRegion || !strings.HasPrefix(resource.Resource, "repository/") || len(name) < 2 || len(name) > 256 || !imageRepositoryNamePattern.MatchString(name) {
			return fmt.Errorf("ECR repository must be an exact ARN in account %s and region %s: %s", connection.TargetID, connection.DefaultRegion, repository)
		}
	case app.CloudConnectionPresetGAR:
		parts := garRepositoryResourcePattern.FindStringSubmatch(repository)
		if parts == nil || parts[1] != connection.TargetID || parts[2] != connection.DefaultRegion {
			return fmt.Errorf("GAR repository must be an exact resource in project %s and location %s: %s", connection.TargetID, connection.DefaultRegion, repository)
		}
	case app.CloudConnectionPresetACR:
		if len(repository) > 256 || !imageRepositoryNamePattern.MatchString(repository) {
			return fmt.Errorf("ACR repository must be an exact lowercase path without hostnames, wildcards, tags, or digests: %s", repository)
		}
	default:
		return fmt.Errorf("repository grants require an ecr, gar, or acr image connection")
	}
	return nil
}
