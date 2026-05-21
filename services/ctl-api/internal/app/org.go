package app

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop/bulk"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/links"
)

type OrgType string

const (
	OrgTypeSandbox     OrgType = "sandbox"
	OrgTypeIntegration OrgType = "integration"
	OrgTypeDefault     OrgType = "default"

	// Legacy
	OrgTypeLegacy OrgType = "real"

	OrgTypeUnknown OrgType = ""
)

type OrgStatus string

const (
	OrgStatusError          OrgStatus = "error"
	OrgStatusActive         OrgStatus = "active"
	OrgStatusProvisioning   OrgStatus = "provisioning"
	OrgStatusDeleting       OrgStatus = "deleting"
	OrgStatusDeprovisioning OrgStatus = "deprovisioning"
	OrgStatusDeprovisioned  OrgStatus = "deprovisioned"
)

// org feature flags
type OrgFeature string

const (
	OrgFeatureOrgDashboard        OrgFeature = "org-dashboard"
	OrgFeatureOrgRunner           OrgFeature = "org-runner"
	OrgFeatureOrgSettings         OrgFeature = "org-settings"
	OrgFeatureAppBranches         OrgFeature = "app-branches"
	OrgFeatureUserManagedFeatures OrgFeature = "user-managed-features"
	OrgFeatureQueues              OrgFeature = "queues"
	OrgFeatureSupportRole         OrgFeature = "support-role"
	OrgFeatureParallelRunnerJobs  OrgFeature = "parallel-runner-jobs"
	OrgFeatureInstallRename       OrgFeature = "install-rename"
	OrgFeatureDeployOutputs       OrgFeature = "deploy-outputs"
	// OrgFeatureTerraformProviderMirror enables build-time vendoring of
	// terraform providers via `terraform providers mirror` and ships the
	// resulting filesystem mirror inside the OCI artifact. The install
	// runner auto-detects the mirror at unpack time, so toggling this
	// flag only affects the build runner.
	OrgFeatureTerraformProviderMirror OrgFeature = "terraform-provider-mirror"
	OrgFeatureAppBranchesUI           OrgFeature = "app-branches-ui"
	OrgFeatureTraceView               OrgFeature = "trace-view"
	OrgFeatureStateGenV2              OrgFeature = "state-gen-v2"
	OrgFeatureAutoSkipNoop            OrgFeature = "auto-skip-noop"
	OrgFeatureSlack                   OrgFeature = "slack"
)

