package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/config"
	configgithub "github.com/nuonco/nuon/pkg/config/github"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	actionsignals "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals"
	componentsignals "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

// CreateAppFromTemplateRequest defines the request body for creating an app from a template
type CreateAppFromTemplateRequest struct {
	Template string `json:"template" validate:"required"`
}

func (c *CreateAppFromTemplateRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	if !configgithub.AllowedTemplates[c.Template] {
		allowedList := make([]string, 0, len(configgithub.AllowedTemplates))
		for t := range configgithub.AllowedTemplates {
			allowedList = append(allowedList, t)
		}
		return fmt.Errorf("template must be one of: %s", strings.Join(allowedList, ", "))
	}

	return nil
}

// @ID						CreateAppFromTemplate
// @Summary				create an app from a template
// @Description			Creates a new app from a predefined template and syncs all configuration
// @Tags					apps
// @Accept					json
// @Param					req	body	CreateAppFromTemplateRequest	true	"Input"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.App
// @Router					/v1/apps/from-template [post]
func (s *service) CreateAppFromTemplate(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	user, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateAppFromTemplateRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	// Fetch and parse the template config from GitHub
	fetcher := configgithub.NewFetcher()
	templateCfg, err := fetcher.FetchTemplateConfig(ctx, req.Template)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to fetch template config: %w", err))
		return
	}

	// Create the app using the template name
	appName := req.Template
	displayName := templateCfg.DisplayName
	if displayName == "" {
		displayName = req.Template
	}

	createAppReq := &CreateAppRequest{
		Name:        appName,
		Description: templateCfg.Description,
		DisplayName: displayName,
	}

	newApp, err := s.createApp(ctx, user, org, createAppReq)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	// Create app config
	// Use service version if it's valid semver, otherwise "development" (matches CLI pattern)
	// In production/staging: s.cfg.Version is a proper semver (e.g., "v1.2.3") → use it
	// In local dev: s.cfg.Version is a git hash (e.g., "75e5ddd") → use "development"
	cliVersion := "development"
	if _, semverErr := semver.NewVersion(s.cfg.Version); semverErr == nil {
		cliVersion = s.cfg.Version
	}
	appConfig, err := s.createAppConfig(ctx, org.ID, newApp.ID, &CreateAppConfigRequest{
		Readme:     templateCfg.Readme,
		CLIVersion: cliVersion,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app config: %w", err))
		return
	}

	// Sync all configuration from the template
	if err := s.syncTemplateConfig(ctx, org, newApp, appConfig, templateCfg); err != nil {
		s.l.Error("failed to sync template config",
			zap.String("app_id", newApp.ID),
			zap.String("template", req.Template),
			zap.Error(err))
		ctx.Error(fmt.Errorf("failed to sync template config: %w", err))
		return
	}

	// Update user journey for first app creation
	if err := s.accountsHelpers.UpdateUserJourneyStepForFirstAppCreate(ctx, user.ID, newApp.ID); err != nil {
		s.l.Warn("failed to update user journey for first app creation",
			zap.String("account_id", user.ID),
			zap.String("app_id", newApp.ID),
			zap.Error(err))
	}

	// Send signals to event loop
	s.evClient.Send(ctx, newApp.ID, &signals.Signal{
		Type: signals.OperationCreated,
	})
	s.evClient.Send(ctx, newApp.ID, &signals.Signal{
		Type: signals.OperationPollDependencies,
	})
	s.evClient.Send(ctx, newApp.ID, &signals.Signal{
		Type: signals.OperationProvision,
	})

	ctx.JSON(http.StatusCreated, newApp)
}

// syncTemplateConfig syncs all configuration from the parsed template config
func (s *service) syncTemplateConfig(ctx context.Context, org *app.Org, newApp *app.App, appConfig *app.AppConfig, cfg *config.AppConfig) error {
	// Create sandbox config if present
	if cfg.Sandbox != nil {
		if err := s.syncSandboxConfig(ctx, org.ID, newApp, appConfig, cfg.Sandbox); err != nil {
			return fmt.Errorf("failed to sync sandbox config: %w", err)
		}
	}

	// Create runner config if present
	if cfg.Runner != nil {
		if err := s.syncRunnerConfig(ctx, org.ID, newApp, appConfig, cfg.Runner); err != nil {
			return fmt.Errorf("failed to sync runner config: %w", err)
		}
	}

	// Create inputs config if present
	if cfg.Inputs != nil {
		if err := s.syncInputsConfig(ctx, org.ID, newApp, appConfig, cfg.Inputs); err != nil {
			return fmt.Errorf("failed to sync inputs config: %w", err)
		}
	}

	// Sync components if present
	if len(cfg.Components) > 0 {
		if err := s.syncComponents(ctx, newApp, appConfig, cfg.Components); err != nil {
			return fmt.Errorf("failed to sync components: %w", err)
		}
	}

	// Sync permissions config if present
	if cfg.Permissions != nil {
		if err := s.syncPermissionsConfig(ctx, org.ID, newApp, appConfig, cfg.Permissions); err != nil {
			return fmt.Errorf("failed to sync permissions config: %w", err)
		}
	}

	// Sync policies config if present
	if cfg.Policies != nil && len(cfg.Policies.Policies) > 0 {
		if err := s.syncPoliciesConfig(ctx, org.ID, newApp, appConfig, cfg.Policies); err != nil {
			return fmt.Errorf("failed to sync policies config: %w", err)
		}
	}

	// Sync break-glass config if present
	if cfg.BreakGlass != nil && len(cfg.BreakGlass.Roles) > 0 {
		if err := s.syncBreakGlassConfig(ctx, org.ID, newApp, appConfig, cfg.BreakGlass); err != nil {
			return fmt.Errorf("failed to sync break glass config: %w", err)
		}
	}

	// Sync actions config if present (MUST be after components for dependency resolution)
	if len(cfg.Actions) > 0 {
		if err := s.syncActionsConfig(ctx, org.ID, newApp, appConfig, cfg.Actions); err != nil {
			return fmt.Errorf("failed to sync actions config: %w", err)
		}
	}

	return nil
}

