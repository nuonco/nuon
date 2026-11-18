package account

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var ErrJobClaimConflict = errors.New("job is not available")

// lockAccount keeps the account from being deleted or having its credentials
// revoked while a credential is issued. KEY SHARE conflicts only with the FOR
// UPDATE that revocation and deletion take, so concurrent issuance does not wait.
func lockAccount(tx *gorm.DB, accountID string) error {
	if accountID == "" {
		return fmt.Errorf("account ID is required")
	}
	return tx.Select("id").Clauses(clause.Locking{Strength: "KEY SHARE"}).Where(app.Account{ID: accountID}).Take(&app.Account{}).Error
}

func (c *Client) insertToken(ctx context.Context, token *app.Token) (*app.Token, error) {
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if token.OrgID != "" {
			if err := lockOrg(tx, token.OrgID); err != nil {
				return err
			}
		}
		if err := lockAccount(tx, token.AccountID); err != nil {
			return err
		}
		token.Token = domains.NewUserTokenID()
		return tx.Create(token).Error
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

type SessionTokenRequest struct {
	AccountID     string
	Issuer        string
	TTL           time.Duration
	SourceTokenID string
}

func (c *Client) IssueSessionToken(ctx context.Context, req SessionTokenRequest) (*app.Token, error) {
	return c.issueAuthToken(ctx, req, app.TokenTypeNuon, "")
}

func (c *Client) IssueOAuthAccessToken(ctx context.Context, req SessionTokenRequest, role string) (*app.Token, error) {
	return c.issueAuthToken(ctx, req, app.TokenTypeOAuth, role)
}

func (c *Client) issueAuthToken(ctx context.Context, req SessionTokenRequest, tokenType app.TokenType, role string) (*app.Token, error) {
	var token *app.Token
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := LockGrantSource(tx, req.AccountID, req.SourceTokenID, true); err != nil {
			return err
		}
		now := time.Now()
		value := &app.Token{AccountID: req.AccountID, CreatedByID: req.AccountID, TokenType: tokenType,
			Issuer: req.Issuer, IssuedAt: now, ExpiresAt: now.Add(req.TTL), Role: role}
		if req.SourceTokenID != "" {
			value.SourceType, value.SourceID = app.TokenSourceTypeToken, req.SourceTokenID
		}
		var err error
		token, err = c.withDB(tx).insertToken(ctx, value)
		return err
	})
	return token, err
}

func LockGrantSource(tx *gorm.DB, accountID, tokenID string, historical bool) error {
	if err := lockAccount(tx, accountID); err != nil {
		return err
	}
	if tokenID == "" {
		return nil
	}
	if historical {
		tx = tx.Unscoped()
	}
	return tx.Select("id").Clauses(clause.Locking{Strength: "SHARE"}).Where(app.Token{ID: tokenID}).Take(&app.Token{}).Error
}

type StaticTokenRequest struct {
	AccountID   string
	OrgID       string
	CreatedByID string
	Name        string
	Role        app.RoleType
	Duration    time.Duration
	Invalidate  bool
}

func (c *Client) CreateDedicatedStaticToken(ctx context.Context, req StaticTokenRequest) (*app.Token, error) {
	var token *app.Token
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		client := c.withDB(tx)
		acct, err := client.CreateOrgServiceAccount(ctx, req.OrgID, DedicatedTokenSubjectPrefix(req.OrgID)+domains.NewAccountID(), req.Name, req.Role)
		if err != nil {
			return err
		}
		req.AccountID = acct.ID
		token, err = client.IssueStaticToken(ctx, req)
		return err
	})
	return token, err
}

func DedicatedTokenSubjectPrefix(orgID string) string {
	return orgID + "-token-"
}

