package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func (s *service) resolveServiceAccountRole(ctx *gin.Context, orgID, role string) (app.RoleType, error) {
	resolved, err := s.authzClient.ResolveAssignableRole(ctx, orgID, app.RoleType(role), app.RoleContextServiceAccount)
	if err != nil {
		return "", stderr.ErrUser{
			Err:         err,
			Description: err.Error(),
		}
	}
	return resolved.RoleType, nil
}

// @ID						ListRoles
// @Summary				List your org's roles
// @Description.markdown	list_roles.md
// @Param					context	query	string	false	"filter to roles assignable on a surface (team, service_account, api_token, oidc_trust_policy)"	extensions(x-go-name=RoleContext)
// @Tags					accounts
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	[]app.Role
// @Router					/v1/roles [GET]
func (s *service) ListRoles(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var roles []app.Role
	res := s.db.WithContext(ctx).
		Where(app.Role{OrgID: generics.NewNullString(org.ID)}).
		Order("role_type").
		Find(&roles)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to list roles for org %s: %w", org.ID, res.Error))
		return
	}

	if roleContext := ctx.Query("context"); roleContext != "" {
		filtered := make([]app.Role, 0, len(roles))
		for _, role := range roles {
			if role.AllowsContext(roleContext) {
				filtered = append(filtered, role)
			}
		}
		roles = filtered
	}

	ctx.JSON(http.StatusOK, roles)
}

// getOrgServiceAccount looks up an account by ID and ensures it is a service
// account that belongs to the given org.
func (s *service) getOrgServiceAccount(ctx context.Context, orgID, accountID string) (*app.Account, error) {
	acct, err := s.acctClient.FindAccount(ctx, accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, stderr.ErrNotFound{
				Err:         fmt.Errorf("service account %q not found", accountID),
				Description: "service account not found",
			}
		}
		return nil, fmt.Errorf("unable to find account: %w", err)
	}

	if acct.AccountType != app.AccountTypeService {
		return nil, stderr.ErrNotFound{
			Err:         fmt.Errorf("account %q is not a service account", accountID),
			Description: "service account not found",
		}
	}

	found := false
	for _, orgIDValue := range acct.OrgIDs {
		if orgIDValue == orgID {
			found = true
			break
		}
	}
	if !found {
		return nil, stderr.ErrAuthorization{
			Err:         fmt.Errorf("service account %q does not belong to org %q", accountID, orgID),
			Description: "service account not found for this org",
		}
	}

	return acct, nil
}

// orgServiceAccountIDs resolves the org's service accounts off the
// account_roles org index. Joining accounts to account_roles in the list query
// instead lets the planner walk the whole accounts table in email order to
// satisfy the ORDER BY and LIMIT, which took tens of seconds on a cold cache.
func (s *service) orgServiceAccountIDs(ctx context.Context, orgID string, includeRunners, includeStacks bool) ([]string, error) {
	accountIDs := []string{}
	if err := s.orgServiceAccounts(ctx, orgID, includeRunners, includeStacks).Pluck("account_roles.account_id", &accountIDs).Error; err != nil {
		return nil, fmt.Errorf("unable to list service account ids for org %s: %w", orgID, err)
	}

	return accountIDs, nil
}

func (s *service) orgServiceAccounts(ctx context.Context, orgID string, includeRunners, includeStacks bool) *gorm.DB {
	tx := s.db.WithContext(ctx).
		Model(&app.AccountRole{}).
		Joins("JOIN accounts ON accounts.id = account_roles.account_id AND accounts.deleted_at = 0 AND accounts.account_type = ?", app.AccountTypeService).
		Where(app.AccountRole{OrgID: generics.NewNullString(orgID)})

	excludedRoleTypes := []app.RoleType{}
	if !includeRunners {
		excludedRoleTypes = append(excludedRoleTypes, app.RoleTypeRunner)
	}
	if !includeStacks {
		excludedRoleTypes = append(excludedRoleTypes, app.RoleTypeStack)
	}
	if len(excludedRoleTypes) > 0 {
		tx = tx.
			Joins("JOIN roles ON roles.id = account_roles.role_id AND roles.deleted_at = 0").
			Where("roles.role_type NOT IN ?", excludedRoleTypes)
	}

	return tx.Distinct("account_roles.account_id")
}

