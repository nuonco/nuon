package authz

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

// EnsureRunnerJobRole grants one job, and the install its runner serves when
// installID is set, in a single private role.
func (h *Client) EnsureRunnerJobRole(ctx context.Context, orgID, jobID, installID, accountID string) error {
	grants := pgtype.Hstore{
		permissions.APIPermissionKey:                                generics.ToPtr(string(permissions.APIRunner)),
		permissions.Object(orgID, permissions.KindRunnerJob, jobID): permissions.PermissionAll.ToStrPtr(),
	}
	if installID != "" {
		grants[permissions.Object(orgID, permissions.KindInstall, installID)] = permissions.PermissionAll.ToStrPtr()
	}
	return h.EnsureManagedServiceAccountRole(ctx, accountID, &app.Role{
		OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypeRunnerJob,
		Version: app.RoleVersionV2, Title: "Runner job", Description: "Scoped access to one runner job.",
		Policies: []app.Policy{{OrgID: generics.NewNullString(orgID), Name: app.PolicyNameRunnerJob, Permissions: grants}},
	})
}

func (h *Client) EnsureLogStreamRole(ctx context.Context, orgID, logStreamID, accountID string) error {
	return h.EnsureManagedServiceAccountRole(ctx, accountID, &app.Role{
		OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypeLogStream,
		Version: app.RoleVersionV2, Title: "Log stream", Description: "Append to one log stream.",
		Policies: []app.Policy{{OrgID: generics.NewNullString(orgID), Name: app.PolicyNameLogStream,
			Permissions: pgtype.Hstore{
				permissions.APIPermissionKey:                                      generics.ToPtr(string(permissions.APIRunner)),
				permissions.Object(orgID, permissions.KindLogStream, logStreamID): permissions.PermissionCreate.ToStrPtr(),
			},
		}},
	})
}

func (h *Client) EnsureTelemetryCollectorRole(ctx context.Context, orgID, installID, accountID string) error {
	return h.EnsureManagedServiceAccountRole(ctx, accountID, &app.Role{
		OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypeTelemetryCollector,
		Version: app.RoleVersionV2, Title: "Telemetry collector", Description: "Scoped access to telemetry endpoints for a single install.",
		Policies: []app.Policy{{OrgID: generics.NewNullString(orgID), Name: app.PolicyNameTelemetryCollector,
			Permissions: pgtype.Hstore{
				permissions.APIPermissionKey:                                    generics.ToPtr(string(permissions.APIRunner)),
				permissions.Object(orgID, permissions.KindTelemetry, installID): permissions.PermissionAll.ToStrPtr(),
			},
		}},
	})
}

func (h *Client) EnsurePhoneHomeRole(ctx context.Context, orgID, versionID, accountID string) error {
	return h.EnsureManagedServiceAccountRole(ctx, accountID, &app.Role{
		OrgID: generics.NewNullString(orgID), RoleType: app.RoleTypePhoneHome,
		Title: "Phone home", Description: "Scoped access to one install stack version callback.",
		Policies: []app.Policy{{OrgID: generics.NewNullString(orgID), Name: app.PolicyNamePhoneHome,
			Permissions: pgtype.Hstore{
				permissions.Object(orgID, permissions.KindPhoneHome, versionID): permissions.PermissionCreate.ToStrPtr(),
			},
		}},
	})
}
