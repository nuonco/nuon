package userorgs

import (
	"context"
	"errors"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"gorm.io/gorm"
)

func (m *middleware) validate(ctx context.Context, orgID, subject string) error {
	var userOrg app.UserOrg
	res := m.db.
		WithContext(ctx).
		Where(&app.UserOrg{
			OrgID:  orgID,
			UserID: subject,
		}).
		First(&userOrg)
	if res.Error == nil {
		return nil
	}

	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return stderr.ErrUser{
			Err:         fmt.Errorf("user does not have access to org"),
			Description: "User does not have access to org",
		}
	}

	return res.Error
}