type ServiceAccountOwnership struct {
	OwnerType   string `json:"owner_type"`
	OwnerID     string `json:"owner_id"`
	Purpose     string `json:"purpose"`
	InstanceKey string `json:"instance_key"`
	OwnerName   string `json:"owner_name,omitempty"`
	InstallID   string `json:"install_id,omitempty"`
}

type ServiceAccount struct {
	app.Account

	SystemAccount         bool                     `json:"system_account"`
	Purposes              []string                 `json:"purposes"`
	ManagedServiceAccount *ServiceAccountOwnership `json:"managed_service_account,omitempty"`
}

type serviceAccountClass struct {
	System   bool
	Purposes []string
	Managed  *ServiceAccountOwnership
}

type serviceAccountClasses map[string]serviceAccountClass

func (c serviceAccountClasses) get(accountID string) serviceAccountClass {
	if class, ok := c[accountID]; ok {
		return class
	}
	return serviceAccountClass{Purposes: []string{}}
}

func (s *service) classifyServiceAccounts(ctx context.Context, orgID string, accounts any) (serviceAccountClasses, error) {
	var roles []struct {
		AccountID string
		RoleType  app.RoleType
	}
	if err := s.db.WithContext(ctx).
		Model(&app.AccountRole{}).
		Select("account_roles.account_id", "roles.role_type").
		Joins("JOIN roles ON roles.id = account_roles.role_id AND roles.deleted_at = 0").
		Where(app.AccountRole{OrgID: generics.NewNullString(orgID)}).
		Where("account_roles.account_id IN (?)", accounts).
		Where("roles.role_type IN ?", []app.RoleType{app.RoleTypeRunner, app.RoleTypeStack}).
		Scan(&roles).Error; err != nil {
		return nil, fmt.Errorf("unable to load service account roles for org %s: %w", orgID, err)
	}
	legacyPurposes := map[string][]string{}
	for _, role := range roles {
		if !slices.Contains(legacyPurposes[role.AccountID], string(role.RoleType)) {
			legacyPurposes[role.AccountID] = append(legacyPurposes[role.AccountID], string(role.RoleType))
		}
	}

	var managed []app.ManagedServiceAccount
	if err := s.db.WithContext(ctx).
		Select("account_id", "owner_type", "owner_id", "purpose", "instance_key").
		Where(app.ManagedServiceAccount{OrgID: orgID}).
		Where("account_id IN (?)", accounts).
		Find(&managed).Error; err != nil {
		return nil, fmt.Errorf("unable to load managed service accounts for org %s: %w", orgID, err)
	}
	ownership := make(map[string]*ServiceAccountOwnership, len(managed))
	for _, binding := range managed {
		ownership[binding.AccountID] = &ServiceAccountOwnership{
			OwnerType:   binding.OwnerType,
			OwnerID:     binding.OwnerID,
			Purpose:     string(binding.Purpose),
			InstanceKey: binding.InstanceKey,
		}
	}

	classes := make(serviceAccountClasses, len(legacyPurposes)+len(ownership))
	for accountID, purposes := range legacyPurposes {
		slices.Sort(purposes)
		classes[accountID] = serviceAccountClass{System: true, Purposes: purposes}
	}
	for accountID, owner := range ownership {
		classes[accountID] = serviceAccountClass{System: true, Purposes: []string{owner.Purpose}, Managed: owner}
	}

	return classes, nil
}

