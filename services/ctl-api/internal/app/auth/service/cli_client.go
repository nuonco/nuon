package service

import (
	"context"
	"errors"

	"github.com/lib/pq"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/nuonauth"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

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
