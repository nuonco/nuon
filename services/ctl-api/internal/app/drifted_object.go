package app

import (
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/bulk"
)

type DriftedObject struct {
	// These fields will be populated from the drifts_view
	TargetType         string  `json:"target_type,omitzero" gorm:"->;-:migration" temporaljson:"target_type,omitzero,omitempty"`
	TargetID           string  `json:"target_id,omitzero" gorm:"->;-:migration" temporaljson:"target_id,omitzero,omitempty"`
	InstallWorkflowID  string  `json:"install_workflow_id,omitzero" gorm:"->;-:migration" temporaljson:"install_workflow_id,omitzero,omitempty"`
	AppSandboxConfigID *string `json:"app_sandbox_config_id,omitzero" gorm:"->;-:migration" temporaljson:"app_sandbox_config_id,omitzero,omitempty"`
	ComponentBuildID   *string `json:"component_build_id,omitzero" gorm:"->;-:migration" temporaljson:"component_build_id,omitzero,omitempty"`
	InstallID          string  `json:"install_id,omitzero" gorm:"->;-:migration" temporaljson:"install_id,omitzero,omitempty"`
	OrgID              string  `json:"org_id,omitzero" gorm:"->;-:migration" temporaljson:"org_id,omitzero,omitempty"`
	InstallComponentID *string `json:"install_component_id,omitzero" gorm:"->;-:migration" temporaljson:"install_component_id,omitzero,omitempty"`
	InstallSandboxID   *string `json:"install_sandbox_id,omitzero" gorm:"->;-:migration" temporaljson:"install_sandbox_id,omitzero,omitempty"`
	ComponentName      string  `json:"component_name,omitzero" gorm:"->;-:migration" temporaljson:"component_name,omitzero,omitempty"`
}

func (d *DriftedObject) UseView() bool {
	return true
}

func (d *DriftedObject) ViewVersion() string {
	return "v1"
}

func (d *DriftedObject) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.DefaultViewName(db, &DriftedObject{}, 1),
			SQL:           viewsql.DriftsViewV1,
			AlwaysReapply: true,
		},
	}
}

func (d *DriftedObject) Indexes(db *gorm.DB) []migrations.Index {
	return nil
}

func (d *DriftedObject) BeforeCreate(tx *gorm.DB) error {

	return nil
}

func (d *DriftedObject) AfterQuery(tx *gorm.DB) error {
	return nil
}

func (d *DriftedObject) EventLoops() []bulk.EventLoop {
	return nil
}
