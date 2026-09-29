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

func generateWebhookSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("unable to generate random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// @temporal-gen-v2 activity
func (a *Activities) CreateWebhookSubscription(ctx context.Context, req CreateWebhookSubscriptionRequest) (*CreateWebhookSubscriptionResponse, error) {
	var vcsConn app.VCSConnection
	if err := a.db.WithContext(ctx).First(&vcsConn, "id = ?", req.VCSConnectionID).Error; err != nil {
		return nil, fmt.Errorf("unable to get vcs connection: %w", err)
	}

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

	secret, err := generateWebhookSecret()
	if err != nil {
		return nil, fmt.Errorf("unable to generate webhook secret: %w", err)
	}

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

	webhookURL := fmt.Sprintf("%s/v1/vcs/webhooks/%s/events", a.cfg.PublicAPIURL, sub.ID)

	hookID, err := a.ghClient.CreateOrgWebhook(ctx, &vcsConn, webhookURL, secret)
	if err != nil {
		a.db.WithContext(ctx).Delete(&sub)
		return nil, fmt.Errorf("unable to create github org webhook: %w", err)
	}

	sub.WebhookURL = webhookURL
	sub.GithubHookID = hookID
	sub.Status = &app.CompositeStatus{
		CreatedAtTS:            time.Now().Unix(),
		Status:                 app.StatusSuccess,
		StatusHumanDescription: "webhook subscription created",
	}
	if err := a.db.WithContext(ctx).
		Model(&sub).
		Where("id = ?", sub.ID).
		Updates(map[string]any{
			"webhook_url":    webhookURL,
			"github_hook_id": hookID,
			"status":         sub.Status,
		}).Error; err != nil {
		return nil, fmt.Errorf("unable to update webhook subscription: %w", err)
	}

	return &CreateWebhookSubscriptionResponse{
		SubscriptionID: sub.ID,
		WebhookURL:     webhookURL,
		AlreadyExisted: false,
	}, nil
}
