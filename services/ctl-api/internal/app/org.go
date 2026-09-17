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

	Apps                  []App                  `faker:"-" swaggerignore:"true" json:"apps,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"apps,omitzero,omitempty"`
	VCSConnections        []VCSConnection        `json:"vcs_connections,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"vcs_connections,omitzero,omitempty"`
	AWSAccountConnections []AWSAccountConnection `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"aws_account_connections,omitzero,omitempty"`
	Invites               []OrgInvite            `faker:"-" swaggerignore:"true" json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"invites,omitzero,omitempty"`
	Features              types.StringBoolMap    `json:"features,omitzero" gorm:"type:jsonb;default null" temporaljson:"features,omitzero,omitempty"`
	Tags                  pq.StringArray         `json:"tags,omitzero" gorm:"type:text[];default '{}'" swaggertype:"array,string" temporaljson:"tags,omitzero,omitempty"`
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
	forced := ForcedFeatures()

	// if active feature not in features, add it
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
