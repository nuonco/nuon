package authz

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// logStreamWriteRole can create logs on one stream and nothing else. The key is
// the stream, so a stolen write token cannot fall back to org-wide access.
func logStreamWriteRole(orgID, logStreamID string) *app.Role {
	return &app.Role{
		OrgID:       generics.NewNullString(orgID),
		RoleType:    app.RoleTypeLogStream,
		Title:       "Log stream",
		Description: "Write logs to a single log stream.",
		Policies: []app.Policy{{
			OrgID: generics.NewNullString(orgID),
			Name:  app.PolicyNameLogStream,
			Permissions: pgtype.Hstore{
				permissions.Object(orgID, permissions.KindLogStream, logStreamID): permissions.PermissionCreate.ToStrPtr(),
			},
		}},
	}
}

// EnsureLogStreamWriteRole gives a log-stream service account a private role
// scoped to writes on that stream.
func (h *Client) EnsureLogStreamWriteRole(ctx context.Context, orgID, logStreamID, accountID string) error {
	ctx = cctx.SetAccountIDContext(ctx, accountID)
	return h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return ensurePrivateRole(tx, accountID, logStreamWriteRole(orgID, logStreamID))
	})
}
