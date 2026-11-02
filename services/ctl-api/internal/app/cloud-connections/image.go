package cloudconnections

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"golang.org/x/oauth2"
	"gorm.io/gorm"

	ecr "github.com/nuonco/nuon/pkg/aws/ecr-authorization"
	"github.com/nuonco/nuon/pkg/oci/metadata"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ImageAuthorizer interface {
	ECRAuthorization(context.Context, string, *app.AWSECRImageConfig, string) (*ecr.Authorization, error)
	GARAuthorization(context.Context, string, *app.GCPGARImageConfig, string) (*oauth2.Token, error)
	ACRAuthorization(context.Context, string, *app.AzureACRImageConfig, string) (*metadata.RegistryAuth, error)
}

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
	if connection == nil {
		return nil
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
	key, _, err := ImageRepositoryCoordinates(&filter, repository)
	if err != nil {
		return nil, err
	}
	if filter.ID == "" {
		var mapping app.ImageRepository
		if err := db.WithContext(ctx).Where(app.ImageRepository{OrgID: filter.OrgID, CanonicalKey: key}).First(&mapping).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("no saved image repository mapping for %s; Add this repository in Settings > Image repositories, then sync", repository)
			}
			return nil, fmt.Errorf("unable to resolve image repository mapping: %w", err)
		}
		filter.ID = mapping.CloudConnectionID
		if filter.ID == "" {
			return nil, fmt.Errorf("image repository mapping has no cloud connection")
		}
	}
	filter.Status = app.CloudConnectionStatusVerified
	var connection app.CloudConnection
	if err := db.WithContext(ctx).Where(filter).First(&connection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("the selected %s image connection for %s is unavailable, does not match the repository, or is not verified; check Settings > Image repositories and verify its Cloud Connection, then sync", filter.Platform, repository)
		}
		return nil, fmt.Errorf("unable to resolve image repository cloud connection: %w", err)
	}
	return &connection, nil
}

func ImageRepositoryCoordinates(connection *app.CloudConnection, repository string) (string, string, error) {
	if err := ValidateImageRepository(connection, repository); err != nil {
		return "", "", err
	}
	key := string(connection.Preset) + ":" + repository
	switch connection.Preset {
	case app.CloudConnectionPresetECR:
		if connection.Platform != app.CloudPlatformAWS {
			return "", "", fmt.Errorf("ecr image repositories require an aws cloud connection")
		}
		resource, _ := arn.Parse(repository)
		prefix := fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s", connection.TargetID, connection.DefaultRegion, strings.TrimPrefix(resource.Resource, "repository/"))
		return key, prefix, nil
	case app.CloudConnectionPresetGAR:
		if connection.Platform != app.CloudPlatformGCP {
			return "", "", fmt.Errorf("gar image repositories require a gcp cloud connection")
		}
		parts := garRepositoryResourcePattern.FindStringSubmatch(repository)
		return key, fmt.Sprintf("%s-docker.pkg.dev/%s/%s", connection.DefaultRegion, connection.TargetID, parts[3]), nil
	case app.CloudConnectionPresetACR:
		if connection.Platform != app.CloudPlatformAzure || connection.AzureRegistryLoginServer == "" {
			return "", "", fmt.Errorf("acr image repositories require an azure cloud connection with a registry login server")
		}
		prefix := connection.AzureRegistryLoginServer + "/" + repository
		return string(connection.Preset) + ":" + prefix, prefix, nil
	default:
		return "", "", fmt.Errorf("image repositories require an ecr, gar, or acr image connection")
	}
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