func (s *service) enrichServiceAccountOwners(ctx context.Context, org *app.Org, owners []*ServiceAccountOwnership) error {
	orgType := plugins.TableName(s.db, app.Org{})
	installType := plugins.TableName(s.db, app.Install{})
	stackType := plugins.TableName(s.db, app.InstallStack{})

	var installIDs, stackIDs []string
	for _, owner := range owners {
		switch owner.OwnerType {
		case installType:
			installIDs = append(installIDs, owner.OwnerID)
		case stackType:
			stackIDs = append(stackIDs, owner.OwnerID)
		}
	}

	stackInstalls := map[string]string{}
	if len(stackIDs) > 0 {
		var stacks []struct{ ID, InstallID string }
		if err := s.db.WithContext(ctx).
			Model(&app.InstallStack{}).
			Select("id", "install_id").
			Where(app.InstallStack{OrgID: org.ID}).
			Where("id IN ?", stackIDs).
			Find(&stacks).Error; err != nil {
			return fmt.Errorf("unable to load service account owner stacks for org %s: %w", org.ID, err)
		}
		for _, stack := range stacks {
			stackInstalls[stack.ID] = stack.InstallID
			installIDs = append(installIDs, stack.InstallID)
		}
	}

	installNames := map[string]string{}
	if len(installIDs) > 0 {
		var installs []struct{ ID, Name string }
		if err := s.db.WithContext(ctx).
			Model(&app.Install{}).
			Scopes(scopes.WithDisableViews).
			Select("id", "name").
			Where(app.Install{OrgID: org.ID}).
			Where("id IN ?", installIDs).
			Find(&installs).Error; err != nil {
			return fmt.Errorf("unable to load service account owner installs for org %s: %w", org.ID, err)
		}
		for _, install := range installs {
			installNames[install.ID] = install.Name
		}
	}

	for _, owner := range owners {
		switch owner.OwnerType {
		case orgType:
			if owner.OwnerID == org.ID {
				owner.OwnerName = org.Name
			}
		case installType:
			owner.OwnerName = installNames[owner.OwnerID]
		case stackType:
			if name, ok := installNames[stackInstalls[owner.OwnerID]]; ok {
				owner.OwnerName = name
				owner.InstallID = stackInstalls[owner.OwnerID]
			}
		}
	}

	return nil
}

func serviceAccountOwners(classes serviceAccountClasses, accountIDs []string) []*ServiceAccountOwnership {
	owners := []*ServiceAccountOwnership{}
	for _, accountID := range accountIDs {
		if owner := classes.get(accountID).Managed; owner != nil {
			owners = append(owners, owner)
		}
	}
	return owners
}

func (s *service) searchServiceAccounts(ctx context.Context, org *app.Org, scope *gorm.DB, accountIDs []string, classes serviceAccountClasses, query string) ([]string, error) {
	if len(accountIDs) == 0 {
		return accountIDs, nil
	}

	var accounts []struct{ ID, Email, Name string }
	if err := s.db.WithContext(ctx).
		Model(&app.Account{}).
		Select("id", "email", "name").
		Where("id IN (?)", scope).
		Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("unable to load service accounts for org %s: %w", org.ID, err)
	}

	if err := s.enrichServiceAccountOwners(ctx, org, serviceAccountOwners(classes, accountIDs)); err != nil {
		return nil, err
	}

	candidates := make(map[string]bool, len(accountIDs))
	for _, accountID := range accountIDs {
		candidates[accountID] = true
	}

	matched := make([]string, 0, len(accountIDs))
	for _, acct := range accounts {
		if !candidates[acct.ID] {
			continue
		}
		fields := []string{acct.ID, acct.Email, acct.Name}
		if owner := classes.get(acct.ID).Managed; owner != nil {
			fields = append(fields, owner.OwnerID, owner.OwnerName)
		}
		for _, field := range fields {
			if strings.Contains(strings.ToLower(field), query) {
				matched = append(matched, acct.ID)
				break
			}
		}
	}

	return matched, nil
}

