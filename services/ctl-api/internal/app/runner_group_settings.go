package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

const (
	DefaultAWSInstanceType = "t3a.medium"
)

// TODO(fd): use the consts
var (
	CommonRunnerGroupSettingsGroups         = [...]string{"operations", "sync"}
	DefaultOrgRunnerGroupSettingsGroups     = [...]string{"build", "sandbox", "runner"}
	DefaultInstallRunnerGroupSettingsGroups = [...]string{"deploys", "action", "sandbox"}
)

type RunnerGroupSettings struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_runner_group_settings,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"index:idx_app_name,unique" temporaljson:"org_id,omitzero,omitempty"`

	RunnerGroupID string `json:"runner_group_id,omitzero" gorm:"index:idx_runner_group_settings,unique" temporaljson:"runner_group_id,omitzero,omitempty"`

	// configuration for deploying the runner
	ContainerImageURL string `json:"container_image_url,omitzero" gorm:"default null;not null" temporaljson:"container_image_url,omitzero,omitempty"`
	ContainerImageTag string `json:"container_image_tag,omitzero" gorm:"default null;not null" temporaljson:"container_image_tag,omitzero,omitempty"`
	ExpectedVersion   string `json:"-" gorm:"-" temporaljson:"expected_version,omitzero,omitempty"`
	RunnerAPIURL      string `json:"runner_api_url,omitzero" gorm:"default null;not null" temporaljson:"runner_apiurl,omitzero,omitempty"`

	// configuration for managing the runner server side
	SandboxMode bool `json:"sandbox_mode,omitzero" temporaljson:"sandbox_mode,omitzero,omitempty"`

	// Various settings for the runner to handle internally
	HeartBeatTimeout           time.Duration `json:"heart_beat_timeout,omitzero" gorm:"default null;" swaggertype:"primitive,integer" temporaljson:"heart_beat_timeout,omitzero,omitempty"`
	OTELCollectorConfiguration string        `json:"otel_collector_config,omitzero" gorm:"default null;not null" temporaljson:"otel_collector_configuration,omitzero,omitempty"`

	EnableSentry  bool           `json:"enable_sentry,omitzero" temporaljson:"enable_sentry,omitzero,omitempty"`
	EnableMetrics bool           `json:"enable_metrics,omitzero" temporaljson:"enable_metrics,omitzero,omitempty"`
	EnableLogging bool           `json:"enable_logging,omitzero" temporaljson:"enable_logging,omitzero,omitempty"`
	LoggingLevel  string         `json:"logging_level,omitzero" temporaljson:"logging_level,omitzero,omitempty"`
	Groups        pq.StringArray `json:"groups,omitzero" gorm:"type:text[];default:'{}'" swaggertype:"array,string" temporaljson:"groups,omitzero,omitempty"` // the job loop groups the runner should poll for

	// Metadata is used as both log and metric tags/attributes in the runner when emitting data
	Metadata pgtype.Hstore `json:"metadata,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"metadata,omitzero,omitempty"`

	// org runner specifics
	OrgAWSIAMRoleARN         string `json:"org_aws_iam_role_arn,omitzero" temporaljson:"org_awsiam_role_arn,omitzero,omitempty"`
	OrgK8sServiceAccountName string `json:"org_k8s_service_account_name,omitzero" temporaljson:"org_k_8_s_service_account_name,omitzero,omitempty"`

	// aws runner specifics runner-v2
	AWSInstanceType            string        `json:"aws_instance_type,omitzero" temporaljson:"aws_instance_type,omitzero,omitempty"`
	AWSMaxInstanceLifetime     int           `json:"aws_max_instance_lifetime" gorm:"not null;default:604800;check:aws_max_instance_lifetime >= 86400 AND aws_max_instance_lifetime <= 31536000" swaggertype:"primitive,integer" temporaljson:"aws_max_instance_lifetime"` // Default: 7 days
	AWSCloudformationStackType string        `json:"aws_cloudformation_stack_type,omitzero" temporaljson:"aws_cloudformation_stack_type,omitzero,omitempty"`
	AWSTags                    pgtype.Hstore `json:"aws_tags,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"aws_tags,omitzero,omitempty"`
	LocalAWSIAMRoleARN         string        `json:"local_aws_iam_role_arn,omitzero" temporaljson:"local_awsiam_role_arn,omitzero,omitempty"`

	// azure runner specifics

	// platform variable for use in the runner
	Platform CloudPlatform `json:"platform" temporaljson:"-" gorm:"-" swaggertype:"string"`
}

func (i *RunnerGroupSettings) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &RunnerGroupSettings{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (i *RunnerGroupSettings) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.CustomViewName(db, &RunnerGroupSettings{}, "settings_v1"),
			SQL:           viewsql.RunnerSettingsV1,
			AlwaysReapply: true, // necessary for this view to be recreated
		},
		{
			Name:          views.CustomViewName(db, &RunnerGroupSettings{}, "wide_v1"),
			SQL:           viewsql.RunnerWideV1,
			AlwaysReapply: true, // necessary for this view to be recreated
		},
	}
}

func (r *RunnerGroupSettings) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerGroupSettingsID()
		r.Metadata["runner_group.id"] = generics.ToPtr(r.ID)
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (r *RunnerGroupSettings) AfterQuery(tx *gorm.DB) error {
	r.ExpectedVersion = r.ContainerImageTag

	if r.AWSInstanceType != "" && r.AWSCloudformationStackType != "" {
		r.Platform = CloudPlatformAWS
	}
	return nil
}
