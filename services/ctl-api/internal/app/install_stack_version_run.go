package app

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/viewsql"
)

type StackVersionRunType string

const (
	StackVersionRunTypeWorkflow  StackVersionRunType = "workflow-run"
	StackVersionRunTypeOutOfBand StackVersionRunType = "out-of-band-update"
)

type StackVersionRunRoleDiff struct {
	Enabled  []string `json:"enabled,omitempty"`
	Disabled []string `json:"disabled,omitempty"`
}

func (d *StackVersionRunRoleDiff) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into StackVersionRunRoleDiff", value)
	}
	if len(bytes) == 0 {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func (d StackVersionRunRoleDiff) Value() (driver.Value, error) {
	return json.Marshal(d)
}

type StackVersionRunInputDiff struct {
	Added   []string `json:"added,omitempty"`
	Removed []string `json:"removed,omitempty"`
	Changed []string `json:"changed,omitempty"`
}

func (d *StackVersionRunInputDiff) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into StackVersionRunInputDiff", value)
	}
	if len(bytes) == 0 {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func (d StackVersionRunInputDiff) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// InstallStackVersionRunKind identifies the operation a stack run represents.
type InstallStackVersionRunKind string

const (
	InstallStackVersionRunKindProvision   InstallStackVersionRunKind = "provision"
	InstallStackVersionRunKindReprovision InstallStackVersionRunKind = "reprovision"
	InstallStackVersionRunKindDeprovision InstallStackVersionRunKind = "deprovision"
)

// Valid reports whether k is a known kind.
func (k InstallStackVersionRunKind) Valid() bool {
	switch k {
	case InstallStackVersionRunKindProvision,
		InstallStackVersionRunKindReprovision,
		InstallStackVersionRunKindDeprovision:
		return true
	}
	return false
}

type InstallStackVersionRun struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	InstallStackVersionID string              `json:"install_stack_version_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"install_stack_version_id,omitzero,omitempty"`
	InstallStackVersion   InstallStackVersion `json:"-" temporaljson:"install_stack_version,omitzero,omitempty"`

	// Kind is the operation this run represents. provision = first-time create,
	// reprovision = idempotent reconcile of an existing install, deprovision =
	// tear-down. default:'provision' lets gorm auto-migrate back-fill historical
	// rows; BeforeCreate validates the value.
	Kind InstallStackVersionRunKind `json:"kind,omitempty" gorm:"type:text;notnull;default:'provision'" temporaljson:"kind,omitzero,omitempty"`

	Data         pgtype.Hstore  `json:"data,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"data,omitzero,omitempty"`
	DataContents map[string]any `json:"data_contents,omitzero" gorm:"-"`

	RunType   StackVersionRunType       `json:"run_type,omitzero" gorm:"type:varchar(50)" temporaljson:"run_type,omitzero,omitempty"`
	RoleDiff  *StackVersionRunRoleDiff  `json:"role_diff,omitzero" gorm:"type:jsonb" temporaljson:"role_diff,omitzero,omitempty"`
	InputDiff *StackVersionRunInputDiff `json:"input_diff,omitzero" gorm:"type:jsonb" temporaljson:"input_diff,omitzero,omitempty"`
	Status    CompositeStatus           `json:"composite_status,omitzero" gorm:"type:jsonb" temporaljson:"status,omitzero,omitempty"`

	// LogStreamID is the OTLP log stream the SDK pushes provisioning logs to.
	// Persisted so the PATCH handler can close the stream on terminal status.
	LogStreamID string `json:"log_stream_id,omitzero" gorm:"default:null" temporaljson:"log_stream_id,omitzero,omitempty"`

	// LogStream is populated transiently:
	//   - On the POST response: with WriteToken + RunnerAPIURL so the SDK can
	//     start pushing logs immediately.
	//   - On GET-runs (via Preload): without WriteToken, so the dashboard can
	//     find the stream to render but can't write to it.
	LogStream *LogStream `json:"log_stream,omitempty" gorm:"foreignKey:LogStreamID;references:ID"`

	// SDKConfig is populated transiently on the POST response so the AWS-native
	// SDK provisioner has everything it needs to apply the stack: runner ID,
	// runner API URL, operation-role permissions, secrets, etc. Mirrors the TF
	// tfvars contract at install-stacks/aws/variables.tf — see
	// sdks/nuon-installer-go/stack/config.go for the SDK-side struct.
	SDKConfig *InstallerSDKConfig `json:"config,omitempty" gorm:"-" swaggerignore:"true"`
}

