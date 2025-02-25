package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type RunnerGroupSettings struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string  `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account `json:"-"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_runner_group_settings,unique"`

	OrgID string `json:"org_id" gorm:"index:idx_app_name,unique"`

	RunnerGroupID string `json:"runner_group_id" gorm:"index:idx_runner_group_settings,unique"`

	// configuration for deploying the runner
	ContainerImageURL string `json:"container_image_url"  gorm:"default null;not null"`
	ContainerImageTag string `json:"container_image_tag"  gorm:"default null;not null"`
	ExpectedVersion   string `json:"-" temporaljson:"expected_version" gorm:"-"`
	RunnerAPIURL      string `json:"runner_api_url" gorm:"default null;not null"`

	// configuration for managing the runner server side
	SandboxMode bool `json:"sandbox_mode"`

	// Various settings for the runner to handle internally
	HeartBeatTimeout           time.Duration `json:"heart_beat_timeout" gorm:"default null;" swaggertype:"primitive,integer"`
	OTELCollectorConfiguration string        `json:"otel_collector_config" gorm:"default null;not null"`

	EnableSentry  bool   `json:"enable_sentry"`
	EnableMetrics bool   `json:"enable_metrics"`
	EnableLogging bool   `json:"enable_logging"`
	LoggingLevel  string `json:"logging_level"`

	// Metadata is used as both log and metric tags/attributes in the runner when emitting data
	Metadata pgtype.Hstore `json:"" gorm:"type:hstore" swaggertype:"object,string"`

	// specific configuration for cloud specific runners, such as AWS or Azure
	AWSIAMRoleARN         string `json:"aws_iam_role_arn"`
	K8sServiceAccountName string `json:"k8s_service_account_name"`
}

func (i *RunnerGroupSettings) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{

		{
			Name: views.CustomViewName(db, &RunnerGroupSettings{}, "settings_v1"),
			SQL:  viewsql.RunnerSettingsV1,
		},
		{
			Name: views.CustomViewName(db, &RunnerGroupSettings{}, "wide_v1"),
			SQL:  viewsql.RunnerWideV1,
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
	return nil
}
