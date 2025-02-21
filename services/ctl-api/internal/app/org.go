package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/bulk"
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
)

type Org struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string  `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account `json:"-"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `gorm:"index:idx_org_name,unique" json:"-"`

	Name              string    `gorm:"index:idx_org_name,unique;notnull" json:"name"`
	Status            OrgStatus `json:"status" gorm:"notnull" swaggertype:"string"`
	StatusDescription string    `json:"status_description" gorm:"notnull"`

	SandboxMode bool `json:"sandbox_mode" gorm:"notnull"`

	OrgType OrgType `json:"-"`

	NotificationsConfig   NotificationsConfig `gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" json:"notifications_config,omitempty"`
	NotificationsConfigID string              `json:"-"`

	RunnerGroup RunnerGroup `json:"runner_group" gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;"`

	LogoURL string `json:"logo_url"`

	Apps           []App           `faker:"-" swaggerignore:"true" json:"apps,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	VCSConnections []VCSConnection `json:"vcs_connections,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	Invites        []OrgInvite     `faker:"-" swaggerignore:"true" json:"-" gorm:"constraint:OnDelete:CASCADE;"`

	// Other relationships as part of the data model

	Runners                   []Runner                   `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	PublicGitVCSConfigs       []PublicGitVCSConfig       `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	ConnectedGithubVCSConfigs []ConnectedGithubVCSConfig `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	VCSConnectionCommits      []VCSConnectionCommit      `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	AWSECRImageConfigs        []AWSECRImageConfig        `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	Installs                  []Install                  `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	Components                []Component                `gorm:"constraint:OnDelete:CASCADE;" json:"-"`

	Installers        []Installer         `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
	InstallerMetadata []InstallerMetadata `gorm:"constraint:OnDelete:CASCADE;" json:"-"`

	Roles        []Role        `faker:"-" swaggerignore:"true" json:"roles,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	Policies     []Policy      `faker:"-" swaggerignore:"true" json:"policies,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	AccountRoles []AccountRole `faker:"-" swaggerignore:"true" json:"account_roles,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
}

func (o *Org) AfterQuery(tx *gorm.DB) error {
	return nil
}

func (o *Org) BeforeCreate(tx *gorm.DB) error {
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
