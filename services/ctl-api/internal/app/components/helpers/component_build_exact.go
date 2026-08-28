package helpers

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// CreateComponentBuildFromConfigConnectionRequest identifies an immutable input
// to a component build. GitRef must already have been resolved by the caller;
// this helper never contacts the VCS provider or substitutes a configured branch.
type CreateComponentBuildFromConfigConnectionRequest struct {
	BuildID                     string
	ComponentConfigConnectionID string
	ResolvedGitCommitSHA        *string
}

// CreateComponentBuildFromConfigConnection creates a build pinned to the exact
// config connection and source ref in req. BuildID may be empty to use the
// normal model-generated ID.
func (s *Helpers) CreateComponentBuildFromConfigConnection(ctx context.Context, req CreateComponentBuildFromConfigConnectionRequest) (*app.ComponentBuild, error) {
	req = normalizeExactBuildRequest(req)
	if err := validateExactBuildRequest(req); err != nil {
		return nil, err
	}

	config, err := s.getComponentConfigConnectionForBuild(ctx, req.ComponentConfigConnectionID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component config connection: %w", err)
	}
	if err := validateExactBuildConfig(req, config); err != nil {
		return nil, err
	}

	build := &app.ComponentBuild{
		ID:                          req.BuildID,
		Status:                      "queued",
		StatusDescription:           "queued and waiting for runner to pick up",
		GitRef:                      req.ResolvedGitCommitSHA,
		ComponentConfigConnectionID: req.ComponentConfigConnectionID,
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if req.BuildID != "" {
			var existing app.ComponentBuild
			err := tx.Where(app.ComponentBuild{ID: req.BuildID}).First(&existing).Error
			if err == nil {
				if existing.ComponentConfigConnectionID != req.ComponentConfigConnectionID || !sameStringPointer(existing.GitRef, req.ResolvedGitCommitSHA) {
					return fmt.Errorf("build ID %s already exists with different immutable inputs", req.BuildID)
				}
				build = &existing
				return nil
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("unable to check existing component build: %w", err)
			}
		}
		if err := tx.Create(build).Error; err != nil {
			return fmt.Errorf("unable to create build for component: %w", err)
		}
		result := tx.Model(&app.ComponentConfigConnection{}).
			Where(app.ComponentConfigConnection{ID: req.ComponentConfigConnectionID}).
			Update("latest_build_id", build.ID)
		if result.Error != nil {
			return fmt.Errorf("unable to set latest_build_id on config connection: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("unable to set latest_build_id: exact config connection is no longer active")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return build, nil
}

func (s *Helpers) getComponentConfigConnectionForBuild(ctx context.Context, id string) (*app.ComponentConfigConnection, error) {
	var config app.ComponentConfigConnection
	res := s.db.WithContext(ctx).
		Preload("Component").
		Preload("Component.Org").
		Preload("TerraformModuleComponentConfig").
		Preload("TerraformModuleComponentConfig.PublicGitVCSConfig").
		Preload("TerraformModuleComponentConfig.ConnectedGithubVCSConfig").
		Preload("HelmComponentConfig").
		Preload("HelmComponentConfig.PublicGitVCSConfig").
		Preload("HelmComponentConfig.ConnectedGithubVCSConfig").
		Preload("DockerBuildComponentConfig").
		Preload("DockerBuildComponentConfig.PublicGitVCSConfig").
		Preload("DockerBuildComponentConfig.ConnectedGithubVCSConfig").
		Preload("ExternalImageComponentConfig").
		Preload("JobComponentConfig").
		Preload("KubernetesManifestComponentConfig").
		Preload("KubernetesManifestComponentConfig.PublicGitVCSConfig").
		Preload("KubernetesManifestComponentConfig.ConnectedGithubVCSConfig").
		Preload("PulumiComponentConfig").
		Preload("PulumiComponentConfig.PublicGitVCSConfig").
		Preload("PulumiComponentConfig.ConnectedGithubVCSConfig").
		Where(app.ComponentConfigConnection{ID: id}).
		First(&config)
	if res.Error != nil {
		return nil, res.Error
	}
	return &config, nil
}

func validateExactBuildRequest(req CreateComponentBuildFromConfigConnectionRequest) error {
	if strings.TrimSpace(req.ComponentConfigConnectionID) == "" {
		return fmt.Errorf("component config connection ID is required")
	}
	return nil
}

func validateExactBuildConfig(req CreateComponentBuildFromConfigConnectionRequest, config *app.ComponentConfigConnection) error {
	if config == nil || config.ID == "" || config.Type == app.ComponentTypeUnknown {
		return stderr.ErrUser{Err: fmt.Errorf("no config found on component"), Description: "please create a component config before building"}
	}
	if config.ID != req.ComponentConfigConnectionID || config.ComponentID == "" || config.Component.ID != config.ComponentID {
		return fmt.Errorf("component config connection does not have a valid owning component")
	}
	if config.VCSConnectionType != app.VCSConnectionTypeNone {
		if req.ResolvedGitCommitSHA == nil || !fullGitCommitSHA.MatchString(strings.TrimSpace(*req.ResolvedGitCommitSHA)) {
			return fmt.Errorf("full resolved Git commit SHA is required for VCS-backed component config")
		}
	}
	return nil
}

var fullGitCommitSHA = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)

func normalizeExactBuildRequest(req CreateComponentBuildFromConfigConnectionRequest) CreateComponentBuildFromConfigConnectionRequest {
	if req.ResolvedGitCommitSHA != nil {
		normalized := strings.ToLower(strings.TrimSpace(*req.ResolvedGitCommitSHA))
		req.ResolvedGitCommitSHA = &normalized
	}
	return req
}

func sameStringPointer(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}
