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

type authResult struct {
	OrgID     string
	AccountID string
}

func (s *Server) authenticate(r *http.Request) (*authResult, error) {
	token := extractBearerToken(r)
	if token == "" {
		s.l.Warn("MCP auth: missing authorization header")
		return nil, fmt.Errorf("missing authorization header")
	}

	orgID := r.Header.Get("X-Nuon-Org-ID")
	if orgID == "" {
		s.l.Warn("MCP auth: missing X-Nuon-Org-ID header")
		return nil, fmt.Errorf("missing X-Nuon-Org-ID header")
	}

	var userToken app.Token
	res := s.db.
		WithContext(r.Context()).
		Where(&app.Token{Token: token}).
		First(&userToken)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		s.l.Warn("MCP auth: token not found in database")
		return nil, fmt.Errorf("invalid token")
	}
	if res.Error != nil {
		s.l.Warn("MCP auth: token lookup error", zap.Error(res.Error))
		return nil, fmt.Errorf("unable to look up token: %w", res.Error)
	}

	if time.Now().After(userToken.ExpiresAt) {
		s.l.Warn("MCP auth: token is expired")
		return nil, fmt.Errorf("token is expired")
	}

	var acct app.Account
	res = s.db.WithContext(r.Context()).
		Preload("Roles").
		Preload("Roles.Org").
		Preload("Roles.Policies").
		First(&acct, "id = ?", userToken.AccountID)
	if res.Error != nil {
		s.l.Warn("MCP auth: unable to fetch account", zap.Error(res.Error))
		return nil, fmt.Errorf("unable to fetch account: %w", res.Error)
	}

	hasAccess := false
	for _, oid := range acct.OrgIDs {
		if oid == orgID {
			hasAccess = true
			break
		}
	}
	if !hasAccess {
		s.l.Warn("MCP auth: account does not have access to org", zap.String("org_id", orgID), zap.Strings("account_org_ids", acct.OrgIDs))
		return nil, fmt.Errorf("account does not have access to org %s", orgID)
	}

	s.l.Info("MCP auth: success", zap.String("account_id", acct.ID), zap.String("org_id", orgID))
	return &authResult{
		OrgID:     orgID,
		AccountID: acct.ID,
	}, nil
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
