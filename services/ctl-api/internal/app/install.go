package app

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/lifecyclephase"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/viewsql"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/links"
)

// Deprecated: Use lifecyclephase.Phase constants directly.
type InstallLifecycleStatus = lifecyclephase.Phase

type Install struct {
	ID             string                        `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID    string                        `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy      Account                       `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt      time.Time                     `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt      time.Time                     `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt      soft_delete.DeletedAt         `gorm:"index:idx_app_install_name,unique" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`
	Metadata       pgtype.Hstore                 `json:"metadata,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"metadata,omitzero,omitempty"`
	LifecyclePhase lifecyclephase.LifecyclePhase `json:"lifecycle_phase,omitzero" gorm:"type:jsonb" swaggertype:"object" temporaljson:"lifecycle_phase,omitzero,omitempty"`
	labels.Labeled

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	Name  string `json:"name,omitzero" gorm:"notnull;index:idx_app_install_name,unique" temporaljson:"name,omitzero,omitempty"`
	App   App    `swaggerignore:"true" json:"app,omitzero" temporaljson:"app,omitzero,omitempty"`
	AppID string `json:"app_id,omitzero" gorm:"notnull;index:idx_app_install_name,unique" temporaljson:"app_id,omitzero,omitempty"`

	SandboxMode sql.NullBool `json:"sandbox_mode,omitempty" gorm:"column:sandbox_mode" temporaljson:"sandbox_mode,omitempty"`

	AppConfigID string    `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`
	AppConfig   AppConfig `json:"-" temporaljson:"app_config,omitzero,omitempty"`

	AppBranchID generics.NullString `json:"app_branch_id,omitzero" gorm:"index" swaggertype:"string" temporaljson:"app_branch_id,omitzero,omitempty"`
	AppBranch   *AppBranch          `json:"app_branch,omitempty" temporaljson:"app_branch,omitzero,omitempty"`

	AppBranchConnections []InstallAppBranchConnection `json:"app_branch_connections,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"app_branch_connections,omitzero,omitempty"`

	AppSandboxConfigID string           `json:"-" swaggerignore:"true" temporaljson:"app_sandbox_config_id,omitzero,omitempty"`
	AppSandboxConfig   AppSandboxConfig `json:"app_sandbox_config,omitzero" temporaljson:"app_sandbox_config,omitzero,omitempty"`

	AppRunnerConfigID string          `json:"-" swaggerignore:"true" temporaljson:"app_runner_config_id,omitzero,omitempty"`
	AppRunnerConfig   AppRunnerConfig `json:"app_runner_config,omitzero" temporaljson:"app_runner_config,omitzero,omitempty"`

	InstallComponents       []InstallComponent        `json:"install_components,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_components,omitzero,omitempty"`
	InstallActionWorkflows  []InstallActionWorkflow   `json:"install_action_workflows,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_action_workflows,omitzero,omitempty"`
	InstallSandboxRuns      []InstallSandboxRun       `json:"install_sandbox_runs,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_sandbox_runs,omitzero,omitempty"`
	InstallInputs           []InstallInputs           `json:"install_inputs,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_inputs,omitzero,omitempty"`
	InstallEvents           []InstallEvent            `json:"install_events,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_events,omitzero,omitempty"`
	InstallIntermediateData []InstallIntermediateData `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_intermediate_data,omitzero,omitempty"`
	InstallSandbox          InstallSandbox            `json:"sandbox" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_sandbox,omitzero,omitempty"`
	InstallConfig           *InstallConfig            `json:"install_config,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_config,omitzero,omitempty"`
	InstallStates           []InstallState            `json:"install_states,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_states,omitzero,omitempty"`

	// InstallRoles is a list of roles associated with that install at given app config ID
	InstallRoles []InstallRoles `json:"install_roles,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_roles,omitzero,omitempty"`

	InstallStack *InstallStack `json:"install_stack,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_stack,omitzero,omitempty"`
	AWSAccount   *AWSAccount   `json:"aws_account,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"aws_account,omitzero,omitempty"`
	AzureAccount *AzureAccount `json:"azure_account,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"azure_account,omitzero,omitempty"`
	GCPAccount   *GCPAccount   `json:"gcp_account,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"gcp_account,omitzero,omitempty"`

	RunnerGroup RunnerGroup `json:"-" gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" temporaljson:"runner_group,omitzero,omitempty"`

	// ComponentHealthContext lets the runner's stateless component-health engine
	// rehydrate cluster access and sandbox helm release ownership after a
	// restart. Opaque to ctl-api; only read/written via the runner-authed
	// component-health-context endpoints. Never serialized on the install
	// itself since it can carry durable cluster access details.
	ComponentHealthContext ComponentHealthContext `gorm:"type:jsonb" json:"-" temporaljson:"-"`

	// SandboxHealthStatus / SandboxHealthMessage are a denormalized rollup of the
	// worst health across the sandbox-owned resources reported by the
	// component-health engine, written on each ingest so every install read can
	// surface a degraded sandbox without querying ClickHouse. Empty until the
	// engine reports.
	SandboxHealthStatus  string `json:"sandbox_health_status,omitzero" gorm:"column:sandbox_health_status;default:''" temporaljson:"sandbox_health_status,omitzero,omitempty"`
	SandboxHealthMessage string `json:"sandbox_health_message,omitzero" gorm:"column:sandbox_health_message;default:''" temporaljson:"sandbox_health_message,omitzero,omitempty"`

	// CloudPlatformMetadata records the cloud account this install is expected to
	// run in, and what it was observed running in. See the type for the trust model.
	CloudPlatformMetadata CloudPlatformMetadata `json:"cloud_platform_metadata,omitzero" gorm:"type:jsonb" swaggertype:"object" temporaljson:"cloud_platform_metadata,omitzero,omitempty"`

	// PhoneHomeAuth carries the salt and root key name used to derive this install's
	// phone-home signing secret. Kept in its own column rather than nested inside
	// CloudPlatformMetadata because that struct is serialized to the wire and these
	// derivation inputs must not be, and json:"-" on a nested field would exclude it
	// from the jsonb value too, not just the API response.
	PhoneHomeAuth *PhoneHomeAuth `json:"-" gorm:"type:jsonb" temporaljson:"-"`

	// generated view current view

	InstallNumber            int                  `json:"install_number,omitzero" gorm:"->;-:migration" temporaljson:"install_number,omitzero,omitempty"`
	SandboxStatus            InstallSandboxStatus `json:"sandbox_status,omitzero" gorm:"->;-:migration" swaggertype:"string" temporaljson:"sandbox_status,omitzero,omitempty"`
	SandboxStatusDescription string               `json:"sandbox_status_description,omitzero" gorm:"-" swaggertype:"string" temporaljson:"sandbox_status_description,omitzero,omitempty"`
	ComponentStatuses        pgtype.Hstore        `json:"component_statuses,omitzero" gorm:"type:hstore;->;-:migration" swaggertype:"object,string" temporaljson:"component_statuses,omitzero,omitempty"`
	ComponentHealthStatuses  pgtype.Hstore        `json:"component_health_statuses,omitzero" gorm:"type:hstore;->;-:migration" swaggertype:"object,string" temporaljson:"component_health_statuses,omitzero,omitempty"`

	Workflows []Workflow `json:"workflows,omitzero" gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" temporaljson:"workflows,omitzero,omitempty"`

	Queues []Queue `json:"queues,omitzero" gorm:"polymorphic:Owner;" temporaljson:"queues,omitzero,omitempty"`

	// after queries

	CurrentInstallInputs                *InstallInputs         `json:"-" gorm:"-" temporaljson:"current_install_inputs,omitzero,omitempty"`
	CompositeComponentStatus            InstallComponentStatus `json:"composite_component_status,omitzero" gorm:"-" swaggertype:"string" temporaljson:"composite_component_status,omitzero,omitempty"`
	CompositeComponentStatusDescription string                 `json:"composite_component_status_description,omitzero" gorm:"-" swaggertype:"string" temporaljson:"composite_component_status_description,omitzero,omitempty"`

	// CompositeHealthStatus is the live-health rollup of the install's
	// components — a parallel axis to CompositeComponentStatus (deploy
	// lifecycle), never merged with it. Empty until the component-health
	// evaluator has produced verdicts.
	CompositeHealthStatus            InstallComponentHealthStatus `json:"composite_health_status,omitzero" gorm:"-" swaggertype:"string" temporaljson:"composite_health_status,omitzero,omitempty"`
	CompositeHealthStatusDescription string                       `json:"composite_health_status_description,omitzero" gorm:"-" swaggertype:"string" temporaljson:"composite_health_status_description,omitzero,omitempty"`
	RunnerStatus                     RunnerStatus                 `json:"runner_status,omitzero" gorm:"-" swaggertype:"string" temporaljson:"runner_status,omitzero,omitempty"`
	RunnerStatusDescription          string                       `json:"runner_status_description,omitzero" gorm:"-" swaggertype:"string" temporaljson:"runner_status_description,omitzero,omitempty"`
	RunnerID                         string                       `json:"runner_id,omitzero" gorm:"-" temporaljson:"runner_id,omitzero,omitempty"`
	CloudPlatform                    CloudPlatform                `json:"cloud_platform,omitzero" gorm:"-" swaggertype:"string" temporaljson:"cloud_platform,omitzero,omitempty"`
	RunnerType                       AppRunnerType                `json:"runner_type,omitzero" gorm:"-" swaggertype:"string" temporaljson:"runner_type,omitzero,omitempty"`
	DriftedObjects                   []DriftedObject              `json:"drifted_objects,omitzero" gorm:"-" temporaljson:"drifted_objects,omitzero,omitempty"`
	Links                            map[string]any               `json:"links,omitzero,omitempty" temporaljson:"-" gorm:"-"`

	// Expected* coalesce the target identifier with the observed one, so callers get
	// the strongest identifier available without caring which is set.
	ExpectedAccountID      string `json:"expected_account_id,omitzero" gorm:"-" temporaljson:"expected_account_id,omitzero,omitempty"`
	ExpectedProjectID      string `json:"expected_project_id,omitzero" gorm:"-" temporaljson:"expected_project_id,omitzero,omitempty"`
	ExpectedSubscriptionID string `json:"expected_subscription_id,omitzero" gorm:"-" temporaljson:"expected_subscription_id,omitzero,omitempty"`

	// WorkflowID is populated by handlers that create a workflow. Not persisted.
	WorkflowID *string `json:"workflow_id,omitempty" gorm:"-"`
}

