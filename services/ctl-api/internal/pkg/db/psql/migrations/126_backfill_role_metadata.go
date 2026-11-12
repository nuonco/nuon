package migrations

import (
	"context"
	"fmt"
	"slices"

	"gorm.io/gorm"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (m *Migrations) Migration126BackfillRoleMetadata(ctx context.Context, db *gorm.DB) error {
	const batchSize = 20
	var offset int

	for {
		var orgs []app.Org
		if err := db.WithContext(ctx).Limit(batchSize).Offset(offset).Find(&orgs).Error; err != nil {
			return fmt.Errorf("unable to fetch orgs: %w", err)
		}
		if len(orgs) == 0 {
			return nil
		}

		for _, org := range orgs {
			if err := reconcileLegacyOrgRoles(ctx, db, org); err != nil {
				return fmt.Errorf("unable to reconcile roles for org %s: %w", org.ID, err)
			}
		}

		offset += batchSize
	}
}

func reconcileLegacyOrgRoles(ctx context.Context, db *gorm.DB, org app.Org) error {
	contexts := []string{app.RoleContextTeam, app.RoleContextServiceAccount, app.RoleContextAPIToken, app.RoleContextTrustPolicy}
	definitions := []struct {
		roleType    app.RoleType
		title       string
		description string
		contexts    []string
		policyName  app.PolicyName
		permission  string
		api         string
	}{
		{app.RoleTypeOrgAdmin, "Admin", "Full access to the organization and all of its resources.", contexts, app.PolicyNameOrgAdmin, "all", "ctl-api"},
		{app.RoleTypeOrgSupport, "Support", "Deprecated full-access role; no longer offered for new assignments.", nil, app.PolicyNameOrgSupport, "all", "ctl-api"},
		{app.RoleTypeOrgReadOnly, "Read-only", "Read-only access to the organization and its resources.", contexts, app.PolicyNameOrgReadOnly, "read", "ctl-api"},
		{app.RoleTypeInstaller, "Installer", "Deprecated full-access role; no longer offered for new assignments.", nil, app.PolicyNameInstaller, "all", "ctl-api"},
		{app.RoleTypeRunner, "Runner", "Permissions for runners executing deployments.", nil, "admin", "all", "runner"},
	}
	var existingRoles []app.Role
	if err := db.WithContext(ctx).Where(app.Role{OrgID: generics.NewNullString(org.ID)}).Find(&existingRoles).Error; err != nil {
		return fmt.Errorf("unable to load roles for org %s: %w", org.ID, err)
	}
	byType := make(map[app.RoleType]app.Role, len(existingRoles))
	for _, role := range existingRoles {
		byType[role.RoleType] = role
	}
	var toCreate, toUpdate []app.Role
	for _, definition := range definitions {
		want := app.Role{
			OrgID:       generics.NewNullString(org.ID),
			RoleType:    definition.roleType,
			Title:       definition.title,
			Description: definition.description,
			Contexts:    definition.contexts,
			Managed:     true,
		}
		existing, ok := byType[want.RoleType]
		if !ok {
			want.Version = app.RoleVersionV2
			want.CreatedByID = org.CreatedByID
			want.Policies = []app.Policy{{
				OrgID:       want.OrgID,
				CreatedByID: org.CreatedByID,
				Name:        definition.policyName,
				Permissions: pgtype.Hstore{org.ID: generics.ToPtr(definition.permission), "api": generics.ToPtr(definition.api)},
			}}
			toCreate = append(toCreate, want)
			continue
		}
		if existing.Title == want.Title && existing.Description == want.Description &&
			existing.Managed == want.Managed && slices.Equal(existing.Contexts, want.Contexts) {
			continue
		}
		want.ID = existing.ID
		toUpdate = append(toUpdate, want)
	}
	if len(toCreate) == 0 && len(toUpdate) == 0 {
		return nil
	}
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(toCreate) > 0 {
			if err := tx.Create(&toCreate).Error; err != nil {
				return fmt.Errorf("unable to create roles for org %s: %w", org.ID, err)
			}
		}
		for i := range toUpdate {
			role := &toUpdate[i]
			if err := tx.Model(role).Select("title", "description", "contexts", "managed").Updates(role).Error; err != nil {
				return fmt.Errorf("unable to update %s role metadata for org %s: %w", role.RoleType, org.ID, err)
			}
		}
		return nil
	})
}
