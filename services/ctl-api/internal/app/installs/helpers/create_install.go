package helpers

import (
	"context"
	"fmt"
	"regexp"

	"gorm.io/gorm"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

type InstallMetadata struct {
	ManagedBy string `json:"managed_by,omitempty"`
}

var awsAccountIDPattern = regexp.MustCompile(`^[0-9]{12}$`)

type CreateInstallAWSAccountParams struct {
	Region       string `json:"region"`
	ConnectionID string `json:"connection_id,omitempty"`

	// AccountID is the AWS account this install targets. Required when the org has
	// the phone-home-auth feature enabled, optional otherwise. Immutable after
	// creation — there is deliberately no equivalent field on UpdateInstallRequest.
	AccountID string `json:"account_id,omitempty"`
}

type CreateInstallAzureAccountParams struct {
	Location string `json:"location"`

	// SubscriptionID is the Azure subscription this install targets. Required when
	// the org has the phone-home-auth feature enabled. Immutable after creation.
	SubscriptionID string `json:"subscription_id,omitempty"`
}

type CreateInstallGCPAccountParams struct {
	// ProjectID is the GCP project this install targets. Required when the org has
	// the phone-home-auth feature enabled. Immutable after creation.
	ProjectID string `json:"project_id"`
	Region    string `json:"region"`
}

type CreateInstallParams struct {
	Name string `json:"name" validate:"required"`

	AWSAccount *CreateInstallAWSAccountParams `json:"aws_account"`

	AzureAccount *CreateInstallAzureAccountParams `json:"azure_account"`

	GCPAccount *CreateInstallGCPAccountParams `json:"gcp_account"`

	Inputs map[string]*string `json:"inputs"`

	InstallConfig *CreateInstallConfigParams `json:"install_config"`

	Metadata InstallMetadata `json:"metadata,omitempty"`

	// Labels are key/value pairs to attach to the install at creation time.
	// They are merged into the install's existing labels (which is empty for a brand-new install).
	Labels map[string]string `json:"labels,omitempty"`

	SandboxMode bool `json:"sandbox_mode,omitempty" swaggerignore:"true"`
}

func (s *Helpers) CreateInstall(ctx context.Context, appID string, req *CreateInstallParams) (*app.Install, error) {
	parentApp := app.App{}
	res := s.db.WithContext(ctx).
		Preload("Components").
		Preload("AppSandboxConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_sandbox_configs.created_at DESC").Limit(1)
		}).
		Preload("AppRunnerConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_runner_configs.created_at DESC").Limit(1)
		}).
		Preload("AppInputConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_input_configs.created_at DESC").Limit(1)
		}).
		Preload("AppConfigs", func(db *gorm.DB) *gorm.DB {
			return db.
				Where(views.TableOrViewName(s.db, &app.AppConfig{}, ".status_v2 ->> 'status' = ?"), string(app.AppConfigStatusActive)).
				Order(views.TableOrViewName(s.db, &app.AppConfig{}, ".created_at DESC")).
				Limit(1)
		}).
		Preload("AppPermissionsConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_permissions_configs.created_at DESC").Limit(1)
		}).
		Preload("AppPermissionsConfigs.Roles").
		First(&parentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	if len(parentApp.AppConfigs) == 0 {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("no active app config found for app %s", appID),
			Description: "No active app config found. Please sync your app configuration before creating an install.",
		}
	}

	// Validate and pin against the input config belonging to the app config this
	// install is pinned to. Using the app's newest input config instead lets the
	// two diverge whenever a newer app config exists, and the config-migration
	// lookup then misses the install's inputs.
	pinnedAppInputConfig, err := s.GetPinnedAppInputConfig(ctx, appID, parentApp.AppConfigs[0].ID)
	if err != nil {
		return nil, fmt.Errorf("unable to get pinned app input config: %w", err)
	}

	if err := s.ValidateInstallInputs(ctx, pinnedAppInputConfig, req.Inputs); err != nil {
		return nil, err
	}
	install := app.Install{
		AppID:              appID,
		Name:               req.Name,
		SandboxMode:        pkggenerics.NewNullBool(req.SandboxMode),
		AppSandboxConfigID: parentApp.AppSandboxConfigs[0].ID,
		AppRunnerConfigID:  parentApp.AppRunnerConfigs[0].ID,
		AppConfigID:        parentApp.AppConfigs[0].ID,
		InstallSandbox: app.InstallSandbox{
			Status: app.InstallSandboxStatusQueued,
			TerraformWorkspace: app.TerraformWorkspace{
				ID: domains.NewTerraformWorkspaceID(),
			},
		},
		Metadata: generics.ToHstore(map[string]string{
			"managed_by": req.Metadata.ManagedBy,
		}),
	}

	if len(req.Labels) > 0 {
		install.Labels = labels.Labels(req.Labels)
	}

	// When enabled, every install must declare which cloud account it targets, so a
	// later phone home can be checked against an identifier the vendor asserted up
	// front rather than one the install reported about itself.
	requireTargetAccount, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeaturePhoneHomeAuth)
	if err != nil {
		return nil, fmt.Errorf("check phone home auth feature: %w", err)
	}

	targetSource := ""

	runnerType := parentApp.AppRunnerConfigs[0].Type
	switch runnerType {
	case app.AppRunnerTypeGCP, app.AppRunnerTypeGCPGKE:
		if req.GCPAccount == nil {
			req.GCPAccount = &CreateInstallGCPAccountParams{}
		}
		req.AWSAccount = nil
		req.AzureAccount = nil

		if requireTargetAccount && req.GCPAccount.ProjectID == "" {
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("gcp_account.project_id is required for GCP installs"),
				Description: "gcp_account.project_id is required because phone home authentication is enabled for this organization",
			}
		}
		if req.GCPAccount.ProjectID != "" {
			targetSource = app.CloudPlatformTargetSourceUser
		}
	case app.AppRunnerTypeAzure, app.AppRunnerTypeAzureAKS, app.AppRunnerTypeAzureACS:
		if req.AzureAccount == nil {
			req.AzureAccount = &CreateInstallAzureAccountParams{}
		}
		req.AWSAccount = nil
		req.GCPAccount = nil

		if requireTargetAccount && req.AzureAccount.SubscriptionID == "" {
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("azure_account.subscription_id is required for Azure installs"),
				Description: "azure_account.subscription_id is required because phone home authentication is enabled for this organization",
			}
		}
		if req.AzureAccount.SubscriptionID != "" {
			targetSource = app.CloudPlatformTargetSourceUser
		}
	case app.AppRunnerTypeAWS, app.AppRunnerTypeAWSEKS, app.AppRunnerTypeAWSECS:
		if req.AWSAccount == nil {
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("aws_account.region is required for AWS installs"),
				Description: "aws_account.region is required for AWS installs",
			}
		}
		if req.AWSAccount.AccountID != "" {
			targetSource = app.CloudPlatformTargetSourceUser
		}
		if req.AWSAccount.ConnectionID != "" {
			if runnerType != app.AppRunnerTypeAWS {
				return nil, stderr.ErrUser{
					Err:         fmt.Errorf("AWS account connections are not supported for runner type %q", runnerType),
					Description: "AWS account connections are only supported for AWS runner installs",
				}
			}
			connection, err := s.validateAWSAccountConnection(ctx, req.AWSAccount.ConnectionID)
			if err != nil {
				return nil, err
			}

			// The connection already names an account, so it is authoritative. An
			// explicit account ID may agree with it but never override it.
			if req.AWSAccount.AccountID != "" && req.AWSAccount.AccountID != connection.AccountID {
				return nil, stderr.ErrUser{
					Err: fmt.Errorf("aws_account.account_id %q conflicts with connection %s account %q",
						req.AWSAccount.AccountID, connection.ID, connection.AccountID),
					Description: "aws_account.account_id does not match the account of the selected AWS account connection",
				}
			}
			req.AWSAccount.AccountID = connection.AccountID
			targetSource = app.CloudPlatformTargetSourceConnection
		}
		if requireTargetAccount && req.AWSAccount.AccountID == "" {
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("aws_account.account_id is required for AWS installs"),
				Description: "aws_account.account_id is required because phone home authentication is enabled for this organization",
			}
		}
		// Only format-checked when the feature is on: the field is advisory otherwise,
		// and rejecting it would change behaviour for organizations that never opted in.
		if requireTargetAccount && !awsAccountIDPattern.MatchString(req.AWSAccount.AccountID) {
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("aws_account.account_id %q is not a 12-digit AWS account ID", req.AWSAccount.AccountID),
				Description: "aws_account.account_id must be exactly 12 digits",
			}
		}
	default:
		if req.AWSAccount == nil && req.AzureAccount == nil && req.GCPAccount == nil {
			return nil, stderr.ErrUser{
				Err:         fmt.Errorf("unable to determine cloud platform from runner type %q", runnerType),
				Description: "unable to determine cloud platform from app runner config",
			}
		}
	}

	if req.AWSAccount != nil {
		install.AWSAccount = &app.AWSAccount{
			Region: req.AWSAccount.Region,
		}
		if req.AWSAccount.ConnectionID != "" {
			install.AWSAccount.AWSAccountConnectionID = &req.AWSAccount.ConnectionID
		}
		install.CloudPlatformMetadata.TargetAccountID = req.AWSAccount.AccountID
	}
	if req.AzureAccount != nil {
		install.AzureAccount = &app.AzureAccount{
			Location:       req.AzureAccount.Location,
			SubscriptionID: req.AzureAccount.SubscriptionID,
		}
		install.CloudPlatformMetadata.TargetSubscriptionID = req.AzureAccount.SubscriptionID
	}
	if req.GCPAccount != nil {
		install.GCPAccount = &app.GCPAccount{
			ProjectID: req.GCPAccount.ProjectID,
			Region:    req.GCPAccount.Region,
		}
		install.CloudPlatformMetadata.TargetProjectID = req.GCPAccount.ProjectID
	}

	// A target is never recorded without its provenance, including on the fallback
	// runner-type branch that does not set targetSource itself.
	if install.CloudPlatformMetadata.HasTarget() {
		if targetSource == "" {
			targetSource = app.CloudPlatformTargetSourceUser
		}
		install.CloudPlatformMetadata.TargetSource = targetSource
	}
	if parentApp.AppPermissionsConfig.ID != "" && len(parentApp.AppPermissionsConfig.Roles) > 0 {
		installRoles := make([]app.InstallRoles, 0)

		for _, role := range parentApp.AppPermissionsConfig.Roles {
			installRoles = append(installRoles, app.InstallRoles{
				AppRoleConfigID: role.ID,
			})
		}

		install.InstallRoles = installRoles
	}

	if pinnedAppInputConfig != nil && pinnedAppInputConfig.ID != "" {
		install.InstallInputs = []app.InstallInputs{
			{
				Values:           req.Inputs,
				AppInputConfigID: pinnedAppInputConfig.ID,
			},
		}
	}

	switch parentApp.AppRunnerConfigs[0].Type {
	case "aws":
		install.InstallStack = &app.InstallStack{
			InstallStackOutputs: app.InstallStackOutputs{
				Data: generics.ToHstore(map[string]string{}),
			},
		}
	case "azure":
		install.InstallStack = &app.InstallStack{
			InstallStackOutputs: app.InstallStackOutputs{
				Data: generics.ToHstore(map[string]string{}),
			},
		}
	case "gcp":
		install.InstallStack = &app.InstallStack{
			InstallStackOutputs: app.InstallStackOutputs{
				Data: generics.ToHstore(map[string]string{}),
			},
		}
	}

	res = s.db.WithContext(ctx).Create(&install)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install: %w", res.Error)
	}

	s.mw.Incr("install.created", metrics.ToTags(map[string]string{
		"org_id":     install.OrgID,
		"app_id":     appID,
		"install_id": install.ID,
	}))

	// Create all install queues (workflows, signals, actions, drift, etc.)
	if err := s.EnsureInstallQueues(ctx, install.ID); err != nil {
		return nil, fmt.Errorf("unable to create install queues: %w", err)
	}

	if req.InstallConfig != nil {
		_, err := s.CreateInstallConfig(ctx, install.ID, req.InstallConfig)
		if err != nil {
			return nil, fmt.Errorf("unable to create install config: %w", err)
		}
	}

	// Install components, actions, and runbooks are created asynchronously
	// by the install-created signal's reconcile activities.

	loadedInstall, err := s.getInstall(ctx, install.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to load all install resources: %w", err)
	}

	if _, err := s.runnersHelpers.CreateInstallRunnerGroup(ctx, loadedInstall); err != nil {
		return nil, fmt.Errorf("unable to create install runner: %w", err)
	}

	return &install, nil
}