type Org struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `gorm:"index:idx_org_name,unique" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	Name              string          `gorm:"index:idx_org_name,unique;notnull" json:"name,omitzero" temporaljson:"name,omitzero,omitempty"`
	Status            OrgStatus       `json:"status,omitzero" gorm:"notnull" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string          `json:"status_description,omitzero" gorm:"notnull" temporaljson:"status_description,omitzero,omitempty"`
	StatusV2          CompositeStatus `json:"status_v2,omitzero" gorm:"type:jsonb" temporaljson:"status_v2,omitzero,omitempty"`

	SandboxMode bool `json:"sandbox_mode,omitzero" gorm:"notnull" temporaljson:"sandbox_mode,omitzero,omitempty"`

	OrgType   OrgType `json:"-" temporaljson:"org_type,omitzero,omitempty"`
	DebugMode bool    `json:"-" temporaljson:"debug_mode,omitzero,omitempty"`

	NotificationsConfig   NotificationsConfig `gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" json:"notifications_config,omitzero,omitempty" temporaljson:"notifications_config,omitzero,omitempty"`
	NotificationsConfigID string              `json:"-" temporaljson:"notifications_config_id,omitzero,omitempty"`

	RunnerGroup RunnerGroup `json:"runner_group,omitzero" gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" temporaljson:"runner_group,omitzero,omitempty"`

	LogoURL string `json:"logo_url,omitzero" temporaljson:"logo_url,omitzero,omitempty"`

	Priority int `json:"-" temporaljson:"priority,omitzero,omitempty"`

	Apps           []App               `faker:"-" swaggerignore:"true" json:"apps,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"apps,omitzero,omitempty"`
	VCSConnections []VCSConnection     `json:"vcs_connections,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"vcs_connections,omitzero,omitempty"`
	Invites        []OrgInvite         `faker:"-" swaggerignore:"true" json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"invites,omitzero,omitempty"`
	Features       types.StringBoolMap `json:"features,omitzero" gorm:"type:jsonb;default null" temporaljson:"features,omitzero,omitempty"`
	Tags           pq.StringArray      `json:"tags,omitzero" gorm:"type:text[];default '{}'" swaggertype:"array,string" temporaljson:"tags,omitzero,omitempty"`
	labels.Labeled

	// Other relationships as part of the data model

	Runners                   []Runner                   `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"runners,omitzero,omitempty"`
	PublicGitVCSConfigs       []PublicGitVCSConfig       `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"public_git_vcs_configs,omitzero,omitempty"`
	ConnectedGithubVCSConfigs []ConnectedGithubVCSConfig `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"connected_github_vcs_configs,omitzero,omitempty"`
	VCSConnectionCommits      []VCSConnectionCommit      `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"vcs_connection_commits,omitzero,omitempty"`
	AWSECRImageConfigs        []AWSECRImageConfig        `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"awsecr_image_configs,omitzero,omitempty"`
	GCPGARImageConfigs        []GCPGARImageConfig        `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"gcp_gar_image_configs,omitzero,omitempty"`
	AzureACRImageConfigs      []AzureACRImageConfig      `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"azure_acr_image_configs,omitzero,omitempty"`
	Installs                  []Install                  `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"installs,omitzero,omitempty"`
	Components                []Component                `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"components,omitzero,omitempty"`

	Installers        []Installer         `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"installers,omitzero,omitempty"`
	InstallerMetadata []InstallerMetadata `gorm:"constraint:OnDelete:CASCADE;" json:"-" temporaljson:"installer_metadata,omitzero,omitempty"`

	Roles        []Role        `faker:"-" swaggerignore:"true" json:"roles,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"roles,omitzero,omitempty"`
	Policies     []Policy      `faker:"-" swaggerignore:"true" json:"policies,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"policies,omitzero,omitempty"`
	AccountRoles []AccountRole `faker:"-" swaggerignore:"true" json:"account_roles,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"account_roles,omitzero,omitempty"`

	// after query

	Links map[string]any `json:"links,omitempty" temporaljson:"-" gorm:"-"`

	// Transient fields for counts (not persisted to database)
	AppCount     int `json:"app_count,omitempty" gorm:"-"`
	InstallCount int `json:"install_count,omitempty" gorm:"-"`
}

func (o *Org) AfterQuery(tx *gorm.DB) error {
	o.Links = links.AppLinks(tx.Statement.Context, o.ID)

	if o.Features == nil {
		o.Features = make(map[string]bool, 0)
	}

	if o.Labels == nil {
		o.Labels = make(labels.Labels)
	}

	actieFeatures := GetFeatures()

	// if active feature not in features, add it
	for _, feature := range actieFeatures {
		if _, ok := o.Features[string(feature)]; !ok {
			o.Features[string(feature)] = false
		}
	}

	afLookup := make(map[string]bool)
	for _, feature := range GetFeatures() {
		afLookup[string(feature)] = true
	}

	// if feature key not in active features, remove it
	for key := range o.Features {
		if !afLookup[key] {
			delete(o.Features, key)
		}
	}

	return nil
}

func (o *Org) BeforeCreate(tx *gorm.DB) error {
	if o.Features == nil {
		o.Features = make(map[string]bool, 0)
	}

	// Set default feature flag values - most features enabled by default
	// except install-break-glass and user-managed-features which remain disabled
	defaultFeatures := map[OrgFeature]bool{
		// Disabled by default
		OrgFeatureInstallRename:           false,
		OrgFeatureDeployOutputs:           true,
		OrgFeatureSupportRole:             false,
		OrgFeatureTerraformProviderMirror: false,
		OrgFeatureAppBranchesUI:           false,
		OrgFeatureTraceView:               false,
		OrgFeatureStateGenV2:              false,
		OrgFeatureSlack:                   false,

		// Enabled by default
		OrgFeatureParallelRunnerJobs: true,
		OrgFeatureQueues:             true,
		OrgFeatureOrgDashboard:       true,
		OrgFeatureOrgRunner:          true,
		OrgFeatureOrgSettings:        true,
		OrgFeatureAppBranches:        true,
	}

	for _, feature := range GetFeatures() {
		if _, ok := o.Features[string(feature)]; !ok {
			o.Features[string(feature)] = defaultFeatures[feature]
		}
	}

	if o.ID == "" {
		o.ID = domains.NewOrgID()
	}

	o.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	return nil
}

