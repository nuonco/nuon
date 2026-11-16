package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CleanupCredentialsRequest struct {
	ServiceAccountSubject string `json:"service_account_subject,omitempty"`
	AccountID             string `json:"account_id,omitempty"`
	OrgID                 string `json:"org_id,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @max-retries 3
func (a *Activities) CleanupCredentials(ctx context.Context, req CleanupCredentialsRequest) error {
	if req.ServiceAccountSubject != "" {
		if err := a.acctClient.RevokeServiceAccountCredentials(ctx, req.ServiceAccountSubject); err != nil {
			return fmt.Errorf("revoke service account credentials: %w", err)
		}
		return nil
	}
	if req.AccountID == "" || req.OrgID == "" {
		return fmt.Errorf("credential cleanup requires a service account subject or an account and org")
	}
	if err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Where(app.Token{OrgID: req.OrgID, AccountID: req.AccountID}).Delete(&app.Token{}).Error
	}); err != nil {
		return fmt.Errorf("revoke organization tokens: %w", err)
	}
	return nil
}
