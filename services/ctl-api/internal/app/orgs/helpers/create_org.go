package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

const OrgSignalsQueueName = "org-signals"

type CreateOrgParams struct {
	Name           string
	UseSandboxMode bool
	Tags           []string
}

func (h *Helpers) CreateOrg(ctx context.Context, acct *app.Account, params *CreateOrgParams) (*app.Org, error) {
	orgTyp := app.OrgTypeDefault
	if params.UseSandboxMode {
		orgTyp = app.OrgTypeSandbox
	}
	if acct.AccountType == app.AccountTypeIntegration {
		orgTyp = app.OrgTypeIntegration
	}
	if h.cfg.ForceSandboxMode {
		orgTyp = app.OrgTypeSandbox
	}

	notificationsCfg := app.NotificationsConfig{
		EnableSlackNotifications: acct.AccountType == app.AccountTypeAuth0,
		EnableEmailNotifications: acct.AccountType == app.AccountTypeAuth0,
		InternalSlackWebhookURL:  h.cfg.InternalSlackWebhookURL,
	}
	org := app.Org{
		Name:                params.Name,
		Status:              "queued",
		StatusDescription:   "waiting for event loop to start and provision org",
		SandboxMode:         params.UseSandboxMode,
		OrgType:             orgTyp,
		NotificationsConfig: notificationsCfg,
		Tags:                params.Tags,
	}
	if h.cfg.ForceSandboxMode {
		org.SandboxMode = true
	}
	if h.cfg.ForceDebugMode {
		org.DebugMode = true
	}

	if err := h.db.WithContext(ctx).Create(&org).Error; err != nil {
		return nil, fmt.Errorf("unable to create org: %w", err)
	}

	// make sure the notifications config orgID is set
	if res := h.db.WithContext(ctx).
		Where(&app.NotificationsConfig{
			OwnerID: org.ID,
		}).
		Updates(app.NotificationsConfig{
			OrgID: org.ID,
		}); res.Error != nil {
		return nil, fmt.Errorf("unable to set org ID on notifications config: %w", res.Error)
	}

	if err := h.authzClient.CreateOrgRoles(ctx, org.ID); err != nil {
		return nil, fmt.Errorf("unable to create org roles: %w", err)
	}

	if err := h.authzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, org.ID, acct.ID); err != nil {
		return nil, fmt.Errorf("unable to add user to org: %w", err)
	}

	if _, err := h.runnersHelpers.CreateOrgRunnerGroup(ctx, &org); err != nil {
		return nil, fmt.Errorf("unable to create org runner group: %w", err)
	}

	// Create the org-signals queue
	_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     org.ID,
		OwnerType:   plugins.TableName(h.db, app.Org{}),
		Namespace:   "orgs",
		Name:        OrgSignalsQueueName,
		MaxInFlight: 10,
		MaxDepth:    50,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create org-signals queue: %w", err)
	}

	return &org, nil
}