func (i *Install) UseView() bool {
	return true
}

func (i *Install) ViewVersion() string {
	return "v9"
}

func (i *Install) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.DefaultViewName(db, &Install{}, 9),
			SQL:           viewsql.InstallsViewV9,
			AlwaysReapply: true,
		},
		{
			Name: views.CustomViewName(db, &Install{}, "migration_test"),
			SQL:  `SELECT * FROM installs ORDER BY created_at DESC`,
		},
	}
}

func (i *Install) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &Install{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (i *Install) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallID()
	}

	i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	i.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

// We want to report the status of the sandbox, the runner, and the components,
// and then roll that up into a high-level status for the install overall.
func (i *Install) AfterQuery(tx *gorm.DB) error {
	i.Links = links.InstallLinks(tx.Statement.Context, i.ID)

	// get the runner status
	i.RunnerStatus = RunnerStatusDeprovisioned
	if len(i.RunnerGroup.Runners) > 0 {
		i.RunnerStatus = i.RunnerGroup.Runners[0].Status
		i.RunnerStatusDescription = i.RunnerGroup.Runners[0].StatusDescription
		i.RunnerID = i.RunnerGroup.Runners[0].ID
	}

	if len(i.InstallInputs) > 0 {
		i.CurrentInstallInputs = &i.InstallInputs[0]
	}

	// get the composite status of all the components
	i.CompositeComponentStatus = compositeComponentStatus(i.ComponentStatuses)
	i.CompositeComponentStatusDescription = compositeComponentStatusDescription(i.ComponentStatuses)

	i.CompositeHealthStatus, i.CompositeHealthStatusDescription = compositeComponentHealthStatus(i.ComponentHealthStatuses)

	// If sandbox mode not explicitly set on the install, inherit from org.
	if !i.SandboxMode.Valid {
		org := i.Org
		if org.Name == "" {
			tx.Where("id = ?", i.OrgID).First(&org)
		}
		i.SandboxMode.Valid = true
		i.SandboxMode.Bool = org.SandboxMode
	}

	if i.AppRunnerConfig.ID != "" {
		i.CloudPlatform = i.AppRunnerConfig.CloudPlatform
		i.RunnerType = i.AppRunnerConfig.Type

	} else {
		i.CloudPlatform = CloudPlatformUnknown
		i.RunnerType = AppRunnerTypeUnknown
	}

	i.setExpectedCloudIdentifiers()

	return nil
}

