package authz

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type CustomerReader struct {
	db         *gorm.DB
	orgID      string
	credential cctx.VerifiedCredential
	grants     []permissions.ScopedPermissions
	ceiling    []permissions.ScopedPermissions
}

func (c *Client) ReadCustomerResources(ctx context.Context, orgID string, read func(tx, apps, installs *gorm.DB) error) error {
	return c.ReadCustomer(ctx, orgID, func(reader *CustomerReader) error {
		return read(reader.DB(), reader.Apps(), reader.Installs())
	})
}

func (c *Client) ReadCustomer(ctx context.Context, orgID string, read func(*CustomerReader) error) error {
	credential, ok := cctx.VerifiedCredentialFromContext(ctx)
	limits := permissions.CredentialLimits{Profile: credential.Profile, API: credential.Audience, Ceiling: credential.Ceiling}
	denied := stderr.ErrAuthorization{Err: fmt.Errorf("customer resource read denied"), Description: "access denied"}
	if !ok || !customerCredentialAllowed(credential) || limits.Validate() != nil ||
		(permissions.Scope{Kind: permissions.KindOrg, OrgID: orgID, ID: orgID}).ValidateShape() != nil {
		return denied
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ceiling, reason, err := loadCustomerCredential(tx, credential, orgID, time.Now())
		if err != nil {
			return err
		}
		if reason != ReasonAllowed {
			return denied
		}
		grants, err := loadScopedGrants(tx, credential.AccountID, orgID)
		if err != nil && !errors.Is(err, errNoScopedGrants) {
			return err
		}
		if err == nil && grants.API != credential.Audience {
			return denied
		}
		return read(&CustomerReader{db: tx, orgID: orgID, credential: credential, grants: grants.Grants, ceiling: ceiling})
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
}

func (r *CustomerReader) DB() *gorm.DB {
	return r.db.Session(&gorm.Session{})
}

func (r *CustomerReader) OrgID() string {
	return r.orgID
}

func (r *CustomerReader) Authorize(permission permissions.PermissionID, scope permissions.Scope) (permissions.ResolvedTarget, error) {
	var denied error = stderr.ErrNotFound{Err: gorm.ErrRecordNotFound, Description: "resource not found"}
	if scope.Kind == permissions.KindOrg {
		denied = stderr.ErrAuthorization{Err: fmt.Errorf("access denied"), Description: "access denied"}
	}
	if scope.OrgID != r.orgID || !permission.AppliesTo(scope.Kind) {
		return permissions.ResolvedTarget{}, denied
	}
	target, err := resolveTarget(r.DB(), scope)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return permissions.ResolvedTarget{}, denied
		}
		return permissions.ResolvedTarget{}, err
	}
	for _, ceiling := range [][]permissions.ScopedPermissions{r.credential.Ceiling, r.ceiling} {
		if !permissions.MatchPermissions(permissions.PermissionCheck{
			Permission: permission, Target: target, Grants: r.grants, Ceiling: ceiling,
		}).Matched {
			return permissions.ResolvedTarget{}, denied
		}
	}
	return target, nil
}

func (r *CustomerReader) Apps() *gorm.DB {
	query := r.DB().Model(&app.App{}).Where(app.App{OrgID: r.orgID})
	for _, source := range [][]permissions.ScopedPermissions{r.grants, r.credential.Ceiling, r.ceiling} {
		query = filterCustomerResources(query, r.DB(), source, r.orgID, permissions.AppRead, "apps")
	}
	return query
}

func (r *CustomerReader) Installs() *gorm.DB {
	query := r.DB().Model(&app.Install{}).Where(app.Install{OrgID: r.orgID})
	table := views.TableOrViewName(r.db, &app.Install{}, "")
	query = query.Where(table+".app_id IN (?)", r.DB().Model(&app.App{}).Select("id").Where(app.App{OrgID: r.orgID}))
	for _, source := range [][]permissions.ScopedPermissions{r.grants, r.credential.Ceiling, r.ceiling} {
		query = filterCustomerResources(query, r.DB(), source, r.orgID, permissions.InstallRead, table)
	}
	return query
}

func (r *CustomerReader) Components() *gorm.DB {
	return r.DB().Model(&app.Component{}).Where(app.Component{OrgID: r.orgID}).
		Where("components.app_id IN (?)", r.Apps().Select("apps.id"))
}

