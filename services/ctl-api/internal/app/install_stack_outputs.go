package app

import (
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
)

type InstallStackOutputs struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	InstallStackID           string              `json:"install_stack_id,omitzero" gorm:"notnull;default null" temporaljson:"install_stack_id,omitzero,omitempty"`
	InstallStackVersionRunID generics.NullString `json:"install_version_run_id,omitzero" swaggertype:"string" temporaljson:"install_stack_version_run_id,omitzero,omitempty"`

	Data         pgtype.Hstore  `json:"data,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"data,omitzero,omitempty"`
	DataContents map[string]any `json:"data_contents,omitzero" gorm:"-"`

	AWSStackOutputs   *AWSStackOutputs   `json:"aws,omitzero" gorm:"-" temporaljson:"aws_stack_outputs,omitzero,omitempty"`
	AzureStackOutputs *AzureStackOutputs `json:"azure,omitzero" gorm:"-" temporaljson:"azure_stack_outputs,omitzero,omitempty"`
	GCPStackOutputs   *GCPStackOutputs   `json:"gcp,omitzero" gorm:"-" temporaljson:"gcp_stack_outputs,omitzero,omitempty"`
}

type StackOutput interface {
	ProvisionRoleID() (string, error)
	DeprovisionRoleID() (string, error)
	MaintenanceRoleID() (string, error)
	CustomRoleID(string) (string, error)
	BreakGlassRoleID(string) (string, error)
	// returns list of custom roles along with their identifier
	CustomRoles() (map[string]string, error)
	// returns list of break glass roles along with their identifier
	BreakGlassRoles() (map[string]string, error)
	// returns list of customer install inputs along with their values in the stack
	InstallInputValues() (map[string]string, error)
}

type AWSStackOutputs struct {
	AccountID             string            `json:"account_id,omitzero" mapstructure:"account_id" temporaljson:"account_id,omitzero,omitempty"`
	Region                string            `json:"region,omitzero" mapstructure:"region" temporaljson:"region,omitzero,omitempty"`
	VPCID                 string            `json:"vpc_id,omitzero" mapstructure:"vpc_id" temporaljson:"vpcid,omitzero,omitempty"`
	RunnerSubnet          string            `json:"runner_subnet,omitzero" mapstructure:"runner_subnet" temporaljson:"runner_subnet,omitzero,omitempty"`
	PublicSubnets         []string          `json:"public_subnets,omitzero" mapstructure:"public_subnets" temporaljson:"public_subnets,omitzero,omitempty"`
	PrivateSubnets        []string          `json:"private_subnets,omitzero" mapstructure:"private_subnets" temporaljson:"private_subnets,omitzero,omitempty"`
	ProvisionIAMRoleARN   string            `json:"provision_iam_role_arn,omitzero" mapstructure:"provision_iam_role_arn" temporaljson:"provision_iam_role_arn,omitzero,omitempty"`
	DeprovisionIAMRoleARN string            `json:"deprovision_iam_role_arn,omitzero" mapstructure:"deprovision_iam_role_arn" temporaljson:"deprovision_iam_role_arn,omitzero,omitempty"`
	MaintenanceIAMRoleARN string            `json:"maintenance_iam_role_arn,omitzero" mapstructure:"maintenance_iam_role_arn" temporaljson:"maintenance_iam_role_arn,omitzero,omitempty"`
	RunnerIAMRoleARN      string            `json:"runner_iam_role_arn,omitzero" mapstructure:"runner_iam_role_arn" temporaljson:"runner_iam_role_arn,omitzero,omitempty"`
	BreakGlassRoleARNs    map[string]string `json:"break_glass_role_arns,omitzero" mapstructure:"break_glass_role_arns" temporaljson:"break_glass_role_arns,omitzero,omitempty"`
	CustomRoleARNs        map[string]string `json:"custom_role_arns,omitzero" mapstructure:"custom_role_arns" temporaljson:"custom_role_arns,omitzero,omitempty"`
	InstallInputs         map[string]string `json:"install_inputs,omitzero" mapstructure:"install_inputs" temporaljson:"install_inputs,omitzero,omitempty"`
}

func (a *AWSStackOutputs) ProvisionRoleID() (string, error)   { return a.ProvisionIAMRoleARN, nil }
func (a *AWSStackOutputs) DeprovisionRoleID() (string, error) { return a.DeprovisionIAMRoleARN, nil }
func (a *AWSStackOutputs) MaintenanceRoleID() (string, error) { return a.MaintenanceIAMRoleARN, nil }