// setExpectedCloudIdentifiers prefers the target identifier supplied at install
// creation over the one a phone home reported, falling back to the latter so
// installs predating the target field still resolve.
func (i *Install) setExpectedCloudIdentifiers() {
	cpm := i.CloudPlatformMetadata
	i.ExpectedAccountID = firstNonEmpty(cpm.TargetAccountID, cpm.ObservedAccountID)
	i.ExpectedProjectID = firstNonEmpty(cpm.TargetProjectID, cpm.ObservedProjectID)
	i.ExpectedSubscriptionID = firstNonEmpty(cpm.TargetSubscriptionID, cpm.ObservedSubscriptionID)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

// compositeComponentStatus coalesces a single status from the statuses of the app's components.
// This is based on the components defined in the app config, not the components present in the install.
// Components may be present in an install's history that have been removed from the app.
func compositeComponentStatus(componentStatuses pgtype.Hstore) InstallComponentStatus {
	// if there are no components, then there are no operations to wait for
	if len(componentStatuses) == 0 {
		return InstallComponentStatusPending
	}

	// check status of each component
	activecount := 0
	for _, status := range componentStatuses {
		switch InstallComponentStatus(*status) {
		case InstallComponentStatusActive:
			activecount++
		case InstallComponentStatusError:
			// if any components have failed, composite status should be "error"
			// we can stop immediately
			return InstallComponentStatusError
		}
	}

	// if all components are active, composite status should be "active"
	if activecount == len(componentStatuses) {
		return InstallComponentStatusActive
	}

	// if any components have not yet succeeded or failed, composite status should be "pending"
	return InstallComponentStatusPending
}

func compositeComponentStatusDescription(componentStatuses pgtype.Hstore) string {
	// if there are no components, then there are no operations to wait for
	if len(componentStatuses) == 0 {
		return "No active components"
	}

	// check status of each component
	activecount := 0
	for _, status := range componentStatuses {
		switch InstallComponentStatus(*status) {
		case InstallComponentStatusActive:
			activecount++
		case InstallComponentStatusError:
			// if any components have failed we can stop immediately
			return "A component is in an error state"
		}
	}

	// if all components are active
	if activecount == len(componentStatuses) {
		return "All components have been deployed"
	}

	// if any components have not yet succeeded or failed
	return "Waiting on components"
}

// compositeComponentHealthStatus rolls the per-component health axis up to a
// single install-level verdict. Unset and not-applicable components carry no
// health signal and are excluded; an install with no evaluated components has
// no composite health (empty), so orgs without the feature surface nothing.
func compositeComponentHealthStatus(componentHealthStatuses pgtype.Hstore) (InstallComponentHealthStatus, string) {
	statuses := make([]InstallComponentHealthStatus, 0, len(componentHealthStatuses))
	for _, status := range componentHealthStatuses {
		if status == nil {
			continue
		}
		statuses = append(statuses, InstallComponentHealthStatus(*status))
	}
	return CompositeComponentHealthStatus(statuses)
}

// CompositeComponentHealthStatus rolls per-component health verdicts up to a
// single install-level verdict and description. Exported so the component
// health evaluator can compute the before/after rollup without a round trip
// through the install view.
func CompositeComponentHealthStatus(statuses []InstallComponentHealthStatus) (InstallComponentHealthStatus, string) {
	counts := map[InstallComponentHealthStatus]int{}
	total := 0
	for _, hs := range statuses {
		if hs == InstallComponentHealthStatusUnset || hs == InstallComponentHealthStatusNotApplicable {
			continue
		}
		counts[hs]++
		total++
	}
	if total == 0 {
		return InstallComponentHealthStatusUnset, ""
	}

	describe := func(n int, verb string) string {
		if n == 1 {
			return fmt.Sprintf("1 component is %s", verb)
		}
		return fmt.Sprintf("%d components are %s", n, verb)
	}

	switch {
	case counts[InstallComponentHealthStatusUnhealthy] > 0:
		return InstallComponentHealthStatusUnhealthy, describe(counts[InstallComponentHealthStatusUnhealthy], "unhealthy")
	case counts[InstallComponentHealthStatusDegraded] > 0:
		return InstallComponentHealthStatusDegraded, describe(counts[InstallComponentHealthStatusDegraded], "degraded")
	case counts[InstallComponentHealthStatusUnknown] > 0:
		return InstallComponentHealthStatusUnknown, describe(counts[InstallComponentHealthStatusUnknown], "in an unknown health state")
	case counts[InstallComponentHealthStatusProgressing] > 0:
		return InstallComponentHealthStatusProgressing, describe(counts[InstallComponentHealthStatusProgressing], "progressing")
	default:
		return InstallComponentHealthStatusHealthy, "All components are healthy"
	}
}

// ComponentHealthContext persists what the runner's component-health engine
// needs to rehydrate cluster access after a restart: a durable, opaque
// (to ctl-api) marshaled kube.ClusterInfo, and the helm release names the
// install's sandbox manages (base infra like external-dns, cert-manager).
type ComponentHealthContext struct {
	ClusterInfoJSON     string   `json:"cluster_info_json"`
	SandboxHelmReleases []string `json:"sandbox_helm_releases"`
}

// Scan implements the database/sql.Scanner interface.
func (c *ComponentHealthContext) Scan(v interface{}) (err error) {
	switch v := v.(type) {
	case nil:
		return nil
	case []byte:
		if err := json.Unmarshal(v, c); err != nil {
			return errors.Wrap(err, "unable to scan component health context")
		}
	}
	return
}

// Value implements the driver.Valuer interface.
func (c *ComponentHealthContext) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (ComponentHealthContext) GormDataType() string {
	return "jsonb"
}

// Where a CloudPlatformMetadata target identifier came from. A backfilled target
// is derived from an unauthenticated phone home, so it pins whatever account last
// phoned home rather than an independently attested one — a weaker control than a
// user- or connection-supplied target.
const (
	CloudPlatformTargetSourceUser       = "user"
	CloudPlatformTargetSourceConnection = "connection"
	CloudPlatformTargetSourceBackfill   = "backfill"
)

// CloudPlatformMetadata records which cloud account an install is expected to run
// in. Target values are supplied at install creation (or derived from an AWS
// account connection); observed values are what the install's stack reported when
// it phoned home. Once phone-home requests are signature-verified this becomes the
// trusted copy, with InstallStackOutputs remaining the untrusted vendor-facing echo.
type CloudPlatformMetadata struct {
	// AWS
	TargetAccountID   string `json:"target_account_id,omitempty"`
	ObservedAccountID string `json:"observed_account_id,omitempty"`

	// GCP
	TargetProjectID   string `json:"target_project_id,omitempty"`
	ObservedProjectID string `json:"observed_project_id,omitempty"`

	// Azure
	TargetSubscriptionID   string `json:"target_subscription_id,omitempty"`
	ObservedSubscriptionID string `json:"observed_subscription_id,omitempty"`

	TargetSource string `json:"target_source,omitempty"`
}

// HasTarget reports whether any cloud's target identifier has been set.
func (c CloudPlatformMetadata) HasTarget() bool {
	return c.TargetAccountID != "" || c.TargetProjectID != "" || c.TargetSubscriptionID != ""
}

// Scan implements the database/sql.Scanner interface.
func (c *CloudPlatformMetadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.Errorf("cannot scan %T into CloudPlatformMetadata", value)
	}
	if len(bytes) == 0 {
		return nil
	}

	return json.Unmarshal(bytes, c)
}