func (r *CustomerReader) Builds() *gorm.DB {
	configs := r.DB().Model(&app.ComponentConfigConnection{}).Scopes(scopes.WithDisableViews).Select("id").
		Where(app.ComponentConfigConnection{OrgID: r.orgID}).
		Where("component_id IN (?)", r.Components().Select("components.id")).
		Where("component_builds.component_id IS NULL OR component_builds.component_id = '' OR component_builds.component_id = component_config_connections.component_id")
	return r.DB().Model(&app.ComponentBuild{}).Where(app.ComponentBuild{OrgID: r.orgID}).
		Where("component_builds.component_config_connection_id IN (?)", configs)
}

func (r *CustomerReader) Resolve(permission permissions.PermissionID, kind permissions.ResourceKind, ref string, parent *permissions.ResolvedTarget) (permissions.ResolvedTarget, error) {
	scope, err := resolveReference(r.DB(), r.orgID, kind, ref, parent)
	if err != nil {
		return permissions.ResolvedTarget{}, err
	}
	return r.Authorize(permission, scope)
}

func resolveReference(db *gorm.DB, orgID string, kind permissions.ResourceKind, ref string, parent *permissions.ResolvedTarget) (permissions.Scope, error) {
	denied := stderr.ErrNotFound{Err: gorm.ErrRecordNotFound, Description: "resource not found"}
	if ref == "" {
		return permissions.Scope{}, denied
	}
	scope := permissions.Scope{Kind: kind, OrgID: orgID, ID: ref}
	query := db.Session(&gorm.Session{}).Scopes(scopes.WithDisableViews).Select("id").Where(map[string]any{"org_id": orgID})
	alias := false
	switch kind {
	case permissions.KindOrg:
		return scope, scope.ValidateShape()
	case permissions.KindApp:
		query, alias = query.Model(&app.App{}), true
	case permissions.KindInstall:
		query, alias = query.Model(&app.Install{}), true
	case permissions.KindComponent:
		query, alias = query.Model(&app.Component{}), true
	case permissions.KindComponentConfig:
		query = query.Model(&app.ComponentConfigConnection{})
	case permissions.KindComponentBuild:
		query = query.Model(&app.ComponentBuild{})
	default:
		return scope, scope.ValidateShape()
	}
	if parent != nil {
		switch {
		case parent.Kind == permissions.KindApp && (kind == permissions.KindComponent || kind == permissions.KindInstall):
			query = query.Where(map[string]any{"app_id": parent.ID})
		case parent.Kind == permissions.KindComponent && kind == permissions.KindComponentConfig:
			query = query.Where(map[string]any{"component_id": parent.ID})
		case parent.Kind == permissions.KindComponent && kind == permissions.KindComponentBuild:
			query = query.Where("component_config_connection_id IN (?)", db.Session(&gorm.Session{}).Model(&app.ComponentConfigConnection{}).
				Scopes(scopes.WithDisableViews).Select("id").Where(app.ComponentConfigConnection{OrgID: orgID, ComponentID: parent.ID}))
		default:
			return permissions.Scope{}, fmt.Errorf("unsupported target parent")
		}
	}
	var rows []struct{ ID string }
	if err := query.Session(&gorm.Session{}).Where(map[string]any{"id": ref}).Limit(1).Find(&rows).Error; err != nil {
		return permissions.Scope{}, err
	}
	if len(rows) == 0 && alias {
		if err := query.Session(&gorm.Session{}).Where(map[string]any{"name": ref}).Limit(2).Find(&rows).Error; err != nil {
			return permissions.Scope{}, err
		}
	}
	if len(rows) != 1 {
		return permissions.Scope{}, denied
	}
	scope.ID = rows[0].ID
	return scope, nil
}

func filterCustomerResources(query, db *gorm.DB, grants []permissions.ScopedPermissions, orgID string, permission permissions.PermissionID, table string) *gorm.DB {
	filter := db.Where("FALSE")
	for _, grant := range grants {
		if grant.Scope.OrgID != orgID {
			continue
		}
		if _, ok := grant.Permissions[permission]; !ok {
			continue
		}
		switch grant.Scope.Kind {
		case permissions.KindOrg:
			return query
		case permissions.KindApp:
			column := table + ".id"
			if permission == permissions.InstallRead {
				column = table + ".app_id"
			}
			filter = filter.Or(map[string]any{column: grant.Scope.ID})
		case permissions.KindInstall:
			if permission == permissions.InstallRead {
				filter = filter.Or(map[string]any{table + ".id": grant.Scope.ID})
			}
		}
	}
	return query.Where(filter)
}