// syncSandboxConfig creates a sandbox config from the template
func (s *service) syncSandboxConfig(ctx context.Context, orgID string, newApp *app.App, appConfig *app.AppConfig, sandbox *config.AppSandboxConfig) error {
	variables := make(map[string]*string)
	for k, v := range sandbox.VarsMap {
		val := v
		variables[k] = &val
	}

	envVars := make(map[string]*string)
	for k, v := range sandbox.EnvVarMap {
		val := v
		envVars[k] = &val
	}

	variablesFiles := make([]string, 0)
	for _, vf := range sandbox.VariablesFiles {
		variablesFiles = append(variablesFiles, vf.Contents)
	}

	references := make([]string, 0)
	for _, ref := range sandbox.References {
		references = append(references, ref.String())
	}

	sandboxConfig := app.AppSandboxConfig{
		OrgID:            orgID,
		AppID:            newApp.ID,
		AppConfigID:      appConfig.ID,
		Variables:        pgtype.Hstore(variables),
		EnvVars:          pgtype.Hstore(envVars),
		VariablesFiles:   pq.StringArray(variablesFiles),
		TerraformVersion: sandbox.TerraformVersion,
		References:       pq.StringArray(references),
	}

	if sandbox.DriftSchedule != nil {
		sandboxConfig.DriftSchedule = *sandbox.DriftSchedule
	}

	// Handle connected repo config
	if sandbox.ConnectedRepo != nil {
		sandboxConfig.ConnectedGithubVCSConfig = &app.ConnectedGithubVCSConfig{
			Repo:      sandbox.ConnectedRepo.Repo,
			Branch:    sandbox.ConnectedRepo.Branch,
			Directory: sandbox.ConnectedRepo.Directory,
		}
	}

	// Handle public repo config
	if sandbox.PublicRepo != nil {
		sandboxConfig.PublicGitVCSConfig = &app.PublicGitVCSConfig{
			Repo:      sandbox.PublicRepo.Repo,
			Branch:    sandbox.PublicRepo.Branch,
			Directory: sandbox.PublicRepo.Directory,
		}
	}

	res := s.db.WithContext(ctx).Create(&sandboxConfig)
	if res.Error != nil {
		return fmt.Errorf("unable to create sandbox config: %w", res.Error)
	}

	// Send signal to update sandbox
	s.evClient.Send(ctx, newApp.ID, &signals.Signal{
		Type:               signals.OperationUpdateSandbox,
		AppSandboxConfigID: sandboxConfig.ID,
	})

	return nil
}

// syncRunnerConfig creates a runner config from the template
func (s *service) syncRunnerConfig(ctx context.Context, orgID string, newApp *app.App, appConfig *app.AppConfig, runner *config.AppRunnerConfig) error {
	envVars := make(map[string]*string)
	for k, v := range runner.EnvVarMap {
		val := v
		envVars[k] = &val
	}

	runnerConfig := app.AppRunnerConfig{
		OrgID:       orgID,
		AppID:       newApp.ID,
		AppConfigID: appConfig.ID,
		Type:        app.AppRunnerType(runner.RunnerType),
		EnvVars:     pgtype.Hstore(envVars),
	}

	if runner.HelmDriver != "" {
		runnerConfig.HelmDriver = app.AppRunnerConfigHelmDriverType(runner.HelmDriver)
	}

	if runner.InitScriptURL != "" {
		runnerConfig.InitScriptURL = runner.InitScriptURL
	}

	res := s.db.WithContext(ctx).Create(&runnerConfig)
	if res.Error != nil {
		return fmt.Errorf("unable to create runner config: %w", res.Error)
	}

	return nil
}

