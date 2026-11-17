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

func MCPAll(ctx context.Context, requirements ...authz.Requirement) (*authz.AccessContext, error) {
	authority, err := authz.RequestAuthority(ctx, permissions.APICtl)
	if err != nil || authority == nil {
		return nil, err
	}
	access, err := authority.Access().Require(requirements...)
	if err != nil {
		return nil, gorm.ErrRecordNotFound
	}
	return access, nil
}

func MCPPoint(ctx context.Context, db *gorm.DB, set authz.ResourceSet, operation permissions.PermissionID, id string) (*authz.AccessContext, error) {
	authority, err := authz.RequestAuthority(ctx, permissions.APICtl)
	if err != nil || authority == nil {
		return nil, err
	}
	orgID, err := MCPRead(ctx)
	if err != nil {
		return nil, err
	}
	resources, err := set.In(db, orgID)
	if err != nil {
		return nil, err
	}
	target, err := resources.Resolve(ctx, id)
	if err != nil {
		return nil, err
	}
	return MCPAll(ctx, authz.Requirement{Operation: operation, Target: target})
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
