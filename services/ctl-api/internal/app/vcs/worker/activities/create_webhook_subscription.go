package activities

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"go.uber.org/zap"
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
	var vcsConn app.VCSConnection
	if err := a.db.WithContext(ctx).First(&vcsConn, "id = ?", req.VCSConnectionID).Error; err != nil {
		return nil, fmt.Errorf("unable to get vcs connection: %w", err)
	}

	var existing app.VCSWebhookSubscription
	err := a.db.WithContext(ctx).
		Where(app.VCSWebhookSubscription{GithubInstallID: vcsConn.GithubInstallID}).
		First(&existing).Error
	if err == nil {
		if existing.GithubHookID != 0 {
			return existingWebhookResponse(&existing), nil
		}
		return a.provisionGithubWebhook(ctx, &vcsConn, &existing, false)
	}
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("unable to check existing webhook subscription: %w", err)
	}

	sub, inserted, err := a.insertWebhookSubscription(ctx, &vcsConn)
	if err != nil {
		return nil, err
	}
	if !inserted {
		if err := a.dropDeletedWebhookSubscription(ctx, vcsConn.GithubInstallID); err != nil {
			return nil, err
		}
		sub, inserted, err = a.insertWebhookSubscription(ctx, &vcsConn)
		if err != nil {
			return nil, err
		}
	}
	if !inserted {
		var winner app.VCSWebhookSubscription
		if err := a.db.WithContext(ctx).Unscoped().
			Where(app.VCSWebhookSubscription{GithubInstallID: vcsConn.GithubInstallID}).
			First(&winner).Error; err != nil {
			return nil, fmt.Errorf("unable to fetch existing webhook subscription: %w", err)
		}
		if winner.DeletedAt != 0 {
			return nil, fmt.Errorf("deleted webhook subscription %s still owns github hook %d", winner.ID, winner.GithubHookID)
		}
		if winner.GithubHookID != 0 {
			return existingWebhookResponse(&winner), nil
		}
		return nil, fmt.Errorf("webhook subscription %s is not ready", winner.ID)
	}

	return a.provisionGithubWebhook(ctx, &vcsConn, sub, true)
}

func existingWebhookResponse(sub *app.VCSWebhookSubscription) *CreateWebhookSubscriptionResponse {
	return &CreateWebhookSubscriptionResponse{
		SubscriptionID: sub.ID,
		WebhookURL:     sub.WebhookURL,
		AlreadyExisted: true,
	}
}

func (a *Activities) insertWebhookSubscription(ctx context.Context, vcsConn *app.VCSConnection) (*app.VCSWebhookSubscription, bool, error) {
	secret, err := generateWebhookSecret()
	if err != nil {
		return nil, false, fmt.Errorf("unable to generate webhook secret: %w", err)
	}

	sub := app.VCSWebhookSubscription{
		OrgID:           vcsConn.OrgID,
		VCSConnectionID: vcsConn.ID,
		GithubInstallID: vcsConn.GithubInstallID,
		WebhookSecret:   secret,
		Status: &app.CompositeStatus{
			CreatedAtTS:            time.Now().Unix(),
			Status:                 app.StatusPending,
			StatusHumanDescription: "creating webhook subscription",
		},
	}
	res := a.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "github_install_id"}},
		DoNothing: true,
	}).Create(&sub)
	if res.Error != nil {
		return nil, false, fmt.Errorf("unable to create webhook subscription: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, false, nil
	}
	return &sub, true, nil
}

func (a *Activities) dropDeletedWebhookSubscription(ctx context.Context, githubInstallID string) error {
	var deleted app.VCSWebhookSubscription
	err := a.db.WithContext(ctx).Unscoped().
		Where(app.VCSWebhookSubscription{GithubInstallID: githubInstallID}).
		First(&deleted).Error
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return fmt.Errorf("unable to find deleted webhook subscription: %w", err)
	}
	if deleted.DeletedAt == 0 || deleted.GithubHookID != 0 {
		return nil
	}
	if err := a.db.WithContext(ctx).Unscoped().Delete(&deleted).Error; err != nil {
		return fmt.Errorf("unable to delete webhook subscription: %w", err)
	}
	return nil
}

func (a *Activities) provisionGithubWebhook(ctx context.Context, vcsConn *app.VCSConnection, sub *app.VCSWebhookSubscription, cleanupOnFailure bool) (*CreateWebhookSubscriptionResponse, error) {
	webhookURL := fmt.Sprintf("%s/v1/vcs/webhooks/%s/events", a.cfg.PublicAPIURL, sub.ID)
	hookID, err := a.ghClient.CreateOrgWebhook(ctx, vcsConn, webhookURL, sub.WebhookSecret)
	if err != nil {
		// GitHub may already hold a hook for this URL: a previous attempt
		// created it and died before recording the ID, or the request timed
		// out after it landed. Adopt it rather than looping on 422 or
		// deleting the row the hook delivers to.
		existingID, findErr := a.ghClient.FindOrgWebhookByURL(ctx, vcsConn, webhookURL)
		if findErr != nil {
			a.l.Warn("unable to look up existing github org webhook",
				zap.String("subscription_id", sub.ID),
				zap.String("vcs_connection_id", vcsConn.ID),
				zap.Error(findErr))
		}
		if findErr != nil || existingID == 0 {
			if cleanupOnFailure {
				if delErr := a.db.WithContext(ctx).Unscoped().Delete(sub).Error; delErr != nil {
					return nil, fmt.Errorf("unable to create github org webhook: %w (cleanup: %v)", err, delErr)
				}
			}
			return nil, fmt.Errorf("unable to create github org webhook: %w", err)
		}
		hookID = existingID
	}

	sub.WebhookURL = webhookURL
	sub.GithubHookID = hookID
	sub.Status = &app.CompositeStatus{
		CreatedAtTS:            time.Now().Unix(),
		Status:                 app.StatusSuccess,
		StatusHumanDescription: "webhook subscription created",
	}
	res := a.db.WithContext(ctx).
		Model(sub).
		Where(app.VCSWebhookSubscription{ID: sub.ID}).
		Updates(map[string]any{
			"webhook_url":    webhookURL,
			"github_hook_id": hookID,
			"status":         sub.Status,
		})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update webhook subscription: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("unable to update webhook subscription %s", sub.ID)
	}

	return &CreateWebhookSubscriptionResponse{
		SubscriptionID: sub.ID,
		WebhookURL:     webhookURL,
		AlreadyExisted: false,
	}, nil
}