// syncInputsConfig creates an inputs config from the template
func (s *service) syncInputsConfig(ctx context.Context, orgID string, newApp *app.App, appConfig *app.AppConfig, inputs *config.AppInputConfig) error {
	// Create input groups
	groups := make([]app.AppInputGroup, 0, len(inputs.Groups))
	for i, grp := range inputs.Groups {
		groups = append(groups, app.AppInputGroup{
			Name:        grp.Name,
			Description: grp.Description,
			DisplayName: grp.DisplayName,
			Index:       i,
		})
	}

	inputConfig := app.AppInputConfig{
		AppConfigID:    appConfig.ID,
		OrgID:          orgID,
		AppID:          newApp.ID,
		AppInputGroups: groups,
	}

	res := s.db.WithContext(ctx).Create(&inputConfig)
	if res.Error != nil {
		return fmt.Errorf("unable to create input config: %w", res.Error)
	}

	// Create inputs
	if len(inputs.Inputs) > 0 {
		appInputs := make([]app.AppInput, 0, len(inputs.Inputs))
		for i, input := range inputs.Inputs {
			var groupID string
			for _, group := range inputConfig.AppInputGroups {
				if group.Name == input.Group {
					groupID = group.ID
					break
				}
			}

			// Convert default value to string
			defaultVal := ""
			if input.Default != nil {
				defaultVal = fmt.Sprintf("%v", input.Default)
			}

			appInputs = append(appInputs, app.AppInput{
				OrgID:            orgID,
				AppInputConfigID: inputConfig.ID,
				AppInputGroupID:  groupID,
				Name:             input.Name,
				Description:      input.Description,
				DisplayName:      input.DisplayName,
				Required:         input.Required,
				Default:          defaultVal,
				Sensitive:        input.Sensitive,
				Type:             app.AppInputType(input.Type),
				Internal:         input.Internal,
				Index:            i,
				Source:           app.AppInputSourceVendor,
			})
		}

		res = s.db.WithContext(ctx).Create(&appInputs)
		if res.Error != nil {
			return fmt.Errorf("unable to create inputs: %w", res.Error)
		}
	}

	return nil
}

// syncComponents creates components and their type-specific configs from the template
func (s *service) syncComponents(ctx context.Context, newApp *app.App, appConfig *app.AppConfig, components []*config.Component) error {
	// Map to track component name -> ID for dependency resolution
	componentIDs := make(map[string]string)

	// First pass: Create all components (without dependencies)
	for _, comp := range components {
		component := app.Component{
			AppID:             newApp.ID,
			Name:              comp.Name,
			VarName:           comp.VarName,
			Status:            "queued",
			StatusDescription: "waiting for event loop to start for component",
		}

		if res := s.db.WithContext(ctx).Create(&component); res.Error != nil {
			return fmt.Errorf("unable to create component %s: %w", comp.Name, res.Error)
		}

		componentIDs[comp.Name] = component.ID

		// Send creation signals to start the event loop
		s.evClient.Send(ctx, component.ID, &componentsignals.Signal{
			Type: componentsignals.OperationCreated,
		})
		s.evClient.Send(ctx, component.ID, &componentsignals.Signal{
			Type: componentsignals.OperationProvision,
		})
		s.evClient.Send(ctx, component.ID, &componentsignals.Signal{
			Type: componentsignals.OperationPollDependencies,
		})
	}

	// Second pass: Create component configs with dependencies resolved
	for _, comp := range components {
		compID := componentIDs[comp.Name]

		// Resolve dependencies to IDs
		depIDs := make([]string, 0, len(comp.Dependencies))
		for _, depName := range comp.Dependencies {
			if depID, ok := componentIDs[depName]; ok {
				depIDs = append(depIDs, depID)
			}
		}

		// Create dependencies if any
		if len(depIDs) > 0 {
			if err := s.componentsHelpers.CreateComponentDependencies(ctx, compID, depIDs); err != nil {
				return fmt.Errorf("unable to create dependencies for %s: %w", comp.Name, err)
			}
		}

		// Create type-specific component config
		if err := s.createComponentConfig(ctx, newApp.ID, compID, appConfig.ID, comp); err != nil {
			return fmt.Errorf("unable to create config for component %s: %w", comp.Name, err)
		}
	}

	// Ensure install components are created for any existing installs
	if err := s.componentsHelpers.EnsureInstallComponents(ctx, newApp.ID, nil); err != nil {
		return fmt.Errorf("unable to ensure install components: %w", err)
	}

	// Update AppConfig with component IDs so GetAppComponents can find them
	componentIDList := make([]string, 0, len(componentIDs))
	for _, id := range componentIDs {
		componentIDList = append(componentIDList, id)
	}

	if err := s.db.WithContext(ctx).Model(appConfig).Update("component_ids", pq.StringArray(componentIDList)).Error; err != nil {
		return fmt.Errorf("unable to update app config with component ids: %w", err)
	}

	return nil
}

// createComponentConfig dispatches to the appropriate type-specific config creation method
func (s *service) createComponentConfig(ctx context.Context, appID, compID, appConfigID string, comp *config.Component) error {
	switch comp.Type {
	case config.TerraformModuleComponentType:
		return s.createTerraformModuleConfig(ctx, compID, appConfigID, comp)
	case config.HelmChartComponentType:
		return s.createHelmChartConfig(ctx, compID, appConfigID, comp)
	case config.DockerBuildComponentType:
		return s.createDockerBuildConfig(ctx, compID, appConfigID, comp)
	case config.ExternalImageComponentType, config.ContainerImageComponentType:
		return s.createExternalImageConfig(ctx, compID, appConfigID, comp)
	case config.KubernetesManifestComponentType:
		return s.createKubernetesManifestConfig(ctx, compID, appConfigID, comp)
	case config.JobComponentType:
		return s.createJobConfig(ctx, compID, appConfigID, comp)
	default:
		return fmt.Errorf("unsupported component type: %s", comp.Type)
	}
}