// Value implements the driver.Valuer interface.
func (c CloudPlatformMetadata) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (CloudPlatformMetadata) GormDataType() string {
	return "jsonb"
}

// PhoneHomeAuth holds the server-side state needed to verify a signed phone home
// for this install: the salt the per-install secret is derived from, the name of
// the root key that derived it, and where the derived secret was published for the
// caller to fetch. Never serialized — the derivation inputs stay control-plane side.
type PhoneHomeAuth struct {
	Salt string `json:"salt"`
	// KeyID is the *name* of the root key that derived this secret, not an index.
	KeyID        string    `json:"key_id"`
	SecretARN    string    `json:"secret_arn,omitempty"`
	SecretRegion string    `json:"secret_region,omitempty"`
	KMSKeyARN    string    `json:"kms_key_arn,omitempty"`
	CreatedAt    time.Time `json:"created_at"`

	LastVerifiedAt *time.Time `json:"last_verified_at,omitempty"`
	LastRejectedAt *time.Time `json:"last_rejected_at,omitempty"`
}

// Scan implements the database/sql.Scanner interface.
func (p *PhoneHomeAuth) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.Errorf("cannot scan %T into PhoneHomeAuth", value)
	}
	if len(bytes) == 0 {
		return nil
	}

	return json.Unmarshal(bytes, p)
}

// Value implements the driver.Valuer interface.
func (p PhoneHomeAuth) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (PhoneHomeAuth) GormDataType() string {
	return "jsonb"
}
