package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/robfig/cron"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
)

type CreateTerraformModuleComponentConfigRequest struct {
	basicVCSConfigRequest

	Version        string             `json:"version"`
	Variables      map[string]*string `json:"variables" validate:"required"`
	VariablesFiles []string           `json:"variables_files,omitempty"`
	EnvVars        map[string]*string `json:"env_vars" validate:"required"`

	AppConfigID string `json:"app_config_id"`

	Dependencies  []string `json:"dependencies"`
	References    []string `json:"references"`
	Checksum      string   `json:"checksum"`
	DriftSchedule *string  `json:"drift_schedule,omitempty"`
}

type LatestTerraformVersion struct {
	Version   string
	Timestamp time.Time
}

const MinTerraformVersion = "1.8.0"

func (c *CreateTerraformModuleComponentConfigRequest) Validate(v *validator.Validate, latestVersion string) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	if c.Version != "" {
		if err := c.validateVersion(latestVersion); err != nil {
			return fmt.Errorf("invalid version %s: %w", c.Version, err)
		}
	}

	if err := c.basicVCSConfigRequest.Validate(); err != nil {
		return err
	}
	return nil
}

func (c *CreateTerraformModuleComponentConfigRequest) validateVersion(latestVersion string) error {
	minConstraint := fmt.Sprintf(">= %s", MinTerraformVersion)
	maxConstraint := fmt.Sprintf("<= %s", latestVersion)
	constraint, err := semver.NewConstraint(fmt.Sprintf("%s, %s", minConstraint, maxConstraint))
	if err != nil {
		return fmt.Errorf("failed to create version constraint: %w", err)
	}

	version, err := semver.NewVersion(c.Version)
	if err != nil {
		return fmt.Errorf("failed to parse version %s: %w", c.Version, err)
	}

	if !constraint.Check(version) {
		return fmt.Errorf("version %s does not satisfy constraint %s", c.Version, constraint)
	}

	return nil
}

// @ID						CreateAppTerraformModuleComponentConfig
// @Summary				create a terraform component config
// @Description.markdown	create_terraform_component_config.md
// @Param					req				body	CreateTerraformModuleComponentConfigRequest	true	"Input"
// @Param					app_id			path	string										true	"app ID"
// @Param					component_id	path	string										true	"component ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.TerraformModuleComponentConfig
// @Router					/v1/apps/{app_id}/components/{component_id}/configs/terraform-module [POST]
func (s *service) CreateAppTerraformModuleComponentConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	cmpID := ctx.Param("component_id")
	_, err := s.getAppComponent(ctx, appID, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component: %w", err))
		return
	}

	// reuse the same logic as non-app scoped endpoint
	s.CreateTerraformModuleComponentConfig(ctx)

}

// @ID						CreateTerraformModuleComponentConfig
// @Summary				create a terraform component config
// @Description.markdown	create_terraform_component_config.md
// @Param					req				body	CreateTerraformModuleComponentConfigRequest	true	"Input"
// @Param					component_id	path	string										true	"component ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated 	  true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.TerraformModuleComponentConfig
// @Router					/v1/components/{component_id}/configs/terraform-module [POST]
func (s *service) CreateTerraformModuleComponentConfig(ctx *gin.Context) {
	cmpID := ctx.Param("component_id")

	var req CreateTerraformModuleComponentConfigRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	latestVersion, err := getLatestTerraformVersion()
	if err != nil {
		ctx.Error(fmt.Errorf("unable to fetch latest terraform version: %w", err))
		return
	}

	if req.Version == "" {
		req.Version = latestVersion
	}

	if err := req.Validate(s.v, latestVersion); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	cfg, err := s.createTerraformModuleComponentConfig(ctx, cmpID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create component cfg: %w", err))
		return
	}

	s.evClient.Send(ctx, cmpID, &signals.Signal{
		Type: signals.OperationConfigCreated,
	})

	s.evClient.Send(ctx, cmpID, &signals.Signal{
		Type:          signals.OperationUpdateComponentType,
		ComponentType: app.ComponentTypeTerraformModule,
	})
	ctx.JSON(http.StatusCreated, cfg)
}

func (s *service) createTerraformModuleComponentConfig(ctx context.Context, cmpID string, req *CreateTerraformModuleComponentConfigRequest) (*app.TerraformModuleComponentConfig, error) {
	parentCmp, err := s.getComponentWithParents(ctx, cmpID)
	if err != nil {
		return nil, err
	}

	depIDs, err := s.helpers.GetComponentIDs(ctx, parentCmp.AppID, req.Dependencies)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component ids")
	}

	// build component config
	connectedGithubVCSConfig, err := req.connectedGithubVCSConfig(ctx, parentCmp, s.vcsHelpers)
	if err != nil {
		return nil, fmt.Errorf("invalid connected github config: %w", err)
	}

	publicGitVCSConfig, err := req.publicGitVCSConfig(ctx, parentCmp, s.vcsHelpers)
	if err != nil {
		return nil, fmt.Errorf("invalid public vcs config: %w", err)
	}

	cfg := app.TerraformModuleComponentConfig{
		Version:                  req.Version,
		PublicGitVCSConfig:       publicGitVCSConfig,
		ConnectedGithubVCSConfig: connectedGithubVCSConfig,
		Variables:                pgtype.Hstore(req.Variables),
		EnvVars:                  pgtype.Hstore(req.EnvVars),
		VariablesFiles:           pq.StringArray(req.VariablesFiles),
	}

	componentConfigConnection := app.ComponentConfigConnection{
		TerraformModuleComponentConfig: &cfg,
		ComponentID:                    parentCmp.ID,
		AppConfigID:                    req.AppConfigID,
		ComponentDependencyIDs:         pq.StringArray(depIDs),
		References:                     pq.StringArray(req.References),
		Checksum:                       req.Checksum,
	}
	if req.DriftSchedule != nil {
		_, err := cron.ParseStandard(*req.DriftSchedule)
		if err != nil {
			return nil, fmt.Errorf("invalid drift schedule: must be a valid cron expression: %s . Error: %s", *req.DriftSchedule, err.Error())
		}
		componentConfigConnection.DriftSchedule = *req.DriftSchedule
	}
	if res := s.db.WithContext(ctx).Create(&componentConfigConnection); res.Error != nil {
		return nil, fmt.Errorf("unable to create terraform component config connection: %w", res.Error)
	}

	err = s.helpers.UpdateComponentType(ctx, cmpID, app.ComponentTypeTerraformModule)
	if err != nil {
		return nil, fmt.Errorf("unable to update component type: %w", err)
	}

	return &cfg, nil
}

type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Draft      bool   `json:"draft"`
	Prerelease bool   `json:"prerelease"`
}

var latestTerraformVersion = LatestTerraformVersion{}

func getLatestTerraformVersion() (string, error) {
	fiveMinutes := 5 * time.Minute
	if latestTerraformVersion.Version != "" && time.Since(latestTerraformVersion.Timestamp) < fiveMinutes {
		return latestTerraformVersion.Version, nil // 🎯 Cache hit - no API call
	}

	url := "https://api.github.com/repos/hashicorp/terraform/releases/latest"

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Remove 'v' prefix if present
	version := strings.TrimPrefix(release.TagName, "v")

	latestTerraformVersion = LatestTerraformVersion{
		Version:   version,
		Timestamp: time.Now(),
	}

	return version, nil
}
