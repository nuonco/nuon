package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
)

var (
	// ErrAccountNotAuthorized is returned when a user tries to authenticate
	// but has no existing account and no pending org invite.
	ErrAccountNotAuthorized = errors.New("account not authorized: no existing account or pending invitation found")

	// ErrEmailDomainNotAllowed is returned when a user tries to authenticate
	// but their email domain is not in the allowed domains list.
	ErrEmailDomainNotAllowed = errors.New("email domain not allowed")
)

// getOrCreateAccountByIdentityStrict looks up an account by (provider_type, sub).
// If found, returns the existing account.
// If not found by sub, checks for an existing account by email or a pending OrgInvite.
// Only creates a new account if there's an existing account (to link) or a pending invite.
func (s *service) getOrCreateAccountByIdentityStrict(
	ctx context.Context,
	providerType app.ProviderType,
	identityProviderID *string,
	userInfo *providers.UserInfo,
) (*app.Account, error) {
	// 1. Look up existing account identity by (provider_type, sub)
	var accountIdentity app.AccountIdentity
	err := s.db.WithContext(ctx).
		Preload("Account").
		Where("provider_type = ? AND sub = ?", providerType, userInfo.Subject).
		First(&accountIdentity).Error

	if err == nil {
		// Found existing identity - check if profile needs update
		needsUpdate := false

		// Update if values have changed (including clearing when provider returns empty)
		if accountIdentity.Name != userInfo.Name {
			accountIdentity.Name = userInfo.Name
			needsUpdate = true
		}
		if accountIdentity.Picture != userInfo.Picture {
			accountIdentity.Picture = userInfo.Picture
			needsUpdate = true
		}

		if needsUpdate {
			if err := s.db.WithContext(ctx).
				Model(&accountIdentity).
				Select("name", "picture").
				Updates(&accountIdentity).Error; err != nil {
				s.l.Warn("failed to update identity profile",
					zap.String("identity_id", accountIdentity.ID),
					zap.Error(err))
				// Don't fail the login - just log the warning
			} else {
				s.l.Debug("updated identity profile",
					zap.String("identity_id", accountIdentity.ID),
					zap.String("name", accountIdentity.Name),
					zap.String("picture", accountIdentity.Picture))
			}
		}

		s.l.Debug("found existing account identity",
			zap.String("account_id", accountIdentity.AccountID),
			zap.String("provider_type", string(providerType)),
			zap.String("sub", userInfo.Subject))
		return accountIdentity.Account, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to lookup account identity: %w", err)
	}

	// 2. No existing identity - check if there's an existing account with this email
	var existingAccount app.Account
	err = s.db.WithContext(ctx).
		Where("email = ?", userInfo.Email).
		First(&existingAccount).Error

	if err == nil {
		// Found existing account by email - link the new identity to it
		s.l.Info("linking new identity to existing account",
			zap.String("account_id", existingAccount.ID),
			zap.String("provider_type", string(providerType)),
			zap.String("sub", userInfo.Subject),
			zap.String("email", userInfo.Email))

		return s.linkIdentityToAccount(ctx, &existingAccount, providerType, identityProviderID, userInfo)
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to lookup account by email: %w", err)
	}

	// 3. No existing account - check for pending OrgInvite
	var pendingInvite app.OrgInvite
	err = s.db.WithContext(ctx).
		Where("email = ? AND status = ?", userInfo.Email, app.OrgInviteStatusPending).
		First(&pendingInvite).Error

	if err == gorm.ErrRecordNotFound {
		// No account and no invite - not authorized
		s.l.Warn("authentication denied: no account or pending invite",
			zap.String("email", userInfo.Email),
			zap.String("provider_type", string(providerType)),
			zap.String("sub", userInfo.Subject))
		return nil, ErrAccountNotAuthorized
	}

	if err != nil {
		return nil, fmt.Errorf("failed to lookup org invite: %w", err)
	}

	// 4. Found pending invite - create account and identity
	s.l.Info("creating account for invited user",
		zap.String("provider_type", string(providerType)),
		zap.String("sub", userInfo.Subject),
		zap.String("email", userInfo.Email),
		zap.String("invite_id", pendingInvite.ID),
		zap.String("org_id", pendingInvite.OrgID))

	return s.createAccountWithIdentity(ctx, providerType, identityProviderID, userInfo)
}

