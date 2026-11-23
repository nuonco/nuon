package service

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/nuonauth"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const pgUndefinedTable = "42P01"

var cliClientRetryInterval = 10 * time.Second

func (s *service) ensureCLIClient(ctx context.Context) error {
	var existing app.OAuthClient
	err := s.db.WithContext(ctx).Where(&app.OAuthClient{ID: nuonauth.CLIClientID}).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	err = s.db.WithContext(ctx).Create(&app.OAuthClient{
		ID:                      nuonauth.CLIClientID,
		ClientName:              nuonauth.CLIClientName,
		RedirectURIs:            pq.StringArray{nuonauth.CLILoopbackRedirectURI},
		TokenEndpointAuthMethod: "none",
	}).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil
	}
	if err != nil {
		return err
	}
	s.l.Info("cli oauth client registered", zap.String("client_id", nuonauth.CLIClientID))
	return nil
}

// startCLIClient registers the CLI client, deferring to a background retry only when
// o_auth_clients does not exist yet. On a fresh install auth starts before migrations create
// it, and failing startup there blocks the deploy that would run those migrations. Every other
// error still fails startup.
func (s *service) startCLIClient(ctx context.Context) error {
	err := s.ensureCLIClient(ctx)
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != pgUndefinedTable {
		return err
	}
	s.l.Warn("o_auth_clients does not exist yet, retrying cli oauth client registration", zap.Error(err))
	go s.retryCLIClient()
	return nil
}

func (s *service) retryCLIClient() {
	ticker := time.NewTicker(cliClientRetryInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCleanup:
			return
		case <-ticker.C:
			if err := s.ensureCLIClient(context.Background()); err != nil {
				s.l.Warn("cli oauth client still not registered", zap.Error(err))
				continue
			}
			return
		}
	}
}
