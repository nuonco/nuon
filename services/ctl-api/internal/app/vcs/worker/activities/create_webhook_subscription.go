package activities

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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

// generateWebhookSecret creates a cryptographically random hex string for HMAC signature verification.
func generateWebhookSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("unable to generate random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// @temporal-gen-v2 activity
func (a *Activities) CreateWebhookSubscription(ctx context.Context, req CreateWebhookSubscriptionRequest) (*CreateWebhookSubscriptionResponse, error) {
	// Load the VCS connection to get github_install_id and org context.
	var vcsConn app.VCSConnection
	if err := a.db.WithContext(ctx).First(&vcsConn, "id = ?", req.VCSConnectionID).Error; err != nil {
		return nil, fmt.Errorf("unable to get vcs connection: %w", err)
	}

	// Check if a webhook subscription already exists for this GitHub installation.
	// This deduplicates across orgs — only one webhook per GitHub org/installation.
	var existing app.VCSWebhookSubscription
	err := a.db.WithContext(ctx).
		Where("github_install_id = ?", vcsConn.GithubInstallID).
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

	// Generate a webhook secret for HMAC signature verification.
	secret, err := generateWebhookSecret()
	if err != nil {
		return nil, fmt.Errorf("unable to generate webhook secret: %w", err)
	}

	// Create the subscription record first to get the ID for the webhook URL.
	// The webhook URL uses the subscription ID (opaque, unguessable) instead of the raw github_install_id.
	sub := app.VCSWebhookSubscription{
		OrgID:           vcsConn.OrgID,
		VCSConnectionID: req.VCSConnectionID,
		GithubInstallID: vcsConn.GithubInstallID,
		WebhookSecret:   secret,
		Status: &app.CompositeStatus{
			CreatedAtTS:            time.Now().Unix(),
			Status:                 app.StatusPending,
			StatusHumanDescription: "creating webhook subscription",
		},
	}

	if err := a.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "github_install_id"}},
		DoNothing: true,
	}).Create(&sub).Error; err != nil {
		return nil, fmt.Errorf("unable to create webhook subscription: %w", err)
	}

	// If conflict (another connection beat us), fetch the existing one.
	if sub.ID == "" {
		if err := a.db.WithContext(ctx).
			Where("github_install_id = ?", vcsConn.GithubInstallID).
			First(&sub).Error; err != nil {
			return nil, fmt.Errorf("unable to fetch existing webhook subscription: %w", err)
		}
		return &CreateWebhookSubscriptionResponse{
			SubscriptionID: sub.ID,
			WebhookURL:     sub.WebhookURL,
			AlreadyExisted: true,
		}, nil
	}

	// Build webhook URL using the subscription ID (opaque identifier).
	webhookURL := fmt.Sprintf("%s/v1/vcs/webhooks/%s/events", a.cfg.PublicAPIURL, sub.ID)

	// Create the GitHub org webhook with the secret for signature verification.
	hookID, err := a.ghClient.CreateOrgWebhook(ctx, &vcsConn, webhookURL, secret)
	if err != nil {
		// Clean up the subscription record on failure.
		a.db.WithContext(ctx).Delete(&sub)
		return nil, fmt.Errorf("unable to create github org webhook: %w", err)
	}

	// Update the subscription with the webhook URL and GitHub hook ID.
	sub.WebhookURL = webhookURL
	sub.GithubHookID = hookID
	sub.Status = &app.CompositeStatus{
		CreatedAtTS:            time.Now().Unix(),
		Status:                 app.StatusSuccess,
		StatusHumanDescription: "webhook subscription created",
	}
	if err := a.db.WithContext(ctx).Save(&sub).Error; err != nil {
		return nil, fmt.Errorf("unable to update webhook subscription: %w", err)
	}

	return &CreateWebhookSubscriptionResponse{
		SubscriptionID: sub.ID,
		WebhookURL:     webhookURL,
		AlreadyExisted: false,
	}, nil
}