func (c *Client) IssueStaticToken(ctx context.Context, req StaticTokenRequest) (*app.Token, error) {
	var token *app.Token
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if req.OrgID != "" {
			if err := lockOrg(tx, req.OrgID); err != nil {
				return err
			}
		}
		// Invalidating revokes under FOR UPDATE, so take it before any weaker lock.
		if req.Invalidate {
			if err := revokeAccountCredentials(tx, req.AccountID); err != nil {
				return err
			}
		} else if err := lockAccount(tx, req.AccountID); err != nil {
			return err
		}
		now := time.Now()
		var err error
		token, err = c.withDB(tx).insertToken(ctx, &app.Token{
			AccountID: req.AccountID, CreatedByID: req.CreatedByID, Name: req.Name, Role: string(req.Role), OrgID: req.OrgID,
			TokenType: app.TokenTypeStatic, Issuer: req.AccountID, IssuedAt: now, ExpiresAt: now.Add(req.Duration),
		})
		return err
	})
	return token, err
}

func (c *Client) IssueFederatedToken(ctx context.Context, verified *app.OIDCTrustPolicy) (*app.Token, error) {
	var token *app.Token
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockOrg(tx, verified.OrgID); err != nil {
			return err
		}
		if err := lockAccount(tx, verified.ServiceAccountID); err != nil {
			return err
		}
		var policy app.OIDCTrustPolicy
		if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Where(app.OIDCTrustPolicy{ID: verified.ID}).Take(&policy).Error; err != nil {
			return err
		}
		if !policy.Enabled || policy.ServiceAccountID != verified.ServiceAccountID || policy.OrgID != verified.OrgID ||
			policy.IssuerURL != verified.IssuerURL || policy.Audience != verified.Audience || !reflect.DeepEqual(policy.ClaimConditions, verified.ClaimConditions) ||
			policy.Role != verified.Role || policy.TokenDurationSeconds != verified.TokenDurationSeconds {
			return fmt.Errorf("trust policy changed during exchange")
		}
		duration := policy.TokenDurationSeconds
		if duration <= 0 {
			duration = app.OIDCTrustPolicyDefaultTokenDuration
		}
		if duration > app.OIDCTrustPolicyMaxTokenDuration {
			duration = app.OIDCTrustPolicyMaxTokenDuration
		}
		now := time.Now()
		var err error
		token, err = c.withDB(tx).insertToken(ctx, &app.Token{
			AccountID: policy.ServiceAccountID, CreatedByID: policy.ServiceAccountID, Name: policy.Name, Role: policy.Role,
			OrgID: policy.OrgID, TokenType: app.TokenTypeFederated, Issuer: policy.IssuerURL, IssuedAt: now,
			ExpiresAt: now.Add(time.Duration(duration) * time.Second), SourceType: app.TokenSourceTypeOIDCTrustPolicy, SourceID: policy.ID,
		})
		return err
	})
	return token, err
}

func (c *Client) ClaimRunnerJobToken(ctx context.Context, orgID, runnerID, jobID string) (*app.Token, error) {
	if orgID == "" || runnerID == "" || jobID == "" {
		return nil, fmt.Errorf("org, runner and job IDs are required")
	}
	var token *app.Token
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// The job row lock serializes claims for one job; the org lock orders this
		// before org cleanup, which locks the org before its jobs.
		if err := lockOrg(tx, orgID); err != nil {
			return err
		}
		var job app.RunnerJob
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.RunnerJob{ID: jobID, RunnerID: runnerID, OrgID: orgID}).Take(&job).Error; err != nil {
			return err
		}
		if job.Status != app.RunnerJobStatusAvailable {
			return ErrJobClaimConflict
		}
		client := c.withDB(tx)
		acct, err := client.FindAccount(ctx, ServiceAccountEmail(job.ID))
		if err != nil {
			return fmt.Errorf("job service account not provisioned: %w", err)
		}
		var existing int64
		if err := tx.Model(&app.Token{}).Where(app.Token{AccountID: acct.ID, SourceType: app.TokenSourceTypeRunnerJob, SourceID: job.ID}).Where("expires_at > ?", time.Now()).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return ErrJobClaimConflict
		}
		token, err = client.IssueToken(ctx, acct.ID, 6*time.Hour, IssueTokenRequest{OrgID: orgID, SourceType: app.TokenSourceTypeRunnerJob, SourceID: job.ID})
		return err
	})
	return token, err
}