// createTerraformModuleConfig creates a terraform module component config
func (s *service) createTerraformModuleConfig(ctx context.Context, compID, appConfigID string, comp *config.Component) error {
	tf := comp.TerraformModule
	if tf == nil {
		return fmt.Errorf("terraform_module config is nil for component %s", comp.Name)
	}

	// Convert variables from map
	variables := make(map[string]*string)
	for k, v := range tf.VarsMap {
		val := v
		variables[k] = &val
	}

	// Convert env vars from map
	envVars := make(map[string]*string)
	for k, v := range tf.EnvVarMap {
		val := v
		envVars[k] = &val
	}

	// Convert variable files
	variablesFiles := make([]string, 0, len(tf.VariablesFiles))
	for _, vf := range tf.VariablesFiles {
		variablesFiles = append(variablesFiles, vf.Contents)
	}

	// Convert references to strings
	references := make([]string, 0, len(comp.References))
	for _, ref := range comp.References {
		references = append(references, ref.String())
	}

	cfg := app.TerraformModuleComponentConfig{
		Version:        tf.TerraformVersion,
		Variables:      pgtype.Hstore(variables),
		EnvVars:        pgtype.Hstore(envVars),
		VariablesFiles: pq.StringArray(variablesFiles),
	}

	// Handle VCS configs
	if tf.PublicRepo != nil {
		cfg.PublicGitVCSConfig = &app.PublicGitVCSConfig{
			Repo:      tf.PublicRepo.Repo,
			Branch:    tf.PublicRepo.Branch,
			Directory: tf.PublicRepo.Directory,
		}
	}
	if tf.ConnectedRepo != nil {
		cfg.ConnectedGithubVCSConfig = &app.ConnectedGithubVCSConfig{
			Repo:      tf.ConnectedRepo.Repo,
			Branch:    tf.ConnectedRepo.Branch,
			Directory: tf.ConnectedRepo.Directory,
		}
	}

	// Create config connection
	configConn := app.ComponentConfigConnection{
		TerraformModuleComponentConfig: &cfg,
		ComponentID:                    compID,
		AppConfigID:                    appConfigID,
		References:                     pq.StringArray(references),
		Checksum:                       comp.Checksum,
	}

	if tf.DriftSchedule != nil {
		configConn.DriftSchedule = *tf.DriftSchedule
	}

	if res := s.db.WithContext(ctx).Create(&configConn); res.Error != nil {
		return fmt.Errorf("unable to create terraform config: %w", res.Error)
	}

	// Update component type
	if err := s.componentsHelpers.UpdateComponentType(ctx, compID, app.ComponentTypeTerraformModule); err != nil {
		return fmt.Errorf("unable to update component type: %w", err)
	}

	// Send signals to trigger build
	s.evClient.Send(ctx, compID, &componentsignals.Signal{
		Type: componentsignals.OperationConfigCreated,
	})
	s.evClient.Send(ctx, compID, &componentsignals.Signal{
		Type:          componentsignals.OperationUpdateComponentType,
		ComponentType: app.ComponentTypeTerraformModule,
	})

	return nil
}

// createHelmChartConfig creates a helm chart component config
func (s *service) createHelmChartConfig(ctx context.Context, compID, appConfigID string, comp *config.Component) error {
	helm := comp.HelmChart
	if helm == nil {
		return fmt.Errorf("helm_chart config is nil for component %s", comp.Name)
	}

	// Convert values from map
	values := make(map[string]*string)
	for k, v := range helm.ValuesMap {
		val := v
		values[k] = &val
	}

	// Convert values files
	valuesFiles := make([]string, 0, len(helm.ValuesFiles))
	for _, vf := range helm.ValuesFiles {
		valuesFiles = append(valuesFiles, vf.Contents)
	}

	// Convert references to strings
	references := make([]string, 0, len(comp.References))
	for _, ref := range comp.References {
		references = append(references, ref.String())
	}

	var helmRepoConfig *app.HelmRepoConfig
	if helm.HelmRepo != nil {
		helmRepoConfig = &app.HelmRepoConfig{
			RepoURL: helm.HelmRepo.RepoURL,
			Chart:   helm.HelmRepo.Chart,
			Version: helm.HelmRepo.Version,
		}
	}

	cfg := app.HelmComponentConfig{
		HelmConfig: &app.HelmConfig{
			ChartName:      helm.ChartName,
			Namespace:      helm.Namespace,
			StorageDriver:  helm.StorageDriver,
			HelmRepoConfig: helmRepoConfig,
			Values:         values,
			ValuesFiles:    valuesFiles,
			TakeOwnership:  helm.TakeOwnership,
		},
	}

	// Handle VCS configs
	if helm.PublicRepo != nil {
		cfg.PublicGitVCSConfig = &app.PublicGitVCSConfig{
			Repo:      helm.PublicRepo.Repo,
			Branch:    helm.PublicRepo.Branch,
			Directory: helm.PublicRepo.Directory,
		}
	}
	if helm.ConnectedRepo != nil {
		cfg.ConnectedGithubVCSConfig = &app.ConnectedGithubVCSConfig{
			Repo:      helm.ConnectedRepo.Repo,
			Branch:    helm.ConnectedRepo.Branch,
			Directory: helm.ConnectedRepo.Directory,
		}
	}

	// Create config connection
	configConn := app.ComponentConfigConnection{
		HelmComponentConfig: &cfg,
		ComponentID:         compID,
		AppConfigID:         appConfigID,
		References:          pq.StringArray(references),
		Checksum:            comp.Checksum,
	}

	if helm.DriftSchedule != nil {
		configConn.DriftSchedule = *helm.DriftSchedule
	}

	if res := s.db.WithContext(ctx).Create(&configConn); res.Error != nil {
		return fmt.Errorf("unable to create helm config: %w", res.Error)
	}

	// Update component type
	if err := s.componentsHelpers.UpdateComponentType(ctx, compID, app.ComponentTypeHelmChart); err != nil {
		return fmt.Errorf("unable to update component type: %w", err)
	}

	// Send signals to trigger build
	s.evClient.Send(ctx, compID, &componentsignals.Signal{
		Type: componentsignals.OperationConfigCreated,
	})
	s.evClient.Send(ctx, compID, &componentsignals.Signal{
		Type:          componentsignals.OperationUpdateComponentType,
		ComponentType: app.ComponentTypeHelmChart,
	})

	return nil
}

