package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// authenticateToken validates the bearer access token and returns the associated
// account plus the token itself. It performs NO org authorization — that is
// handled separately so a valid-but-org-less request (e.g. an OAuth client
// before selecting an org) still authenticates rather than triggering a re-auth.
// A nil error means the token is good; a non-nil error should result in a 401.
func (s *Server) authenticateToken(r *http.Request) (*app.Account, *app.Token, error) {
	token := extractBearerToken(r)
	if token == "" {
		return nil, nil, fmt.Errorf("missing authorization header")
	}

	var userToken app.Token
	res := s.db.
		WithContext(r.Context()).
		Where(&app.Token{Token: token}).
		First(&userToken)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		s.l.Warn("MCP auth: token not found in database")
		return nil, nil, fmt.Errorf("invalid token")
	}
	if res.Error != nil {
		s.l.Warn("MCP auth: token lookup error", zap.Error(res.Error))
		return nil, nil, fmt.Errorf("unable to look up token: %w", res.Error)
	}

	if time.Now().After(userToken.ExpiresAt) {
		s.l.Warn("MCP auth: token is expired")
		return nil, nil, fmt.Errorf("token is expired")
	}

	var acct app.Account
	res = s.db.WithContext(r.Context()).
		Preload("Roles").
		Preload("Roles.Org").
		Preload("Roles.Policies").
		First(&acct, "id = ?", userToken.AccountID)
	if res.Error != nil {
		s.l.Warn("MCP auth: unable to fetch account", zap.Error(res.Error))
		return nil, nil, fmt.Errorf("unable to fetch account: %w", res.Error)
	}

	return &acct, &userToken, nil
}

// accountHasOrgAccess reports whether the account can access the given org.
func accountHasOrgAccess(acct *app.Account, orgID string) bool {
	for _, oid := range acct.OrgIDs {
		if oid == orgID {
			return true
		}
	}
	return false
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return parts[1]
}