// @ID						ListServiceAccounts
// @Summary				List service accounts for the current org
// @Description.markdown	list_service_accounts.md
// @Param					offset			query	int		false	"offset of results to return"	Default(0)
// @Param					limit			query	int		false	"limit of results to return"	Default(10)
// @Param					page			query	int		false	"page number of results to return"	Default(0)
// @Param					include_runners	query	bool	false	"include service accounts with the runner role (excluded by default; ignored when management is set)"
// @Param					include_stacks	query	bool	false	"include service accounts with the stack role (excluded by default; ignored when management is set)"
// @Param					management		query	string	false	"filter by who manages the account"	Enums(user, system, all)
// @Param					purpose			query	string	false	"filter to service accounts with this exact purpose"
// @Param					q				query	string	false	"case-insensitive substring match on name, email, ID, or owner name and ID"
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	[]ServiceAccount
// @Router					/v1/service-accounts [GET]
func (s *service) ListServiceAccounts(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	management := ctx.Query("management")
	includeRunners := ctx.Query("include_runners") == "true"
	includeStacks := ctx.Query("include_stacks") == "true"
	switch management {
	case "":
	case "user", "system", "all":
		includeRunners, includeStacks = true, true
	default:
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid management filter %q", management),
			Description: "management must be one of user, system or all",
		})
		return
	}
	purpose := ctx.Query("purpose")
	query := strings.ToLower(ctx.Query("q"))

	accountIDs, err := s.orgServiceAccountIDs(ctx, org.ID, includeRunners, includeStacks)
	if err != nil {
		ctx.Error(err)
		return
	}

	filtered := management == "user" || management == "system" || purpose != "" || query != ""
	var classes serviceAccountClasses
	if filtered {
		scope := s.orgServiceAccounts(ctx, org.ID, includeRunners, includeStacks)
		classes, err = s.classifyServiceAccounts(ctx, org.ID, scope)
		if err != nil {
			ctx.Error(err)
			return
		}

		candidates := make([]string, 0, len(accountIDs))
		for _, accountID := range accountIDs {
			class := classes.get(accountID)
			if (management == "user" && class.System) || (management == "system" && !class.System) {
				continue
			}
			if purpose != "" && !slices.Contains(class.Purposes, purpose) {
				continue
			}
			candidates = append(candidates, accountID)
		}
		accountIDs = candidates

		if query != "" {
			accountIDs, err = s.searchServiceAccounts(ctx, org, scope, accountIDs, classes, query)
			if err != nil {
				ctx.Error(err)
				return
			}
		}
	}

	accounts := []app.Account{}
	if len(accountIDs) > 0 {
		res := s.db.WithContext(ctx).
			Model(&app.Account{}).
			Where("accounts.id IN ?", accountIDs).
			Order("accounts.email").
			Order("accounts.id").
			Scopes(scopes.WithOffsetPagination).
			Preload("Roles", "org_id = ?", org.ID).
			Preload("Roles.Org").
			Preload("Roles.Policies").
			Find(&accounts)
		if res.Error != nil {
			ctx.Error(fmt.Errorf("unable to list service accounts for org %s: %w", org.ID, res.Error))
			return
		}
	}

	accounts, err = db.HandlePaginatedResponse(ctx, accounts)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to handle paginated response: %w", err))
		return
	}

	pageIDs := make([]string, 0, len(accounts))
	for _, acct := range accounts {
		pageIDs = append(pageIDs, acct.ID)
	}
	if !filtered {
		classes, err = s.classifyServiceAccounts(ctx, org.ID, pageIDs)
		if err != nil {
			ctx.Error(err)
			return
		}
	}
	if query == "" {
		if err := s.enrichServiceAccountOwners(ctx, org, serviceAccountOwners(classes, pageIDs)); err != nil {
			ctx.Error(err)
			return
		}
	}

	resp := make([]ServiceAccount, 0, len(accounts))
	for _, acct := range accounts {
		class := classes.get(acct.ID)
		resp = append(resp, ServiceAccount{
			Account:               acct,
			SystemAccount:         class.System,
			Purposes:              class.Purposes,
			ManagedServiceAccount: class.Managed,
		})
	}

	ctx.JSON(http.StatusOK, resp)
}