// createDockerBuildConfig creates a docker build component config
func (s *service) createDockerBuildConfig(ctx context.Context, compID, appConfigID string, comp *config.Component) error {
	docker := comp.DockerBuild
	if docker == nil {
		return fmt.Errorf("docker_build config is nil for component %s", comp.Name)
	}

	// Convert env vars from map
	envVars := make(map[string]*string)
	for k, v := range docker.EnvVarMap {
		val := v
		envVars[k] = &val
	}

	// Convert references to strings
	references := make([]string, 0, len(comp.References))
	for _, ref := range comp.References {
		references = append(references, ref.String())
	}

	cfg := app.DockerBuildComponentConfig{
		Dockerfile: docker.Dockerfile,
		EnvVars:    pgtype.Hstore(envVars),
	}

	// Handle VCS configs
	if docker.PublicRepo != nil {
		cfg.PublicGitVCSConfig = &app.PublicGitVCSConfig{
			Repo:      docker.PublicRepo.Repo,
			Branch:    docker.PublicRepo.Branch,
			Directory: docker.PublicRepo.Directory,
		}
	}
	if docker.ConnectedRepo != nil {
		cfg.ConnectedGithubVCSConfig = &app.ConnectedGithubVCSConfig{
			Repo:      docker.ConnectedRepo.Repo,
			Branch:    docker.ConnectedRepo.Branch,
			Directory: docker.ConnectedRepo.Directory,
		}
	}

	// Create config connection
	configConn := app.ComponentConfigConnection{
		DockerBuildComponentConfig: &cfg,
		ComponentID:                compID,
		AppConfigID:                appConfigID,
		References:                 pq.StringArray(references),
		Checksum:                   comp.Checksum,
	}

	if res := s.db.WithContext(ctx).Create(&configConn); res.Error != nil {
		return fmt.Errorf("unable to create docker build config: %w", res.Error)
	}

	// Update component type
	if err := s.componentsHelpers.UpdateComponentType(ctx, compID, app.ComponentTypeDockerBuild); err != nil {
		return fmt.Errorf("unable to update component type: %w", err)
	}

	// Send signals to trigger build
	s.evClient.Send(ctx, compID, &componentsignals.Signal{
		Type: componentsignals.OperationConfigCreated,
	})
	s.evClient.Send(ctx, compID, &componentsignals.Signal{
		Type:          componentsignals.OperationUpdateComponentType,
		ComponentType: app.ComponentTypeDockerBuild,
	})

	return nil
}

// createExternalImageConfig creates an external image component config
func (s *service) createExternalImageConfig(ctx context.Context, compID, appConfigID string, comp *config.Component) error {
	extImg := comp.ExternalImage
	if extImg == nil {
		return fmt.Errorf("external_image config is nil for component %s", comp.Name)
	}

	// Convert references to strings
	references := make([]string, 0, len(comp.References))
	for _, ref := range comp.References {
		references = append(references, ref.String())
	}

	cfg := app.ExternalImageComponentConfig{}

	// Handle AWS ECR config
	if extImg.AWSECRImageConfig != nil {
		cfg.ImageURL = extImg.AWSECRImageConfig.ImageURL
		cfg.Tag = extImg.AWSECRImageConfig.Tag
		cfg.AWSECRImageConfig = &app.AWSECRImageConfig{
			IAMRoleARN: extImg.AWSECRImageConfig.IAMRoleARN,
			AWSRegion:  extImg.AWSECRImageConfig.AWSRegion,
		}
	}

	// Handle public image config
	if extImg.PublicImageConfig != nil {
		cfg.ImageURL = extImg.PublicImageConfig.ImageURL
		cfg.Tag = extImg.PublicImageConfig.Tag
	}

	// Create config connection
	configConn := app.ComponentConfigConnection{
		ExternalImageComponentConfig: &cfg,
		ComponentID:                  compID,
		AppConfigID:                  appConfigID,
		References:                   pq.StringArray(references),
		Checksum:                     comp.Checksum,
	}

	if res := s.db.WithContext(ctx).Create(&configConn); res.Error != nil {
		return fmt.Errorf("unable to create external image config: %w", res.Error)
	}

	// Update component type
	if err := s.componentsHelpers.UpdateComponentType(ctx, compID, app.ComponentTypeExternalImage); err != nil {
		return fmt.Errorf("unable to update component type: %w", err)
	}

	// Send signals to trigger build
	s.evClient.Send(ctx, compID, &componentsignals.Signal{
		Type: componentsignals.OperationConfigCreated,
	})
	s.evClient.Send(ctx, compID, &componentsignals.Signal{
		Type:          componentsignals.OperationUpdateComponentType,
		ComponentType: app.ComponentTypeExternalImage,
	})

	return nil
}

