package service

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetOrgMembers
// @Summary				Get current org members and pending invites
// @Description			Returns current-org members and pending invites. Active members include all current-org role_bindings; role_type describes only their org-scoped role and is empty for scoped-only members. Role filters match any binding before pagination. Pending invites have no bindings yet.
// @Param					q			query	string	false	"search query to filter members by email or name"
// @Param					status		query	string	false	"comma-separated statuses: active and/or invited"
// @Param					role_type	query	string	false	"comma-separated role types"
// @Param					offset		query	int		false	"offset of results to return"	Default(0)
// @Param					limit		query	int		false	"limit of results to return"	Default(20)
// @Param					page		query	int		false	"page number of results to return"	Default(0)
// @Tags					orgs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.OrgMember
// @Router					/v1/orgs/current/members [GET]
func (s *service) GetOrgMembers(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	pagination := cctx.OffsetPaginationFromContext(ctx)
	if ctx.Query("limit") == "" && pagination != nil {
		pagination.Limit = 20
		cctx.SetOffPaginationGinCtx(ctx, *pagination)
	}

	members, err := s.getOrgMembers(ctx, org.ID, ctx.Query("q"), ctx.Query("status"), ctx.Query("role_type"))
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, members)
}

func (s *service) getOrgMembers(ctx *gin.Context, orgID, q, status, roleType string) ([]app.OrgMember, error) {
	acct, err := cctx.AccountFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get account from context: %w", err)
	}

	statuses, err := parseOrgMemberStatuses(status)
	if err != nil {
		return nil, stderr.ErrUser{
			Err:         err,
			Description: err.Error(),
		}
	}

	roleTypes := parseOrgMemberRoleTypes(roleType)

	membersQuery := s.db.WithContext(ctx).
		Model(&app.Account{}).
		Select(`
			accounts.id AS id,
			accounts.id AS account_id,
			CAST('' AS text) AS invite_id,
			accounts.email AS email,
			COALESCE(accounts.name, '') AS name,
			CAST('active' AS text) AS status,
			COALESCE(MAX(roles.role_type) FILTER (WHERE account_roles.scope_kind = 'org' AND account_roles.scope_id = ?), '') AS role_type,
			CAST(NULL AS jsonb) AS invite_scope,
			MIN(account_roles.created_at) AS joined_at,
			accounts.created_at AS created_at`, orgID).
		Joins("JOIN account_roles ON account_roles.account_id = accounts.id AND account_roles.org_id = ? AND account_roles.deleted_at = 0", orgID).
		Joins("JOIN roles ON roles.id = account_roles.role_id AND roles.deleted_at = 0").
		Where("accounts.account_type != ?", app.AccountTypeService).
		Group("accounts.id")

	invitesQuery := s.db.WithContext(ctx).
		Model(&app.OrgInvite{}).
		Select(`org_invites.id AS id,
			CAST('' AS text) AS account_id,
			org_invites.id AS invite_id,
			org_invites.email AS email,
			CAST('' AS text) AS name,
			CAST('invited' AS text) AS status,
			org_invites.role_type AS role_type,
			org_invites.scope AS invite_scope,
			CAST(NULL AS timestamp) AS joined_at,
			org_invites.created_at AS created_at`).
		Where(app.OrgInvite{OrgID: orgID, Status: app.OrgInviteStatusPending}).
		Where(`NOT EXISTS (
			SELECT 1 FROM accounts
			JOIN account_roles ON account_roles.account_id = accounts.id AND account_roles.org_id = org_invites.org_id AND account_roles.deleted_at = 0
			WHERE accounts.email = org_invites.email
			AND accounts.account_type != ?
			AND accounts.deleted_at = 0
		)`, app.AccountTypeService)

	if len(roleTypes) > 0 {
		membersQuery = membersQuery.Having("BOOL_OR(roles.role_type IN ?)", roleTypes)
		invitesQuery = invitesQuery.Where("org_invites.role_type IN ?", roleTypes)
	}
	unionQuery := s.db.WithContext(ctx).Raw("(?) UNION ALL (?)", membersQuery, invitesQuery)

	members := []app.OrgMember{}
	tx := s.db.WithContext(ctx).
		Table("(?) AS org_members", unionQuery).
		Scopes(scopes.WithOffsetPagination).
		Order("org_members.email ASC").
		Order("org_members.id ASC")

	if !strings.HasSuffix(acct.Email, "nuon.co") {
		tx = tx.Where("org_members.email NOT LIKE ?", "%nuon.co")
	}

	if q != "" {
		queryPattern := "%" + q + "%"
		tx = tx.Where("org_members.email ILIKE ? OR org_members.name ILIKE ?", queryPattern, queryPattern)
	}

	if len(statuses) > 0 {
		tx = tx.Where("org_members.status IN ?", statuses)
	}

	if err := tx.Find(&members).Error; err != nil {
		return nil, fmt.Errorf("unable to get org members %s: %w", orgID, err)
	}

	members, err = db.HandlePaginatedResponse(ctx, members)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	ids := make([]string, 0, len(members))
	for _, member := range members {
		if member.AccountID != "" {
			ids = append(ids, member.AccountID)
		}
	}
	bindings, err := authz.New(authz.Params{DB: s.db}).AccountRoleBindings(ctx, orgID, ids)
	if err != nil {
		return nil, err
	}
	for i := range members {
		members[i].RoleBindings = []app.RoleBinding{}
		if members[i].AccountID != "" {
			members[i].RoleBindings = bindings[members[i].AccountID]
		}
	}
	return members, nil
}

func parseOrgMemberStatuses(raw string) ([]app.OrgMemberStatus, error) {
	if raw == "" {
		return nil, nil
	}

	var statuses []app.OrgMemberStatus
	for _, part := range strings.Split(raw, ",") {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		status := app.OrgMemberStatus(p)
		switch status {
		case app.OrgMemberStatusActive, app.OrgMemberStatusInvited:
			statuses = append(statuses, status)
		default:
			return nil, fmt.Errorf("unrecognised status %q", p)
		}
	}

	return statuses, nil
}

func parseOrgMemberRoleTypes(raw string) []app.RoleType {
	if raw == "" {
		return nil
	}

	var roleTypes []app.RoleType
	for _, part := range strings.Split(raw, ",") {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		roleTypes = append(roleTypes, app.RoleType(p))
	}

	return roleTypes
}