func (a *AWSStackOutputs) CustomRoleID(name string) (string, error) {
	arn, ok := a.CustomRoleARNs[name]
	if !ok {
		return "", fmt.Errorf("custom role %q does not exist in stack outputs", name)
	}
	return arn, nil
}

func (a *AWSStackOutputs) BreakGlassRoleID(name string) (string, error) {
	arn, ok := a.BreakGlassRoleARNs[name]
	if !ok {
		return "", fmt.Errorf("break glass role %q does not exist in stack outputs", name)
	}
	return arn, nil
}

func (a *AWSStackOutputs) CustomRoles() (map[string]string, error) {
	return a.CustomRoleARNs, nil
}

func (a *AWSStackOutputs) BreakGlassRoles() (map[string]string, error) {
	return a.BreakGlassRoleARNs, nil
}

func (a *AWSStackOutputs) InstallInputValues() (map[string]string, error) {
	return a.InstallInputs, nil
}

type AzureStackOutputs struct {
	ResourceGroupID       string `json:"resource_group_id,omitzero" mapstructure:"resource_group_id" temporaljson:"resource_group_id,omitzero,omitempty"`
	ResourceGroupName     string `json:"resource_group_name,omitzero" mapstructure:"resource_group_name" temporaljson:"resource_group_name,omitzero,omitempty"`
	ResourceGroupLocation string `json:"resource_group_location,omitzero" mapstructure:"resource_group_location" temporaljson:"resource_group_location,omitzero,omitempty"`

	SubscriptionID       string `cty:"subscription_id" json:"subscription_id" hcl:"subscription_id" temporaljson:"subscription_id,omitzero,omitempty" mapstructure:"subscription_id"`
	SubscriptionTenantID string `cty:"subscription_tenant_id" json:"subscription_tenant_id" hcl:"subscription_tenant_id" temporaljson:"subscription_tenant_id,omitzero,omitempty" mapstructure:"subscription_tenant_id"`

	NetworkID   string `json:"network_id,omitzero" mapstructure:"network_id" temporaljson:"network_id,omitzero,omitempty"`
	NetworkName string `json:"network_name,omitzero" mapstructure:"network_name" temporaljson:"network_name,omitzero,omitempty"`

	PublicSubnetIDs   []string `json:"public_subnet_ids,omitzero" mapstructure:"public_subnet_ids" temporaljson:"public_subnet_ids,omitzero,omitempty"`
	PublicSubnetNames []string `json:"public_subnet_names,omitzero" mapstructure:"public_subnet_names" temporaljson:"public_subnet_names,omitzero,omitempty"`

	PrivateSubnetIDs   []string `json:"private_subnet_ids,omitzero" mapstructure:"private_subnet_ids" temporaljson:"private_subnet_ids,omitzero,omitempty"`
	PrivateSubnetNames []string `json:"private_subnet_names,omitzero" mapstructure:"private_subnet_names" temporaljson:"private_subnet_names,omitzero,omitempty"`

	KeyVaultID   string `json:"key_vault_id,omitzero" mapstructure:"key_vault_id" temporaljson:"key_vault_id,omitzero,omitempty"`
	KeyVaultName string `json:"key_vault_name,omitzero" mapstructure:"key_vault_name" temporaljson:"key_vault_name,omitzero,omitempty"`
}

func (a *AzureStackOutputs) ProvisionRoleID() (string, error)   { return "", nil }
func (a *AzureStackOutputs) DeprovisionRoleID() (string, error) { return "", nil }
func (a *AzureStackOutputs) MaintenanceRoleID() (string, error) { return "", nil }

func (a *AzureStackOutputs) CustomRoleID(_ string) (string, error) {
	return "", fmt.Errorf("not supported on azure")
}

func (a *AzureStackOutputs) BreakGlassRoleID(_ string) (string, error) {
	return "", fmt.Errorf("not supported on azure")
}

func (a *AzureStackOutputs) CustomRoles() (map[string]string, error) {
	return nil, nil
}

func (a *AzureStackOutputs) BreakGlassRoles() (map[string]string, error) {
	return nil, nil
}

func (a *AzureStackOutputs) InstallInputValues() (map[string]string, error) {
	return nil, nil
}

