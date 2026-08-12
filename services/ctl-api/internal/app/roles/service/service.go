package service

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type Params struct {
	fx.In

	V  *validator.Validate
	DB *gorm.DB `name:"psql"`
	L  *zap.Logger
}

type service struct {
	v  *validator.Validate
	db *gorm.DB
	l  *zap.Logger
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		v:  params.V,
		db: params.DB,
		l:  params.L,
	}
}

func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	roles := ge.Group("/v1/roles")
	{
		roles.GET("", s.ListRoles)
		roles.POST("", s.CreateRole)
		roles.GET("/:role_id", s.GetRole)
		roles.PATCH("/:role_id", s.UpdateRole)
		roles.DELETE("/:role_id", s.DeleteRole)
	}

	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) requireOrgAdmin(ctx *gin.Context) (*app.Org, error) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		return nil, err
	}

	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		return nil, err
	}

	for _, role := range acct.Roles {
		if role.RoleType == app.RoleTypeOrgAdmin && role.OrgID.ValueString() == org.ID {
			return org, nil
		}
	}

	return nil, stderr.ErrAuthorization{
		Err:         fmt.Errorf("only org admins can manage roles"),
		Description: "only org admins can manage roles",
	}
}

// requireAvailableTitle rejects a title already used by any role in the org,
// managed ones included: titles are what the role pickers display, so a
// duplicate leaves users unable to tell two roles apart. The comparison is
// case-insensitive, since "Release manager" and "release manager" are not
// distinguishable in a dropdown either. Pass excludeRoleID when updating so a
// role does not collide with itself.
func (s *service) requireAvailableTitle(ctx *gin.Context, orgID, title, excludeRoleID string) error {
	query := s.db.WithContext(ctx).
		Model(&app.Role{}).
		Where(app.Role{OrgID: generics.NewNullString(orgID)}).
		Where("lower(title) = lower(?)", title)
	if excludeRoleID != "" {
		query = query.Where("id != ?", excludeRoleID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return fmt.Errorf("unable to check for an existing role named %q: %w", title, err)
	}
	if count > 0 {
		return userErr(fmt.Errorf("a role named %q already exists in this org", title))
	}

	return nil
}

func (s *service) getOrgRole(ctx *gin.Context, orgID, roleID string) (*app.Role, error) {
	var role app.Role
	res := s.db.WithContext(ctx).
		Preload("Policies").
		Where(app.Role{ID: roleID, OrgID: generics.NewNullString(orgID)}).
		First(&role)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, stderr.ErrNotFound{
				Err:         fmt.Errorf("role %q not found", roleID),
				Description: "role not found",
			}
		}
		return nil, fmt.Errorf("unable to find role: %w", res.Error)
	}

	return &role, nil
}

// requireCustomRole loads the role and refuses managed roles: those are owned
// and kept in sync by the authz reconciler, and this API never competes with
// it.
func (s *service) requireCustomRole(ctx *gin.Context, orgID, roleID string) (*app.Role, error) {
	role, err := s.getOrgRole(ctx, orgID, roleID)
	if err != nil {
		return nil, err
	}

	if role.Managed {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("role %q is managed", roleID),
			Description: "managed roles cannot be edited or deleted",
		}
	}

	return role, nil
}
