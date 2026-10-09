package oauthclients

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

const (
	secretPrefix     = "nuon_ocs_"
	MaxActiveSecrets = 2
)

var (
	ErrInvalidClient   = errors.New("invalid client credentials")
	ErrTooManySecrets  = fmt.Errorf("a client may hold at most %d active secrets; revoke one before creating another", MaxActiveSecrets)
	ErrSecretInThePast = errors.New("expires_at must be in the future")
	ErrClientNotFound  = errors.New("oauth client not found")
	ErrSecretNotFound  = errors.New("oauth client secret not found")
	ErrNotConfidential = errors.New("oauth client is not a confidential client")
)

func Issuer(cfg *internal.Config) string {
	if cfg.RootDomain == "localhost" {
		return "http://localhost:8084"
	}
	return fmt.Sprintf("https://auth.%s", cfg.RootDomain)
}

func TokenEndpoint(cfg *internal.Config) string {
	return Issuer(cfg) + "/oauth/token"
}

func CreateConfidentialClient(ctx context.Context, db *gorm.DB, orgID, accountID, name string) (*app.OAuthClient, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "service-account-client"
	}

	client := app.OAuthClient{
		ClientName:              name,
		RedirectURIs:            pq.StringArray{},
		TokenEndpointAuthMethod: app.OAuthTokenEndpointAuthMethodClientSecretBasic,
		OrgID:                   generics.NewNullString(orgID),
		AccountID:               generics.NewNullString(accountID),
		GrantTypes:              pq.StringArray{app.OAuthGrantTypeClientCredentials},
	}
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockAccount(tx, accountID); err != nil {
			return err
		}
		return tx.Create(&client).Error
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create oauth client")
	}

	return &client, nil
}

func GetAccountClient(ctx context.Context, db *gorm.DB, orgID, accountID, clientID string) (*app.OAuthClient, error) {
	var client app.OAuthClient
	res := db.WithContext(ctx).
		Preload("Secrets", orderByCreatedAt).
		Where(app.OAuthClient{ID: clientID, OrgID: generics.NewNullString(orgID), AccountID: generics.NewNullString(accountID)}).
		First(&client)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, ErrClientNotFound
	}
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get oauth client")
	}

	return &client, nil
}

func ListAccountClients(ctx context.Context, db *gorm.DB, orgID, accountID string) ([]app.OAuthClient, error) {
	clients := []app.OAuthClient{}
	if res := db.WithContext(ctx).
		Preload("Secrets", orderByCreatedAt).
		Where(app.OAuthClient{OrgID: generics.NewNullString(orgID), AccountID: generics.NewNullString(accountID)}).
		Order("created_at").
		Find(&clients); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to list oauth clients")
	}

	return clients, nil
}

func ListSecrets(ctx context.Context, db *gorm.DB, clientID string) ([]app.OAuthClientSecret, error) {
	secrets := []app.OAuthClientSecret{}
	if res := db.WithContext(ctx).
		Where(app.OAuthClientSecret{ClientID: clientID}).
		Order("created_at").
		Find(&secrets); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to list oauth client secrets")
	}

	return secrets, nil
}

func lockAccount(tx *gorm.DB, accountID string) error {
	var acct app.Account
	if err := tx.Clauses(clause.Locking{Strength: "KEY SHARE"}).
		Select("id", "account_type").
		Where(app.Account{ID: accountID}).First(&acct).Error; err != nil {
		return err
	}
	if acct.AccountType != app.AccountTypeService {
		return ErrInvalidClient
	}
	return nil
}

func lockClient(tx *gorm.DB, clientID, strength string) (*app.OAuthClient, error) {
	var client app.OAuthClient
	res := tx.Where(app.OAuthClient{ID: clientID}).First(&client)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, ErrClientNotFound
	}
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to look up oauth client")
	}
	if !client.IsConfidential() {
		return nil, ErrNotConfidential
	}
	if err := lockAccount(tx, client.AccountID.String); err != nil {
		return nil, err
	}
	res = tx.Clauses(clause.Locking{Strength: strength}).
		Where(app.OAuthClient{ID: clientID}).First(&client)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, ErrClientNotFound
	}
	if res.Error != nil {
		return nil, res.Error
	}

	return &client, nil
}

