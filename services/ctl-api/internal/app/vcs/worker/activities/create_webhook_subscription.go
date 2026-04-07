package activities

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateWebhookSubscriptionRequest struct {
	VCSConnectionID string `validate:"required"`
}

type CreateWebhookSubscriptionResponse struct {
	SubscriptionID string `json:"subscription_id"`
	WebhookURL     string `json:"webhook_url"`
	AlreadyExisted bool   `json:"already_existed"`
}

// @temporal-gen-v2 activity
func (a *Activities) CreateWebhookSubscription(ctx context.Context, req CreateWebhookSubscriptionRequest) (*CreateWebhookSubscriptionResponse, error) {
	// Check if subscription already exists for this VCS connection.
	var existing app.VCSWebhookSubscription
	err := a.db.WithContext(ctx).
		Where("vcs_connection_id = ?", req.VCSConnectionID).
		First(&existing).Error
	if err == nil {
		return &CreateWebhookSubscriptionResponse{
			SubscriptionID: existing.ID,
			WebhookURL:     existing.WebhookURL,
			AlreadyExisted: true,
		}, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("unable to check existing webhook subscription: %w", err)
	}

	// Load the VCS connection to get org context.
	var vcsConn app.VCSConnection
	if err := a.db.WithContext(ctx).First(&vcsConn, "id = ?", req.VCSConnectionID).Error; err != nil {
		return nil, fmt.Errorf("unable to get vcs connection: %w", err)
	}

	// Generic webhook URL — events are routed by installation ID in the payload.
	webhookURL := fmt.Sprintf("%s/v1/vcs/events", a.cfg.PublicAPIURL)

	// Check if another VCS connection with the same github_install_id already
	// has a webhook subscription. If so, skip the GitHub API call to avoid
	// duplicate webhook registration errors.
	var existingSibling app.VCSWebhookSubscription
	err = a.db.WithContext(ctx).
		Joins("JOIN vcs_connections ON vcs_connections.id = vcs_webhook_subscriptions.vcs_connection_id").
		Where("vcs_connections.github_install_id = ?", vcsConn.GithubInstallID).
		First(&existingSibling).Error

	var hookID int64
	if err == nil {
		// Reuse the existing GitHub webhook hook ID.
		hookID = existingSibling.GithubHookID
	} else if err == gorm.ErrRecordNotFound {
		// No existing webhook for this GitHub installation — create one.
		hookID, err = a.ghClient.CreateOrgWebhook(ctx, &vcsConn, webhookURL)
		if err != nil {
			return nil, fmt.Errorf("unable to create github org webhook: %w", err)
		}
	} else {
		return nil, fmt.Errorf("unable to check existing webhook for github install: %w", err)
	}

	// Persist the subscription.
	sub := app.VCSWebhookSubscription{
		OrgID:           vcsConn.OrgID,
		VCSConnectionID: req.VCSConnectionID,
		GithubHookID:    hookID,
		WebhookURL:      webhookURL,
		Status: &app.CompositeStatus{
			CreatedAtTS:            time.Now().Unix(),
			Status:                 app.StatusSuccess,
			StatusHumanDescription: "webhook subscription created",
		},
	}

	if err := a.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "vcs_connection_id"}},
		DoNothing: true,
	}).Create(&sub).Error; err != nil {
		return nil, fmt.Errorf("unable to create webhook subscription: %w", err)
	}

	return &CreateWebhookSubscriptionResponse{
		SubscriptionID: sub.ID,
		WebhookURL:     webhookURL,
		AlreadyExisted: false,
	}, nil
}
