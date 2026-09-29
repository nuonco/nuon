package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
)

var (
	ErrAccountNotAuthorized = errors.New("account not authorized: no existing account or pending invitation found")

	ErrEmailDomainNotAllowed = errors.New("email domain not allowed")
)

func (s *service) getOrCreateAccountByIdentityStrict(
	ctx context.Context,
	identityProvider *app.IdentityProvider,
	userInfo *providers.UserInfo,
) (*app.Account, error) {
	var accountIdentity app.AccountIdentity
	err := s.db.WithContext(ctx).
		Preload("Account").
		Where(&app.AccountIdentity{IdentityProviderID: identityProvider.ID, Sub: userInfo.Subject}).
		First(&accountIdentity).Error

	if err == nil {
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
			zap.String("provider_id", identityProvider.ID),
			zap.String("sub", userInfo.Subject))
		return accountIdentity.Account, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to lookup account identity: %w", err)
	}

	var existingAccount app.Account
	err = s.db.WithContext(ctx).
		Where("email = ?", strings.ToLower(userInfo.Email)).
		First(&existingAccount).Error

	if err == nil {
		s.l.Info("linking new identity to existing account",
			zap.String("account_id", existingAccount.ID),
			zap.String("provider_id", identityProvider.ID),
			zap.String("sub", userInfo.Subject),
			zap.String("email", userInfo.Email))

		return s.linkIdentityToAccount(ctx, &existingAccount, identityProvider, userInfo)
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to lookup account by email: %w", err)
	}

	var pendingInvite app.OrgInvite
	err = s.db.WithContext(ctx).
		Where("email = ? AND status = ?", userInfo.Email, app.OrgInviteStatusPending).
		First(&pendingInvite).Error

	if err == gorm.ErrRecordNotFound {
		s.l.Warn("authentication denied: no account or pending invite",
			zap.String("email", userInfo.Email),
			zap.String("provider_id", identityProvider.ID),
			zap.String("sub", userInfo.Subject))
		return nil, ErrAccountNotAuthorized
	}

	if err != nil {
		return nil, fmt.Errorf("failed to lookup org invite: %w", err)
	}

	s.l.Info("creating account for invited user",
		zap.String("provider_id", identityProvider.ID),
		zap.String("sub", userInfo.Subject),
		zap.String("email", userInfo.Email),
		zap.String("invite_id", pendingInvite.ID),
		zap.String("org_id", pendingInvite.OrgID))

	return s.createAccountWithIdentity(ctx, identityProvider, userInfo, true)
}

func (s *service) getOrCreateAccountByIdentity(
	ctx context.Context,
	identityProvider *app.IdentityProvider,
	userInfo *providers.UserInfo,
) (*app.Account, error) {
	var accountIdentity app.AccountIdentity
	err := s.db.WithContext(ctx).
		Preload("Account").
		Where(&app.AccountIdentity{IdentityProviderID: identityProvider.ID, Sub: userInfo.Subject}).
		First(&accountIdentity).Error

	if err == nil {
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
			zap.String("provider_id", identityProvider.ID),
			zap.String("sub", userInfo.Subject))
		return accountIdentity.Account, nil
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to lookup account identity: %w", err)
	}

	var existingAccount app.Account
	err = s.db.WithContext(ctx).
		Where("email = ?", strings.ToLower(userInfo.Email)).
		First(&existingAccount).Error

	if err == nil {
		s.l.Info("linking new identity to existing account",
			zap.String("account_id", existingAccount.ID),
			zap.String("provider_id", identityProvider.ID),
			zap.String("sub", userInfo.Subject),
			zap.String("email", userInfo.Email))

		return s.linkIdentityToAccount(ctx, &existingAccount, identityProvider, userInfo)
	}

	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to lookup account by email: %w", err)
	}

	if !s.isEmailDomainAllowed(userInfo.Email) {
		s.l.Warn("authentication denied: email domain not allowed",
			zap.String("email", userInfo.Email),
			zap.String("provider_id", identityProvider.ID),
			zap.String("sub", userInfo.Subject))
		return nil, ErrEmailDomainNotAllowed
	}

	s.l.Info("creating account for user with allowed domain",
		zap.String("provider_id", identityProvider.ID),
		zap.String("sub", userInfo.Subject),
		zap.String("email", userInfo.Email))

	return s.createAccountWithIdentity(ctx, identityProvider, userInfo, false)
}

func (s *service) createAccountWithIdentity(
	ctx context.Context,
	identityProvider *app.IdentityProvider,
	userInfo *providers.UserInfo,
	isInvitedUser bool,
) (*app.Account, error) {
	var userJourneys app.UserJourneys
	if isInvitedUser {
		userJourneys = account.NoUserJourneys()
	} else if s.cfg.EvaluationJourneyEnabled {
		userJourneys = account.DefaultEvaluationJourney("dashboard")
	} else {
		userJourneys = account.NoUserJourneys()
	}

	acct, err := s.acctClient.CreateAuthAccount(ctx, userInfo.Email, userInfo.Subject, userJourneys)
	if err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	accountIdentity := &app.AccountIdentity{
		AccountID:          acct.ID,
		IdentityProviderID: identityProvider.ID,
		ProviderType:       identityProvider.ProviderType,
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
		zap.String("provider_id", identityProvider.ID),
		zap.String("email", userInfo.Email))

	return acct, nil
}

func (s *service) linkIdentityToAccount(
	ctx context.Context,
	account *app.Account,
	identityProvider *app.IdentityProvider,
	userInfo *providers.UserInfo,
) (*app.Account, error) {
	accountIdentity := &app.AccountIdentity{
		AccountID:          account.ID,
		IdentityProviderID: identityProvider.ID,
		ProviderType:       identityProvider.ProviderType,
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
		zap.String("provider_id", identityProvider.ID),
		zap.String("email", userInfo.Email))

	return account, nil
}