// @ID						ListServiceAccountPurposes
// @Summary				List the purposes of the current org's service accounts
// @Description.markdown	list_service_account_purposes.md
// @Tags					accounts
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		string
// @Router					/v1/service-accounts/purposes [GET]
func (s *service) ListServiceAccountPurposes(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	classes, err := s.classifyServiceAccounts(ctx, org.ID, s.orgServiceAccounts(ctx, org.ID, true, true))
	if err != nil {
		ctx.Error(err)
		return
	}

	purposes := []string{}
	for _, class := range classes {
		purposes = append(purposes, class.Purposes...)
	}
	slices.Sort(purposes)

	ctx.JSON(http.StatusOK, slices.Compact(purposes))
}

type CreateServiceAccountRequest struct {
	// Name is a human-friendly label for the service account.
	Name string `json:"name" binding:"required"`
	// Role must be one of the service account roles returned by GET /v1/roles.
	Role string `json:"role" binding:"required"`
}

// @ID						CreateServiceAccount
// @Summary				Create a service account for the current org
// @Description.markdown	create_service_account.md
// @Param					req	body	CreateServiceAccountRequest	true	"Input"
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.Account
// @Router					/v1/service-accounts [POST]
func (s *service) CreateServiceAccount(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateServiceAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		})
		return
	}

	roleType, err := s.resolveServiceAccountRole(ctx, org.ID, req.Role)
	if err != nil {
		ctx.Error(err)
		return
	}

	acct, err := s.acctClient.CreateServiceAccount(ctx, domains.NewAccountID(), req.Name)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create service account: %w", err))
		return
	}

	if err := s.authzClient.SetAccountOrgRole(ctx, org.ID, acct.ID, roleType); err != nil {
		ctx.Error(fmt.Errorf("unable to assign role to service account: %w", err))
		return
	}

	acct, err = s.acctClient.FindAccount(ctx, acct.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to reload service account: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, acct)
}

type UpdateServiceAccountRequest struct {
	// Name is a human-friendly label for the service account.
	Name string `json:"name" binding:"required"`
}

// @ID						UpdateServiceAccount
// @Summary				Update a service account for the current org
// @Description.markdown	update_service_account.md
// @Param					account_id	path	string						true	"service account ID"
// @Param					req			body	UpdateServiceAccountRequest	true	"Input"
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Account
// @Router					/v1/service-accounts/{account_id} [PATCH]
func (s *service) UpdateServiceAccount(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	accountID := ctx.Param("account_id")

	var req UpdateServiceAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		})
		return
	}

	acct, err := s.getOrgServiceAccount(ctx, org.ID, accountID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := authz.RequireUserManaged(tx, acct.ID); err != nil {
			return err
		}
		return tx.Model(&app.Account{}).Where(app.Account{ID: acct.ID}).Update("name", req.Name).Error
	}); err != nil {
		ctx.Error(fmt.Errorf("unable to update service account: %w", err))
		return
	}

	acct, err = s.acctClient.FindAccount(ctx, acct.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to reload service account: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, acct)
}

type UpdateServiceAccountRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// @ID						UpdateServiceAccountRole
// @Summary				Update the role of a service account for the current org
// @Description.markdown	update_service_account_role.md
// @Param					account_id	path	string							true	"service account ID"
// @Param					req			body	UpdateServiceAccountRoleRequest	true	"Input"
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Account
// @Router					/v1/service-accounts/{account_id}/role [PATCH]
func (s *service) UpdateServiceAccountRole(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	accountID := ctx.Param("account_id")

	var req UpdateServiceAccountRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		})
		return
	}

	roleType, err := s.resolveServiceAccountRole(ctx, org.ID, req.Role)
	if err != nil {
		ctx.Error(err)
		return
	}

	acct, err := s.getOrgServiceAccount(ctx, org.ID, accountID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := authz.RequireUserManaged(tx, acct.ID); err != nil {
			return err
		}
		return authz.New(authz.Params{DB: tx}).SetAccountOrgRole(ctx, org.ID, acct.ID, roleType)
	}); err != nil {
		ctx.Error(fmt.Errorf("unable to update service account role: %w", err))
		return
	}

	acct, err = s.acctClient.FindAccount(ctx, acct.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to reload service account: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, acct)
}

