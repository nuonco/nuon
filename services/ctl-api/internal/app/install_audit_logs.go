package app

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/viewsql"
)

type InstallAuditLog struct {
	InstallID string    `json:"install_id,omitzero" gorm:"->;-:migration"`
	Type      string    `json:"type,omitzero" gorm:"->;-:migration"`
	TimeStamp time.Time `json:"time_stamp,omitzero" gorm:"->;-:migration"`
	LogLine   string    `json:"log_line,omitzero" gorm:"->;-:migration"`
}

const installAuditLogCurrentVersion = 2

func (*InstallAuditLog) UseView() bool {
	return true
}

func (*InstallAuditLog) CurrentViewVersion() int {
	return installAuditLogCurrentVersion
}

func (i *InstallAuditLog) ViewVersion() string {
	return fmt.Sprintf("v%d", i.CurrentViewVersion())
}

func (i *InstallAuditLog) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name: views.DefaultViewName(db, &InstallAuditLog{}, 1),
			SQL:  viewsql.InstallAuditLogsViewV1,
		},
		{
			Name: views.DefaultViewName(db, &InstallAuditLog{}, 2),
			SQL:  viewsql.InstallAuditLogsViewV2,
		},
	}
}

func (m InstallAuditLog) GetTableOptions() (string, bool) {
	return "", false
}

func (r InstallAuditLog) MigrateDB(db *gorm.DB) *gorm.DB {
	return db
}
