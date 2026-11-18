package service

import (
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

func (s *service) issueFederatedToken(tx *gorm.DB, policy *app.OIDCTrustPolicy) (*app.Token, error) {
	return account.New(account.Params{DB: tx}).IssueFederatedToken(tx.Statement.Context, policy)
}