// createKubernetesManifestConfig creates a kubernetes manifest component config
func (s *service) createKubernetesManifestConfig(ctx context.Context, compID, appConfigID string, comp *config.Component) error {
	k8s := comp.KubernetesManifest
	if k8s == nil {
		return fmt.Errorf("kubernetes_manifest config is nil for component %s", comp.Name)
	}

	// Convert references to strings
	references := make([]string, 0, len(comp.References))
	for _, ref := range comp.References {
		references = append(references, ref.String())
	}

	cfg := app.KubernetesManifestComponentConfig{
		Manifest:  k8s.Manifest,
		Namespace: k8s.Namespace,
	}

	// Handle kustomize config
	if k8s.Kustomize != nil {
		cfg.Kustomize = &app.KustomizeConfig{
			Path:           k8s.Kustomize.Path,
			Patches:        k8s.Kustomize.Patches,
			EnableHelm:     k8s.Kustomize.EnableHelm,
			LoadRestrictor: k8s.Kustomize.LoadRestrictor,
		}
	}

	// Handle VCS configs
	if k8s.PublicRepo != nil {
		cfg.PublicGitVCSConfig = &app.PublicGitVCSConfig{
			Repo:      k8s.PublicRepo.Repo,
			Branch:    k8s.PublicRepo.Branch,
			Directory: k8s.PublicRepo.Directory,
		}
	}
	if k8s.ConnectedRepo != nil {
		cfg.ConnectedGithubVCSConfig = &app.ConnectedGithubVCSConfig{
			Repo:      k8s.ConnectedRepo.Repo,
			Branch:    k8s.ConnectedRepo.Branch,
			Directory: k8s.ConnectedRepo.Directory,
		}
	}

	// Create config connection
	configConn := app.ComponentConfigConnection{
		KubernetesManifestComponentConfig: &cfg,
		ComponentID:                       compID,
		AppConfigID:                       appConfigID,
		References:                        pq.StringArray(references),
		Checksum:                          comp.Checksum,
	}

	if k8s.DriftSchedule != nil {
		configConn.DriftSchedule = *k8s.DriftSchedule
	}

	if res := s.db.WithContext(ctx).Create(&configConn); res.Error != nil {
		return fmt.Errorf("unable to create kubernetes manifest config: %w", res.Error)
	}

	// Update component type
	if err := s.componentsHelpers.UpdateComponentType(ctx, compID, app.ComponentTypeKubernetesManifest); err != nil {
		return fmt.Errorf("unable to update component type: %w", err)
	}

	// Send signals to trigger build
	s.evClient.Send(ctx, compID, &componentsignals.Signal{
		Type: componentsignals.OperationConfigCreated,
	})
	s.evClient.Send(ctx, compID, &componentsignals.Signal{
		Type:          componentsignals.OperationUpdateComponentType,
		ComponentType: app.ComponentTypeKubernetesManifest,
	})

	return nil
}

// createJobConfig creates a job component config
func (s *service) createJobConfig(ctx context.Context, compID, appConfigID string, comp *config.Component) error {
	job := comp.Job
	if job == nil {
		return fmt.Errorf("job config is nil for component %s", comp.Name)
	}

	// Convert env vars from map
	envVars := make(map[string]*string)
	for k, v := range job.EnvVarMap {
		val := v
		envVars[k] = &val
	}

	// Convert references to strings
	references := make([]string, 0, len(comp.References))
	for _, ref := range comp.References {
		references = append(references, ref.String())
	}

	cfg := app.JobComponentConfig{
		ImageURL: job.ImageURL,
		Tag:      job.Tag,
		Cmd:      job.Cmd,
		EnvVars:  pgtype.Hstore(envVars),
		Args:     job.Args,
	}

	// Create config connection
	configConn := app.ComponentConfigConnection{
		JobComponentConfig: &cfg,
		ComponentID:        compID,
		AppConfigID:        appConfigID,
		References:         pq.StringArray(references),
		Checksum:           comp.Checksum,
	}

	if res := s.db.WithContext(ctx).Create(&configConn); res.Error != nil {
		return fmt.Errorf("unable to create job config: %w", res.Error)
	}

	// Update component type
	if err := s.componentsHelpers.UpdateComponentType(ctx, compID, app.ComponentTypeJob); err != nil {
		return fmt.Errorf("unable to update component type: %w", err)
	}

	// Send signals to trigger build
	s.evClient.Send(ctx, compID, &componentsignals.Signal{
		Type: componentsignals.OperationConfigCreated,
	})
	s.evClient.Send(ctx, compID, &componentsignals.Signal{
		Type:          componentsignals.OperationUpdateComponentType,
		ComponentType: app.ComponentTypeJob,
	})

	return nil
}

