package access

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func ExecutionConfigTarget(ctx context.Context, db *gorm.DB, target permissions.Target) (permissions.Target, error) {
	runner, err := cctx.RunnerFromContext(ctx)
	if err != nil {
		return target, nil
	}
	resources, err := Runners.In(db, target.OrgID)
	if err != nil {
		return target, err
	}
	owner, err := resources.Resolve(ctx, runner.ID)
	if err != nil {
		return target, err
	}
	if owner.InstallID != "" {
		if owner.AppID != target.AppID {
			return target, gorm.ErrRecordNotFound
		}
		target.InstallID = owner.InstallID
	}
	return target, nil
}
