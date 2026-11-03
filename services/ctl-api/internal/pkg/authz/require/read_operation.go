package require

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type ReadTarget struct {
	Name   string
	Kind   permissions.ResourceKind
	Parent string
}

type ReadContract struct {
	Permission permissions.PermissionID
	Targets    []ReadTarget
	Collection permissions.ResourceKind
}

type ReadAccess struct {
	reader   *authz.CustomerReader
	db       *gorm.DB
	contract ReadContract
	targets  map[string]permissions.ResolvedTarget
}

type readAccessKey struct{}

func ReadAccessFromContext(ctx context.Context) (*ReadAccess, error) {
	access, ok := customerContext(ctx).Value(readAccessKey{}).(*ReadAccess)
	if !ok || access == nil {
		return nil, fmt.Errorf("authorized read context is required")
	}
	return access, nil
}

func (a *ReadAccess) DB() *gorm.DB {
	return a.db.Session(&gorm.Session{})
}

func (a *ReadAccess) OrgID() string {
	return a.reader.OrgID()
}

func (a *ReadAccess) Target(name string) (permissions.ResolvedTarget, error) {
	target, ok := a.targets[name]
	if !ok {
		return permissions.ResolvedTarget{}, fmt.Errorf("authorized target %q is required", name)
	}
	return target, nil
}

func (a *ReadAccess) Query() (*gorm.DB, error) {
	var query *gorm.DB
	switch a.contract.Collection {
	case permissions.KindApp:
		query = a.reader.Apps()
	case permissions.KindInstall:
		query = a.reader.Installs()
	case permissions.KindComponent:
		query = a.reader.Components()
	case permissions.KindComponentBuild:
		query = a.reader.Builds()
	default:
		return nil, fmt.Errorf("authorized collection is required")
	}
	return query.WithContext(a.db.Statement.Context), nil
}

func (c ReadContract) validate() error {
	if len(c.Targets) == 0 && c.Collection == "" {
		return fmt.Errorf("read contract requires a target or collection")
	}
	switch c.Permission {
	case permissions.OrgRead, permissions.OrgWrite, permissions.AppRead, permissions.InstallRead,
		permissions.RunnerRead, permissions.RunnerJobRead, permissions.LogStreamRead, permissions.TelemetryRead:
	default:
		return fmt.Errorf("unsupported read permission %q", c.Permission)
	}
	if c.Collection != "" {
		expected := permissions.AppRead
		switch c.Collection {
		case permissions.KindApp, permissions.KindComponent, permissions.KindComponentBuild:
		case permissions.KindInstall:
			expected = permissions.InstallRead
		default:
			return fmt.Errorf("unsupported read collection %q", c.Collection)
		}
		if c.Permission != expected {
			return fmt.Errorf("collection %q requires %q", c.Collection, expected)
		}
	}
	seen := make(map[string]permissions.ResourceKind, len(c.Targets))
	for _, target := range c.Targets {
		if target.Name == "" || seen[target.Name] != "" || !c.Permission.AppliesTo(target.Kind) {
			return fmt.Errorf("invalid declared read target %q", target.Name)
		}
		if target.Parent != "" {
			parent := seen[target.Parent]
			if !((parent == permissions.KindApp && (target.Kind == permissions.KindInstall || target.Kind == permissions.KindComponent)) ||
				(parent == permissions.KindComponent && (target.Kind == permissions.KindComponentBuild || target.Kind == permissions.KindComponentConfig))) {
				return fmt.Errorf("invalid declared target parent %q", target.Parent)
			}
		}
		seen[target.Name] = target.Kind
	}
	return nil
}

func ReadOperation[T any](ctx context.Context, db *gorm.DB, contract ReadContract, refs map[string]string, read func(context.Context) (T, error)) (T, error) {
	var result T
	if err := contract.validate(); err != nil {
		return result, err
	}
	ctx = customerContext(ctx)
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return result, err
	}
	contract.Targets = slices.Clone(contract.Targets)
	err = authz.New(authz.Params{DB: db}).ReadCustomer(ctx, orgID, func(reader *authz.CustomerReader) error {
		access := &ReadAccess{reader: reader, contract: contract, targets: make(map[string]permissions.ResolvedTarget, len(contract.Targets))}
		for _, requirement := range contract.Targets {
			var parent *permissions.ResolvedTarget
			if requirement.Parent != "" {
				value := access.targets[requirement.Parent]
				parent = &value
			}
			ref := refs[requirement.Name]
			if requirement.Kind == permissions.KindOrg {
				ref = orgID
			}
			target, err := reader.Resolve(contract.Permission, requirement.Kind, ref, parent)
			if err != nil {
				return err
			}
			access.targets[requirement.Name] = target
		}
		opCtx := context.WithValue(reader.DB().Statement.Context, readAccessKey{}, access)
		access.db = reader.DB().WithContext(opCtx)
		var err error
		result, err = read(opCtx)
		return err
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return result, nil
}

type ReadRoutes struct {
	router   *gin.RouterGroup
	db       *gorm.DB
	contract ReadContract
}

func ReadRoute(router gin.IRouter, db *gorm.DB, contract ReadContract) *ReadRoutes {
	if err := contract.validate(); err != nil {
		panic(err)
	}
	contract.Targets = slices.Clone(contract.Targets)
	return &ReadRoutes{router: CustomerRoute(router), db: db, contract: contract}
}

func (r *ReadRoutes) GET(path string, handler func(*gin.Context) (any, error)) {
	r.router.GET(path, func(ctx *gin.Context) {
		request := ctx.Request
		defer func() { ctx.Request = request }()
		refs := make(map[string]string, len(r.contract.Targets))
		for _, target := range r.contract.Targets {
			refs[target.Name] = ctx.Param(target.Name)
		}
		result, err := ReadOperation(customerContext(ctx), r.db, r.contract, refs, func(opCtx context.Context) (any, error) {
			ctx.Request = ctx.Request.WithContext(opCtx)
			access, err := ReadAccessFromContext(opCtx)
			if err != nil {
				return nil, err
			}
			for i := range ctx.Params {
				if target, ok := access.targets[ctx.Params[i].Key]; ok {
					ctx.Params[i].Value = target.ID
				}
			}
			return handler(ctx)
		})
		if err != nil {
			ctx.Error(err)
			ctx.Abort()
			return
		}
		ctx.JSON(http.StatusOK, result)
	})
}