// syncPermissionsConfig creates a permissions config with IAM roles from the template
func (s *service) syncPermissionsConfig(ctx context.Context, orgID string, newApp *app.App, appConfig *app.AppConfig, permissions *config.PermissionsConfig) error {
	// Build the roles array from permission config
	roles := make([]app.AppAWSIAMRoleConfig, 0)

	// Add provision role
	if permissions.ProvisionRole != nil {
		roles = append(roles, s.convertIAMRole(appConfig.ID, app.AWSIAMRoleTypeRunnerProvision, permissions.ProvisionRole))
	}

	// Add maintenance role
	if permissions.MaintenanceRole != nil {
		roles = append(roles, s.convertIAMRole(appConfig.ID, app.AWSIAMRoleTypeRunnerMaintenance, permissions.MaintenanceRole))
	}

	// Add deprovision role
	if permissions.DeprovisionRole != nil {
		roles = append(roles, s.convertIAMRole(appConfig.ID, app.AWSIAMRoleTypeRunnerDeprovision, permissions.DeprovisionRole))
	}

	permissionsConfig := app.AppPermissionsConfig{
		OrgID:       orgID,
		AppID:       newApp.ID,
		AppConfigID: appConfig.ID,
		Roles:       roles,
	}

	res := s.db.WithContext(ctx).Create(&permissionsConfig)
	if res.Error != nil {
		return fmt.Errorf("unable to create permissions config: %w", res.Error)
	}

	return nil
}

// convertIAMRole converts a config IAM role to an app IAM role config
func (s *service) convertIAMRole(appConfigID string, roleType app.AWSIAMRoleType, role *config.AppAWSIAMRole) app.AppAWSIAMRoleConfig {
	policies := make([]app.AppAWSIAMPolicyConfig, 0, len(role.Policies))
	for _, policy := range role.Policies {
		policies = append(policies, app.AppAWSIAMPolicyConfig{
			AppConfigID:       appConfigID,
			ManagedPolicyName: policy.ManagedPolicyName,
			Name:              policy.Name,
			Contents:          dbgenerics.ToJSON(policy.Contents),
		})
	}

	return app.AppAWSIAMRoleConfig{
		AppConfigID:             appConfigID,
		Type:                    roleType,
		Name:                    role.Name,
		Description:             role.Description,
		DisplayName:             role.DisplayName,
		PermissionsBoundaryJSON: dbgenerics.ToJSON(role.PermissionsBoundary),
		Policies:                policies,
	}
}

// syncPoliciesConfig creates a policies config from the template
func (s *service) syncPoliciesConfig(ctx context.Context, orgID string, newApp *app.App, appConfig *app.AppConfig, policies *config.PoliciesConfig) error {
	// Build the policy configs
	policyConfigs := make([]app.AppPolicyConfig, 0, len(policies.Policies))
	for _, policy := range policies.Policies {
		policyConfigs = append(policyConfigs, app.AppPolicyConfig{
			OrgID:       orgID,
			AppID:       newApp.ID,
			AppConfigID: appConfig.ID,
			Type:        policy.Type,
			Engine:      policy.Engine,
			Contents:    policy.Contents,
			Components:  policy.Components,
		})
	}

	policiesConfig := app.AppPoliciesConfig{
		OrgID:       orgID,
		AppID:       newApp.ID,
		AppConfigID: appConfig.ID,
		Policies:    policyConfigs,
	}

	res := s.db.WithContext(ctx).Create(&policiesConfig)
	if res.Error != nil {
		return fmt.Errorf("unable to create policies config: %w", res.Error)
	}

	return nil
}

// syncBreakGlassConfig creates a break-glass config from the template
func (s *service) syncBreakGlassConfig(ctx context.Context, orgID string, newApp *app.App, appConfig *app.AppConfig, breakGlass *config.BreakGlass) error {
	// Build the roles array for break-glass
	roles := make([]app.AppAWSIAMRoleConfig, 0, len(breakGlass.Roles))
	for _, role := range breakGlass.Roles {
		roles = append(roles, s.convertIAMRole(appConfig.ID, app.AWSIAMRoleTypeBreakGlass, role))
	}

	breakGlassConfig := app.AppBreakGlassConfig{
		OrgID:       orgID,
		AppID:       newApp.ID,
		AppConfigID: appConfig.ID,
		Roles:       roles,
	}

	res := s.db.WithContext(ctx).Create(&breakGlassConfig)
	if res.Error != nil {
		return fmt.Errorf("unable to create break glass config: %w", res.Error)
	}

	return nil
}