func CreateSecret(ctx context.Context, db *gorm.DB, clientID string, expiresAt *time.Time) (*app.OAuthClientSecret, string, error) {
	now := time.Now()
	if expiresAt != nil && !expiresAt.After(now) {
		return nil, "", ErrSecretInThePast
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, "", errors.Wrap(err, "unable to generate client secret")
	}
	plaintext := secretPrefix + base64.RawURLEncoding.EncodeToString(raw)

	var secret app.OAuthClientSecret
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		client, err := lockClient(tx, clientID, "UPDATE")
		if err != nil {
			return err
		}
		now := time.Now()
		if expiresAt != nil && !expiresAt.After(now) {
			return ErrSecretInThePast
		}

		existing, err := ListSecrets(ctx, tx, client.ID)
		if err != nil {
			return err
		}
		active := 0
		for i := range existing {
			if existing[i].Active(now) {
				active++
			}
		}
		if active >= MaxActiveSecrets {
			return ErrTooManySecrets
		}

		secret = app.OAuthClientSecret{
			ClientID:   client.ID,
			SecretHash: hashSecret(plaintext),
			Hint:       plaintext[len(plaintext)-4:],
			ExpiresAt:  expiresAt,
		}
		if res := tx.Create(&secret); res.Error != nil {
			return errors.Wrap(res.Error, "unable to create oauth client secret")
		}
		return nil
	})
	if err != nil {
		return nil, "", err
	}

	return &secret, plaintext, nil
}

func RevokeSecret(ctx context.Context, db *gorm.DB, clientID, secretID string) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := lockClient(tx, clientID, "UPDATE"); err != nil {
			return err
		}

		res := tx.Where(app.OAuthClientSecret{ID: secretID, ClientID: clientID}).Delete(&app.OAuthClientSecret{})
		if res.Error != nil {
			return errors.Wrap(res.Error, "unable to revoke oauth client secret")
		}
		if res.RowsAffected == 0 {
			return ErrSecretNotFound
		}

		return account.RevokeDerivedCredentials(tx, app.TokenSourceTypeOAuthClientSecret, []string{secretID})
	})
}

func DeleteClient(ctx context.Context, db *gorm.DB, clientID string) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := lockClient(tx, clientID, "UPDATE"); err != nil {
			return err
		}

		var secretIDs []string
		if res := tx.Unscoped().Model(&app.OAuthClientSecret{}).
			Where(app.OAuthClientSecret{ClientID: clientID}).
			Pluck("id", &secretIDs); res.Error != nil {
			return errors.Wrap(res.Error, "unable to look up oauth client secrets")
		}
		if len(secretIDs) > 0 {
			if res := tx.Where("id = ANY(?)", pq.Array(secretIDs)).Delete(&app.OAuthClientSecret{}); res.Error != nil {
				return errors.Wrap(res.Error, "unable to delete oauth client secrets")
			}
			if err := account.RevokeDerivedCredentials(tx, app.TokenSourceTypeOAuthClientSecret, secretIDs); err != nil {
				return err
			}
		}

		if res := tx.Delete(&app.OAuthClient{ID: clientID}); res.Error != nil {
			return errors.Wrap(res.Error, "unable to delete oauth client")
		}
		return nil
	})
}

func Authenticate(tx *gorm.DB, clientID, plaintext string) (*app.OAuthClient, *app.OAuthClientSecret, error) {
	if clientID == "" || plaintext == "" {
		return nil, nil, ErrInvalidClient
	}

	client, err := lockClient(tx, clientID, "SHARE")
	if errors.Is(err, ErrClientNotFound) || errors.Is(err, ErrNotConfidential) || errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, ErrInvalidClient
	}
	if err != nil {
		return nil, nil, err
	}
	if !client.AllowsGrantType(app.OAuthGrantTypeClientCredentials) {
		return nil, nil, ErrInvalidClient
	}

	var secret app.OAuthClientSecret
	res := tx.Clauses(clause.Locking{Strength: "SHARE"}).
		Where(app.OAuthClientSecret{ClientID: client.ID, SecretHash: hashSecret(plaintext)}).
		First(&secret)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, nil, ErrInvalidClient
	}
	if res.Error != nil {
		return nil, nil, errors.Wrap(res.Error, "unable to look up oauth client secret")
	}

	if !secret.Active(time.Now()) {
		return nil, nil, ErrInvalidClient
	}

	return client, &secret, nil
}

func orderByCreatedAt(db *gorm.DB) *gorm.DB {
	return db.Order("created_at")
}

func hashSecret(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}
