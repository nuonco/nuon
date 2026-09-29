package app

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/links"
)

type OrgType string

const (
	OrgTypeSandbox     OrgType = "sandbox"
	OrgTypeIntegration OrgType = "integration"
	OrgTypeDefault     OrgType = "default"

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

type OrgFeature string

const (
	OrgFeatureAppBranches              OrgFeature = "app-branches"
	OrgFeatureUserManagedFeatures      OrgFeature = "user-managed-features"
	OrgFeatureSupportRole              OrgFeature = "support-role"
	OrgFeatureInstallRename            OrgFeature = "install-rename"
	OrgFeatureTerraformProviderMirror  OrgFeature = "terraform-provider-mirror"
	OrgFeatureAppBranchesUI            OrgFeature = "app-branches-ui"
	OrgFeatureTraceView                OrgFeature = "trace-view"
	OrgFeatureAutoSkipNoop             OrgFeature = "auto-skip-noop"
	OrgFeatureSlack                    OrgFeature = "slack"
	OrgFeaturePulumiSandbox            OrgFeature = "pulumi-sandbox"
	OrgFeaturePulumiUpdatePlans        OrgFeature = "pulumi-update-plans"
	OrgFeatureNotebooks                OrgFeature = "notebooks"
	OrgFeatureVersionsUI               OrgFeature = "enable-versions-ui"
	OrgFeatureSpaceliftInstallStacks   OrgFeature = "spacelift-install-stacks"
	OrgFeatureAWSAccountConnections    OrgFeature = "aws-account-connections"
	OrgFeatureServiceAccountsAndTokens OrgFeature = "service-accounts-and-tokens"
	OrgFeaturePhoneHomeAuth            OrgFeature = "phone-home-auth"
	OrgFeatureRunbookStudio            OrgFeature = "runbook-studio"
	OrgFeatureCronNamespaceIsolation   OrgFeature = "cron-namespace-isolation"
	OrgFeatureNewAppIA                 OrgFeature = "new-app-ia"
	OrgFeatureOrgHealthcheckSweeps     OrgFeature = "org-healthcheck-sweeps"
	OrgFeatureAppInstallSyncing        OrgFeature = "app-install-syncing"
	OrgFeatureSandboxOCIArtifacts      OrgFeature = "sandbox-oci-artifacts"
	OrgFeatureDefaultAppBranches       OrgFeature = "default-app-branches"
	OrgFeatureNewInstallIA             OrgFeature = "new-install-ia"
	OrgFeatureDisableAppSync           OrgFeature = "disable-app-sync"
)

type OrgTelemetrySettings struct {
	Enabled bool `json:"enabled" gorm:"not null;default:false" temporaljson:"enabled,omitempty"`
}

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

	Telemetry OrgTelemetrySettings `json:"telemetry" gorm:"embedded;embeddedPrefix:telemetry_" temporaljson:"telemetry,omitempty"`

	OrgType   OrgType `json:"-" temporaljson:"org_type,omitzero,omitempty"`
	DebugMode bool    `json:"-" temporaljson:"debug_mode,omitzero,omitempty"`

	NotificationsConfig   NotificationsConfig `gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" json:"notifications_config,omitzero,omitempty" temporaljson:"notifications_config,omitzero,omitempty"`
	NotificationsConfigID string              `json:"-" temporaljson:"notifications_config_id,omitzero,omitempty"`

	RunnerGroup RunnerGroup `json:"runner_group,omitzero" gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" temporaljson:"runner_group,omitzero,omitempty"`

	LogoURL string `json:"logo_url,omitzero" temporaljson:"logo_url,omitzero,omitempty"`

	Priority int `json:"-" temporaljson:"priority,omitzero,omitempty"`

	Apps                  []App                  `faker:"-" swaggerignore:"true" json:"apps,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"apps,omitzero,omitempty"`
	VCSConnections        []VCSConnection        `json:"vcs_connections,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"vcs_connections,omitzero,omitempty"`
	AWSAccountConnections []AWSAccountConnection `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"aws_account_connections,omitzero,omitempty"`
	Invites               []OrgInvite            `faker:"-" swaggerignore:"true" json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"invites,omitzero,omitempty"`
	Features              types.StringBoolMap    `json:"features,omitzero" gorm:"type:jsonb;default null" temporaljson:"features,omitzero,omitempty"`
	Tags                  pq.StringArray         `json:"tags,omitzero" gorm:"type:text[];default '{}'" swaggertype:"array,string" temporaljson:"tags,omitzero,omitempty"`
	labels.Labeled

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
	forced := ForcedFeatures()

	for _, feature := range actieFeatures {
		if forced[string(feature)] {
			o.Features[string(feature)] = true
			continue
		}
		if _, ok := o.Features[string(feature)]; !ok {
			o.Features[string(feature)] = false
		}
	}

	afLookup := make(map[string]bool)
	for _, feature := range GetFeatures() {
		afLookup[string(feature)] = true
	}

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

	defaultFeatures := DefaultFeatures()
	forced := ForcedFeatures()

	for _, feature := range GetFeatures() {
		if forced[string(feature)] {
			o.Features[string(feature)] = true
			continue
		}
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

func DefaultFeatures() map[OrgFeature]bool {
	return map[OrgFeature]bool{
		OrgFeatureInstallRename:           false,
		OrgFeatureSupportRole:             false,
		OrgFeatureTerraformProviderMirror: false,
		OrgFeatureTraceView:               false,
		OrgFeatureSlack:                   false,
		OrgFeaturePulumiSandbox:           false,
		OrgFeaturePulumiUpdatePlans:       false,
		OrgFeatureNotebooks:               false,
		OrgFeatureSpaceliftInstallStacks:  false,
		OrgFeatureAWSAccountConnections:   false,
		OrgFeaturePhoneHomeAuth:           false,
		OrgFeatureRunbookStudio:           false,
		OrgFeatureCronNamespaceIsolation:  false,
		OrgFeatureNewAppIA:                false,
		OrgFeatureOrgHealthcheckSweeps:    false,
		OrgFeatureAppInstallSyncing:       false,
		OrgFeatureSandboxOCIArtifacts:     false,
		OrgFeatureDefaultAppBranches:      false,
		OrgFeatureNewInstallIA:            false,
		OrgFeatureDisableAppSync:          false,

		OrgFeatureAppBranches:   true,
		OrgFeatureAppBranchesUI: true,
	}
}

func GetFeatures() []OrgFeature {
	return []OrgFeature{
		OrgFeatureAppBranches,
		OrgFeatureUserManagedFeatures,
		OrgFeatureSupportRole,
		OrgFeatureInstallRename,
		OrgFeatureTerraformProviderMirror,
		OrgFeatureAppBranchesUI,
		OrgFeatureTraceView,
		OrgFeatureAutoSkipNoop,
		OrgFeatureSlack,
		OrgFeaturePulumiSandbox,
		OrgFeaturePulumiUpdatePlans,
		OrgFeatureNotebooks,
		OrgFeatureVersionsUI,
		OrgFeatureSpaceliftInstallStacks,
		OrgFeatureAWSAccountConnections,
		OrgFeatureServiceAccountsAndTokens,
		OrgFeaturePhoneHomeAuth,
		OrgFeatureRunbookStudio,
		OrgFeatureCronNamespaceIsolation,
		OrgFeatureNewAppIA,
		OrgFeatureOrgHealthcheckSweeps,
		OrgFeatureAppInstallSyncing,
		OrgFeatureSandboxOCIArtifacts,
		OrgFeatureDefaultAppBranches,
		OrgFeatureNewInstallIA,
		OrgFeatureDisableAppSync,
	}
}

// OrgFeatureInfo contains metadata about a feature flag
type OrgFeatureInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// Forced marks a flag this deployment pins on for every org, which callers
	// cannot toggle off.
	Forced bool `json:"forced"`
}

func GetFeatureDescriptions() map[OrgFeature]string {
	return map[OrgFeature]string{
		OrgFeatureAppBranches:              "Support for multiple application branches allowing parallel development and testing",
		OrgFeatureUserManagedFeatures:      "Allow organization users to manage feature flags through the public API (admin-only flag)",
		OrgFeatureSupportRole:              "Enable the support role option when inviting users to the organization",
		OrgFeatureInstallRename:            "Allow renaming installs from the dashboard edit install modal",
		OrgFeatureTerraformProviderMirror:  "Vendor terraform providers at build time and ship them inside the OCI artifact so install runners can `terraform init` without reaching registry.terraform.io",
		OrgFeatureAppBranchesUI:            "Enable the app branches UI in the dashboard for managing and switching between app branches",
		OrgFeatureTraceView:                "Enable the trace view tab on action runs, deploys, and sandbox runs to visualize OTEL spans emitted by the runner",
		OrgFeatureAutoSkipNoop:             "Automatically skip noop plans without requiring approval, overriding per-component skip_noops settings",
		OrgFeatureSlack:                    "Enable the Slack integration, including the Slack link in the dashboard sidebar and per-org Slack workspace/channel subscriptions",
		OrgFeaturePulumiSandbox:            "Enable Pulumi-typed app sandboxes (sandbox type=pulumi) in addition to Terraform",
		OrgFeaturePulumiUpdatePlans:        "Pin Pulumi applies to the approved preview via saved update plans; leave off for stacks using helm (the helm Release resource fails plan validation)",
		OrgFeatureNotebooks:                "Enable install-scoped Notebooks — a Jupyter-style surface where each cell runs a command on the install's runner via a long-lived, warm per-notebook Temporal workflow, skipping the cold install-workflow step tree for near-real-time adhoc execution.",
		OrgFeatureVersionsUI:               "Enable the install app config versions tab in the dashboard, showing the history of config updates and component diffs for each install.",
		OrgFeatureSpaceliftInstallStacks:   "Surface the Spacelift options (blueprint and administrative stack) on the install stack await step, so customers can provision the Terraform install stack through Spacelift instead of running Terraform locally.",
		OrgFeatureAWSAccountConnections:    "Enable organization-owned cross-account AWS connections with external ID trust verification.",
		OrgFeatureServiceAccountsAndTokens: "Enable the API tokens and service accounts management pages in the dashboard settings navigation.",
		OrgFeaturePhoneHomeAuth:            "Require install phone-home requests to carry an HMAC signature derived from a per-install secret, and require a target cloud account identifier (AWS account ID, GCP project ID, or Azure subscription ID) at install creation. Depends on the phone-home CMK and management-role IAM grants being in place.",
		OrgFeatureRunbookStudio:            "Enable the runbook studio in the dashboard — a literate editor for authoring runbook markdown around executable steps with a live install-state preview.",
		OrgFeatureCronNamespaceIsolation:   "Route the org's runner-healthcheck and install cron queues into dedicated Temporal namespaces + task queues polled by their own workers, isolating cron load from the api task queue.",
		OrgFeatureNewAppIA:                 "Enable the branch-centric app information architecture in the dashboard: branches as the app landing page, grouped navigation, and the app source header. Requires app-branches-ui.",
		OrgFeatureOrgHealthcheckSweeps:     "Replace per-runner and per-process healthcheck cron emitters with two per-org sweep emitters that check all runners/processes in paginated batches. Toggle via POST /v1/orgs/{org_id}/migrate-healthcheck-sweeps, which also migrates the emitters.",
		OrgFeatureAppInstallSyncing:        "Enable app install config syncing: point an app at a git repo of per-install configs so pushes to that repo sync every install's config and create missing installs behind an approval step. Gates the install syncs API, the VCS push fan-out, and the dashboard install syncs tab.",
		OrgFeatureSandboxOCIArtifacts:      "Build the app sandbox into an OCI artifact during branch runs and resolve sandbox runs against that artifact instead of cloning the sandbox git source. With it off, sandbox runs always clone git.",
		OrgFeatureDefaultAppBranches:       "Route `nuon apps sync` through an app branch run: every app gets a `default` branch covering all of its installs, and the sync hands its config to a run on that branch instead of the standalone config sync plus install rollout. Requires app-branches.",
		OrgFeatureNewInstallIA:             "Enable the new install information architecture in the dashboard. Requires app-branches-ui.",
		OrgFeatureDisableAppSync:           "Block standalone `nuon apps sync`. Config changes ship through config-managed app branches (`nuon branches sync`) instead; on a TTY the CLI offers a wizard that creates a branch config file and moves the app's installs onto it.",
	}
}

func GetFeaturesWithDescriptions() []OrgFeatureInfo {
	features := GetFeatures()
	descriptions := GetFeatureDescriptions()
	forced := ForcedFeatures()
	result := make([]OrgFeatureInfo, 0, len(features))

	for _, feature := range features {
		result = append(result, OrgFeatureInfo{
			Name:        string(feature),
			Description: descriptions[feature],
			Forced:      forced[string(feature)],
		})
	}

	return result
}

// why: adminOnlyFeatures are never exposed to org users via the public API, either
// because they gate the flag system itself or because enabling them depends on
// infrastructure prerequisites outside the org's control.
var adminOnlyFeatures = map[OrgFeature]struct{}{
	OrgFeatureUserManagedFeatures:   {},
	OrgFeatureAWSAccountConnections: {},
	OrgFeaturePhoneHomeAuth:         {},
}

func GetUserManageableFeatures() []OrgFeature {
	allFeatures := GetFeatures()
	forced := ForcedFeatures()
	manageable := make([]OrgFeature, 0, len(allFeatures)-len(adminOnlyFeatures))

	for _, feature := range allFeatures {
		if _, ok := adminOnlyFeatures[feature]; ok {
			continue
		}
		if forced[string(feature)] {
			continue
		}
		manageable = append(manageable, feature)
	}

	return manageable
}
