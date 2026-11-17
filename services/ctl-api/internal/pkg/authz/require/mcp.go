package require

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func MCPRead(ctx context.Context) (string, error) {
	return mcpRequest(ctx, false)
}

func MCPWrite(ctx context.Context) (string, error) {
	return mcpRequest(ctx, true)
}

func mcpRequest(ctx context.Context, write bool) (string, error) {
	orgID, err := Org(ctx)
	if err != nil {
		return "", err
	}
	authority, err := authz.RequestAuthority(ctx, permissions.APICtl)
	if err != nil {
		return "", err
	}
	if authority == nil {
		if write {
			return Write(ctx)
		}
		return Read(ctx)
	}
	if write && keys.TokenRoleFromContext(ctx) == string(app.RoleTypeOrgReadOnly) {
		return "", fmt.Errorf("this access token is read-only")
	}
	return orgID, nil
}

func MCPOperation(ctx context.Context, operation permissions.PermissionID, target permissions.Target) error {
	authority, err := authz.RequestAuthority(ctx, permissions.APICtl)
	if err != nil || authority == nil {
		return err
	}
	if err := authority.Require(authz.Requirement{Operation: operation, Target: target}); err != nil {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func MCPPoint(ctx context.Context, db *gorm.DB, set authz.ResourceSet, operation permissions.PermissionID, id string) error {
	authority, err := authz.RequestAuthority(ctx, permissions.APICtl)
	if err != nil || authority == nil {
		return err
	}
	orgID, err := MCPRead(ctx)
	if err != nil {
		return err
	}
	resources, err := set.In(db, orgID)
	if err != nil {
		return err
	}
	target, err := resources.Resolve(ctx, id)
	if err != nil {
		return err
	}
	return MCPOperation(ctx, operation, target)
}

func MCPCollection(ctx context.Context, source *gorm.DB, set authz.ResourceSet, operation permissions.PermissionID) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		authority, err := authz.RequestAuthority(ctx, permissions.APICtl)
		if err != nil {
			db.AddError(err)
			return db
		}
		if authority == nil {
			return db
		}
		orgID, err := MCPRead(ctx)
		if err != nil {
			db.AddError(err)
			return db
		}
		resources, err := set.In(source, orgID)
		if err != nil {
			db.AddError(err)
			return db
		}
		selection, err := authority.Select(ctx, operation, resources)
		if err != nil {
			db.AddError(err)
			return db
		}
		return selection.Scope()(db)
	}
}
