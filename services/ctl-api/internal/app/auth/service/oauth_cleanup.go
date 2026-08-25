package service

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const oauthCleanupInterval = time.Hour

// runOAuthCleanup periodically purges spent OAuth artifacts (expired or consumed
// authorization codes and refresh tokens) so those tables stay bounded. It runs
// once at startup and then on an interval until the service stops.
func (s *service) runOAuthCleanup() {
	s.purgeExpiredOAuth(context.Background())

	ticker := time.NewTicker(oauthCleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCleanup:
			return
		case <-ticker.C:
			s.purgeExpiredOAuth(context.Background())
		}
	}
}

func (s *service) purgeExpiredOAuth(ctx context.Context) {
	now := time.Now()

	codeRes := s.db.WithContext(ctx).
		Unscoped().
		Where("expires_at < ? OR consumed = ?", now, true).
		Delete(&app.OAuthAuthorizationCode{})
	if codeRes.Error != nil {
		s.l.Warn("failed to purge oauth authorization codes", zap.Error(codeRes.Error))
	}

	refreshRes := s.db.WithContext(ctx).
		Unscoped().
		Where("expires_at < ? OR consumed = ?", now, true).
		Delete(&app.OAuthRefreshToken{})
	if refreshRes.Error != nil {
		s.l.Warn("failed to purge oauth refresh tokens", zap.Error(refreshRes.Error))
	}

	s.l.Debug("purged spent oauth artifacts",
		zap.Int64("authorization_codes", codeRes.RowsAffected),
		zap.Int64("refresh_tokens", refreshRes.RowsAffected))
}
