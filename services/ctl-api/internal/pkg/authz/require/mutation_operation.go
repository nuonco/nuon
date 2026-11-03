package require

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type MutationContract struct {
	Permission          permissions.PermissionID
	Target              ReadTarget
	SubjectAccountParam string
}

type MutationAccess struct {
	db               *gorm.DB
	target           permissions.ResolvedTarget
	subjectAccountID string
	permission       permissions.PermissionID
}

type mutationAccessKey struct{}

func MutationAccessFromContext(ctx context.Context) (*MutationAccess, error) {
	access, ok := customerContext(ctx).Value(mutationAccessKey{}).(*MutationAccess)
	if !ok || access == nil {
		return nil, fmt.Errorf("authorized mutation context is required")
	}
	return access, nil
}

func (a *MutationAccess) DB() *gorm.DB {
	return a.db.Session(&gorm.Session{})
}

func (a *MutationAccess) Target() permissions.ResolvedTarget {
	return a.target
}

func (a *MutationAccess) SubjectAccountID() string {
	return a.subjectAccountID
}

func (a *MutationAccess) AcceptWorkflow(wf *app.Workflow, expiresAt time.Time) error {
	decision, err := authz.New(authz.Params{DB: a.db}).AcceptWorkflow(a.db.Statement.Context, authz.MutationRequest{
		Permission: a.permission, Target: a.target.Scope,
	}, wf, expiresAt)
	if err != nil {
		return err
	}
	if !decision.Allowed {
		return customerDenied(a.target.Kind)
	}
	return nil
}

func (c MutationContract) validate() error {
	switch c.Permission {
	case permissions.OrgWrite, permissions.AppWrite, permissions.AppDelete, permissions.InstallWrite, permissions.InstallDelete:
	default:
		return fmt.Errorf("unsupported mutation permission %q", c.Permission)
	}
	if c.Target.Name == "" || c.Target.Parent != "" || !c.Permission.AppliesTo(c.Target.Kind) {
		return fmt.Errorf("invalid mutation target")
	}
	if c.SubjectAccountParam != "" && (c.Permission != permissions.OrgWrite || c.SubjectAccountParam == c.Target.Name) {
		return fmt.Errorf("invalid mutation subject")
	}
	return nil
}

func MutationOperation[T any](ctx context.Context, db *gorm.DB, contract MutationContract, ref, subjectAccountID string, mutate func(context.Context) (T, error)) (T, error) {
	var result T
	if err := contract.validate(); err != nil {
		return result, err
	}
	if contract.SubjectAccountParam != "" && subjectAccountID == "" {
		return result, fmt.Errorf("mutation subject is required")
	}
	ctx = customerContext(ctx)
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if contract.Target.Kind == permissions.KindOrg {
		ref = orgID
	}
	decision, err := authz.New(authz.Params{DB: db}).MutateCustomer(ctx, authz.MutationRequest{
		Permission: contract.Permission, Target: permissions.Scope{Kind: contract.Target.Kind, OrgID: orgID},
		TargetReference: ref, SubjectAccountID: subjectAccountID,
	}, func(tx *gorm.DB, target permissions.ResolvedTarget) error {
		access := &MutationAccess{target: target, subjectAccountID: subjectAccountID, permission: contract.Permission}
		opCtx := context.WithValue(tx.Statement.Context, mutationAccessKey{}, access)
		access.db = tx.WithContext(opCtx)
		var err error
		result, err = mutate(opCtx)
		return err
	})
	if err == nil && !decision.Allowed {
		err = customerDenied(contract.Target.Kind)
	}
	if err != nil {
		var zero T
		return zero, err
	}
	return result, nil
}

type Response struct {
	Status int
	Body   any
}

type PreparedMutation struct {
	SubjectAccountID string
	Mutate           func(*gin.Context) (Response, error)
}

type MutationRoutes struct {
	router   *gin.RouterGroup
	db       *gorm.DB
	contract MutationContract
}

func MutationRoute(router gin.IRouter, db *gorm.DB, contract MutationContract) *MutationRoutes {
	if err := contract.validate(); err != nil {
		panic(err)
	}
	return &MutationRoutes{router: CustomerRoute(router), db: db, contract: contract}
}

func (r *MutationRoutes) Handle(method, path string, handler func(*gin.Context) (Response, error)) {
	r.HandlePrepared(method, path, func(ctx *gin.Context) (PreparedMutation, error) {
		return PreparedMutation{SubjectAccountID: ctx.Param(r.contract.SubjectAccountParam), Mutate: handler}, nil
	})
}

func (r *MutationRoutes) HandlePrepared(method, path string, prepare func(*gin.Context) (PreparedMutation, error)) {
	r.router.Handle(method, path, func(ctx *gin.Context) {
		request := ctx.Request
		defer func() { ctx.Request = request }()
		prepared, err := prepare(ctx)
		if err != nil {
			ctx.Error(err)
			ctx.Abort()
			return
		}
		response, err := MutationOperation(customerContext(ctx), r.db, r.contract, ctx.Param(r.contract.Target.Name), prepared.SubjectAccountID, func(opCtx context.Context) (Response, error) {
			ctx.Request = ctx.Request.WithContext(opCtx)
			access, err := MutationAccessFromContext(opCtx)
			if err != nil {
				return Response{}, err
			}
			for i := range ctx.Params {
				if ctx.Params[i].Key == r.contract.Target.Name {
					ctx.Params[i].Value = access.target.ID
				}
			}
			return prepared.Mutate(ctx)
		})
		if err != nil {
			ctx.Error(err)
			ctx.Abort()
			return
		}
		if response.Body == nil {
			ctx.Status(response.Status)
			return
		}
		ctx.JSON(response.Status, response.Body)
	})
}