// InstallerSDKConfig is the JSON shape the SDK expects on CreateRunResponse.
// Field tags match sdks/nuon-installer-go/stack/config.go exactly. Only
// populated on the create-run POST response — never persisted, never returned
// on the GET path.
type InstallerSDKConfig struct {
	InstallID    string `json:"install_id,omitempty"`
	RunnerID     string `json:"runner_id,omitempty"`
	RunnerAPIURL string `json:"runner_api_url,omitempty"`

	NuonSupportIAMRoleARNs []string `json:"nuon_support_iam_role_arns,omitempty"`

	InstallInputs map[string]string `json:"install_inputs,omitempty"`

	AutoGenerateSecrets []string                          `json:"auto_generate_secrets,omitempty"`
	Secrets             map[string]InstallerSDKSecret     `json:"secrets,omitempty"`
	BreakGlassRoles     map[string]InstallerSDKRoleConfig `json:"break_glass_roles,omitempty"`
	CustomRoles         map[string]InstallerSDKRoleConfig `json:"custom_roles,omitempty"`

	ProvisionPermissions          []string `json:"provision_permissions,omitempty"`
	ProvisionInlinePolicyDocument string   `json:"provision_inline_policy_document,omitempty"`
	ProvisionManagedPolicyARNs    []string `json:"provision_managed_policy_arns,omitempty"`

	MaintenancePermissions          []string `json:"maintenance_permissions,omitempty"`
	MaintenanceInlinePolicyDocument string   `json:"maintenance_inline_policy_document,omitempty"`
	MaintenanceManagedPolicyARNs    []string `json:"maintenance_managed_policy_arns,omitempty"`

	DeprovisionPermissions          []string `json:"deprovision_permissions,omitempty"`
	DeprovisionInlinePolicyDocument string   `json:"deprovision_inline_policy_document,omitempty"`
	DeprovisionManagedPolicyARNs    []string `json:"deprovision_managed_policy_arns,omitempty"`
}

// InstallerSDKSecret is the customer-provided secret shape.
type InstallerSDKSecret struct {
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Value       string `json:"value,omitempty"`
}

// InstallerSDKRoleConfig is the per-role payload for break-glass/custom roles.
type InstallerSDKRoleConfig struct {
	Permissions          []string `json:"permissions,omitempty"`
	InlinePolicyDocument string   `json:"inline_policy_document,omitempty"`
	ManagedPolicyARNs    []string `json:"managed_policy_arns,omitempty"`
	Enabled              bool     `json:"enabled,omitempty"`
}

func (i *InstallStackVersionRun) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallStackVersionRun{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (a *InstallStackVersionRun) AfterQuery(tx *gorm.DB) error {
	if len(a.Data) < 1 {
		return nil
	}

	a.DataContents = map[string]any{}
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToSliceHookFunc(","),
			generics.StringToMapDecodeHook(),
			mapstructure.StringToTimeDurationHookFunc(),
		),
		WeaklyTypedInput: true,
		Result:           &a.DataContents,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create gcp decoder")
	}
	if err := decoder.Decode(a.Data); err != nil {
		return errors.Wrap(err, "unable to parse gcp outputs")
	}

	return nil
}

func (i *InstallStackVersionRun) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallStackVersionRunID()
	}

	if i.CreatedByID == "" {
		i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if i.OrgID == "" {
		i.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	if i.Kind == "" {
		i.Kind = InstallStackVersionRunKindProvision
	}
	if !i.Kind.Valid() {
		return fmt.Errorf("invalid install stack version run kind: %q", i.Kind)
	}

	return nil
}

func (i *InstallStackVersionRun) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.CustomViewName(db, &InstallStackVersionRun{}, "state_view_v1"),
			SQL:           viewsql.InstallStackVersionRunsStateViewV1,
			AlwaysReapply: true,
		},
	}
}
