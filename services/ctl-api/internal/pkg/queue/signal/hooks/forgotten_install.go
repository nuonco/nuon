package hooks

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func eventInstallOwnerID(event signal.SignalPhaseEvent) string {
	if event.OwnerType == "installs" && event.OwnerID != "" {
		return event.OwnerID
	}
	if event.InstallID != nil && *event.InstallID != "" {
		return *event.InstallID
	}
	return ""
}

// isForgottenInstall reports whether installID is soft-deleted (forgotten) or
// already gone. Empty id is not forgotten.
func isForgottenInstall(ctx context.Context, db *gorm.DB, installID string) (bool, error) {
	if installID == "" || db == nil {
		return false, nil
	}

	var install app.Install
	err := db.WithContext(ctx).
		Unscoped().
		Select("id", "deleted_at").
		Where("id = ?", installID).
		First(&install).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return install.DeletedAt != 0, nil
}
