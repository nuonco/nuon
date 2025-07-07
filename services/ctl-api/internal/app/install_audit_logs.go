package app

import (
	"time"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type InstallAuditLog struct {
	InstallID string    `json:"install_id,omitzero" gorm:"->;-:migration"`
	Type      string    `json:"type,omitzero" gorm:"->;-:migration"`
	TimeStamp time.Time `json:"time_stamp,omitzero" gorm:"->;-:migration"`
	LogLine   string    `json:"log_line,omitzero" gorm:"->;-:migration"`
}

func (*InstallAuditLog) UseView() bool {
	return true
}

func (*InstallAuditLog) ViewVersion() string {
	return "v1"
}

func (i *InstallAuditLog) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name: views.DefaultViewName(db, &InstallAuditLog{}, 1),
			SQL:  viewsql.InstallAuditLogsViewV1,
		},
	}
}

func (m InstallAuditLog) GetTableOptions() (string, bool) {
	return "", false
}

func (r InstallAuditLog) MigrateDB(db *gorm.DB) *gorm.DB {
	return db
}