func (o *Org) EventLoops() []bulk.EventLoop {
	evs := make([]bulk.EventLoop, 0)
	evs = append(evs, bulk.EventLoop{
		Namespace: "orgs",
		ID:        o.ID,
	})
	evs = append(evs, o.RunnerGroup.EventLoops()...)

	for _, app := range o.Apps {
		evs = append(evs, app.EventLoops()...)
	}

	return evs
}

// active feature flags for an orgs
func GetFeatures() []OrgFeature {
	return []OrgFeature{
		OrgFeatureOrgDashboard,
		OrgFeatureOrgRunner,
		OrgFeatureOrgSettings,
		OrgFeatureAppBranches,
		OrgFeatureQueues,
		OrgFeatureUserManagedFeatures,
		OrgFeatureParallelRunnerJobs,
		OrgFeatureSupportRole,
		OrgFeatureInstallRename,
		OrgFeatureDeployOutputs,
		OrgFeatureTerraformProviderMirror,
		OrgFeatureStateGenV2,
		OrgFeatureAppBranchesUI,
		OrgFeatureTraceView,
		OrgFeatureAutoSkipNoop,
		OrgFeatureSlack,
	}
}

// OrgFeatureInfo contains metadata about a feature flag
type OrgFeatureInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetFeatureDescriptions returns a map of feature names to their descriptions
func GetFeatureDescriptions() map[OrgFeature]string {
	return map[OrgFeature]string{
		OrgFeatureOrgDashboard:            "Access to the organization dashboard interface for managing org-wide settings and analytics",
		OrgFeatureOrgRunner:               "Enable organization-specific runner functionality for executing deployments",
		OrgFeatureOrgSettings:             "Access to organization settings management interface",
		OrgFeatureAppBranches:             "Support for multiple application branches allowing parallel development and testing",
		OrgFeatureUserManagedFeatures:     "Allow organization users to manage feature flags through the public API (admin-only flag)",
		OrgFeatureQueues:                  "Enable queue-based workflow execution for improved task scheduling and resource management",
		OrgFeatureSupportRole:             "Enable the support role option when inviting users to the organization",
		OrgFeatureParallelRunnerJobs:      "Enable parallel runner job execution via per-job-group queues (opt-in, requires runner reprovisioning)",
		OrgFeatureInstallRename:           "Allow renaming installs from the dashboard edit install modal",
		OrgFeatureDeployOutputs:           "Enable tabbed deploy detail page with plan, variables, state, and outputs tabs",
		OrgFeatureTerraformProviderMirror: "Vendor terraform providers at build time and ship them inside the OCI artifact so install runners can `terraform init` without reaching registry.terraform.io",
		OrgFeatureAppBranchesUI:           "Enable the app branches UI in the dashboard for managing and switching between app branches",
		OrgFeatureTraceView:               "Enable the trace view tab on action runs, deploys, and sandbox runs to visualize OTEL spans emitted by the runner",
		OrgFeatureStateGenV2:              "Use the new queue-based partial state regeneration system instead of the legacy full-regeneration workflow",
		OrgFeatureAutoSkipNoop:            "Automatically skip noop plans without requiring approval, overriding per-component skip_noops settings",
		OrgFeatureSlack:                   "Enable the Slack integration, including the Slack link in the dashboard sidebar and per-org Slack workspace/channel subscriptions",
	}
}

// GetFeaturesWithDescriptions returns all features with their descriptions
func GetFeaturesWithDescriptions() []OrgFeatureInfo {
	features := GetFeatures()
	descriptions := GetFeatureDescriptions()
	result := make([]OrgFeatureInfo, 0, len(features))

	for _, feature := range features {
		result = append(result, OrgFeatureInfo{
			Name:        string(feature),
			Description: descriptions[feature],
		})
	}

	return result
}

// GetUserManageableFeatures returns features that users are allowed to toggle
// (excludes the user-managed-features flag itself, which is admin-only)
func GetUserManageableFeatures() []OrgFeature {
	allFeatures := GetFeatures()
	manageable := make([]OrgFeature, 0, len(allFeatures)-1)

	for _, feature := range allFeatures {
		if feature != OrgFeatureUserManagedFeatures {
			manageable = append(manageable, feature)
		}
	}

	return manageable
}