// syncActionsConfig creates action workflows and their configs from the template
func (s *service) syncActionsConfig(ctx context.Context, orgID string, newApp *app.App, appConfig *app.AppConfig, actions []*config.ActionConfig) error {
	for _, action := range actions {
		// Phase A: Create ActionWorkflow
		actionWorkflow := app.ActionWorkflow{
			OrgID:  orgID,
			AppID:  newApp.ID,
			Name:   action.Name,
			Status: app.ActionWorkflowStatusActive,
		}

		res := s.db.WithContext(ctx).Create(&actionWorkflow)
		if res.Error != nil {
			return fmt.Errorf("unable to create action workflow %s: %w", action.Name, res.Error)
		}

		// Phase B: Create ActionWorkflowConfig
		timeout := 5 * time.Minute // default timeout
		if action.Timeout != "" {
			parsedTimeout, err := time.ParseDuration(action.Timeout)
			if err == nil {
				timeout = parsedTimeout
			}
		}

		// Resolve component dependencies to IDs
		depIDs, err := s.componentsHelpers.GetComponentIDs(ctx, newApp.ID, action.Dependencies)
		if err != nil {
			return fmt.Errorf("unable to resolve component dependencies for action %s: %w", action.Name, err)
		}

		// Convert references to strings
		references := make([]string, 0, len(action.References))
		for _, ref := range action.References {
			references = append(references, ref.String())
		}

		actionWorkflowConfig := app.ActionWorkflowConfig{
			OrgID:                  orgID,
			AppID:                  newApp.ID,
			AppConfigID:            appConfig.ID,
			ActionWorkflowID:       actionWorkflow.ID,
			Timeout:                timeout,
			ComponentDependencyIDs: pq.StringArray(depIDs),
			References:             pq.StringArray(references),
			BreakGlassRoleARN:      generics.NewNullString(action.BreakGlassRole),
		}

		res = s.db.WithContext(ctx).Create(&actionWorkflowConfig)
		if res.Error != nil {
			return fmt.Errorf("unable to create action workflow config for %s: %w", action.Name, res.Error)
		}

		// Phase C: Create Triggers
		if err := s.createActionWorkflowTriggers(ctx, orgID, newApp.ID, appConfig.ID, actionWorkflowConfig.ID, action.Triggers); err != nil {
			return fmt.Errorf("unable to create triggers for action %s: %w", action.Name, err)
		}

		// Phase D: Create Steps
		if err := s.createActionWorkflowSteps(ctx, orgID, newApp.ID, appConfig.ID, actionWorkflowConfig.ID, action.Steps); err != nil {
			return fmt.Errorf("unable to create steps for action %s: %w", action.Name, err)
		}

		// Send signal for action creation
		s.evClient.Send(ctx, actionWorkflow.ID, &actionsignals.Signal{
			Type: actionsignals.OperationCreated,
		})
		s.evClient.Send(ctx, actionWorkflow.ID, &actionsignals.Signal{
			Type:     actionsignals.OperationConfigCreated,
			ConfigID: actionWorkflowConfig.ID,
		})
	}

	// Phase E: Propagate to Installs
	if err := s.actionsHelpers.EnsureInstallAction(ctx, newApp.ID, nil); err != nil {
		return fmt.Errorf("unable to ensure install action workflows: %w", err)
	}

	return nil
}

// createActionWorkflowTriggers creates triggers for an action workflow config
func (s *service) createActionWorkflowTriggers(ctx context.Context, orgID, appID, appConfigID, awcID string, triggers []*config.ActionTriggerConfig) error {
	for _, trigger := range triggers {
		var componentID string
		if trigger.ComponentName != "" {
			comp, err := s.componentsHelpers.GetComponentByName(ctx, appID, trigger.ComponentName)
			if err != nil {
				return fmt.Errorf("unable to find component %s: %w", trigger.ComponentName, err)
			}
			componentID = comp.ID
		}

		newTrigger := app.ActionWorkflowTriggerConfig{
			OrgID:                  orgID,
			AppID:                  appID,
			AppConfigID:            appConfigID,
			ActionWorkflowConfigID: awcID,
			Type:                   app.ActionWorkflowTriggerType(trigger.Type),
			CronSchedule:           trigger.CronSchedule,
			Index:                  int(trigger.Index),
			ComponentID:            generics.NewNullString(componentID),
		}

		res := s.db.WithContext(ctx).Create(&newTrigger)
		if res.Error != nil {
			return fmt.Errorf("unable to create action workflow trigger: %w", res.Error)
		}
	}

	return nil
}

// createActionWorkflowSteps creates steps for an action workflow config
func (s *service) createActionWorkflowSteps(ctx context.Context, orgID, appID, appConfigID, awcID string, steps []*config.ActionStepConfig) error {
	stepIdx := 0
	prevStepID := ""

	for _, step := range steps {
		// Convert env vars
		envVars := make(map[string]*string)
		for k, v := range step.EnvVarMap {
			val := v
			envVars[k] = &val
		}

		// Convert references to strings
		references := make([]string, 0, len(step.References))
		for _, ref := range step.References {
			references = append(references, ref.String())
		}

		newStep := app.ActionWorkflowStepConfig{
			OrgID:                  orgID,
			AppID:                  appID,
			AppConfigID:            appConfigID,
			ActionWorkflowConfigID: awcID,
			Name:                   step.Name,
			EnvVars:                pgtype.Hstore(envVars),
			Command:                step.Command,
			InlineContents:         step.InlineContents,
			Idx:                    stepIdx,
			PreviousStepID:         prevStepID,
			References:             pq.StringArray(references),
		}

		// Handle VCS configs
		if step.PublicRepo != nil {
			newStep.PublicGitVCSConfig = &app.PublicGitVCSConfig{
				Repo:      step.PublicRepo.Repo,
				Branch:    step.PublicRepo.Branch,
				Directory: step.PublicRepo.Directory,
			}
		}
		if step.ConnectedRepo != nil {
			newStep.ConnectedGithubVCSConfig = &app.ConnectedGithubVCSConfig{
				Repo:      step.ConnectedRepo.Repo,
				Branch:    step.ConnectedRepo.Branch,
				Directory: step.ConnectedRepo.Directory,
			}
		}

		res := s.db.WithContext(ctx).Create(&newStep)
		if res.Error != nil {
			return fmt.Errorf("unable to create action workflow step: %w", res.Error)
		}

		prevStepID = newStep.ID
		stepIdx++
	}

	return nil
}