type GCPStackOutputs struct {
	ProjectID                 string            `json:"project_id,omitzero" mapstructure:"project_id" temporaljson:"project_id,omitzero,omitempty"`
	Region                    string            `json:"region,omitzero" mapstructure:"region" temporaljson:"region,omitzero,omitempty"`
	NetworkName               string            `json:"network_name,omitzero" mapstructure:"network_name" temporaljson:"network_name,omitzero,omitempty"`
	NetworkID                 string            `json:"network_id,omitzero" mapstructure:"network_id" temporaljson:"network_id,omitzero,omitempty"`
	PublicSubnetName          string            `json:"public_subnet_name,omitzero" mapstructure:"public_subnet_name" temporaljson:"public_subnet_name,omitzero,omitempty"`
	PrivateSubnetName         string            `json:"private_subnet_name,omitzero" mapstructure:"private_subnet_name" temporaljson:"private_subnet_name,omitzero,omitempty"`
	RunnerSubnetName          string            `json:"runner_subnet_name,omitzero" mapstructure:"runner_subnet_name" temporaljson:"runner_subnet_name,omitzero,omitempty"`
	RunnerServiceAccountEmail string            `json:"runner_service_account_email,omitzero" mapstructure:"runner_service_account_email" temporaljson:"runner_service_account_email,omitzero,omitempty"`
	ProvisionSAEmail          string            `json:"provision_sa_email,omitzero" mapstructure:"provision_sa_email" temporaljson:"provision_sa_email,omitzero,omitempty"`
	MaintenanceSAEmail        string            `json:"maintenance_sa_email,omitzero" mapstructure:"maintenance_sa_email" temporaljson:"maintenance_sa_email,omitzero,omitempty"`
	DeprovisionSAEmail        string            `json:"deprovision_sa_email,omitzero" mapstructure:"deprovision_sa_email" temporaljson:"deprovision_sa_email,omitzero,omitempty"`
	BreakGlassSAEmails        map[string]string `json:"break_glass_sa_emails,omitzero" mapstructure:"break_glass_sa_emails" temporaljson:"break_glass_sa_emails,omitzero,omitempty"`
	CustomSAEmails            map[string]string `json:"custom_sa_emails,omitzero" mapstructure:"custom_sa_emails" temporaljson:"custom_sa_emails,omitzero,omitempty"`
	InstallInputs             map[string]string `json:"install_inputs,omitzero" mapstructure:"install_inputs" temporaljson:"install_inputs,omitzero,omitempty"`
}

func (a *GCPStackOutputs) ProvisionRoleID() (string, error)   { return a.ProvisionSAEmail, nil }
func (a *GCPStackOutputs) DeprovisionRoleID() (string, error) { return a.DeprovisionSAEmail, nil }
func (a *GCPStackOutputs) MaintenanceRoleID() (string, error) { return a.MaintenanceSAEmail, nil }

func (a *GCPStackOutputs) CustomRoleID(name string) (string, error) {
	email, ok := a.CustomSAEmails[name]
	if !ok {
		return "", fmt.Errorf("custom service account %q does not exist in stack outputs", name)
	}
	return email, nil
}

func (a *GCPStackOutputs) BreakGlassRoleID(name string) (string, error) {
	email, ok := a.BreakGlassSAEmails[name]
	if !ok {
		return "", fmt.Errorf("break glass service account %q does not exist in stack outputs", name)
	}
	return email, nil
}

func (a *GCPStackOutputs) CustomRoles() (map[string]string, error) {
	return a.BreakGlassSAEmails, nil
}

func (a *GCPStackOutputs) BreakGlassRoles() (map[string]string, error) {
	return a.CustomSAEmails, nil
}

func (a *GCPStackOutputs) InstallInputValues() (map[string]string, error) {
	return a.InstallInputs, nil
}

func (a *InstallStackOutputs) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallStackOutputs{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (a *InstallStackOutputs) AfterQuery(tx *gorm.DB) error {
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
	}

	// decode into content map[string]interface
	decoderConfig.Result = &a.DataContents
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create gcp decoder")
	}
	if err := decoder.Decode(a.Data); err != nil {
		return errors.Wrap(err, "unable to parse gcp outputs")
	}

	// detect cloud platform from output keys
	// ja/sk/am: what have we become
	if _, isGCP := a.Data["runner_service_account_email"]; isGCP {
		a.GCPStackOutputs = &GCPStackOutputs{}
		decoderConfig.Result = a.GCPStackOutputs
	} else if _, isAzure := a.Data["resource_group_id"]; isAzure {
		a.AzureStackOutputs = &AzureStackOutputs{}
		decoderConfig.Result = a.AzureStackOutputs
	} else {
		a.AWSStackOutputs = &AWSStackOutputs{}
		decoderConfig.Result = a.AWSStackOutputs
	}

	// decode into cloud stack output
	decoder, err = mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create gcp decoder")
	}
	if err := decoder.Decode(a.Data); err != nil {
		return errors.Wrap(err, "unable to parse gcp outputs")
	}

	return nil
}

func (a *InstallStackOutputs) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppCfgID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