// getOrCreateAccountByIdentity looks up an account by (provider_type, sub).
// If found, returns the existing account.
// If not found by sub, checks for an existing account by email.
// If no existing account, creates a new account if the email domain is allowed.
func (s *service) getOrCreateAccountByIdentity(
	ctx context.Context,
	providerType app.ProviderType,
	identityProviderID *string,
	userInfo *providers.UserInfo,
) (*app.Account, error) {
	// 1. Look up existing account identity by (provider_type, sub)
	var accountIdentity app.AccountIdentity
	err := s.db.WithContext(ctx).
		Preload("Account").
		Where("provider_type = ? AND sub = ?", providerType, userInfo.Subject).
		First(&accountIdentity).Error

	if err == nil {
		// Found existing identity - check if profile needs update
		needsUpdate := false

		if accountIdentity.Name != userInfo.Name {
			accountIdentity.Name = userInfo.Name
			needsUpdate = true
		}
		if accountIdentity.Picture != userInfo.Picture {
			accountIdentity.Picture = userInfo.Picture
			needsUpdate = true
		}

		if needsUpdate {
			if err := s.db.WithContext(ctx).
				Model(&accountIdentity).
				Select("name", "picture").
				Updates(&accountIdentity).Error; err != nil {
				s.l.Warn("failed to update identity profile",
					zap.String("identity_id", accountIdentity.ID),
					zap.Error(err))
			} else {
				s.l.Debug("updated identity profile",
					zap.String("identity_id", accountIdentity.ID),
					zap.String("name", accountIdentity.Name),
					zap.String("picture", accountIdentity.Picture))
			}
		}

		s.l.Debug("found existing account identity",
			zap.String("account_id", accountIdentity.AccountID),
			zap.String("provider_type", string(providerType)),
			zap.String("sub", userInfo.Subject))
		return accountIdentity.Account, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to lookup account identity: %w", err)
	}

	// 2. No existing identity - check if there's an existing account with this email
	var existingAccount app.Account
	err = s.db.WithContext(ctx).
		Where("email = ?", userInfo.Email).
		First(&existingAccount).Error

	if err == nil {
		// Found existing account by email - link the new identity to it
		s.l.Info("linking new identity to existing account",
			zap.String("account_id", existingAccount.ID),
			zap.String("provider_type", string(providerType)),
			zap.String("sub", userInfo.Subject),
			zap.String("email", userInfo.Email))

		return s.linkIdentityToAccount(ctx, &existingAccount, providerType, identityProviderID, userInfo)
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to lookup account by email: %w", err)
	}

	// 3. No existing account - check if email domain is allowed
	if !s.isEmailDomainAllowed(userInfo.Email) {
		s.l.Warn("authentication denied: email domain not allowed",
			zap.String("email", userInfo.Email),
			zap.String("provider_type", string(providerType)),
			zap.String("sub", userInfo.Subject))
		return nil, ErrEmailDomainNotAllowed
	}

	// 4. Email domain is allowed - create account and identity
	s.l.Info("creating account for user with allowed domain",
		zap.String("provider_type", string(providerType)),
		zap.String("sub", userInfo.Subject),
		zap.String("email", userInfo.Email))

	return s.createAccountWithIdentity(ctx, providerType, identityProviderID, userInfo)
}

// createAccountWithIdentity creates a new account and links it to the identity provider.
func (s *service) createAccountWithIdentity(
	ctx context.Context,
	providerType app.ProviderType,
	identityProviderID *string,
	userInfo *providers.UserInfo,
) (*app.Account, error) {
	// Create the account using the account client
	// NOTE: at this time, we are not enabling user journeys for users created through this flow.
	acct, err := s.acctClient.CreateAuthAccount(ctx, userInfo.Email, userInfo.Subject, account.NoUserJourneys())
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	// Create the account identity
	accountIdentity := &app.AccountIdentity{
		AccountID:          acct.ID,
		IdentityProviderID: identityProviderID,
		ProviderType:       providerType,
		Sub:                userInfo.Subject,
		Name:               userInfo.Name,
		Picture:            userInfo.Picture,
	}

	if err := s.db.WithContext(ctx).Create(accountIdentity).Error; err != nil {
		return nil, fmt.Errorf("failed to create account identity: %w", err)
	}

	s.l.Info("created new account with identity",
		zap.String("account_id", acct.ID),
		zap.String("identity_id", accountIdentity.ID),
		zap.String("provider_type", string(providerType)),
		zap.String("email", userInfo.Email))

	return acct, nil
}

// linkIdentityToAccount creates an account_identity record linking an existing account
// to a new identity provider. This is used when a user with an existing account
// authenticates via a new provider.
func (s *service) linkIdentityToAccount(
	ctx context.Context,
	account *app.Account,
	providerType app.ProviderType,
	identityProviderID *string,
	userInfo *providers.UserInfo,
) (*app.Account, error) {
	accountIdentity := &app.AccountIdentity{
		AccountID:          account.ID,
		IdentityProviderID: identityProviderID,
		ProviderType:       providerType,
		Sub:                userInfo.Subject,
		Name:               userInfo.Name,
		Picture:            userInfo.Picture,
	}

	if err := s.db.WithContext(ctx).Create(accountIdentity).Error; err != nil {
		return nil, fmt.Errorf("failed to create account identity: %w", err)
	}

	s.l.Info("linked identity to existing account",
		zap.String("account_id", account.ID),
		zap.String("identity_id", accountIdentity.ID),
		zap.String("provider_type", string(providerType)),
		zap.String("email", userInfo.Email))

	return account, nil
}

// getAccountIdentityByProviderAndSub looks up an account identity by provider type and subject.
func (s *service) getAccountIdentityByProviderAndSub(
	ctx context.Context,
	providerType app.ProviderType,
	sub string,
) (*app.AccountIdentity, error) {
	var accountIdentity app.AccountIdentity
	err := s.db.WithContext(ctx).
		Preload("Account").
		Where("provider_type = ? AND sub = ?", providerType, sub).
		First(&accountIdentity).Error

	if err != nil {
		return nil, err
	}

	return &accountIdentity, nil
}