// @ID						DeleteServiceAccount
// @Summary				Delete a service account for the current org
// @Description.markdown	delete_service_account.md
// @Param					account_id	path	string	true	"service account ID"
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				202
// @Router					/v1/service-accounts/{account_id} [DELETE]
func (s *service) DeleteServiceAccount(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	accountID := ctx.Param("account_id")

	acct, err := s.getOrgServiceAccount(ctx, org.ID, accountID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := authz.RequireUserManaged(tx, acct.ID); err != nil {
			return err
		}
		if err := authz.New(authz.Params{DB: tx}).RemoveAccountOrgRoles(ctx, org.ID, acct.ID); err != nil {
			return fmt.Errorf("unable to remove roles: %w", err)
		}
		return account.New(account.Params{DB: tx}).InvalidateTokens(ctx, acct.Email)
	}); err != nil {
		ctx.Error(fmt.Errorf("unable to delete service account: %w", err))
		return
	}

	ctx.Status(http.StatusAccepted)
}

type CreateServiceAccountTokenRequest struct {
	// Duration defaults to one year.
	Duration string `json:"duration" default:"8760h"`

	// Name labels the token where it is listed; defaults to the account's identity.
	Name string `json:"name"`

	Invalidate bool `json:"invalidate"`
}

type CreateServiceAccountTokenResponse struct {
	Token string `json:"token,omitzero"`
}

// @ID						CreateServiceAccountToken
// @Summary				Create a token for a service account in the current org
// @Description.markdown	create_service_account_token.md
// @Param					account_id	path	string								true	"service account ID"
// @Param					req			body	CreateServiceAccountTokenRequest	true	"Input"
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	CreateServiceAccountTokenResponse
// @Router					/v1/service-accounts/{account_id}/tokens [POST]
func (s *service) CreateServiceAccountToken(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	accountID := ctx.Param("account_id")

	var req CreateServiceAccountTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, http.ErrBodyNotAllowed) {
		if err.Error() != "EOF" {
			ctx.Error(stderr.ErrUser{
				Err:         fmt.Errorf("unable to parse request: %w", err),
				Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
			})
			return
		}
	}

	duration := req.Duration
	if duration == "" {
		duration = "8760h"
	}

	dur, err := time.ParseDuration(duration)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid duration: %w", err),
			Description: fmt.Sprintf("invalid duration: %s", err.Error()),
		})
		return
	}

	acct, err := s.getOrgServiceAccount(ctx, org.ID, accountID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if req.Invalidate {
		if err := s.acctClient.InvalidateTokens(ctx, acct.Email); err != nil {
			ctx.Error(fmt.Errorf("unable to invalidate tokens: %w", err))
			return
		}
	}

	caller, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = acct.Email
	}

	// createStaticToken, not acctClient.CreateToken: only this one stamps the columns
	// ListStaticTokens filters on, so tokens made the other way cannot be revoked
	// from the org's API tokens page.
	token, err := s.createStaticToken(ctx, acct, org.ID, caller.ID, name, orgRoleType(acct, org.ID), dur)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create token: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, CreateServiceAccountTokenResponse{
		Token: token.Token,
	})
}

// orgRoleType reports the account's org role for display only.
func orgRoleType(acct *app.Account, orgID string) app.RoleType {
	for _, role := range acct.Roles {
		if role.OrgID.ValueString() == orgID {
			return role.RoleType
		}
	}

	return ""
}
